package service

import (
	"context"
	"strings"

	"trade-chain/internal/domain"
	"trade-chain/internal/repository"
)

type wishlistService struct {
	repo     repository.WishlistRepository
	products repository.ProductRepository
}

func NewWishlistService(r repository.WishlistRepository, p repository.ProductRepository) WishlistService {
	return &wishlistService{r, p}
}
func (s *wishlistService) Create(ctx context.Context, v *domain.Wishlist) (*domain.Wishlist, error) {
	if v == nil || blank(v.ProductID) || blank(v.Name) {
		return nil, ErrInvalidInput
	}
	if _, e := s.products.GetByID(ctx, v.ProductID); e != nil {
		return nil, normalizeError(e)
	}
	v.Name = strings.TrimSpace(v.Name)
	out, e := s.repo.Create(ctx, v)
	return out, normalizeError(e)
}
func (s *wishlistService) GetByID(ctx context.Context, id string) (*domain.Wishlist, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	v, e := s.repo.GetByID(ctx, id)
	return v, normalizeError(e)
}
func (s *wishlistService) GetByProductID(ctx context.Context, id string) (*domain.Wishlist, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	v, e := s.repo.GetByProductID(ctx, id)
	return v, normalizeError(e)
}
func (s *wishlistService) AddCategoryOption(ctx context.Context, w, c string) error {
	if blank(w) || blank(c) {
		return ErrInvalidInput
	}
	return normalizeError(s.repo.AddCategoryOption(ctx, w, c))
}
func (s *wishlistService) RemoveCategoryOption(ctx context.Context, w, c string) error {
	if blank(w) || blank(c) {
		return ErrInvalidInput
	}
	return normalizeError(s.repo.RemoveCategoryOption(ctx, w, c))
}
func (s *wishlistService) GetOptions(ctx context.Context, id string) ([]domain.Category, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	v, e := s.repo.GetOptions(ctx, id)
	return v, normalizeError(e)
}
func (s *wishlistService) Delete(ctx context.Context, id string) error {
	if blank(id) {
		return ErrInvalidInput
	}
	return normalizeError(s.repo.Delete(ctx, id))
}

func (s *wishlistService) UpdateByProductID(
	ctx context.Context,
	productID string,
	dto *domain.CreateWishlistDTO,
) (*domain.Wishlist, error) {
	if blank(productID) || dto == nil || blank(dto.Name) {
		return nil, ErrInvalidInput
	}

	dto.Name = strings.TrimSpace(dto.Name)

	if _, err := s.products.GetByID(ctx, productID); err != nil {
		return nil, normalizeError(err)
	}

	out, err := s.repo.UpdateByProductID(
		ctx,
		productID,
		dto.Name,
		dto.CategoryIDs,
	)

	return out, normalizeError(err)
}
