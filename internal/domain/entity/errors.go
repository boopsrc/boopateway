package entity

import (
	"errors"
	"fmt"
)

var (
	ErrPersonNotFound          = errors.New("pessoa não encontrada")
	ErrEmailAlreadyExists      = errors.New("email já cadastrado")
	ErrPersonAlreadyRegistered = errors.New("esta pessoa já está cadastrada")
	ErrTransactionNotFound     = errors.New("transação não encontrada")
	ErrTransactionNotPending   = errors.New("transação não está pendente de confirmação")
	ErrPartnerUnavailable      = errors.New("falha ao comunicar com o serviço de reconhecimento facial")
)

// ValidationError representa entrada inválida do cliente.
type ValidationError string

func (e ValidationError) Error() string { return string(e) }

// InvalidCodeError é retornado quando o código de confirmação não confere.
type InvalidCodeError struct {
	RemainingAttempts int
	Status            TransactionStatus
}

func (e *InvalidCodeError) Error() string {
	return fmt.Sprintf("código inválido (%d tentativas restantes)", e.RemainingAttempts)
}
