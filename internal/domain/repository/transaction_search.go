package repository

import "github.com/boopateway/boopateway/internal/domain/entity"

// TransactionFilter filtra a busca de transações. Campos vazios/nil são ignorados.
type TransactionFilter struct {
	Name        string // contém, sem diferenciar maiúsculas
	Email       string // contém, sem diferenciar maiúsculas
	AmountCents *int64 // valor exato
	Status      entity.TransactionStatus
}

// TransactionSearchResult é a transação com os dados da pessoa (nil se não houver pessoa vinculada).
type TransactionSearchResult struct {
	entity.Transaction
	PersonName  *string
	PersonEmail *string
}
