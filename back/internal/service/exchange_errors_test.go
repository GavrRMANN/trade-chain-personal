package service

import (
	"errors"
	"testing"

	"trade-chain/internal/domain"
)

func TestProductUnavailableMapsToConflict(t *testing.T) {
	err := mapExchangeError(domain.ErrProductUnavailable)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("ошибка %v, ожидался ErrConflict", err)
	}
	if errors.Is(err, ErrInvalidInput) {
		t.Fatalf("ошибка недоступного товара не должна быть ErrInvalidInput: %v", err)
	}
}
