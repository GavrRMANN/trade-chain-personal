package service

import (
	"context"
	"strings"

	"trade-chain/internal/domain"
	"trade-chain/internal/repository"
)

type categoryService struct{ repo repository.CategoryRepository }

func NewCategoryService(r repository.CategoryRepository) CategoryService { return &categoryService{r} }
func (s *categoryService) Create(ctx context.Context, v *domain.Category) (*domain.Category, error) {
	if v == nil || blank(v.Name) {
		return nil, ErrInvalidInput
	}
	v.Name = strings.TrimSpace(v.Name)
	out, e := s.repo.Create(ctx, v)
	return out, normalizeError(e)
}
func (s *categoryService) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	v, e := s.repo.GetByID(ctx, id)
	return v, normalizeError(e)
}
func (s *categoryService) GetSubcategories(ctx context.Context, id string) ([]domain.Category, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	v, e := s.repo.GetSubcategories(ctx, id)
	return v, normalizeError(e)
}
func (s *categoryService) Update(ctx context.Context, id string, v *domain.Category) (*domain.Category, error) {
	if blank(id) || v == nil || blank(v.Name) {
		return nil, ErrInvalidInput
	}
	out, e := s.repo.Update(ctx, id, v)
	return out, normalizeError(e)
}
func (s *categoryService) Delete(ctx context.Context, id string) error {
	if blank(id) {
		return ErrInvalidInput
	}
	return normalizeError(s.repo.Delete(ctx, id))
}

func (s *categoryService) Search(
	ctx context.Context,
	search string,
) ([]domain.Category, error) {
	if blank(search) {
		return nil, ErrInvalidInput
	}
	v, e := s.repo.Search(ctx, search)
	if e != nil {
		return nil, e
	}
	return v, nil
}

func (s *categoryService) List(ctx context.Context) ([]domain.Category, error) {
	v, e := s.repo.List(ctx)
	return v, normalizeError(e)
}
