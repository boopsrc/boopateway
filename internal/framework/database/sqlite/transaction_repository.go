package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

type TransactionRepository struct{ db *sql.DB }

var _ repository.TransactionRepository = (*TransactionRepository)(nil)

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, t *entity.Transaction) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO transactions (id, amount_cents, partner_person_id, person_id, status)
		 VALUES (?, ?, ?, ?, ?) RETURNING created_at, updated_at`,
		t.ID, t.AmountCents, t.PartnerPersonID, t.PersonID, t.Status,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
}

func (r *TransactionRepository) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	var t entity.Transaction
	err := r.db.QueryRowContext(ctx,
		`SELECT id, amount_cents, partner_person_id, person_id, status, attempts, created_at, updated_at
		 FROM transactions WHERE id = ?`, id,
	).Scan(&t.ID, &t.AmountCents, &t.PartnerPersonID, &t.PersonID, &t.Status, &t.Attempts, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.ErrTransactionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TransactionRepository) Confirm(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE transactions SET status = ?, updated_at = `+nowExpr+`
		 WHERE id = ? AND status = ?`,
		entity.StatusConfirmed, id, entity.StatusPendingConfirmation)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrTransactionNotPending
	}
	return nil
}

func (r *TransactionRepository) RegisterFailedAttempt(ctx context.Context, id string, maxAttempts int) (int, entity.TransactionStatus, error) {
	var (
		attempts int
		status   entity.TransactionStatus
	)
	err := r.db.QueryRowContext(ctx,
		`UPDATE transactions
		 SET attempts = attempts + 1,
		     status = CASE WHEN attempts + 1 >= ? THEN ? ELSE status END,
		     updated_at = `+nowExpr+`
		 WHERE id = ? AND status = ?
		 RETURNING attempts, status`,
		maxAttempts, entity.StatusCancelInvalidCode, id, entity.StatusPendingConfirmation,
	).Scan(&attempts, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", entity.ErrTransactionNotPending
	}
	return attempts, status, err
}
