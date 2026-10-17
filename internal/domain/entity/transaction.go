package entity

type TransactionStatus string

const (
	StatusPendingConfirmation TransactionStatus = "pending_confirmation"
	StatusConfirmed           TransactionStatus = "confirmed"
	StatusCancelNotFound      TransactionStatus = "cancel_notfound"
	StatusCancelInvalidCode   TransactionStatus = "cancel_invalid_code"

	MaxConfirmAttempts = 3
	MaxAmountCents     = 100_000_000_00
)

type Transaction struct {
	ID              string
	AmountCents     int64
	PartnerPersonID *int64 // id da pessoa no parceiro
	PersonID        *int64 // id da pessoa do nosso lado
	Status          TransactionStatus
	Attempts        int
	CreatedAt       string
	UpdatedAt       string
}

func ValidateAmount(amountCents int64) error {
	if amountCents <= 0 || amountCents > MaxAmountCents {
		return ValidationError("amount_cents deve ser um inteiro positivo (em centavos)")
	}
	return nil
}

func (t *Transaction) IsPending() bool {
	return t.Status == StatusPendingConfirmation && t.PersonID != nil
}
