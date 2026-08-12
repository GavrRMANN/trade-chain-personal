package repository

import (
	"context"
	"database/sql"
	"errors"

	"trade-chain/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type reviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) ReviewRepository {
	return &reviewRepository{db: db}
}

// reviewColumns держит список в одном месте: comment уже был в доменной
// модели, но ни в одном запросе — текст отзыва молча терялся.
const reviewColumns = `review_id, chain_id, from_customer_id, to_customer_id, product_id, rating, comment, created_at, updated_at`

func scanReview(row rowScanner) (domain.Review, error) {
	var review domain.Review
	err := row.Scan(
		&review.ReviewID,
		&review.ChainID,
		&review.FromCustomerID,
		&review.ToCustomerID,
		&review.ProductID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	return review, err
}

func (r *reviewRepository) Create(ctx context.Context, review *domain.Review) (*domain.Review, error) {
	query := `
		INSERT INTO reviews (chain_id, from_customer_id, to_customer_id, product_id, rating, comment)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + reviewColumns

	created, err := scanReview(r.db.QueryRow(ctx, query,
		review.ChainID,
		review.FromCustomerID,
		review.ToCustomerID,
		review.ProductID,
		review.Rating,
		review.Comment,
	))
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *reviewRepository) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	query := `SELECT ` + reviewColumns + ` FROM reviews WHERE review_id = $1`

	review, err := scanReview(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &review, nil
}

func (r *reviewRepository) GetByCustomerID(ctx context.Context, customerID string) ([]domain.Review, error) {
	query := `
		SELECT ` + reviewColumns + `
		FROM reviews
		WHERE to_customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]domain.Review, 0)
	for rows.Next() {
		review, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *reviewRepository) GetAverageRating(ctx context.Context, customerID string) (float64, error) {
	query := `
		SELECT COALESCE(AVG(rating), 0)::float
		FROM reviews
		WHERE to_customer_id = $1
	`

	var avg float64
	err := r.db.QueryRow(ctx, query, customerID).Scan(&avg)
	if err != nil {
		return 0, err
	}
	return avg, nil
}

func (r *reviewRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM reviews WHERE review_id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}
