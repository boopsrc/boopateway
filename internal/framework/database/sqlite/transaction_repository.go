package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

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

func (r *TransactionRepository) ListByPersonID(ctx context.Context, personID int64, limit, offset int) ([]entity.Transaction, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM transactions WHERE person_id = ?`, personID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, amount_cents, partner_person_id, person_id, status, attempts, created_at, updated_at
		 FROM transactions WHERE person_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		personID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []entity.Transaction{}
	for rows.Next() {
		var t entity.Transaction
		if err := rows.Scan(&t.ID, &t.AmountCents, &t.PartnerPersonID, &t.PersonID, &t.Status, &t.Attempts, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, rows.Err()
}

func (r *TransactionRepository) Search(ctx context.Context, f repository.TransactionFilter, limit, offset int) ([]repository.TransactionSearchResult, int, error) {
	where := []string{"1=1"}
	var args []any
	if f.Name != "" {
		where = append(where, "p.name LIKE ? ESCAPE '\\'")
		args = append(args, "%"+likeEscape(f.Name)+"%")
	}
	if f.Email != "" {
		where = append(where, "p.email LIKE ? ESCAPE '\\'")
		args = append(args, "%"+likeEscape(f.Email)+"%")
	}
	if f.AmountCents != nil {
		where = append(where, "t.amount_cents = ?")
		args = append(args, *f.AmountCents)
	}
	if f.Status != "" {
		where = append(where, "t.status = ?")
		args = append(args, f.Status)
	}
	from := ` FROM transactions t LEFT JOIN persons p ON p.id = t.person_id WHERE ` + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+from, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT t.id, t.amount_cents, t.partner_person_id, t.person_id, t.status, t.attempts, t.created_at, t.updated_at,
		        p.name, p.email`+from+` ORDER BY t.created_at DESC LIMIT ? OFFSET ?`,
		append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []repository.TransactionSearchResult{}
	for rows.Next() {
		var it repository.TransactionSearchResult
		t := &it.Transaction
		if err := rows.Scan(&t.ID, &t.AmountCents, &t.PartnerPersonID, &t.PersonID, &t.Status, &t.Attempts,
			&t.CreatedAt, &t.UpdatedAt, &it.PersonName, &it.PersonEmail); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}
	return items, total, rows.Err()
}

// likeEscape evita que % e _ digitados pelo usuário virem curingas do LIKE.
func likeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
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
