package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"trade-chain/internal/domain"
	"trade-chain/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type customerService struct{ repo repository.CustomerRepository }

func NewCustomerService(repo repository.CustomerRepository) CustomerService {
	return &customerService{repo: repo}
}

func (s *customerService) Create(ctx context.Context, dto *domain.CreateCustomerDTO) (*domain.Customer, error) {
	if dto == nil || blank(dto.Email) || len(dto.Password) < 8 {
		return nil, ErrInvalidInput
	}
	dto.Email = strings.ToLower(strings.TrimSpace(dto.Email))
	dto.FullName = strings.TrimSpace(dto.FullName)
	if _, err := s.repo.GetByEmail(ctx, dto.Email); err == nil {
		return nil, ErrConflict
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, normalizeError(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	copyDTO := *dto
	copyDTO.Password = string(hash)
	created, err := s.repo.Create(ctx, &copyDTO)
	return created, normalizeError(err)
}
func (s *customerService) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	v, e := s.repo.GetByID(ctx, id)
	return v, normalizeError(e)
}
func (s *customerService) Update(ctx context.Context, id string, dto *domain.UpdateCustomerDTO) (*domain.Customer, error) {
	if blank(id) || dto == nil {
		return nil, ErrInvalidInput
	}
	copyDTO := *dto
	if copyDTO.Email != nil {
		v := strings.ToLower(strings.TrimSpace(*copyDTO.Email))
		if blank(v) {
			return nil, ErrInvalidInput
		}
		copyDTO.Email = &v
	}
	if copyDTO.FullName != nil {
		// Пустое ФИО — это осознанное «убрать имя», а не пропуск поля:
		// пропуск приходит как отсутствующий ключ и остаётся nil.
		v := strings.TrimSpace(*copyDTO.FullName)
		copyDTO.FullName = &v
	}
	if copyDTO.Password != nil {
		if len(*copyDTO.Password) < 8 {
			return nil, ErrInvalidInput
		}
		h, e := bcrypt.GenerateFromPassword([]byte(*copyDTO.Password), bcrypt.DefaultCost)
		if e != nil {
			return nil, e
		}
		v := string(h)
		copyDTO.Password = &v
	}
	v, e := s.repo.Update(ctx, id, &copyDTO)
	return v, normalizeError(e)
}
func (s *customerService) Delete(ctx context.Context, id string) error {
	if blank(id) {
		return ErrInvalidInput
	}
	return normalizeError(s.repo.Delete(ctx, id))
}
func (s *customerService) List(ctx context.Context, offset, limit int) ([]domain.Customer, error) {
	o, l, e := validatePage(offset, limit)
	if e != nil {
		return nil, e
	}
	v, e := s.repo.List(ctx, o, l)
	return v, normalizeError(e)
}

// ListOverview отдаёт участников вместе с их показателями на площадке.
func (s *customerService) ListOverview(ctx context.Context, offset, limit int) ([]domain.CustomerOverview, error) {
	o, l, e := validatePage(offset, limit)
	if e != nil {
		return nil, e
	}
	v, e := s.repo.ListOverview(ctx, o, l)
	return v, normalizeError(e)
}

func (s *customerService) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	if blank(email) {
		return nil, ErrInvalidInput
	}
	v, err := s.repo.GetByEmail(ctx, email)
	return v, normalizeError(err)
}

// Функции для работы с личным вишлистом

func (s *customerService) GetCustomerWishlistOptions(
	ctx context.Context,
	customerID string,
) ([]domain.CustomerWishlistOption, error) {
	if blank(customerID) {
		return nil, ErrInvalidInput
	}

	v, err := s.repo.GetCustomerWishlistOptions(ctx, customerID)
	return v, normalizeError(err)
}

func (s *customerService) AddCustomerWishlistOption(
	ctx context.Context,
	customerID string,
	categoryID string,
) error {
	if blank(customerID) || blank(categoryID) {
		return ErrInvalidInput
	}

	err := s.repo.AddCustomerWishlistOption(
		ctx,
		customerID,
		categoryID,
	)

	return normalizeError(err)
}

func (s *customerService) DeleteCustomerWishlistOption(
	ctx context.Context,
	customerID string,
	categoryID string,
) error {
	if blank(customerID) || blank(categoryID) {
		return ErrInvalidInput
	}

	err := s.repo.DeleteCustomerWishlistOption(
		ctx,
		customerID,
		categoryID,
	)

	return normalizeError(err)
}

func (s *customerService) ReplaceCustomerWishlistOptions(
	ctx context.Context,
	customerID string,
	dto *domain.UpdateCustomerWishlistDTO,
) error {
	if blank(customerID) || dto == nil {
		return ErrInvalidInput
	}

	for _, categoryID := range dto.CategoryIDs {
		if blank(categoryID) {
			return ErrInvalidInput
		}
	}

	err := s.repo.ReplaceCustomerWishlistOptions(
		ctx,
		customerID,
		dto,
	)

	return normalizeError(err)
}
