package usecase

import (
	"context"

	"github.com/boopateway/boopateway/internal/domain/repository"
)

// GetPersonImageUseCase devolve o caminho da foto da pessoa para o portal administrativo.
type GetPersonImageUseCase struct {
	validate *ValidateAdminTokenUseCase
	persons  repository.PersonRepository
}

func NewGetPersonImageUseCase(validate *ValidateAdminTokenUseCase, persons repository.PersonRepository) *GetPersonImageUseCase {
	return &GetPersonImageUseCase{validate: validate, persons: persons}
}

func (uc *GetPersonImageUseCase) Execute(ctx context.Context, token string, id int64) (string, error) {
	if _, err := uc.validate.Execute(ctx, token); err != nil {
		return "", err
	}
	p, err := uc.persons.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	return p.ImagePath, nil
}
