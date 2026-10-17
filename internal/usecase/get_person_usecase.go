package usecase

import (
	"context"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

type GetPersonUseCase struct {
	persons repository.PersonRepository
}

func NewGetPersonUseCase(persons repository.PersonRepository) *GetPersonUseCase {
	return &GetPersonUseCase{persons: persons}
}

func (uc *GetPersonUseCase) Execute(ctx context.Context, id int64) (*entity.Person, error) {
	return uc.persons.FindByID(ctx, id)
}
