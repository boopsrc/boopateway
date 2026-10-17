package repository

import "context"

type IdempotencyRecord struct {
	RequestHash string
	StatusCode  int // 0 = em processamento
	Response    []byte
}

type IdempotencyRepository interface {
	// Begin reserva a chave. started=true: esta requisição é dona da chave e deve processar.
	Begin(ctx context.Context, key, scope, requestHash string) (rec *IdempotencyRecord, started bool, err error)
	Complete(ctx context.Context, key, scope string, statusCode int, response []byte) error
	Release(ctx context.Context, key, scope string) error
}
