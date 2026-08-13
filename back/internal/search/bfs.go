package search

import (
	"context"

	"trade-chain/internal/domain"
	"trade-chain/internal/service"
)

type queueNode struct {
	Product domain.Product
	Depth   int
}

func findChainBFS(
	ctx context.Context,
	service service.ProductService,
	target domain.Product,
	source domain.Product,
	maxDepth int,
) (*ProductSearchResult, error) {

	queue := []queueNode{
		{
			Product: source,
			Depth:   0,
		},
	}

	visited := make(map[string]bool)
	parent := make(map[string]string)
	productMap := make(map[string]domain.Product)

	visited[source.ProductID] = true
	productMap[source.ProductID] = source

	for len(queue) > 0 {

		current := queue[0]
		queue = queue[1:]

		if current.Product.ProductID == target.ProductID {

			path := restorePath(
				current.Product.ProductID,
				parent,
				productMap,
			)

			return &ProductSearchResult{
				Products: reverseProducts(path),
				Length:   len(path) - 1,
			}, nil
		}

		if current.Depth >= maxDepth {
			continue
		}

		neighbors, err := service.GetExchangeCandidates(
			ctx,
			current.Product.ProductID,
		)
		if err != nil {
			return nil, err
		}

		for _, next := range neighbors {

			if visited[next.ProductID] {
				continue
			}

			visited[next.ProductID] = true

			parent[next.ProductID] = current.Product.ProductID

			productMap[next.ProductID] = next

			queue = append(queue, queueNode{
				Product: next,
				Depth:   current.Depth + 1,
			})
		}
	}

	return nil, nil
}

func findLegacyChainBFS(
	ctx context.Context,
	service service.ProductService,
	target domain.Product,
	userProducts []domain.Product,
	maxDepth int,
) (*ProductSearchResult, error) {
	myProducts := make(map[string]domain.Product)
	for _, product := range userProducts {
		myProducts[product.ProductID] = product
	}

	queue := []queueNode{{Product: target, Depth: 0}}
	visited := map[string]bool{target.ProductID: true}
	parent := make(map[string]string)
	productMap := map[string]domain.Product{target.ProductID: target}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, ok := myProducts[current.Product.ProductID]; ok {
			path := restorePath(current.Product.ProductID, parent, productMap)
			return &ProductSearchResult{Products: reverseProducts(path), Length: len(path) - 1}, nil
		}
		if current.Depth >= maxDepth {
			continue
		}
		neighbors, err := service.GetExchangeCandidates(ctx, current.Product.ProductID)
		if err != nil {
			return nil, err
		}
		for _, next := range neighbors {
			if visited[next.ProductID] {
				continue
			}
			visited[next.ProductID] = true
			parent[next.ProductID] = current.Product.ProductID
			productMap[next.ProductID] = next
			queue = append(queue, queueNode{Product: next, Depth: current.Depth + 1})
		}
	}
	return nil, nil
}

func restorePath(
	productID string,
	parent map[string]string,
	products map[string]domain.Product,
) []domain.Product {

	path := make([]domain.Product, 0)

	current := productID

	for {

		product, ok := products[current]
		if !ok {
			break
		}

		path = append(path, product)

		prev, ok := parent[current]
		if !ok {
			break
		}

		current = prev
	}

	return path
}

func reverseProducts(products []domain.Product) []domain.Product {
	for i, j := 0, len(products)-1; i < j; i, j = i+1, j-1 {
		products[i], products[j] = products[j], products[i]
	}
	return products
}
