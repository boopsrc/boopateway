package usecase

import (
	"context"
	"strings"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

type SearchTransactionsInput struct {
	Token  string
	Filter repository.TransactionFilter
	Page   int
	Limit  int
}

type SearchTransactionsOutput struct {
	Items []repository.TransactionSearchResult
	Total int
	Page  int
	Limit int
}

// SearchTransactionsUseCase busca transações de todas as pessoas por nome, e-mail, valor ou status.
type SearchTransactionsUseCase struct {
	validate     *ValidateAdminTokenUseCase
	transactions repository.TransactionRepository
}

func NewSearchTransactionsUseCase(validate *ValidateAdminTokenUseCase, transactions repository.TransactionRepository) *SearchTransactionsUseCase {
	return &SearchTransactionsUseCase{validate: validate, transactions: transactions}
}

var validStatuses = map[entity.TransactionStatus]bool{
	entity.StatusPendingConfirmation: true,
	entity.StatusConfirmed:           true,
	entity.StatusCancelNotFound:      true,
	entity.StatusCancelInvalidCode:   true,
	entity.StatusCancelPersonDeleted: true,
}

func (uc *SearchTransactionsUseCase) Execute(ctx context.Context, in SearchTransactionsInput) (*SearchTransactionsOutput, error) {
	if _, err := uc.validate.Execute(ctx, in.Token); err != nil {
		return nil, err
	}
	f := in.Filter
	f.Name = strings.TrimSpace(f.Name)
	f.Email = strings.TrimSpace(f.Email)
	if f.Status != "" && !validStatuses[f.Status] {
		return nil, entity.ValidationError("status inválido")
	}
	if f.AmountCents != nil && *f.AmountCents <= 0 {
		return nil, entity.ValidationError("amount_cents deve ser um inteiro positivo (em centavos)")
	}

	page := max(in.Page, 1)
	limit := in.Limit
	if limit <= 0 {
		limit = DefaultPageSize
	}
	limit = min(limit, MaxPageSize)

	items, total, err := uc.transactions.Search(ctx, f, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	return &SearchTransactionsOutput{Items: items, Total: total, Page: page, Limit: limit}, nil
}
