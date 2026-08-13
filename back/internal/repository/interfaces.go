package repository

import (
	"context"

	"trade-chain/internal/domain"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *domain.CreateCustomerDTO) (*domain.Customer, error)
	GetByID(ctx context.Context, id string) (*domain.Customer, error)
	GetByEmail(ctx context.Context, email string) (*domain.Customer, error)
	Update(ctx context.Context, id string, customer *domain.UpdateCustomerDTO) (*domain.Customer, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]domain.Customer, error)
	ListOverview(ctx context.Context, offset, limit int) ([]domain.CustomerOverview, error)

	// Работа с вишлистами пользователя
	GetCustomerWishlistOptions(ctx context.Context, customerID string) ([]domain.CustomerWishlistOption, error)
	AddCustomerWishlistOption(ctx context.Context, customerID, categoryID string) error
	DeleteCustomerWishlistOption(ctx context.Context, customerID, categoryID string) error
	ReplaceCustomerWishlistOptions(ctx context.Context, customerID string, dto *domain.UpdateCustomerWishlistDTO) error
}

type ProductRepository interface {
	Create(ctx context.Context, product *domain.CreateProductDTO) (*domain.Product, error)
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetByCustomerID(ctx context.Context, customerID string) ([]domain.Product, error)
	GetOwnByCustomerID(ctx context.Context, customerID string) ([]domain.Product, error)
	Update(ctx context.Context, id string, product *domain.UpdateProductDTO) (*domain.Product, error)
	Delete(ctx context.Context, id string, customerID string) error
	List(ctx context.Context, customerID *string, q string, categoryID *string, page int, limit int) ([]domain.Product, error)
	GetExchangeCandidates(ctx context.Context, productID string) ([]domain.Product, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) (*domain.Category, error)
	GetByID(ctx context.Context, id string) (*domain.Category, error)
	GetSubcategories(ctx context.Context, parentID string) ([]domain.Category, error)
	Update(ctx context.Context, id string, category *domain.Category) (*domain.Category, error)
	Search(ctx context.Context, search string) ([]domain.Category, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.Category, error)
}

type WishlistRepository interface {
	Create(ctx context.Context, wishlist *domain.Wishlist) (*domain.Wishlist, error)
	GetByID(ctx context.Context, id string) (*domain.Wishlist, error)
	GetByProductID(ctx context.Context, productID string) (*domain.Wishlist, error)
	UpdateByProductID(ctx context.Context, productID string, name string, categoryIDs []string) (*domain.Wishlist, error)
	AddCategoryOption(ctx context.Context, wishlistID, categoryID string) error
	RemoveCategoryOption(ctx context.Context, wishlistID, categoryID string) error
	GetOptions(ctx context.Context, wishlistID string) ([]domain.Category, error)
	Delete(ctx context.Context, id string) error
}

// ChainFilter — отбор сделок для списков предложений.
//
// Пустые поля означают «без ограничения»: так один запрос покрывает и экран
// «мои обмены» целиком, и входящие в ожидании ответа.
type ChainFilter struct {
	CustomerID string
	Role       domain.OfferRole
	Statuses   []domain.ChainStatus
}

type ChainRepository interface {
	Create(ctx context.Context, chain *domain.Chain) (*domain.Chain, error)
	GetByID(ctx context.Context, id string, customerID string) (*domain.Chain, error)
	GetByProductID(ctx context.Context, productID string, customerID string) ([]domain.Chain, error)
	GetByCustomerID(ctx context.Context, customerID string) ([]domain.Chain, error)
	List(ctx context.Context, filter ChainFilter) ([]domain.Chain, error)
	GetFullChain(ctx context.Context, chainID string) ([]domain.Chain, error)
	UpdateStatus(ctx context.Context, id string, status domain.ChainStatus) error
	UpdateStatusIfCurrent(ctx context.Context, id string, customerID string, current, next domain.ChainStatus) error
	CompleteExchange(ctx context.Context, chainID string) error // добавить
	ExpirePending(ctx context.Context) ([]domain.Chain, error)
	Delete(ctx context.Context, id, initiatorID string) error
}

// NegotiationRepository хранит то, чем звено обрастает в переговорах:
// переписку сторон и их решения об итоге обмена.
type NegotiationRepository interface {
	AddMessage(ctx context.Context, message *domain.ChainMessage) (*domain.ChainMessage, error)
	ListMessages(ctx context.Context, chainID string) ([]domain.ChainMessage, error)
	Confirm(ctx context.Context, confirmation *domain.ChainConfirmation) error
	ListConfirmations(ctx context.Context, chainID string) ([]domain.ChainConfirmation, error)
}

type NotificationRepository interface {
	ListReads(ctx context.Context, customerID string) ([]domain.NotificationRead, error)
	MarkRead(ctx context.Context, customerID, chainID string, kind domain.NotificationKind) error
}

type ReviewRepository interface {
	Create(ctx context.Context, review *domain.Review) (*domain.Review, error)
	GetByID(ctx context.Context, id string) (*domain.Review, error)
	GetByCustomerID(ctx context.Context, customerID string) ([]domain.Review, error)
	GetAverageRating(ctx context.Context, customerID string) (float64, error)
	Delete(ctx context.Context, id string) error
}
