package search

import (
	"context"
	"sort"

	"trade-chain/internal/domain"
	"trade-chain/internal/service"
)

// defaultCandidatesLimit — сколько кандидатов отдавать, если фронт не
// указал limit. Совпадает с порогом, который раньше был зашит на фронте при
// ручном отборе рекомендаций по совпадению категории.
const defaultCandidatesLimit = 8

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

// FindCandidates подбирает следующий шаг обмена для конкретного товара:
// сперва вещи, которые владельцы явно хотят получить взамен (совпадение по
// вишлисту, см. GetExchangeCandidates), отсортированные по CalculateScore,
// а если их не хватает до limit — остальные активные товары каталога,
// кроме собственных вещей владельца source-товара.
//
// directOnly отключает добор каталогом. Совпадение по вишлисту — это ребро
// того же графа, по которому ищется цепочка: только до таких вещей от
// текущего товара есть прямой путь. Остальной каталог — просто товары
// рядом, и предлагать их следующим шагом маршрута нельзя: обмен с ними
// никуда не ведёт, а человек видит их в одном ряду с подтверждённым шагом.
func (s *SearchService) FindCandidates(
	ctx context.Context,
	productID string,
	limit int,
	directOnly bool,
) ([]domain.Product, error) {
	if limit <= 0 {
		limit = defaultCandidatesLimit
	}

	source, err := s.productService.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	wishlisted, err := s.productService.GetExchangeCandidates(ctx, productID)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(wishlisted, func(i, j int) bool {
		return CalculateScore(*source, wishlisted[i]) > CalculateScore(*source, wishlisted[j])
	})

	seen := map[string]bool{source.ProductID: true}
	result := make([]domain.Product, 0, limit)
	for _, candidate := range wishlisted {
		if seen[candidate.ProductID] {
			continue
		}
		seen[candidate.ProductID] = true
		result = append(result, candidate)
		if len(result) >= limit {
			return result, nil
		}
	}

	if directOnly {
		return result, nil
	}

	// Кандидатов по вишлисту не хватило — дозаполняем остальными активными
	// товарами каталога. List, получив CustomerID владельца source-товара,
	// сам исключит его собственные вещи и поднимет вверх то, что совпадает
	// с его вишлистом — тот же принцип релевантности, что и в основной ленте.
	rest, err := s.productService.List(ctx, &source.CustomerID, "", nil, 0, limit*2)
	if err != nil {
		return nil, err
	}
	for _, product := range rest {
		if seen[product.ProductID] || product.Status != domain.ProductActive {
			continue
		}
		seen[product.ProductID] = true
		result = append(result, product)
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}
