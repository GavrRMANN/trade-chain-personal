package repository

import (
	"context"
	"database/sql"
	"errors"

	"trade-chain/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type wishlistRepository struct {
	db *pgxpool.Pool
}

func NewWishlistRepository(db *pgxpool.Pool) WishlistRepository {
	return &wishlistRepository{db: db}
}

func (r *wishlistRepository) Create(ctx context.Context, wishlist *domain.Wishlist) (*domain.Wishlist, error) {
	query := `
		INSERT INTO wishlists (product_id, name)
		VALUES ($1, $2)
		RETURNING wishlist_id, product_id, name, created_at, updated_at
	`

	var created domain.Wishlist
	err := r.db.QueryRow(ctx, query, wishlist.ProductID, wishlist.Name).Scan(
		&created.WishlistID,
		&created.ProductID,
		&created.Name,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *wishlistRepository) GetByID(ctx context.Context, id string) (*domain.Wishlist, error) {
	query := `
		SELECT wishlist_id, product_id, name, created_at, updated_at
		FROM wishlists
		WHERE wishlist_id = $1
	`

	var wishlist domain.Wishlist
	err := r.db.QueryRow(ctx, query, id).Scan(
		&wishlist.WishlistID,
		&wishlist.ProductID,
		&wishlist.Name,
		&wishlist.CreatedAt,
		&wishlist.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &wishlist, nil
}

func (r *wishlistRepository) GetByProductID(
	ctx context.Context,
	productID string,
) (*domain.Wishlist, error) {
	query := `
        SELECT wishlist_id, product_id, name, created_at, updated_at
        FROM wishlists
        WHERE product_id = $1
    `

	var wishlist domain.Wishlist

	err := r.db.QueryRow(ctx, query, productID).Scan(
		&wishlist.WishlistID,
		&wishlist.ProductID,
		&wishlist.Name,
		&wishlist.CreatedAt,
		&wishlist.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &wishlist, nil
}

func (r *wishlistRepository) UpdateByProductID(
	ctx context.Context,
	productID string,
	name string,
	categoryIDs []string,
) (*domain.Wishlist, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE wishlists
		SET name = $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE product_id = $2
		RETURNING wishlist_id, product_id, name, created_at, updated_at
	`

	var wishlist domain.Wishlist

	err = tx.QueryRow(ctx, query, name, productID).Scan(
		&wishlist.WishlistID,
		&wishlist.ProductID,
		&wishlist.Name,
		&wishlist.CreatedAt,
		&wishlist.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	_, err = tx.Exec(
		ctx,
		`DELETE FROM wishlist_options WHERE wishlist_id = $1`,
		wishlist.WishlistID,
	)
	if err != nil {
		return nil, err
	}

	for _, categoryID := range categoryIDs {
		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO wishlist_options (wishlist_id, category_id)
			VALUES ($1, $2)
			`,
			wishlist.WishlistID,
			categoryID,
		)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &wishlist, nil
}

func (r *wishlistRepository) AddCategoryOption(ctx context.Context, wishlistID, categoryID string) error {
	query := `
		INSERT INTO wishlist_options (wishlist_id, category_id)
		VALUES ($1, $2)
	`
	_, err := r.db.Exec(ctx, query, wishlistID, categoryID)
	return err
}

func (r *wishlistRepository) RemoveCategoryOption(ctx context.Context, wishlistID, categoryID string) error {
	query := `
		DELETE FROM wishlist_options
		WHERE wishlist_id = $1 AND category_id = $2
	`
	result, err := r.db.Exec(ctx, query, wishlistID, categoryID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *wishlistRepository) GetOptions(ctx context.Context, wishlistID string) ([]domain.Category, error) {
	query := `
		SELECT c.category_id, c.name, c.parent_id, c.created_at, c.updated_at
		FROM wishlist_options wo
		JOIN categories c ON wo.category_id = c.category_id
		WHERE wo.wishlist_id = $1
		ORDER BY c.name
	`

	rows, err := r.db.Query(ctx, query, wishlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var cat domain.Category
		err := rows.Scan(
			&cat.CategoryID,
			&cat.Name,
			&cat.ParentID,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (r *wishlistRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM wishlists WHERE wishlist_id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}
