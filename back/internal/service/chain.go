package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"trade-chain/internal/domain"
	"trade-chain/internal/events"
	"trade-chain/internal/exchange"
	"trade-chain/internal/repository"
)

type chainService struct {
	repo         repository.ChainRepository
	products     repository.ProductRepository
	negotiations repository.NegotiationRepository
	events       *events.Broker
}

func NewChainService(
	repo repository.ChainRepository,
	products repository.ProductRepository,
	negotiations repository.NegotiationRepository,
	brokers ...*events.Broker,
) ChainService {
	var broker *events.Broker
	if len(brokers) > 0 {
		broker = brokers[0]
	}
	return &chainService{repo: repo, products: products, negotiations: negotiations, events: broker}
}

func (s *chainService) publish(eventType string, chain *domain.Chain) {
	if s.events == nil || chain == nil {
		return
	}
	recipients := []string{chain.InitiatorID}
	if chain.RecipientID != nil && *chain.RecipientID != chain.InitiatorID {
		recipients = append(recipients, *chain.RecipientID)
	}
	s.events.Publish(events.Event{Type: eventType, ChainID: chain.ChainID}, recipients...)
}

// dealOf собирает взгляд на звено, с которым работают правила согласования.
//
// Стороны берутся из самого звена, а не из текущих владельцев товаров:
// успешный обмен меняет владельцев местами, и вычисление на лету после
// завершения сделки указывало бы обеими сторонами на одного человека.
func dealOf(chain *domain.Chain) exchange.Deal {
	var recipientID string
	if chain.RecipientID != nil {
		recipientID = *chain.RecipientID
	}
	return exchange.Deal{
		ChainID:     chain.ChainID,
		InitiatorID: chain.InitiatorID,
		RecipientID: recipientID,
		Status:      domain.ChainStatus(chain.Status),
		ExpiresAt:   chain.ExpiresAt,
	}
}

func (s *chainService) Create(ctx context.Context, c *domain.Chain) (*domain.Chain, error) {
	if c == nil || blank(c.FromProductID) || blank(c.InitiatorID) {
		return nil, ErrInvalidInput
	}

	// Фронтенд присылает пустые строки для необязательных UUID-полей, а база
	// ждёт NULL: пустая строка в UUID-колонке — ошибка. Нормализуем один раз.
	c.ToProductID = nilIfEmpty(c.ToProductID)
	c.ToCategoryID = nilIfEmpty(c.ToCategoryID)
	c.ExchangeGoalID = nilIfEmpty(c.ExchangeGoalID)
	c.RouteStepID = nilIfEmpty(c.RouteStepID)
	c.RecipientID = nilIfEmpty(c.RecipientID)

	hasProductGoal := c.ToProductID != nil
	hasCategoryGoal := c.ToCategoryID != nil
	if !hasProductGoal && !hasCategoryGoal {
		return nil, ErrInvalidInput
	}

	offered, err := s.products.GetByID(ctx, c.FromProductID)
	if err != nil {
		return nil, normalizeError(err)
	}

	// Новое предложение всегда начинается с ожидания ответа.
	c.Status = string(domain.ChainPending)
	if c.ExpiresAt.IsZero() {
		c.ExpiresAt = time.Now().Add(exchange.DefaultTTL)
	}

	if hasProductGoal {
		// Цель — конкретный товар: валидация обмена между двумя товарами.
		requested, err := s.products.GetByID(ctx, *c.ToProductID)
		if err != nil {
			return nil, normalizeError(err)
		}

		deal := exchange.Deal{
			ChainID:     c.ChainID,
			InitiatorID: c.InitiatorID,
			RecipientID: requested.CustomerID,
		}
		if err := exchange.Validate(deal, *offered, *requested); err != nil {
			return nil, mapExchangeError(err)
		}

		c.Surcharge = exchange.NormalizeSurcharge(c.Surcharge)
		if err := exchange.ValidateSurcharge(deal, c.Surcharge); err != nil {
			return nil, mapExchangeError(err)
		}

		c.RecipientID = &requested.CustomerID
	}

	// Цель-категория: получатель неизвестен, обмен не с конкретным человеком.
	// Валидация товаров и доплат не нужна.

	v, err := s.repo.Create(ctx, c)
	if errors.Is(err, domain.ErrOfferDuplicate) {
		return nil, mapExchangeError(err)
	}
	if err != nil {
		return nil, normalizeError(err)
	}
	s.publish(events.ExchangeOfferCreated, v)
	return v, nil
}

func (s *chainService) GetByID(ctx context.Context, id string, customerID string) (*domain.Chain, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	v, err := s.repo.GetByID(ctx, id, customerID)
	return v, normalizeError(err)
}

func (s *chainService) GetByProductID(ctx context.Context, id string, customerID string) ([]domain.Chain, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	v, err := s.repo.GetByProductID(ctx, id, customerID)
	return v, normalizeError(err)
}

func (s *chainService) GetByCustomerID(ctx context.Context, customerID string) ([]domain.Chain, error) {
	if blank(customerID) {
		return nil, ErrInvalidInput
	}
	v, err := s.repo.GetByCustomerID(ctx, customerID)
	return v, normalizeError(err)
}

func (s *chainService) GetFullChain(ctx context.Context, id string) ([]domain.Chain, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	v, err := s.repo.GetFullChain(ctx, id)
	return v, normalizeError(err)
}

// statusActions переводит запрошенный статус в действие стейт-машины.
//
// completed сюда не входит намеренно: обмен считается состоявшимся только
// после подтверждения обеими сторонами, и путь к нему один — через Confirm.
var statusActions = map[domain.ChainStatus]exchange.Action{
	domain.ChainActive:    exchange.ActionAccept,
	domain.ChainRejected:  exchange.ActionDecline,
	domain.ChainCountered: exchange.ActionCounter,
	domain.ChainCancelled: exchange.ActionCancel,
}

// UpdateStatus оставлен ради существующей ручки PATCH /chains/{id}/status:
// фронт присылает желаемый статус, а решение принимает стейт-машина.
func (s *chainService) UpdateStatus(ctx context.Context, id string, status domain.ChainStatus, userID string) error {
	action, ok := statusActions[status]
	if !ok {
		return ErrInvalidInput
	}
	_, err := s.Decide(ctx, id, action, userID)
	return err
}

// Decide применяет к звену действие одной из сторон.
func (s *chainService) Decide(ctx context.Context, id string, action exchange.Action, actorID string) (*domain.Chain, error) {
	if blank(id) || blank(actorID) {
		return nil, ErrInvalidInput
	}

	chain, err := s.repo.GetByID(ctx, id, actorID)
	if err != nil {
		return nil, normalizeError(err)
	}
	deal := dealOf(chain)
	if action == exchange.ActionAccept {
		if err := s.validateAcceptableProducts(ctx, chain, deal); err != nil {
			return nil, mapExchangeError(err)
		}
	}

	next, err := exchange.Apply(deal, action, actorID, time.Now())
	if err != nil {
		return nil, mapExchangeError(err)
	}
	if err := s.repo.UpdateStatusIfCurrent(ctx, id, actorID, domain.ChainStatus(chain.Status), next); err != nil {
		return nil, normalizeError(err)
	}

	updated, err := s.repo.GetByID(ctx, id, actorID)
	if err != nil {
		return nil, normalizeError(err)
	}
	s.publish(events.ExchangeChainUpdated, updated)
	return updated, nil
}

// validateAcceptableProducts не даёт принять предложение после того, как
// товар уже ушёл в другой завершённый обмен. Такое предложение сохраняется
// в истории и получает отдельный статус unavailable при завершении сделки.
func (s *chainService) validateAcceptableProducts(
	ctx context.Context,
	chain *domain.Chain,
	deal exchange.Deal,
) error {
	offered, err := s.products.GetByID(ctx, chain.FromProductID)
	if err != nil {
		return normalizeError(err)
	}

	log.Printf(
		"ACCEPT CHECK offered: product=%s status=%s owner=%s expectedOwner=%s",
		offered.ProductID,
		offered.Status,
		offered.CustomerID,
		deal.InitiatorID,
	)

	if offered.Status != domain.ProductActive ||
		offered.CustomerID != deal.InitiatorID {
		return domain.ErrProductUnavailable
	}

	if chain.ToProductID == nil {
		return nil
	}

	requested, err := s.products.GetByID(ctx, *chain.ToProductID)
	if err != nil {
		return normalizeError(err)
	}

	log.Printf(
		"ACCEPT CHECK requested: product=%s status=%s owner=%s expectedOwner=%s",
		requested.ProductID,
		requested.Status,
		requested.CustomerID,
		deal.RecipientID,
	)

	if requested.Status != domain.ProductActive ||
		requested.CustomerID != deal.RecipientID {
		return domain.ErrProductUnavailable
	}

	return nil
}

// Confirm записывает решение стороны об итоге обмена и, если высказались оба,
// закрывает сделку.
//
// Подтверждения перечитываются из базы после записи своего: две стороны могут
// нажать кнопку одновременно, и решение по локальному списку оставило бы обмен
// незавершённым, хотя согласились оба.
func (s *chainService) Confirm(ctx context.Context, id, actorID string, success bool, reason string) (*domain.Chain, error) {
	if blank(id) || blank(actorID) {
		return nil, ErrInvalidInput
	}

	chain, err := s.repo.GetByID(ctx, id, actorID)
	if err != nil {
		return nil, normalizeError(err)
	}
	deal := dealOf(chain)

	existing, err := s.negotiations.ListConfirmations(ctx, id)
	if err != nil {
		return nil, normalizeError(err)
	}
	if err := exchange.CanConfirm(deal, actorID, existing); err != nil {
		// Подтверждение могло сохраниться, а финализация сделки упасть позже
		// (например, из-за ограничения базы). Повторный запрос должен повторить
		// финализацию, а не оставлять обмен навсегда активным.
		if errors.Is(err, domain.ErrAlreadyConfirmed) {
			if status, settled := exchange.Resolve(deal, existing); settled {
				related, settleErr := s.settle(ctx, id, actorID, status)
				if settleErr != nil {
					return nil, settleErr
				}
				updated, getErr := s.repo.GetByID(ctx, id, actorID)
				if getErr != nil {
					return nil, normalizeError(getErr)
				}
				s.publish(events.ExchangeConfirmationCreated, updated)
				if updated.Status == string(domain.ChainCompleted) {
					s.publish(events.ExchangeCompleted, updated)
				}
				s.publishRelatedUpdates(id, related)
				return updated, nil
			}
		}
		return nil, mapExchangeError(err)
	}

	// Причина нужна только при отказе: у состоявшегося обмена объяснять нечего,
	// и текст рядом с «да» читался бы как условие.
	if success {
		reason = ""
	}

	err = s.negotiations.Confirm(ctx, &domain.ChainConfirmation{
		ChainID:    id,
		CustomerID: actorID,
		Success:    success,
		Reason:     strings.TrimSpace(reason),
	})
	if err != nil {
		return nil, normalizeError(err)
	}

	all, err := s.negotiations.ListConfirmations(ctx, id)
	if err != nil {
		return nil, normalizeError(err)
	}

	var related []domain.Chain
	if status, settled := exchange.Resolve(deal, all); settled {
		var settleErr error
		related, settleErr = s.settle(ctx, id, actorID, status)
		if settleErr != nil {
			return nil, settleErr
		}
	}

	updated, err := s.repo.GetByID(ctx, id, actorID)
	if err != nil {
		return nil, normalizeError(err)
	}
	s.publish(events.ExchangeConfirmationCreated, updated)
	if updated.Status == string(domain.ChainCompleted) {
		s.publish(events.ExchangeCompleted, updated)
	}
	s.publishRelatedUpdates(id, related)
	return updated, nil
}

func (s *chainService) publishRelatedUpdates(currentChainID string, chains []domain.Chain) {
	for i := range chains {
		if chains[i].ChainID != currentChainID {
			s.publish(events.ExchangeChainUpdated, &chains[i])
		}
	}
}

// settle закрывает звено итоговым статусом.
//
// Успешный обмен идёт через CompleteExchange: он в одной транзакции меняет
// владельцев товаров, поэтому вещи не могут разъехаться со статусом сделки.
func (s *chainService) settle(ctx context.Context, id string, actorID string, status domain.ChainStatus) ([]domain.Chain, error) {
	if status != domain.ChainCompleted {
		return nil, normalizeError(s.repo.UpdateStatus(ctx, id, status))
	}
	chain, err := s.repo.GetByID(ctx, id, actorID)
	if err != nil {
		return nil, normalizeError(err)
	}
	related, err := s.relatedChains(ctx, chain, actorID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CompleteExchange(ctx, id); err != nil {
		// Обе стороны могли подтвердить одновременно: тот, кто пришёл вторым,
		// увидит звено уже завершённым, и это не ошибка.
		chain, getErr := s.repo.GetByID(ctx, id, actorID)
		if getErr == nil && chain.Status == string(domain.ChainCompleted) {
			return nil, nil
		}
		return nil, normalizeError(err)
	}

	updated := make([]domain.Chain, 0)
	for chainID, previousStatus := range related {
		chain, err := s.repo.GetByID(ctx, chainID, actorID)
		if err != nil {
			return nil, normalizeError(err)
		}
		if chain.Status != previousStatus {
			updated = append(updated, *chain)
		}
	}
	return updated, nil
}

func (s *chainService) relatedChains(ctx context.Context, chain *domain.Chain, customerID string) (map[string]string, error) {
	productIDs := []string{chain.FromProductID}
	if chain.ToProductID != nil {
		productIDs = append(productIDs, *chain.ToProductID)
	}

	chains := make(map[string]string)
	for _, productID := range productIDs {
		found, err := s.repo.GetByProductID(ctx, productID, customerID)
		if err != nil {
			return nil, normalizeError(err)
		}
		for _, related := range found {
			chains[related.ChainID] = related.Status
		}
	}
	return chains, nil
}

// Messages отдаёт переписку по звену. Читать её может только участник:
// договорённости о встрече — не публичная часть объявления.
func (s *chainService) Messages(ctx context.Context, id, actorID string) ([]domain.ChainMessage, error) {
	if blank(id) || blank(actorID) {
		return nil, ErrInvalidInput
	}

	chain, err := s.repo.GetByID(ctx, id, actorID)
	if err != nil {
		return nil, normalizeError(err)
	}
	deal := dealOf(chain)
	if !deal.Involves(actorID) {
		return nil, mapExchangeError(domain.ErrNotParticipant)
	}

	v, err := s.negotiations.ListMessages(ctx, id)
	return v, normalizeError(err)
}

func (s *chainService) SendMessage(ctx context.Context, id, actorID, body string) (*domain.ChainMessage, error) {
	body = strings.TrimSpace(body)
	if blank(id) || blank(actorID) || body == "" {
		return nil, ErrInvalidInput
	}

	chain, err := s.repo.GetByID(ctx, id, actorID)
	if err != nil {
		return nil, normalizeError(err)
	}
	deal := dealOf(chain)
	if err := exchange.CanWrite(deal, actorID); err != nil {
		return nil, mapExchangeError(err)
	}

	v, err := s.negotiations.AddMessage(ctx, &domain.ChainMessage{
		ChainID:    id,
		CustomerID: actorID,
		Body:       body,
	})
	if err != nil {
		return nil, normalizeError(err)
	}
	if s.events != nil {
		s.events.Publish(
			events.Event{Type: events.ExchangeMessageCreated, ChainID: chain.ChainID},
			deal.Counterparty(actorID),
		)
	}
	return v, nil
}

// CanReview сообщает, вправе ли пользователь оценить вторую сторону, и кого
// именно он оценивает. Нужен сервису отзывов, чтобы оценка опиралась
// на состоявшийся обмен, а не на желание поставить звёзды.
func (s *chainService) CanReview(ctx context.Context, id, actorID string) (string, error) {
	if blank(id) || blank(actorID) {
		return "", ErrInvalidInput
	}

	chain, err := s.repo.GetByID(ctx, id, actorID)
	if err != nil {
		return "", normalizeError(err)
	}
	deal := dealOf(chain)
	if err := exchange.CanReview(deal, actorID); err != nil {
		return "", mapExchangeError(err)
	}
	return deal.Counterparty(actorID), nil
}

func (s *chainService) ExpireOffers(ctx context.Context) error {
	chains, err := s.repo.ExpirePending(ctx)
	if err != nil {
		return normalizeError(err)
	}
	for i := range chains {
		s.publish(events.ExchangeChainUpdated, &chains[i])
	}
	return nil
}

func (s *chainService) Delete(ctx context.Context, id, actorID string) error {
	if blank(id) || blank(actorID) {
		return ErrInvalidInput
	}
	chain, err := s.repo.GetByID(ctx, id, actorID)
	if err != nil {
		return normalizeError(err)
	}
	if err := s.repo.Delete(ctx, id, actorID); err != nil {
		return normalizeError(err)
	}
	s.publish(events.ExchangeChainDeleted, chain)
	return nil
}
