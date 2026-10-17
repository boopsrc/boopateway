package usecase

import (
	"context"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type ListPersonsInput struct {
	Token string
	Page  int
	Limit int
}

type ListPersonsOutput struct {
	Items []entity.Person
	Total int
	Page  int
	Limit int
}

type ListPersonsUseCase struct {
	validate *ValidateAdminTokenUseCase
	persons  repository.PersonRepository
}

func NewListPersonsUseCase(validate *ValidateAdminTokenUseCase, persons repository.PersonRepository) *ListPersonsUseCase {
	return &ListPersonsUseCase{validate: validate, persons: persons}
}

func (uc *ListPersonsUseCase) Execute(ctx context.Context, in ListPersonsInput) (*ListPersonsOutput, error) {
	if _, err := uc.validate.Execute(ctx, in.Token); err != nil {
		return nil, err
	}
	page := max(in.Page, 1)
	limit := in.Limit
	if limit <= 0 {
		limit = DefaultPageSize
	}
	limit = min(limit, MaxPageSize)

	items, total, err := uc.persons.List(ctx, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	return &ListPersonsOutput{Items: items, Total: total, Page: page, Limit: limit}, nil
}
