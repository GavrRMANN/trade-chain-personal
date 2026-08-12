package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"trade-chain/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(
	ctx context.Context,
	dto *domain.CreateProductDTO,
) (*domain.Product, error) {
	query := `
		INSERT INTO products (
			customer_id,
			category_id,
			title,
			description,
			image,
			price,
			location
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			product_id,
			customer_id,
			COALESCE(category_id::text, ''),
			title,
			COALESCE(description, ''),
			COALESCE(image, ''),
			price,
			COALESCE(location, ''),
			status,
			created_at,
			updated_at
	`

	var created domain.Product

	err := r.db.QueryRow(
		ctx,
		query,
		dto.CustomerID,
		dto.CategoryID,
		dto.Title,
		dto.Description,
		dto.Image,
		dto.Price,
		dto.Location,
	).Scan(
		&created.ProductID,
		&created.CustomerID,
		&created.CategoryID,
		&created.Title,
		&created.Description,
		&created.Image,
		&created.Price,
		&created.Location,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *productRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `
		SELECT
			product_id,
			customer_id,
			COALESCE(category_id::text, ''),
			title,
			COALESCE(description, ''),
			COALESCE(image, ''),
			price,
			COALESCE(location, ''),
			status,
			created_at,
			updated_at
		FROM products
		WHERE product_id = $1
	`

	var product domain.Product
	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ProductID,
		&product.CustomerID,
		&product.CategoryID,
		&product.Title,
		&product.Description,
		&product.Image,
		&product.Price,
		&product.Location,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetByCustomerID(ctx context.Context, customerID string) ([]domain.Product, error) {
	query := `
		SELECT
			product_id,
			customer_id,
			COALESCE(category_id::text, ''),
			title,
			COALESCE(description, ''),
			COALESCE(image, ''),
			price,
			COALESCE(location, ''),
			status,
			created_at,
			updated_at
		FROM products
		WHERE customer_id = $1 AND status != 'archived'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ProductID,
			&p.CustomerID,
			&p.CategoryID,
			&p.Title,
			&p.Description,
			&p.Image,
			&p.Price,
			&p.Location,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

// GetOwnByCustomerID возвращает все товары владельца, включая архивные.
func (r *productRepository) GetOwnByCustomerID(ctx context.Context, customerID string) ([]domain.Product, error) {
	query := `
		SELECT
			product_id,
			customer_id,
			COALESCE(category_id::text, ''),
			title,
			COALESCE(description, ''),
			COALESCE(image, ''),
			price,
			COALESCE(location, ''),
			status,
			created_at,
			updated_at
		FROM products
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ProductID,
			&p.CustomerID,
			&p.CategoryID,
			&p.Title,
			&p.Description,
			&p.Image,
			&p.Price,
			&p.Location,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *productRepository) Update(ctx context.Context, id string, dto *domain.UpdateProductDTO) (*domain.Product, error) {
	query := `
		UPDATE products
		SET
			title = COALESCE($1, title),
			description = COALESCE($2, description),
			category_id = COALESCE($3, category_id),
			image = COALESCE($4, image),
			price = COALESCE($5, price),
			location = COALESCE($6, location),
			status = COALESCE($7, status)
		WHERE product_id = $8 AND status != 'archived'
		RETURNING
			product_id,
			customer_id,
			COALESCE(category_id::text, ''),
			title,
			COALESCE(description, ''),
			COALESCE(image, ''),
			price,
			COALESCE(location, ''),
			status,
			created_at,
			updated_at	`

	var updated domain.Product
	err := r.db.QueryRow(ctx, query,
		dto.Title,
		dto.Description,
		dto.CategoryID,
		dto.Image,
		dto.Price,
		dto.Location,
		dto.Status,
		id,
	).Scan(
		&updated.ProductID,
		&updated.CustomerID,
		&updated.CategoryID,
		&updated.Title,
		&updated.Description,
		&updated.Image,
		&updated.Price,
		&updated.Location,
		&updated.Status,
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

func (r *productRepository) Delete(
	ctx context.Context,
	productID string,
	customerID string,
) error {
	query := `
		UPDATE products
		SET status = 'archived'
		WHERE product_id = $1
		  AND customer_id = $2
	`

	result, err := r.db.Exec(ctx, query, productID, customerID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *productRepository) List(
	ctx context.Context,
	customerID *string,
	q string,
	categoryID *string,
	offset int,
	limit int,
) ([]domain.Product, error) {
	args := []interface{}{}
	argIndex := 1

	// Товар зрителя, который закрывает желание владельца карточки. Считается
	// в самом запросе, а не отдельным проходом по ленте: иначе на каждую
	// страницу выдачи пришлось бы делать ещё один поход в базу. LATERAL, а не
	// подзапрос в SELECT: результат нужен ещё и в ORDER BY, а на выражение из
	// списка выборки там ссылаться нельзя.
	matchColumn := `NULL::text`
	matchJoin := ``
	matchOrder := ``

	var viewerArgIndex int

	if customerID != nil {
		viewerArgIndex = argIndex
		args = append(args, *customerID)
		argIndex++

		matchColumn = `matched.product_id::text`
		matchJoin = fmt.Sprintf(`
			LEFT JOIN LATERAL (
				SELECT mine.product_id
				FROM wishlists w
				JOIN wishlist_options wo
					ON wo.wishlist_id = w.wishlist_id
				JOIN products mine
					ON mine.category_id = wo.category_id
				WHERE w.product_id = p.product_id
				  AND mine.customer_id = $%d
				  AND mine.status = 'active'
				ORDER BY mine.created_at DESC
				LIMIT 1
			) matched ON TRUE
		`, viewerArgIndex)
		matchOrder = `(matched.product_id IS NOT NULL) DESC,`
	}

	query := fmt.Sprintf(`
		SELECT
			p.product_id,
			p.customer_id,
			COALESCE(p.category_id::text, ''),
			p.title,
			COALESCE(p.description, ''),
			COALESCE(p.image, ''),
			p.price,
			COALESCE(p.location, ''),
			p.status,
			p.created_at,
			p.updated_at,
			%s AS matched_by_product_id
		FROM products p
		%s
		WHERE p.status NOT IN ('archived', 'exchanged')
	`, matchColumn, matchJoin)

	if customerID != nil {
		query += fmt.Sprintf(`
			AND p.customer_id != $%d
		`, viewerArgIndex)
	}

	var searchArgIndex int

	if q != "" {
		searchArgIndex = argIndex

		query += fmt.Sprintf(`
			AND (
				COALESCE(p.search_vector, ''::tsvector)
					@@ websearch_to_tsquery('simple', $%d)
				OR p.title %% $%d
				OR COALESCE(p.description, '') %% $%d
			)
		`, searchArgIndex, searchArgIndex, searchArgIndex)

		args = append(args, q)
		argIndex++
	}

	if categoryID != nil {
		query += fmt.Sprintf(`
			AND p.category_id = $%d
		`, argIndex)

		args = append(args, *categoryID)
		argIndex++
	}

	// Подходящие карточки идут первыми во всей выдаче, а не только на текущей
	// странице: иначе при бесконечной прокрутке блок «Вам подойдёт» пополнялся
	// бы новыми карточками уже после того, как пользователь его пролистал.
	if q != "" {
		query += fmt.Sprintf(`
			ORDER BY
				%s
				(
					0.60 * ts_rank_cd(
						COALESCE(p.search_vector, ''::tsvector),
						websearch_to_tsquery('simple', $%d)
					) +
					0.25 * similarity(p.title, $%d) +
					0.15 * similarity(
						COALESCE(p.description, ''),
						$%d
					)
				) DESC,
				p.created_at DESC
		`, matchOrder, searchArgIndex, searchArgIndex, searchArgIndex)
	} else {
		query += fmt.Sprintf(`
			ORDER BY
				%s
				p.created_at DESC
		`, matchOrder)
	}

	limitArgIndex := argIndex
	offsetArgIndex := argIndex + 1

	query += fmt.Sprintf(`
		LIMIT $%d OFFSET $%d
	`, limitArgIndex, offsetArgIndex)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.Printf("PRODUCT LIST QUERY ERROR: %v", err)
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product

	for rows.Next() {
		var p domain.Product
		var matchedBy *string

		if err := rows.Scan(
			&p.ProductID,
			&p.CustomerID,
			&p.CategoryID,
			&p.Title,
			&p.Description,
			&p.Image,
			&p.Price,
			&p.Location,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
			&matchedBy,
		); err != nil {
			log.Printf("PRODUCT LIST SCAN ERROR: %v", err)
			return nil, err
		}

		p.MatchedByProductID = matchedBy
		p.Matched = matchedBy != nil

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		log.Printf("PRODUCT LIST ROWS ERROR: %v", err)
		return nil, err
	}

	log.Printf(
		"PRODUCT LIST SUCCESS: customerID=%v q=%q category=%v offset=%d limit=%d got=%d",
		customerID,
		q,
		categoryID,
		offset,
		limit,
		len(products),
	)

	return products, nil
}

func (r *productRepository) GetExchangeCandidates(ctx context.Context, productID string) ([]domain.Product, error) {
	query := `
		SELECT DISTINCT
			p.product_id,
			p.customer_id,
			COALESCE(p.category_id::text, ''),
			p.title,
			COALESCE(p.description, ''),
			COALESCE(p.image, ''),
			p.price,
			COALESCE(p.location, ''),
			p.status,
			p.created_at,
			p.updated_at
		FROM products source
		JOIN wishlists w
			ON w.product_id = source.product_id
		JOIN wishlist_options wo
			ON wo.wishlist_id = w.wishlist_id
		JOIN products p
			ON p.category_id = wo.category_id
		WHERE
			source.product_id = $1
			AND p.status != 'archived'
			AND p.product_id <> source.product_id
			AND p.customer_id <> source.customer_id
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ProductID,
			&p.CustomerID,
			&p.CategoryID,
			&p.Title,
			&p.Description,
			&p.Image,
			&p.Price,
			&p.Location,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}
