package search

import (
	"context"

	"trade-chain/internal/domain"
	"trade-chain/internal/service"
)

type SearchService struct {
	productService  service.ProductService
	categoryService service.CategoryService
}

type ProductSearchResult struct {
	Products []domain.Product
	Length   int
}

type CategorySearchResult struct {
	Categories []domain.Category
	Length     int
}

func NewSearchService(
	productService service.ProductService,
	categoryService service.CategoryService) *SearchService {
	return &SearchService{
		productService:  productService,
		categoryService: categoryService,
	}
}

func (s *SearchService) FindChain(
	ctx context.Context,
	customerID string,
	sourceProductID string,
	targetProductID string,
	maxDepth int,
) (*ProductSearchResult, error) {

	source, err := s.productService.GetByID(ctx, sourceProductID)
	if err != nil {
		return nil, err
	}
	if source.CustomerID != customerID || source.Status != domain.ProductActive {
		return nil, service.ErrForbidden
	}

	target, err := s.productService.GetByID(ctx, targetProductID)
	if err != nil {
		return nil, err
	}

	if source.ProductID == target.ProductID {
		return &ProductSearchResult{Products: []domain.Product{*source}, Length: 0}, nil
	}

	return findChainBFS(
		ctx,
		s.productService,
		*source,
		*target,
		maxDepth,
	)
}

// FindChainToTarget keeps the recommendation endpoint's legacy behaviour.
func (s *SearchService) FindChainToTarget(
	ctx context.Context,
	customerID string,
	targetProductID string,
	maxDepth int,
) (*ProductSearchResult, error) {
	target, err := s.productService.GetByID(ctx, targetProductID)
	if err != nil {
		return nil, err
	}
	myProducts, err := s.productService.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if len(myProducts) == 0 {
		return &ProductSearchResult{
			Products: []domain.Product{},
			Length:   0,
		}, nil
	}
	return findLegacyChainBFS(ctx, s.productService, *target, myProducts, maxDepth)
}
