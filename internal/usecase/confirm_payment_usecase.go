package usecase

import (
	"context"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/gateway"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

type ConfirmPaymentInput struct {
	TransactionID string
	Code          string
}

type ConfirmPaymentOutput struct {
	Transaction *entity.Transaction
	Person      *entity.Person
}

type ConfirmPaymentUseCase struct {
	persons      repository.PersonRepository
	transactions repository.TransactionRepository
	hasher       gateway.CodeHasher
}

func NewConfirmPaymentUseCase(
	persons repository.PersonRepository,
	transactions repository.TransactionRepository,
	hasher gateway.CodeHasher,
) *ConfirmPaymentUseCase {
	return &ConfirmPaymentUseCase{persons: persons, transactions: transactions, hasher: hasher}
}

func (uc *ConfirmPaymentUseCase) Execute(ctx context.Context, in ConfirmPaymentInput) (*ConfirmPaymentOutput, error) {
	if err := entity.ValidateCode(in.Code); err != nil {
		return nil, err
	}

	t, err := uc.transactions.FindByID(ctx, in.TransactionID)
	if err != nil {
		return nil, err
	}
	if !t.IsPending() {
		return nil, entity.ErrTransactionNotPending
	}

	person, err := uc.persons.FindByID(ctx, *t.PersonID)
	if err != nil {
		return nil, err
	}

	if !uc.hasher.Compare(person.CodeHash, in.Code) {
		attempts, status, err := uc.transactions.RegisterFailedAttempt(ctx, t.ID, entity.MaxConfirmAttempts)
		if err != nil {
			return nil, err
		}
		return nil, &entity.InvalidCodeError{
			RemainingAttempts: max(0, entity.MaxConfirmAttempts-attempts),
			Status:            status,
		}
	}

	if err := uc.transactions.Confirm(ctx, t.ID); err != nil {
		return nil, err
	}
	t.Status = entity.StatusConfirmed
	return &ConfirmPaymentOutput{Transaction: t, Person: person}, nil
}
