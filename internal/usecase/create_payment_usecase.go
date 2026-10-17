package usecase

import (
	"context"
	"crypto/rand"
	"errors"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/gateway"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

type CreatePaymentInput struct {
	AmountCents int64
	Image       []byte
}

type CreatePaymentOutput struct {
	Transaction *entity.Transaction
	Person      *entity.Person // nil quando não encontrada
}

type CreatePaymentUseCase struct {
	persons      repository.PersonRepository
	transactions repository.TransactionRepository
	recognition  gateway.FaceRecognitionGateway
}

func NewCreatePaymentUseCase(
	persons repository.PersonRepository,
	transactions repository.TransactionRepository,
	recognition gateway.FaceRecognitionGateway,
) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{persons: persons, transactions: transactions, recognition: recognition}
}

func (uc *CreatePaymentUseCase) Execute(ctx context.Context, in CreatePaymentInput) (*CreatePaymentOutput, error) {
	if err := entity.ValidateAmount(in.AmountCents); err != nil {
		return nil, err
	}
	img, err := entity.NewImage(in.Image)
	if err != nil {
		return nil, err
	}

	person, err := uc.identify(ctx, img)
	if err != nil {
		return nil, err
	}

	t := &entity.Transaction{
		ID:          "txn_" + rand.Text(),
		AmountCents: in.AmountCents,
		Status:      entity.StatusCancelNotFound,
	}
	if person != nil {
		t.Status = entity.StatusPendingConfirmation
		t.PersonID = &person.ID
		t.PartnerPersonID = &person.PartnerPersonID
	}
	if err := uc.transactions.Create(ctx, t); err != nil {
		return nil, err
	}
	return &CreatePaymentOutput{Transaction: t, Person: person}, nil
}

// identify busca a pessoa no parceiro e depois na nossa base. nil = não encontrada.
func (uc *CreatePaymentUseCase) identify(ctx context.Context, img entity.Image) (*entity.Person, error) {
	partnerID, found, err := uc.recognition.SearchPerson(ctx, img)
	if err != nil || !found {
		return nil, err
	}
	person, err := uc.persons.FindByPartnerPersonID(ctx, partnerID)
	if errors.Is(err, entity.ErrPersonNotFound) {
		return nil, nil
	}
	return person, err
}
