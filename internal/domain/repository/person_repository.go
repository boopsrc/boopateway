package repository

import (
	"context"

	"github.com/boopateway/boopateway/internal/domain/entity"
)

type PersonRepository interface {
	Create(ctx context.Context, p *entity.Person) error
	FindByID(ctx context.Context, id int64) (*entity.Person, error)
	FindByPartnerPersonID(ctx context.Context, partnerID int64) (*entity.Person, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	List(ctx context.Context, limit, offset int) (items []entity.Person, total int, err error)
	// Delete remove a pessoa mantendo o histórico de transações (pendentes viram cancel_person_deleted).
	Delete(ctx context.Context, id int64) error
}
