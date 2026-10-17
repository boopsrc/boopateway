package usecase

import (
	"context"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

type GetPaymentUseCase struct {
	transactions repository.TransactionRepository
}

func NewGetPaymentUseCase(transactions repository.TransactionRepository) *GetPaymentUseCase {
	return &GetPaymentUseCase{transactions: transactions}
}

func (uc *GetPaymentUseCase) Execute(ctx context.Context, id string) (*entity.Transaction, error) {
	return uc.transactions.FindByID(ctx, id)
}
