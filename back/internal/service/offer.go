package service

import (
	"context"
	"time"

	"trade-chain/internal/domain"
	"trade-chain/internal/exchange"
	"trade-chain/internal/repository"
)

// Слой предложений не заводит собственных правил согласования: принять,
// отклонить и отозвать умеет chainService, и второе место, где решается,
// кому что можно, рано или поздно разошлось бы с первым. Здесь остаётся
// прикладная работа — собрать список, собрать карточку, перевести
// пожелания клиента в отбор для базы.

type offerService struct {
	chains       ChainService
	repo         repository.ChainRepository
	negotiations repository.NegotiationRepository
	now          func() time.Time
}

func NewOfferService(
	chains ChainService,
	repo repository.ChainRepository,
	negotiations repository.NegotiationRepository,
) OfferService {
	return &offerService{chains: chains, repo: repo, negotiations: negotiations, now: time.Now}
}

// CreateOfferInput — предложение обмена, каким его отправляет клиент.
type CreateOfferInput struct {
	InitiatorID         string
	OfferedProductID    string
	RequestedProductID  string
	RequestedCategoryID string
	ExchangeGoalID      *string
	RouteStepID         *string
	Surcharge           domain.Surcharge
	Comment             string
}

// Offer — предложение и, если оно принято, состояние обмена по нему.
type Offer struct {
	Chain    domain.Chain
	Status   exchange.OfferStatus
	Exchange exchange.ExchangeStatus // пусто, пока предложение не принято
}

// OfferDetails — карточка предложения: состояние, переписка и решения сторон.
type OfferDetails struct {
	Offer
	Messages      []domain.ChainMessage
	Confirmations []domain.ChainConfirmation
}

func (s *offerService) Create(ctx context.Context, in CreateOfferInput) (*Offer, error) {
	var toProductID *string
	var toCategoryID *string
	if in.RequestedProductID != "" {
		toProductID = &in.RequestedProductID
	}
	if in.RequestedCategoryID != "" {
		toCategoryID = &in.RequestedCategoryID
	}

	chain, err := s.chains.Create(ctx, &domain.Chain{
		FromProductID:  in.OfferedProductID,
		ToProductID:    toProductID,
		ToCategoryID:   toCategoryID,
		InitiatorID:    in.InitiatorID,
		ExchangeGoalID: in.ExchangeGoalID,
		RouteStepID:    in.RouteStepID,
		Surcharge:      in.Surcharge,
		Message:        in.Comment,
	})
	if err != nil {
		return nil, err
	}
	return s.view(ctx, chain)
}

// List отдаёт предложения человека, отобранные по стороне и состоянию.
func (s *offerService) List(ctx context.Context, actorID string, role domain.OfferRole, statuses []exchange.OfferStatus) ([]Offer, error) {
	if blank(actorID) {
		return nil, ErrInvalidInput
	}

	chains, err := s.repo.List(ctx, repository.ChainFilter{
		CustomerID: actorID,
		Role:       role,
		Statuses:   exchange.ChainStatusesFor(statuses),
	})
	if err != nil {
		return nil, normalizeError(err)
	}

	// База отбирает по своим статусам, а истёкший срок виден только на чтении,
	// поэтому окончательный отбор — здесь: иначе просроченное предложение
	// пришло бы в ответ на запрос pending.
	offers := make([]Offer, 0, len(chains))
	for i := range chains {
		offer := s.offerOf(&chains[i], nil)
		if !matchesOfferStatus(offer.Status, statuses) {
			continue
		}
		offers = append(offers, offer)
	}
	return offers, nil
}

// Details собирает карточку предложения целиком. Доступна только участникам:
// внутри переписка о встрече.
func (s *offerService) Details(ctx context.Context, offerID, actorID string) (*OfferDetails, error) {
	if blank(offerID) || blank(actorID) {
		return nil, ErrInvalidInput
	}

	chain, err := s.chains.GetByID(ctx, offerID, actorID)
	if err != nil {
		return nil, err
	}
	if !dealOf(chain).Involves(actorID) {
		return nil, mapExchangeError(domain.ErrNotParticipant)
	}

	messages, err := s.chains.Messages(ctx, offerID, actorID)
	if err != nil {
		return nil, err
	}
	confirmations, err := s.negotiations.ListConfirmations(ctx, offerID)
	if err != nil {
		return nil, normalizeError(err)
	}

	return &OfferDetails{
		Offer:         s.offerOf(chain, confirmations),
		Messages:      messages,
		Confirmations: confirmations,
	}, nil
}

func (s *offerService) Accept(ctx context.Context, offerID, actorID string) (*Offer, error) {
	return s.decide(ctx, offerID, exchange.ActionAccept, actorID)
}

func (s *offerService) Decline(ctx context.Context, offerID, actorID string) (*Offer, error) {
	return s.decide(ctx, offerID, exchange.ActionDecline, actorID)
}

func (s *offerService) Cancel(ctx context.Context, offerID, actorID string) (*Offer, error) {
	return s.decide(ctx, offerID, exchange.ActionCancel, actorID)
}

// Confirm подтверждает итог обмена от лица одной из сторон.
func (s *offerService) Confirm(ctx context.Context, exchangeID, actorID string, success bool, reason string) (*Offer, error) {
	chain, err := s.chains.Confirm(ctx, exchangeID, actorID, success, reason)
	if err != nil {
		return nil, err
	}
	return s.view(ctx, chain)
}

func (s *offerService) decide(ctx context.Context, offerID string, action exchange.Action, actorID string) (*Offer, error) {
	chain, err := s.chains.Decide(ctx, offerID, action, actorID)
	if err != nil {
		return nil, err
	}
	return s.view(ctx, chain)
}

// view достраивает звено до предложения, дочитывая подтверждения.
//
// Они нужны только принятым предложениям: пока никто не согласился,
// подтверждать нечего, и лишний запрос к базе ничего не добавит.
func (s *offerService) view(ctx context.Context, chain *domain.Chain) (*Offer, error) {
	var confirmations []domain.ChainConfirmation
	if domain.ChainStatus(chain.Status) == domain.ChainActive {
		var err error
		confirmations, err = s.negotiations.ListConfirmations(ctx, chain.ChainID)
		if err != nil {
			return nil, normalizeError(err)
		}
	}

	offer := s.offerOf(chain, confirmations)
	return &offer, nil
}

func (s *offerService) offerOf(chain *domain.Chain, confirmations []domain.ChainConfirmation) Offer {
	deal := dealOf(chain)
	status, _ := exchange.ExchangeStatusOf(deal, confirmations)

	return Offer{
		Chain:    *chain,
		Status:   exchange.OfferStatusOf(deal, s.now()),
		Exchange: status,
	}
}

func matchesOfferStatus(status exchange.OfferStatus, wanted []exchange.OfferStatus) bool {
	if len(wanted) == 0 {
		return true
	}
	for _, w := range wanted {
		if w == status {
			return true
		}
	}
	return false
}
