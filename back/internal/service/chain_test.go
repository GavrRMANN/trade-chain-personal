package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"trade-chain/internal/domain"
	"trade-chain/internal/exchange"
	"trade-chain/internal/repository"
)

// Подставные репозитории держат состояние в памяти: правила согласования
// проверяются целиком, вплоть до подмены статуса и обмена владельцами,
// но без базы — иначе эти проверки некуда было бы запускать в CI.

const (
	initiator = "cust-initiator"
	recipient = "cust-recipient"
	stranger  = "cust-stranger"

	offeredID   = "prod-offered"
	requestedID = "prod-requested"
	strangerID  = "prod-stranger"
	chainID     = "chain-1"
)

type fakeProductRepo struct {
	products map[string]domain.Product
}

func (f *fakeProductRepo) GetByID(_ context.Context, id string) (*domain.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return nil, errNoRows
	}
	return &p, nil
}

func (f *fakeProductRepo) Create(context.Context, *domain.CreateProductDTO) (*domain.Product, error) {
	return nil, nil
}
func (f *fakeProductRepo) GetByCustomerID(context.Context, string) ([]domain.Product, error) {
	return nil, nil
}
func (f *fakeProductRepo) GetOwnByCustomerID(context.Context, string) ([]domain.Product, error) {
	return nil, nil
}
func (f *fakeProductRepo) Update(context.Context, string, *domain.UpdateProductDTO) (*domain.Product, error) {
	return nil, nil
}
func (f *fakeProductRepo) Delete(context.Context, string, string) error { return nil }
func (f *fakeProductRepo) List(ctx context.Context, customerID *string, q string, categoryID *string, page int, limit int) ([]domain.Product, error) {
	return nil, nil
}
func (f *fakeProductRepo) Search(context.Context, string, *string) ([]domain.Product, error) {
	return nil, nil
}
func (f *fakeProductRepo) GetExchangeCandidates(context.Context, string) ([]domain.Product, error) {
	return nil, nil
}

type fakeChainRepo struct {
	chains           map[string]domain.Chain
	products         *fakeProductRepo
	completed        int  // сколько раз обмен доводился до конца
	rejectDuplicates bool // повторяет уникальный индекс на предложениях в ожидании
}

func (f *fakeChainRepo) Create(_ context.Context, c *domain.Chain) (*domain.Chain, error) {
	if f.rejectDuplicates {
		for _, existing := range f.chains {
			toProductMatch := existing.ToProductID == nil && c.ToProductID == nil ||
				existing.ToProductID != nil && c.ToProductID != nil && *existing.ToProductID == *c.ToProductID
			if existing.Status == string(domain.ChainPending) &&
				existing.InitiatorID == c.InitiatorID &&
				existing.FromProductID == c.FromProductID &&
				toProductMatch {
				return nil, domain.ErrOfferDuplicate
			}
		}
	}

	stored := *c
	stored.ChainID = fmt.Sprintf("chain-%d", len(f.chains)+1)
	f.chains[stored.ChainID] = stored
	return &stored, nil
}

func (f *fakeChainRepo) GetByID(_ context.Context, id string, customerID string) (*domain.Chain, error) {
	c, ok := f.chains[id]
	if !ok {
		return nil, errNoRows
	}
	return &c, nil
}

func (f *fakeChainRepo) UpdateStatus(_ context.Context, id string, status domain.ChainStatus) error {
	c, ok := f.chains[id]
	if !ok {
		return errNoRows
	}
	c.Status = string(status)
	f.chains[id] = c
	return nil
}

func (f *fakeChainRepo) UpdateStatusIfCurrent(ctx context.Context, id string, customerID string, current, next domain.ChainStatus) error {
	chain, err := f.GetByID(ctx, id, customerID)
	if err != nil {
		return err
	}
	if chain.Status != string(current) {
		return errNoRows
	}
	return f.UpdateStatus(ctx, id, next)
}

// CompleteExchange повторяет главное свойство настоящего: меняет владельцев
// товаров местами и закрывает звено.
func (f *fakeChainRepo) CompleteExchange(_ context.Context, id string) error {
	c, ok := f.chains[id]
	if !ok {
		return errNoRows
	}
	if c.Status != string(domain.ChainActive) {
		return errors.New("chain must be active to complete")
	}

	from := f.products.products[c.FromProductID]
	if c.ToProductID == nil {
		return errors.New("chain must have ToProductID to complete")
	}
	to := f.products.products[*c.ToProductID]
	from.CustomerID, to.CustomerID = to.CustomerID, from.CustomerID
	from.Status, to.Status = domain.ProductExchanged, domain.ProductExchanged
	f.products.products[c.FromProductID] = from
	f.products.products[*c.ToProductID] = to

	c.Status = string(domain.ChainCompleted)
	f.chains[id] = c
	f.completed++

	// Настоящий репозиторий закрывает конкурирующие предложения по тем же
	// товарам в этой же транзакции; фейк повторяет это, иначе правило негде
	// проверить.
	for otherID, other := range f.chains {
		if otherID == id || (other.Status != string(domain.ChainPending) && other.Status != string(domain.ChainActive)) {
			continue
		}
		if touchesSameProducts(other, c) {
			if isOutgoingCompetingOffer(other, c) {
				other.Status = string(domain.ChainCancelled)
			} else {
				other.Status = string(domain.ChainUnavailable)
			}
			f.chains[otherID] = other
		}
	}
	return nil
}

func touchesSameProducts(a, b domain.Chain) bool {
	products := []string{b.FromProductID}
	if b.ToProductID != nil {
		products = append(products, *b.ToProductID)
	}
	for _, product := range products {
		if a.FromProductID == product || (a.ToProductID != nil && *a.ToProductID == product) {
			return true
		}
	}
	return false
}

func isOutgoingCompetingOffer(other, completed domain.Chain) bool {
	if other.Status == string(domain.ChainPending) {
		return true
	}
	if other.InitiatorID == completed.InitiatorID && other.FromProductID == completed.FromProductID {
		return true
	}
	return completed.ToProductID != nil &&
		completed.RecipientID != nil &&
		other.InitiatorID == *completed.RecipientID &&
		other.FromProductID == *completed.ToProductID
}

func (f *fakeChainRepo) GetByProductID(context.Context, string, string) ([]domain.Chain, error) {
	return nil, nil
}
func (f *fakeChainRepo) GetByCustomerID(ctx context.Context, customerID string) ([]domain.Chain, error) {
	return f.List(ctx, repository.ChainFilter{CustomerID: customerID})
}

func (f *fakeChainRepo) List(_ context.Context, filter repository.ChainFilter) ([]domain.Chain, error) {
	found := make([]domain.Chain, 0, len(f.chains))
	for _, c := range f.chains {
		if !matchesRole(c, filter) || !matchesStatus(c, filter.Statuses) {
			continue
		}
		found = append(found, c)
	}
	return found, nil
}

func matchesRole(c domain.Chain, filter repository.ChainFilter) bool {
	switch filter.Role {
	case domain.RoleIncoming:
		return c.RecipientID != nil && *c.RecipientID == filter.CustomerID
	case domain.RoleOutgoing:
		return c.InitiatorID == filter.CustomerID
	default:
		return c.InitiatorID == filter.CustomerID || (c.RecipientID != nil && *c.RecipientID == filter.CustomerID)
	}
}

func matchesStatus(c domain.Chain, statuses []domain.ChainStatus) bool {
	if len(statuses) == 0 {
		return true
	}
	for _, status := range statuses {
		if c.Status == string(status) {
			return true
		}
	}
	return false
}
func (f *fakeChainRepo) GetFullChain(context.Context, string) ([]domain.Chain, error) {
	return nil, nil
}
func (f *fakeChainRepo) ExpirePending(context.Context) ([]domain.Chain, error) { return nil, nil }
func (f *fakeChainRepo) Delete(context.Context, string, string) error          { return nil }

type fakeNegotiationRepo struct {
	messages      []domain.ChainMessage
	confirmations []domain.ChainConfirmation
}

func (f *fakeNegotiationRepo) AddMessage(_ context.Context, m *domain.ChainMessage) (*domain.ChainMessage, error) {
	stored := *m
	f.messages = append(f.messages, stored)
	return &stored, nil
}

func (f *fakeNegotiationRepo) ListMessages(context.Context, string) ([]domain.ChainMessage, error) {
	return f.messages, nil
}

func (f *fakeNegotiationRepo) Confirm(_ context.Context, c *domain.ChainConfirmation) error {
	for _, existing := range f.confirmations {
		if existing.CustomerID == c.CustomerID {
			return errors.New("duplicate key value violates unique constraint")
		}
	}
	f.confirmations = append(f.confirmations, *c)
	return nil
}

func (f *fakeNegotiationRepo) ListConfirmations(context.Context, string) ([]domain.ChainConfirmation, error) {
	return f.confirmations, nil
}

var errNoRows = errors.New("sql: no rows in result set")

// strPtr возвращает указатель на строку — нужен для инициализации *string полей
// из констант, адрес которых брать нельзя.
func strPtr(s string) *string { return &s }

type fixture struct {
	service      ChainService
	chains       *fakeChainRepo
	products     *fakeProductRepo
	negotiations *fakeNegotiationRepo
}

// newFixture готовит принятое предложение: инициатор предложил свой товар
// за чужой, владелец согласился, стороны договариваются о встрече.
func newFixture(status domain.ChainStatus) *fixture {
	products := &fakeProductRepo{products: map[string]domain.Product{
		offeredID:   {ProductID: offeredID, CustomerID: initiator, Status: domain.ProductActive},
		requestedID: {ProductID: requestedID, CustomerID: recipient, Status: domain.ProductActive},
		strangerID:  {ProductID: strangerID, CustomerID: stranger, Status: domain.ProductActive},
	}}
	chains := &fakeChainRepo{
		products: products,
		chains: map[string]domain.Chain{
			chainID: {
				ChainID:       chainID,
				FromProductID: offeredID,
				ToProductID:   strPtr(requestedID),
				InitiatorID:   initiator,
				RecipientID:   strPtr(recipient),
				Status:        string(status),
			},
		},
	}
	negotiations := &fakeNegotiationRepo{}

	// Создаём реальный сервис с фиктивными репозиториями
	service := NewChainService(chains, products, negotiations)

	return &fixture{
		service:      service,
		chains:       chains,
		products:     products,
		negotiations: negotiations,
	}
}

func TestCreateStartsFromPendingAndRemembersRecipient(t *testing.T) {
	f := newFixture(domain.ChainPending)

	// Попытка создать сразу завершённым не должна проходить: согласие второй
	// стороны иначе становится необязательным.
	created, err := f.service.Create(context.Background(), &domain.Chain{
		FromProductID: offeredID,
		ToProductID:   strPtr(requestedID),
		InitiatorID:   initiator,
		Status:        string(domain.ChainCompleted),
	})
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if created.Status != string(domain.ChainPending) {
		t.Errorf("статус %q, ожидался %q", created.Status, domain.ChainPending)
	}
	if created.RecipientID == nil || *created.RecipientID != recipient {
		if created.RecipientID == nil {
			t.Errorf("вторая сторона nil, ожидалась %q", recipient)
		} else {
			t.Errorf("вторая сторона %q, ожидалась %q", *created.RecipientID, recipient)
		}
	}
	if created.ExpiresAt.IsZero() {
		t.Error("предложению не проставлен срок ответа")
	}
}

func TestCreateRejectsSomeoneElsesProduct(t *testing.T) {
	f := newFixture(domain.ChainPending)

	// Инициатор предлагает чужую вещь третьему человеку.
	_, err := f.service.Create(context.Background(), &domain.Chain{
		FromProductID: requestedID, // товар получателя, не инициатора
		ToProductID:   strPtr(strangerID),
		InitiatorID:   initiator,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("ошибка %v, ожидалась ErrForbidden", err)
	}
}

func TestOnlyRecipientAcceptsOffer(t *testing.T) {
	f := newFixture(domain.ChainPending)

	if err := f.service.UpdateStatus(context.Background(), chainID, domain.ChainActive, initiator); !errors.Is(err, ErrForbidden) {
		t.Fatalf("инициатор не может принять своё предложение: %v", err)
	}
	if err := f.service.UpdateStatus(context.Background(), chainID, domain.ChainActive, recipient); err != nil {
		t.Fatalf("получатель должен уметь принять предложение: %v", err)
	}
	if got := f.chains.chains[chainID].Status; got != string(domain.ChainActive) {
		t.Errorf("статус %q, ожидался %q", got, domain.ChainActive)
	}
}

// Ради этого всё и затевалось: закрыть обмен в одиночку больше нельзя.
func TestCompletedRequiresBothConfirmations(t *testing.T) {
	f := newFixture(domain.ChainActive)
	ctx := context.Background()

	if err := f.service.UpdateStatus(ctx, chainID, domain.ChainCompleted, initiator); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("завершить обмен сменой статуса нельзя, получено: %v", err)
	}

	chain, err := f.service.Confirm(ctx, chainID, initiator, true, "")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if chain.Status != string(domain.ChainActive) {
		t.Errorf("после одного подтверждения статус %q, ожидался %q", chain.Status, domain.ChainActive)
	}
	if f.chains.completed != 0 {
		t.Error("обмен проведён по одному подтверждению")
	}

	chain, err = f.service.Confirm(ctx, chainID, recipient, true, "")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if chain.Status != string(domain.ChainCompleted) {
		t.Errorf("статус %q, ожидался %q", chain.Status, domain.ChainCompleted)
	}
	if f.products.products[offeredID].CustomerID != recipient {
		t.Error("товары не поменяли владельцев после завершения обмена")
	}
}

// Состоявшийся обмен закрывает предложения по тем же вещам: принять их уже
// нельзя, вещи у новых владельцев.
func TestCompletionClosesCompetingOffers(t *testing.T) {
	f := newFixture(domain.ChainActive)
	ctx := context.Background()

	competing, err := f.service.Create(ctx, &domain.Chain{
		FromProductID: strangerID,
		ToProductID:   strPtr(requestedID),
		InitiatorID:   stranger,
	})
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if _, err := f.service.Confirm(ctx, chainID, initiator, true, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Confirm(ctx, chainID, recipient, true, ""); err != nil {
		t.Fatal(err)
	}

	if got := f.chains.chains[competing.ChainID].Status; got != string(domain.ChainCancelled) {
		t.Errorf("конкурирующее предложение в статусе %q, ожидался %q", got, domain.ChainCancelled)
	}
}

func TestCompletionMarksIncomingOfferAsUnavailable(t *testing.T) {
	f := newFixture(domain.ChainActive)
	ctx := context.Background()

	incoming := domain.Chain{
		ChainID:       "chain-incoming",
		FromProductID: strangerID,
		ToProductID:   strPtr(requestedID),
		InitiatorID:   stranger,
		RecipientID:   strPtr(recipient),
		Status:        string(domain.ChainActive),
	}
	f.chains.chains[incoming.ChainID] = incoming

	if _, err := f.service.Confirm(ctx, chainID, initiator, true, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Confirm(ctx, chainID, recipient, true, ""); err != nil {
		t.Fatal(err)
	}

	if got := f.chains.chains[incoming.ChainID].Status; got != string(domain.ChainUnavailable) {
		t.Errorf("входящий обмен в статусе %q, ожидался %q", got, domain.ChainUnavailable)
	}
	if f.products.products[requestedID].Status != domain.ProductExchanged {
		t.Errorf("статус товара %q, ожидался %q", f.products.products[requestedID].Status, domain.ProductExchanged)
	}
}

func TestIncomingOfferCannotBeAcceptedAfterProductIsExchanged(t *testing.T) {
	f := newFixture(domain.ChainActive)
	ctx := context.Background()
	incomingID := "chain-incoming-pending"
	f.chains.chains[incomingID] = domain.Chain{
		ChainID:       incomingID,
		FromProductID: strangerID,
		ToProductID:   strPtr(requestedID),
		InitiatorID:   stranger,
		RecipientID:   strPtr(recipient),
		Status:        string(domain.ChainPending),
	}
	f.products.products[requestedID] = domain.Product{
		ProductID: requestedID, CustomerID: recipient, Status: domain.ProductExchanged,
	}

	if _, err := f.service.Decide(ctx, incomingID, exchange.ActionAccept, recipient); !errors.Is(err, ErrConflict) {
		t.Fatalf("принятие предложения с недоступным товаром вернуло %v, ожидался конфликт", err)
	}
	if got := f.chains.chains[incomingID].Status; got != string(domain.ChainPending) {
		t.Errorf("входящее предложение изменило статус на %q", got)
	}
}

func TestCompletionCancelsCompetingActiveOffers(t *testing.T) {
	f := newFixture(domain.ChainActive)
	ctx := context.Background()

	// Второй обмен уже принят получателем, поэтому он не должен остаться
	// доступным после успешного завершения первого.
	f.chains.chains["chain-active-competing"] = domain.Chain{
		ChainID:       "chain-active-competing",
		FromProductID: requestedID,
		ToProductID:   strPtr(strangerID),
		InitiatorID:   recipient,
		RecipientID:   strPtr(stranger),
		Status:        string(domain.ChainActive),
	}

	if _, err := f.service.Confirm(ctx, chainID, initiator, true, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Confirm(ctx, chainID, recipient, true, ""); err != nil {
		t.Fatal(err)
	}

	if got := f.chains.chains["chain-active-competing"].Status; got != string(domain.ChainCancelled) {
		t.Errorf("принятый конкурирующий обмен в статусе %q, ожидался %q", got, domain.ChainCancelled)
	}
}

func TestSingleNegativeConfirmationFailsExchange(t *testing.T, customerID string) {
	f := newFixture(domain.ChainActive)
	ctx := context.Background()

	if _, err := f.service.Confirm(ctx, chainID, initiator, false, ""); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	chain, _ := f.service.GetByID(ctx, chainID, customerID)
	if chain.Status != string(domain.ChainFailed) {
		t.Errorf("статус %q, ожидался %q — не состоявшийся обмен решает одна сторона", chain.Status, domain.ChainFailed)
	}
	if f.products.products[offeredID].CustomerID != initiator {
		t.Error("товары поменяли владельцев, хотя обмен не состоялся")
	}
}

func TestSecondConfirmationFromSameSideIsConflict(t *testing.T) {
	f := newFixture(domain.ChainActive)
	ctx := context.Background()

	if _, err := f.service.Confirm(ctx, chainID, initiator, true, ""); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if _, err := f.service.Confirm(ctx, chainID, initiator, true, ""); !errors.Is(err, ErrConflict) {
		t.Fatalf("ошибка %v, ожидалась ErrConflict", err)
	}
}

func TestRepeatedConfirmationRetriesFailedSettlement(t *testing.T) {
	f := newFixture(domain.ChainActive)
	ctx := context.Background()
	f.negotiations.confirmations = []domain.ChainConfirmation{
		{ChainID: chainID, CustomerID: initiator, Success: true},
		{ChainID: chainID, CustomerID: recipient, Success: true},
	}

	chain, err := f.service.Confirm(ctx, chainID, recipient, true, "")
	if err != nil {
		t.Fatalf("повторное подтверждение должно завершить обмен: %v", err)
	}
	if chain.Status != string(domain.ChainCompleted) {
		t.Errorf("статус %q, ожидался %q", chain.Status, domain.ChainCompleted)
	}
}

func TestStrangerCannotConfirmOrRead(t *testing.T) {
	f := newFixture(domain.ChainActive)
	ctx := context.Background()

	if _, err := f.service.Confirm(ctx, chainID, stranger, true, ""); !errors.Is(err, ErrForbidden) {
		t.Fatalf("ошибка %v, ожидалась ErrForbidden", err)
	}
	if _, err := f.service.Messages(ctx, chainID, stranger); !errors.Is(err, ErrForbidden) {
		t.Fatalf("переписка не должна быть видна постороннему, получено: %v", err)
	}
}

func TestChatClosesWithTheDeal(t *testing.T) {
	f := newFixture(domain.ChainActive)
	ctx := context.Background()

	if _, err := f.service.SendMessage(ctx, chainID, initiator, "  во сколько удобно?  "); err != nil {
		t.Fatalf("по активной сделке писать можно: %v", err)
	}
	if got := f.negotiations.messages[0].Body; got != "во сколько удобно?" {
		t.Errorf("тело сообщения %q, ожидалось без обрамляющих пробелов", got)
	}
	if _, err := f.service.SendMessage(ctx, chainID, initiator, "   "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("пустое сообщение не должно проходить, получено: %v", err)
	}

	if err := f.chains.UpdateStatus(ctx, chainID, domain.ChainCompleted); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.SendMessage(ctx, chainID, initiator, "ещё вопрос"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("по закрытой сделке писать нельзя, получено: %v", err)
	}
}

// После обмена владельцы товаров меняются местами, поэтому вторую сторону
// нельзя вычислять по текущему владельцу — иначе оценивать будет некого.
func TestCanReviewAfterCompletedExchange(t *testing.T) {
	f := newFixture(domain.ChainActive)
	ctx := context.Background()

	if _, err := f.service.Confirm(ctx, chainID, initiator, true, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Confirm(ctx, chainID, recipient, true, ""); err != nil {
		t.Fatal(err)
	}

	counterparty, err := f.service.CanReview(ctx, chainID, initiator)
	if err != nil {
		t.Fatalf("после состоявшегося обмена отзыв разрешён: %v", err)
	}
	if counterparty != recipient {
		t.Errorf("оценивается %q, ожидался %q", counterparty, recipient)
	}
}

func TestCannotReviewUnfinishedExchange(t *testing.T) {
	f := newFixture(domain.ChainActive)

	if _, err := f.service.CanReview(context.Background(), chainID, initiator); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("ошибка %v, ожидалась ErrInvalidInput", err)
	}
}

func TestDecideRejectsUnknownAction(t *testing.T) {
	f := newFixture(domain.ChainPending)

	if _, err := f.service.Decide(context.Background(), chainID, exchange.Action("approve"), recipient); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("ошибка %v, ожидалась ErrInvalidInput", err)
	}
}
