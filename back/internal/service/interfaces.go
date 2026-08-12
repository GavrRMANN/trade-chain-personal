package service

import (
	"context"

	"trade-chain/internal/domain"
	"trade-chain/internal/exchange"
)

type CustomerService interface {
	Create(context.Context, *domain.CreateCustomerDTO) (*domain.Customer, error)
	GetByID(context.Context, string) (*domain.Customer, error)
	GetByEmail(context.Context, string) (*domain.Customer, error) // добавить
	Update(context.Context, string, *domain.UpdateCustomerDTO) (*domain.Customer, error)
	Delete(context.Context, string) error
	List(context.Context, int, int) ([]domain.Customer, error)
	ListOverview(context.Context, int, int) ([]domain.CustomerOverview, error)
}

type ProductService interface {
	Create(context.Context, *domain.CreateProductDTO) (*domain.Product, error)
	GetByID(context.Context, string) (*domain.Product, error)
	GetByCustomerID(context.Context, string) ([]domain.Product, error)
	GetOwnByCustomerID(context.Context, string) ([]domain.Product, error)
	Update(context.Context, string, *domain.UpdateProductDTO) (*domain.Product, error)
	Delete(context.Context, string, string) error
	List(context.Context, *string, string, *string, int, int) ([]domain.Product, error)
	GetExchangeCandidates(context.Context, string) ([]domain.Product, error)
}

type ChainService interface {
	Create(context.Context, *domain.Chain) (*domain.Chain, error)
	GetByID(context.Context, string, string) (*domain.Chain, error)
	GetByProductID(context.Context, string, string) ([]domain.Chain, error)
	GetByCustomerID(context.Context, string) ([]domain.Chain, error)
	GetFullChain(context.Context, string) ([]domain.Chain, error)
	UpdateStatus(context.Context, string, domain.ChainStatus, string) error // добавили userID
	Decide(context.Context, string, exchange.Action, string) (*domain.Chain, error)
	Confirm(ctx context.Context, chainID, actorID string, success bool, reason string) (*domain.Chain, error)
	Messages(ctx context.Context, chainID, actorID string) ([]domain.ChainMessage, error)
	SendMessage(ctx context.Context, chainID, actorID, body string) (*domain.ChainMessage, error)
	ExpireOffers(ctx context.Context) error
	CanReview(ctx context.Context, chainID, actorID string) (string, error)
	Delete(ctx context.Context, chainID, actorID string) error
}

type NotificationService interface {
	ListReads(ctx context.Context, customerID string) ([]domain.NotificationRead, error)
	MarkRead(ctx context.Context, customerID, chainID string, kind domain.NotificationKind) error
	MarkAllRead(ctx context.Context, customerID string) error
}

// OfferService — экран предложений: отправить, ответить, довести до итога.
//
// Отдельный интерфейс от ChainService нужен транспорту: маршруты
// /exchange-offers работают в терминах предложения и обмена, а не звена,
// и не должны иметь доступа к операциям вроде удаления цепочки.
type OfferService interface {
	Create(ctx context.Context, in CreateOfferInput) (*Offer, error)
	List(ctx context.Context, actorID string, role domain.OfferRole, statuses []exchange.OfferStatus) ([]Offer, error)
	Details(ctx context.Context, offerID, actorID string) (*OfferDetails, error)
	Accept(ctx context.Context, offerID, actorID string) (*Offer, error)
	Decline(ctx context.Context, offerID, actorID string) (*Offer, error)
	Cancel(ctx context.Context, offerID, actorID string) (*Offer, error)
	Confirm(ctx context.Context, exchangeID, actorID string, success bool, reason string) (*Offer, error)
}

type ReviewService interface {
	Create(context.Context, *domain.Review) (*domain.Review, error)
	GetByID(context.Context, string) (*domain.Review, error)
	GetByCustomerID(context.Context, string) ([]domain.Review, error)
	GetAverageRating(context.Context, string) (float64, error)
	Delete(context.Context, string) error
}

type CategoryService interface {
	Create(context.Context, *domain.Category) (*domain.Category, error)
	GetByID(context.Context, string) (*domain.Category, error)
	GetSubcategories(context.Context, string) ([]domain.Category, error)
	Update(context.Context, string, *domain.Category) (*domain.Category, error)
	Delete(context.Context, string) error
	Search(context.Context, string) ([]domain.Category, error)
	List(context.Context) ([]domain.Category, error)
}

type WishlistService interface {
	Create(context.Context, *domain.Wishlist) (*domain.Wishlist, error)
	GetByID(context.Context, string) (*domain.Wishlist, error)
	GetByProductID(context.Context, string) (*domain.Wishlist, error)
	AddCategoryOption(context.Context, string, string) error
	RemoveCategoryOption(context.Context, string, string) error
	GetOptions(context.Context, string) ([]domain.Category, error)
	Delete(context.Context, string) error
	UpdateByProductID(ctx context.Context, productID string, dto *domain.CreateWishlistDTO) (*domain.Wishlist, error)
}
