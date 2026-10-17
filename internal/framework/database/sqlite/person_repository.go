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
