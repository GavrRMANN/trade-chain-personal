package repository

import (
	"context"

	"trade-chain/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type negotiationRepository struct {
	db *pgxpool.Pool
}

func NewNegotiationRepository(db *pgxpool.Pool) NegotiationRepository {
	return &negotiationRepository{db: db}
}

func (r *negotiationRepository) AddMessage(ctx context.Context, message *domain.ChainMessage) (*domain.ChainMessage, error) {
	query := `
		INSERT INTO chain_messages (chain_id, customer_id, body)
		VALUES ($1, $2, $3)
		RETURNING message_id, chain_id, customer_id, body, created_at
	`

	var created domain.ChainMessage
	err := r.db.QueryRow(ctx, query,
		message.ChainID,
		message.CustomerID,
		message.Body,
	).Scan(
		&created.MessageID,
		&created.ChainID,
		&created.CustomerID,
		&created.Body,
		&created.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// ListMessages отдаёт переписку в порядке написания: читать разговор с конца
// неудобно, а объёмы здесь такие, что постраничность только мешает.
func (r *negotiationRepository) ListMessages(ctx context.Context, chainID string) ([]domain.ChainMessage, error) {
	query := `
		SELECT message_id, chain_id, customer_id, body, created_at
		FROM chain_messages
		WHERE chain_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]domain.ChainMessage, 0)
	for rows.Next() {
		var message domain.ChainMessage
		if err := rows.Scan(
			&message.MessageID,
			&message.ChainID,
			&message.CustomerID,
			&message.Body,
			&message.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

// Confirm записывает решение стороны об итоге обмена.
//
// Повторное подтверждение отбивается первичным ключом и превращается
// в конфликт, а не в тихую перезапись: передумать задним числом нельзя,
// иначе подтверждение перестаёт что-либо значить.
func (r *negotiationRepository) Confirm(ctx context.Context, confirmation *domain.ChainConfirmation) error {
	query := `
		INSERT INTO chain_confirmations (chain_id, customer_id, success, reason)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query,
		confirmation.ChainID,
		confirmation.CustomerID,
		confirmation.Success,
		confirmation.Reason,
	)
	return err
}

func (r *negotiationRepository) ListConfirmations(ctx context.Context, chainID string) ([]domain.ChainConfirmation, error) {
	query := `
		SELECT chain_id, customer_id, success, reason, created_at
		FROM chain_confirmations
		WHERE chain_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	confirmations := make([]domain.ChainConfirmation, 0)
	for rows.Next() {
		var confirmation domain.ChainConfirmation
		if err := rows.Scan(
			&confirmation.ChainID,
			&confirmation.CustomerID,
			&confirmation.Success,
			&confirmation.Reason,
			&confirmation.CreatedAt,
		); err != nil {
			return nil, err
		}
		confirmations = append(confirmations, confirmation)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return confirmations, nil
}
