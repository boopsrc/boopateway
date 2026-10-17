package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const nowExpr = `strftime('%Y-%m-%dT%H:%M:%fZ','now')`

const schema = `
CREATE TABLE IF NOT EXISTS persons (
	id                INTEGER PRIMARY KEY AUTOINCREMENT,
	name              TEXT    NOT NULL,
	email             TEXT    NOT NULL UNIQUE,
	partner_person_id INTEGER NOT NULL UNIQUE,
	image_path        TEXT    NOT NULL,
	code_hash         BLOB    NOT NULL,
	created_at        TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS transactions (
	id                TEXT    PRIMARY KEY,
	amount_cents      INTEGER NOT NULL CHECK (amount_cents > 0),
	partner_person_id INTEGER,
	person_id         INTEGER REFERENCES persons(id),
	status            TEXT    NOT NULL,
	attempts          INTEGER NOT NULL DEFAULT 0,
	created_at        TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
	updated_at        TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
	key          TEXT    NOT NULL,
	scope        TEXT    NOT NULL,
	request_hash TEXT    NOT NULL,
	status_code  INTEGER NOT NULL DEFAULT 0,
	response     BLOB,
	created_at   TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
	PRIMARY KEY (key, scope)
);

CREATE TABLE IF NOT EXISTS request_logs (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
	method      TEXT    NOT NULL,
	path        TEXT    NOT NULL,
	route       TEXT    NOT NULL DEFAULT '',
	query       TEXT    NOT NULL DEFAULT '',
	status      INTEGER NOT NULL,
	duration_ms INTEGER NOT NULL,
	client_ip   TEXT    NOT NULL DEFAULT '',
	user_agent  TEXT    NOT NULL DEFAULT '',
	error       TEXT    NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs (created_at);
`

func Open(ctx context.Context, path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, err
	}
	dsn := "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite tem um único escritor; uma conexão evita SQLITE_BUSY.
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}
