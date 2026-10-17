package usecase

import (
	"context"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

type ListPersonTransactionsInput struct {
	Token    string
	PersonID int64
	Page     int
	Limit    int
}

type ListPersonTransactionsOutput struct {
	Items []entity.Transaction
	Total int
	Page  int
	Limit int
}

// ListPersonTransactionsUseCase mostra o histórico de transações de uma pessoa no portal.
type ListPersonTransactionsUseCase struct {
	validate     *ValidateAdminTokenUseCase
	persons      repository.PersonRepository
	transactions repository.TransactionRepository
}

func NewListPersonTransactionsUseCase(
	validate *ValidateAdminTokenUseCase,
	persons repository.PersonRepository,
	transactions repository.TransactionRepository,
) *ListPersonTransactionsUseCase {
	return &ListPersonTransactionsUseCase{validate: validate, persons: persons, transactions: transactions}
}

func (uc *ListPersonTransactionsUseCase) Execute(ctx context.Context, in ListPersonTransactionsInput) (*ListPersonTransactionsOutput, error) {
	if _, err := uc.validate.Execute(ctx, in.Token); err != nil {
		return nil, err
	}
	if _, err := uc.persons.FindByID(ctx, in.PersonID); err != nil {
		return nil, err
	}
	page := max(in.Page, 1)
	limit := in.Limit
	if limit <= 0 {
		limit = DefaultPageSize
	}
	limit = min(limit, MaxPageSize)

	items, total, err := uc.transactions.ListByPersonID(ctx, in.PersonID, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	return &ListPersonTransactionsOutput{Items: items, Total: total, Page: page, Limit: limit}, nil
}
