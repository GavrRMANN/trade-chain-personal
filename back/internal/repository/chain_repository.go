package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"trade-chain/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// uniqueViolationCode — код ошибки Postgres при нарушении уникального индекса.
const uniqueViolationCode = "23505"

type chainRepository struct {
	db *pgxpool.Pool
}

func NewChainRepository(db *pgxpool.Pool) ChainRepository {
	return &chainRepository{db: db}
}

// chainColumns перечислены один раз намеренно: список повторялся в четырёх
// запросах, и добавление колонки требовало не забыть ни один из них.
const chainColumns = `chain_id, from_product_id, to_product_id, to_category_id, initiator_id, recipient_id,
	previous_chain_id, next_chain_id, status, message, exchange_goal_id, route_step_id,
	surcharge_amount, surcharge_currency, surcharge_payer, expires_at, created_at, updated_at`

// chainColumnsOf повторяет тот же список с именем таблицы.
//
// Обе ветки UNION в рекурсивном запросе обязаны совпадать по составу колонок,
// а второй список, набранный руками, разъезжается с первым при первом же
// добавленном поле — и запрос падает уже в базе.
func chainColumnsOf(alias string) string {
	columns := strings.Split(chainColumns, ",")
	for i, column := range columns {
		columns[i] = alias + "." + strings.TrimSpace(column)
	}
	return strings.Join(columns, ", ")
}

// rowScanner покрывает и одиночную строку, и строку выборки: у pgx.Row
// и pgx.Rows одинаковый Scan, так что разбор звена пишется один раз.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanChain(row rowScanner) (domain.Chain, error) {
	var chain domain.Chain
	err := row.Scan(
		&chain.ChainID,
		&chain.FromProductID,
		&chain.ToProductID,
		&chain.ToCategoryID,
		&chain.InitiatorID,
		&chain.RecipientID,
		&chain.PreviousChainID,
		&chain.NextChainID,
		&chain.Status,
		&chain.Message,
		&chain.ExchangeGoalID,
		&chain.RouteStepID,
		&chain.Surcharge.Amount,
		&chain.Surcharge.Currency,
		&chain.Surcharge.Payer,
		&chain.ExpiresAt,
		&chain.CreatedAt,
		&chain.UpdatedAt,
	)
	return chain, err
}

func (r *chainRepository) queryChains(ctx context.Context, query string, args ...any) ([]domain.Chain, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chains := make([]domain.Chain, 0)
	for rows.Next() {
		chain, err := scanChain(rows)
		if err != nil {
			return nil, err
		}
		chains = append(chains, chain)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chains, nil
}

func (r *chainRepository) Create(ctx context.Context, chain *domain.Chain) (*domain.Chain, error) {
	query := `
			INSERT INTO chains (from_product_id, to_product_id, to_category_id, initiator_id, recipient_id, previous_chain_id, next_chain_id, status, message,
				exchange_goal_id, route_step_id, surcharge_amount, surcharge_currency, surcharge_payer, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, COALESCE($15, CURRENT_TIMESTAMP + INTERVAL '72 hours'))
			RETURNING ` + chainColumns

	// Нулевое время означает «срок не задан» — базе передаётся NULL,
	// и она подставляет свой срок по умолчанию.
	var expiresAt *time.Time
	if !chain.ExpiresAt.IsZero() {
		expiresAt = &chain.ExpiresAt
	}

	created, err := scanChain(r.db.QueryRow(ctx, query,
		chain.FromProductID,
		chain.ToProductID,
		chain.ToCategoryID,
		chain.InitiatorID,
		chain.RecipientID,
		chain.PreviousChainID,
		chain.NextChainID,
		chain.Status,
		chain.Message,
		chain.ExchangeGoalID,
		chain.RouteStepID,
		chain.Surcharge.Amount,
		chain.Surcharge.Currency,
		chain.Surcharge.Payer,
		expiresAt,
	))
	if err != nil {
		// Второе предложение по той же паре товаров отсекается уникальным
		// индексом. Ошибка базы здесь ожидаемая, а не поломка: человек просто
		// нажал кнопку дважды, и ему нужен понятный ответ, а не 500.
		if isUniqueViolation(err) {
			return nil, domain.ErrOfferDuplicate
		}
		return nil, err
	}
	return &created, nil
}

// isUniqueViolation отличает нарушение уникальности от прочих ошибок базы.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

func (r *chainRepository) GetByID(
	ctx context.Context,
	id string,
	customerID string,
) (*domain.Chain, error) {
	query := `SELECT ` + chainColumns + ` FROM chains WHERE chain_id = $1`

	chain, err := scanChain(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &chain, nil
}

func (r *chainRepository) GetByProductID(
	ctx context.Context,
	productID string,
	customerID string,
) ([]domain.Chain, error) {
	query := `
		SELECT ` + chainColumns + `
		FROM chains
		WHERE from_product_id = $1 OR to_product_id = $1
		ORDER BY created_at DESC
	`

	return r.queryChains(ctx, query, productID)
}

// GetByCustomerID отдаёт все сделки человека — и те, что он предложил сам,
// и те, что предложили ему. Без этого запроса на фронте нет входящих:
// пользователь узнаёт о предложении, только если знает его идентификатор.
func (r *chainRepository) GetByCustomerID(ctx context.Context, customerID string) ([]domain.Chain, error) {
	return r.List(ctx, ChainFilter{CustomerID: customerID})
}

// List отбирает сделки человека по стороне и состоянию.
//
// Отбор делает база, а не Go: списки входящих и исходящих открывают на каждом
// заходе в приложение, и вычитывать ради них всю историю пользователя, чтобы
// отфильтровать в памяти, — лишний трафик, который растёт вместе с историей.
func (r *chainRepository) List(ctx context.Context, filter ChainFilter) ([]domain.Chain, error) {
	// Пустая роль означает «любая сторона»: у списка «мои обмены» разделения
	// на входящие и исходящие нет.
	asInitiator := filter.Role == domain.RoleAny || filter.Role == domain.RoleOutgoing
	asRecipient := filter.Role == domain.RoleAny || filter.Role == domain.RoleIncoming

	statuses := make([]string, 0, len(filter.Statuses))
	for _, status := range filter.Statuses {
		statuses = append(statuses, string(status))
	}

	query := `
		SELECT ` + chainColumns + `
		FROM chains
		WHERE (($2 AND initiator_id = $1) OR ($3 AND recipient_id = $1))
		  AND (cardinality($4::text[]) = 0 OR status = ANY($4::text[]))
		ORDER BY created_at DESC
	`
	chains, err := r.queryChains(
		ctx,
		query,
		filter.CustomerID,
		asInitiator,
		asRecipient,
		statuses,
	)
	if err != nil {
		return nil, err
	}

	return chains, nil
}

func (r *chainRepository) GetFullChain(ctx context.Context, chainID string) ([]domain.Chain, error) {
	query := `
		WITH RECURSIVE chain_path AS (
			SELECT ` + chainColumns + `
			FROM chains
			WHERE chain_id = $1
			UNION ALL
			SELECT ` + chainColumnsOf("c") + `
			FROM chains c
			INNER JOIN chain_path cp ON c.chain_id = cp.next_chain_id OR c.chain_id = cp.previous_chain_id
		)
		SELECT ` + chainColumns + `
		FROM chain_path
		ORDER BY created_at
	`
	return r.queryChains(ctx, query, chainID)
}

func (r *chainRepository) UpdateStatus(ctx context.Context, id string, status domain.ChainStatus) error {
	query := `UPDATE chains SET status = $1 WHERE chain_id = $2`
	result, err := r.db.Exec(ctx, query, string(status), id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *chainRepository) UpdateStatusIfCurrent(ctx context.Context, id string, customerID string, current, next domain.ChainStatus) error {
	result, err := r.db.Exec(ctx, `
		UPDATE chains
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE chain_id = $2 AND status = $3
		  AND (status <> $4 OR expires_at > CURRENT_TIMESTAMP)
	`, next, id, current, domain.ChainPending)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CompleteExchange завершает обмен: создаёт копии товаров у новых владельцев,
// архивирует исходные записи и обновляет статус цепочки.
func (r *chainRepository) CompleteExchange(ctx context.Context, chainID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Получить цепочку
	var chain domain.Chain
	err = tx.QueryRow(ctx, `
			SELECT chain_id, from_product_id, to_product_id, initiator_id, recipient_id, status
			FROM chains
			WHERE chain_id = $1
			FOR UPDATE
		`, chainID).Scan(
		&chain.ChainID,
		&chain.FromProductID,
		&chain.ToProductID,
		&chain.InitiatorID,
		&chain.RecipientID,
		&chain.Status,
	)
	if err != nil {
		return err
	}

	if chain.Status != string(domain.ChainActive) {
		return errors.New("chain must be active to complete")
	}

	// Цель-категория: целевого товара нет, обмен владельцев невозможен.
	if chain.ToProductID == nil {
		return errors.New("chain with category goal cannot complete exchange")
	}

	toProductID := *chain.ToProductID

	// 2. Заблокировать товары и прочитать текущих владельцев.
	//
	// Без блокировки два обмена, завершающихся одновременно, читают одного и
	// того же владельца и записывают результат поверх друг друга: вещь уезжает
	// дважды. Порядок фиксирован по product_id — иначе обмены с общим товаром
	// берут блокировки крест-накрест и встают в дедлок.
	owners := make(map[string]string, 2)
	rows, err := tx.Query(ctx, `
			SELECT product_id, customer_id
			FROM products
			WHERE product_id IN ($1, $2)
			  AND status = $3
			ORDER BY product_id
			FOR UPDATE
		`, chain.FromProductID, toProductID, string(domain.ProductActive))
	if err != nil {
		return err
	}
	for rows.Next() {
		var productID, ownerID string
		if err := rows.Scan(&productID, &ownerID); err != nil {
			rows.Close()
			return err
		}
		owners[productID] = ownerID
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	fromOwner, ok := owners[chain.FromProductID]
	if !ok {
		return sql.ErrNoRows
	}
	toOwner, ok := owners[toProductID]
	if !ok {
		return sql.ErrNoRows
	}

	// 3. Создать новые активные карточки у получателей. Исходные карточки
	// сохраняются у прежних владельцев для истории обменов.
	_, err = tx.Exec(ctx, `
			INSERT INTO products (customer_id, category_id, title, description, image, price, location, status)
			SELECT $1, category_id, title, description, image, price, location, $3
			FROM products
			WHERE product_id = $2
		`, toOwner, chain.FromProductID, string(domain.ProductActive))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
			INSERT INTO products (customer_id, category_id, title, description, image, price, location, status)
			SELECT $1, category_id, title, description, image, price, location, $3
			FROM products
			WHERE product_id = $2
		`, fromOwner, toProductID, string(domain.ProductActive))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
			UPDATE products
			SET status = $1, updated_at = CURRENT_TIMESTAMP
			WHERE product_id IN ($2, $3)
		`, string(domain.ProductArchived), chain.FromProductID, toProductID)
	if err != nil {
		return err
	}

	// 4. Обновить статус цепочки
	_, err = tx.Exec(ctx, `
			UPDATE chains SET status = $1 WHERE chain_id = $2
		`, string(domain.ChainCompleted), chainID)
	if err != nil {
		return err
	}

	// 5. Закрыть конкурирующие предложения по тем же товарам.
	//
	// Вещи уже уехали к новым владельцам, и остальные предложения обещают то,
	// чего у людей больше нет. Оставить их висеть — значит дать второй стороне
	// принять предложение и приехать на встречу впустую.
	// Всё в той же транзакции: иначе между сменой владельца и закрытием
	// предложений существует момент, когда чужой оффер ещё можно принять.
	_, err = tx.Exec(ctx, `
		UPDATE chains
		SET status = CASE
			WHEN (from_product_id = $4 AND initiator_id = $7)
			  OR (from_product_id = $5 AND initiator_id = $8)
			THEN $1
			ELSE $9
		END,
		updated_at = CURRENT_TIMESTAMP
		WHERE chain_id <> $2
		  AND status IN ($3, $6)
		  AND (from_product_id IN ($4, $5) OR to_product_id IN ($4, $5))
		`, string(domain.ChainCancelled), chainID, string(domain.ChainPending),
		chain.FromProductID, toProductID, string(domain.ChainActive),
		chain.InitiatorID, *chain.RecipientID, string(domain.ChainUnavailable))
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *chainRepository) ExpirePending(ctx context.Context) ([]domain.Chain, error) {
	return r.queryChains(ctx, `
		UPDATE chains
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE status = $2 AND expires_at <= CURRENT_TIMESTAMP
		RETURNING `+chainColumns,
		domain.ChainExpired,
		domain.ChainPending,
	)
}

func (r *chainRepository) Delete(ctx context.Context, id, initiatorID string) error {
	result, err := r.db.Exec(ctx, `
		DELETE FROM chains
		WHERE chain_id = $1 AND initiator_id = $2 AND status = $3
	`, id, initiatorID, domain.ChainPending)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}
