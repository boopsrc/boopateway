package sqlite

import (
	"context"
	"database/sql"
	"strings"

	"github.com/boopateway/boopateway/internal/domain/repository"
)

// requestLogRetention é por quanto tempo os logs de requisição são mantidos.
const requestLogRetention = "-30 days"

type RequestLogRepository struct{ db *sql.DB }

var _ repository.RequestLogRepository = (*RequestLogRepository)(nil)

func NewRequestLogRepository(db *sql.DB) *RequestLogRepository {
	return &RequestLogRepository{db: db}
}

func (r *RequestLogRepository) Insert(ctx context.Context, l *repository.RequestLog) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO request_logs (method, path, route, query, status, duration_ms, client_ip, user_agent, error)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.Method, l.Path, l.Route, l.Query, l.Status, l.DurationMs, l.ClientIP, l.UserAgent, l.Error)
	if err != nil {
		return err
	}
	l.ID, _ = res.LastInsertId()
	// Limpeza ocasional (~1 a cada 200 inserts) para não crescer sem limite.
	if l.ID%200 == 0 {
		_, err = r.db.ExecContext(ctx,
			`DELETE FROM request_logs WHERE created_at < strftime('%Y-%m-%dT%H:%M:%fZ','now',?)`, requestLogRetention)
	}
	return err
}

func (r *RequestLogRepository) Search(ctx context.Context, f repository.RequestLogFilter, limit, offset int) ([]repository.RequestLog, int, error) {
	var (
		where []string
		args  []any
	)
	if f.Method != "" {
		where, args = append(where, "method = ?"), append(args, f.Method)
	}
	if f.Path != "" {
		where, args = append(where, "path LIKE ? ESCAPE '\\'"), append(args, "%"+likeEscape(f.Path)+"%")
	}
	if f.StatusClass != 0 {
		where, args = append(where, "status / 100 = ?"), append(args, f.StatusClass)
	}
	if f.From != "" {
		where, args = append(where, "created_at >= ?"), append(args, f.From)
	}
	if f.To != "" {
		where, args = append(where, "created_at <= ?"), append(args, f.To)
	}
	cond := ""
	if len(where) > 0 {
		cond = " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM request_logs`+cond, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, created_at, method, path, route, query, status, duration_ms, client_ip, user_agent, error
		 FROM request_logs`+cond+` ORDER BY id DESC LIMIT ? OFFSET ?`,
		append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []repository.RequestLog{}
	for rows.Next() {
		var l repository.RequestLog
		if err := rows.Scan(&l.ID, &l.CreatedAt, &l.Method, &l.Path, &l.Route, &l.Query, &l.Status,
			&l.DurationMs, &l.ClientIP, &l.UserAgent, &l.Error); err != nil {
			return nil, 0, err
		}
		items = append(items, l)
	}
	return items, total, rows.Err()
}
