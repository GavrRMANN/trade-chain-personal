package search

import "trade-chain/internal/domain"

// Оценка перехода. В дальнешйем добавлю сравнение по оценке продавца, дате появления объявления, по имени
func CalculateScore(from domain.Product, to domain.Product) float64 {
	if from.CategoryID == nil || to.CategoryID == nil {
		return 0
	}

	if *from.CategoryID == *to.CategoryID {
		return 1
	}

	return 0
}
