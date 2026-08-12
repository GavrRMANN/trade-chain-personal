package service

import (
	"context"

	"trade-chain/internal/domain"
	"trade-chain/internal/repository"
)

type reviewService struct {
	repo      repository.ReviewRepository
	customers repository.CustomerRepository
	products  repository.ProductRepository
	chains    ChainService
}

func NewReviewService(
	r repository.ReviewRepository,
	c repository.CustomerRepository,
	p repository.ProductRepository,
	chains ChainService,
) ReviewService {
	return &reviewService{repo: r, customers: c, products: p, chains: chains}
}

// Create принимает отзыв по итогам конкретной сделки.
//
// Кого оценивают, сервис определяет сам — по звену обмена, а не по телу
// запроса: раньше получателя оценки называл автор, и поставить единицу можно
// было незнакомому человеку, с которым никогда не менялся.
func (s *reviewService) Create(ctx context.Context, v *domain.Review) (*domain.Review, error) {
	if v == nil || blank(v.FromCustomerID) || v.Rating < 1 || v.Rating > 5 {
		return nil, ErrInvalidInput
	}
	if v.ChainID == nil || blank(*v.ChainID) {
		return nil, ErrInvalidInput
	}

	counterparty, err := s.chains.CanReview(ctx, *v.ChainID, v.FromCustomerID)
	if err != nil {
		return nil, err
	}
	v.ToCustomerID = counterparty

	if v.ProductID != nil {
		if e := s.checkProduct(ctx, *v.ProductID); e != nil {
			return nil, e
		}
	}

	out, e := s.repo.Create(ctx, v)
	return out, normalizeError(e)
}

func (s *reviewService) checkProduct(ctx context.Context, id string) error {
	_, err := s.products.GetByID(ctx, id)
	return normalizeError(err)
}
func (s *reviewService) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	v, e := s.repo.GetByID(ctx, id)
	return v, normalizeError(e)
}
func (s *reviewService) GetByCustomerID(ctx context.Context, id string) ([]domain.Review, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	v, e := s.repo.GetByCustomerID(ctx, id)
	return v, normalizeError(e)
}
func (s *reviewService) GetAverageRating(ctx context.Context, id string) (float64, error) {
	if blank(id) {
		return 0, ErrInvalidInput
	}
	v, e := s.repo.GetAverageRating(ctx, id)
	return v, normalizeError(e)
}
func (s *reviewService) Delete(ctx context.Context, id string) error {
	if blank(id) {
		return ErrInvalidInput
	}
	return normalizeError(s.repo.Delete(ctx, id))
}
