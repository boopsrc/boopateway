package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

type PersonRepository struct{ db *sql.DB }

var _ repository.PersonRepository = (*PersonRepository)(nil)

func NewPersonRepository(db *sql.DB) *PersonRepository { return &PersonRepository{db: db} }

const personCols = `id, name, email, partner_person_id, image_path, code_hash, created_at`

func (r *PersonRepository) Create(ctx context.Context, p *entity.Person) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO persons (name, email, partner_person_id, image_path, code_hash)
		 VALUES (?, ?, ?, ?, ?) RETURNING id, created_at`,
		p.Name, p.Email, p.PartnerPersonID, p.ImagePath, p.CodeHash,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		if strings.Contains(err.Error(), "email") {
			return entity.ErrEmailAlreadyExists
		}
		return entity.ErrPersonAlreadyRegistered
	}
	return err
}

func (r *PersonRepository) FindByID(ctx context.Context, id int64) (*entity.Person, error) {
	return r.scan(r.db.QueryRowContext(ctx, `SELECT `+personCols+` FROM persons WHERE id = ?`, id))
}

func (r *PersonRepository) FindByPartnerPersonID(ctx context.Context, partnerID int64) (*entity.Person, error) {
	return r.scan(r.db.QueryRowContext(ctx, `SELECT `+personCols+` FROM persons WHERE partner_person_id = ?`, partnerID))
}

func (r *PersonRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM persons WHERE email = ?`, email).Scan(&n)
	return n > 0, err
}

func (r *PersonRepository) List(ctx context.Context, limit, offset int) ([]entity.Person, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM persons`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+personCols+` FROM persons ORDER BY id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []entity.Person{}
	for rows.Next() {
		var p entity.Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Email, &p.PartnerPersonID, &p.ImagePath, &p.CodeHash, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, rows.Err()
}

func (r *PersonRepository) Delete(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE transactions SET status = ?, updated_at = `+nowExpr+`
		 WHERE person_id = ? AND status = ?`,
		entity.StatusCancelPersonDeleted, id, entity.StatusPendingConfirmation); err != nil {
		return err
	}
	// Mantém partner_person_id como histórico; só desfaz o vínculo com a pessoa removida.
	if _, err := tx.ExecContext(ctx, `UPDATE transactions SET person_id = NULL WHERE person_id = ?`, id); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM persons WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrPersonNotFound
	}
	return tx.Commit()
}

func (r *PersonRepository) scan(row *sql.Row) (*entity.Person, error) {
	var p entity.Person
	err := row.Scan(&p.ID, &p.Name, &p.Email, &p.PartnerPersonID, &p.ImagePath, &p.CodeHash, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.ErrPersonNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
