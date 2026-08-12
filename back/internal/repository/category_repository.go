package repository

import (
	"context"
	"database/sql"
	"errors"

	"trade-chain/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type categoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	query := `
		INSERT INTO categories (name, icon, parent_id)
		VALUES ($1, $2, $3)
		RETURNING category_id, name, icon, parent_id, created_at, updated_at
	`

	var created domain.Category
	err := r.db.QueryRow(ctx, query, category.Name, category.Icon, category.ParentID).Scan(
		&created.CategoryID,
		&created.Name,
		&created.Icon,
		&created.ParentID,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	query := `
		SELECT category_id, name, icon, parent_id, created_at, updated_at
		FROM categories
		WHERE category_id = $1
	`

	var category domain.Category
	err := r.db.QueryRow(ctx, query, id).Scan(
		&category.CategoryID,
		&category.Name,
		&category.Icon,
		&category.ParentID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetSubcategories(ctx context.Context, parentID string) ([]domain.Category, error) {
	query := `
		SELECT category_id, name, icon, parent_id, created_at, updated_at
		FROM categories
		WHERE parent_id = $1
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, parentID)
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
			&cat.Icon,
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

func (r *categoryRepository) Update(ctx context.Context, id string, category *domain.Category) (*domain.Category, error) {
	query := `
		UPDATE categories
		SET name = COALESCE($1, name),
		    icon = COALESCE($2, icon),
		    parent_id = COALESCE($3, parent_id)
		WHERE category_id = $4
		RETURNING category_id, name, icon, parent_id, created_at, updated_at
	`

	var updated domain.Category
	err := r.db.QueryRow(ctx, query, category.Name, category.Icon, category.ParentID, id).Scan(
		&updated.CategoryID,
		&updated.Name,
		&updated.Icon,
		&updated.ParentID,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &updated, nil
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM categories WHERE category_id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *categoryRepository) Search(
	ctx context.Context,
	search string,
) ([]domain.Category, error) {

	query := `
SELECT
	category_id,
	name,
	icon,
	parent_id,
	updated_at,

	(
		0.75 * ts_rank_cd(
			search_vector,
			websearch_to_tsquery('simple', $1)
		)
		+
		0.25 * similarity(name, $1)
	) AS score

FROM categories

WHERE
	search_vector @@ websearch_to_tsquery('simple', $1)
	OR
	name % $1

ORDER BY
	score DESC,
	name ASC;
`

	rows, err := r.db.Query(ctx, query, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)

	for rows.Next() {
		var category domain.Category
		var score float64

		err := rows.Scan(
			&category.CategoryID,
			&category.Name,
			&category.Icon,
			&category.ParentID,
			&category.UpdatedAt,
			&score,
		)

		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, rows.Err()
}

func (r *categoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	query := `
		SELECT category_id, name, icon, parent_id, created_at, updated_at
		FROM categories
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
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
			&cat.Icon,
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
