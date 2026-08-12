package search

import "trade-chain/internal/domain"

type Edge struct {
	From domain.Product

	To domain.Product

	Score float64
}
