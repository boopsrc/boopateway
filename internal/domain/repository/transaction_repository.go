package repository

import (
	"context"

	"github.com/boopateway/boopateway/internal/domain/entity"
)

type TransactionRepository interface {
	Create(ctx context.Context, t *entity.Transaction) error
	FindByID(ctx context.Context, id string) (*entity.Transaction, error)
	// ListByPersonID lista as transações da pessoa, da mais recente para a mais antiga.
	ListByPersonID(ctx context.Context, personID int64, limit, offset int) (items []entity.Transaction, total int, err error)
	// Search busca transações de todas as pessoas aplicando os filtros, da mais recente para a mais antiga.
	Search(ctx context.Context, f TransactionFilter, limit, offset int) (items []TransactionSearchResult, total int, err error)
	// Confirm só altera se a transação ainda estiver pendente; caso contrário ErrTransactionNotPending.
	Confirm(ctx context.Context, id string) error
	// RegisterFailedAttempt incrementa tentativas e cancela ao atingir o limite.
	RegisterFailedAttempt(ctx context.Context, id string, maxAttempts int) (attempts int, status entity.TransactionStatus, err error)
}
