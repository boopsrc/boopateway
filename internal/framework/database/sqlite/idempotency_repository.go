package sqlite

import (
	"context"
	"database/sql"

	"github.com/boopateway/boopateway/internal/domain/repository"
)

type IdempotencyRepository struct{ db *sql.DB }

var _ repository.IdempotencyRepository = (*IdempotencyRepository)(nil)

func NewIdempotencyRepository(db *sql.DB) *IdempotencyRepository {
	return &IdempotencyRepository{db: db}
}

func (r *IdempotencyRepository) Begin(ctx context.Context, key, scope, hash string) (*repository.IdempotencyRecord, bool, error) {
	// Chaves expiram após 24h.
	if _, err := r.db.ExecContext(ctx,
		`DELETE FROM idempotency_keys WHERE created_at < strftime('%Y-%m-%dT%H:%M:%fZ','now','-1 day')`); err != nil {
		return nil, false, err
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO idempotency_keys (key, scope, request_hash) VALUES (?, ?, ?) ON CONFLICT DO NOTHING`,
		key, scope, hash)
	if err != nil {
		return nil, false, err
	}
	if n, _ := res.RowsAffected(); n == 1 {
		return nil, true, nil
	}
	var rec repository.IdempotencyRecord
	err = r.db.QueryRowContext(ctx,
		`SELECT request_hash, status_code, response FROM idempotency_keys WHERE key = ? AND scope = ?`,
		key, scope).Scan(&rec.RequestHash, &rec.StatusCode, &rec.Response)
	return &rec, false, err
}

func (r *IdempotencyRepository) Complete(ctx context.Context, key, scope string, status int, body []byte) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE idempotency_keys SET status_code = ?, response = ? WHERE key = ? AND scope = ?`,
		status, body, key, scope)
	return err
}

func (r *IdempotencyRepository) Release(ctx context.Context, key, scope string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM idempotency_keys WHERE key = ? AND scope = ?`, key, scope)
	return err
}
