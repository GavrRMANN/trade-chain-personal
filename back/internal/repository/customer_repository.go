package repository

import (
	"context"
	"database/sql"
	"errors"

	"trade-chain/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type customerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) CustomerRepository {
	return &customerRepository{db: db}
}

// customerColumns держит список в одном месте: колонка full_name появилась
// после всех запросов, и добавлять её пришлось бы в пять мест сразу.
const customerColumns = `customer_id, email, full_name, password_hash, created_at, updated_at`

func scanCustomer(row rowScanner) (domain.Customer, error) {
	var customer domain.Customer
	err := row.Scan(
		&customer.CustomerID,
		&customer.Email,
		&customer.FullName,
		&customer.PasswordHash,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	return customer, err
}

func (r *customerRepository) Create(ctx context.Context, customer *domain.CreateCustomerDTO) (*domain.Customer, error) {
	query := `
		INSERT INTO customers (email, password_hash, full_name)
		VALUES ($1, $2, $3)
		RETURNING ` + customerColumns

	created, err := scanCustomer(r.db.QueryRow(ctx, query, customer.Email, customer.Password, customer.FullName))
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *customerRepository) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	query := `
		SELECT ` + customerColumns + `
		FROM customers
		WHERE customer_id = $1 AND is_active = TRUE
	`

	customer, err := scanCustomer(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &customer, nil
}

func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	query := `
		SELECT ` + customerColumns + `
		FROM customers
		WHERE email = $1 AND is_active = TRUE
	`

	customer, err := scanCustomer(r.db.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &customer, nil
}

func (r *customerRepository) Update(ctx context.Context, id string, customer *domain.UpdateCustomerDTO) (*domain.Customer, error) {
	query := `
		UPDATE customers
		SET email = COALESCE($1, email),
		    password_hash = COALESCE($2, password_hash),
		    full_name = COALESCE($3, full_name)
		WHERE customer_id = $4
		RETURNING ` + customerColumns

	updated, err := scanCustomer(r.db.QueryRow(ctx, query, customer.Email, customer.Password, customer.FullName, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &updated, nil
}

func (r *customerRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE customers SET is_active = FALSE WHERE customer_id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *customerRepository) List(ctx context.Context, offset, limit int) ([]domain.Customer, error) {
	query := `
		SELECT ` + customerColumns + `
		FROM customers
		WHERE is_active = TRUE
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []domain.Customer
	for rows.Next() {
		customer, err := scanCustomer(rows)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}

	return customers, rows.Err()
}

// ListOverview собирает профиль вместе с производными показателями.
//
// Показатели считаются скалярными подзапросами, а не набором JOIN с
// GROUP BY: агрегаты идут по трём независимым таблицам, и объединение их в
// одном соединении перемножило бы строки — товары размножились бы на число
// отзывов. Отдельные подзапросы дают каждому агрегату свой индекс.
func (r *customerRepository) ListOverview(ctx context.Context, offset, limit int) ([]domain.CustomerOverview, error) {
	query := `
		SELECT
			c.customer_id,
			c.email,
			c.full_name,
			c.created_at,
			COALESCE((SELECT AVG(rating)::float8 FROM reviews WHERE to_customer_id = c.customer_id), 0),
			(SELECT COUNT(*) FROM reviews   WHERE to_customer_id = c.customer_id),
			(SELECT COUNT(*) FROM products  WHERE customer_id = c.customer_id),
			(SELECT COUNT(*) FROM products  WHERE customer_id = c.customer_id AND status = 'active'),
			(SELECT COUNT(*) FROM chains    WHERE initiator_id = c.customer_id OR recipient_id = c.customer_id)
		FROM customers c
		WHERE c.is_active = TRUE
		ORDER BY c.created_at DESC, c.customer_id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Пустой срез, а не nil: список участников уходит в JSON, и `[]`
	// клиенту понятнее, чем `null`.
	overviews := make([]domain.CustomerOverview, 0)
	for rows.Next() {
		var overview domain.CustomerOverview
		err := rows.Scan(
			&overview.CustomerID,
			&overview.Email,
			&overview.FullName,
			&overview.CreatedAt,
			&overview.Rating,
			&overview.ReviewCount,
			&overview.ProductCount,
			&overview.ActiveProductCount,
			&overview.ChainCount,
		)
		if err != nil {
			return nil, err
		}
		overviews = append(overviews, overview)
	}

	return overviews, rows.Err()
}

// Методы для работы с вишлистами пользователя

func (r *customerRepository) GetCustomerWishlistOptions(
	ctx context.Context,
	customerID string,
) ([]domain.CustomerWishlistOption, error) {
	query := `
		SELECT customer_id, category_id
		FROM customer_wishlist_options
		WHERE customer_id = $1
		ORDER BY category_id
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]domain.CustomerWishlistOption, 0)

	for rows.Next() {
		var option domain.CustomerWishlistOption

		if err := rows.Scan(
			&option.CustomerID,
			&option.CategoryID,
		); err != nil {
			return nil, err
		}

		options = append(options, option)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return options, nil
}

func (r *customerRepository) AddCustomerWishlistOption(
	ctx context.Context,
	customerID string,
	categoryID string,
) error {
	query := `
		INSERT INTO customer_wishlist_options (
			customer_id,
			category_id
		)
		VALUES ($1, $2)
		ON CONFLICT (customer_id, category_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, customerID, categoryID)
	return err
}

func (r *customerRepository) DeleteCustomerWishlistOption(
	ctx context.Context,
	customerID string,
	categoryID string,
) error {
	query := `
		DELETE FROM customer_wishlist_options
		WHERE customer_id = $1
		  AND category_id = $2
	`

	result, err := r.db.Exec(ctx, query, customerID, categoryID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *customerRepository) ReplaceCustomerWishlistOptions(
	ctx context.Context,
	customerID string,
	dto *domain.UpdateCustomerWishlistDTO,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	deleteQuery := `
		DELETE FROM customer_wishlist_options
		WHERE customer_id = $1
	`

	if _, err := tx.Exec(ctx, deleteQuery, customerID); err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO customer_wishlist_options (
			customer_id,
			category_id
		)
		VALUES ($1, $2)
		ON CONFLICT (customer_id, category_id) DO NOTHING
	`

	for _, categoryID := range dto.CategoryIDs {
		if _, err := tx.Exec(ctx, insertQuery, customerID, categoryID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
