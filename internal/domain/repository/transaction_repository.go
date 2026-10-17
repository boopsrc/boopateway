package repository

import (
	"context"

	"github.com/boopateway/boopateway/internal/domain/entity"
)

type TransactionRepository interface {
	Create(ctx context.Context, t *entity.Transaction) error
	FindByID(ctx context.Context, id string) (*entity.Transaction, error)
	// Confirm só altera se a transação ainda estiver pendente; caso contrário ErrTransactionNotPending.
	Confirm(ctx context.Context, id string) error
	// RegisterFailedAttempt incrementa tentativas e cancela ao atingir o limite.
	RegisterFailedAttempt(ctx context.Context, id string, maxAttempts int) (attempts int, status entity.TransactionStatus, err error)
}
