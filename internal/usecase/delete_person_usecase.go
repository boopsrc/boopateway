package usecase

import (
	"context"
	"log/slog"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/gateway"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

type DeletePersonUseCase struct {
	validate    *ValidateAdminTokenUseCase
	persons     repository.PersonRepository
	recognition gateway.AdminRecognitionGateway
	images      gateway.ImageStorage
}

func NewDeletePersonUseCase(
	validate *ValidateAdminTokenUseCase,
	persons repository.PersonRepository,
	recognition gateway.AdminRecognitionGateway,
	images gateway.ImageStorage,
) *DeletePersonUseCase {
	return &DeletePersonUseCase{validate: validate, persons: persons, recognition: recognition, images: images}
}

// Execute remove a pessoa no parceiro, na base e a foto. As transações são mantidas.
func (uc *DeletePersonUseCase) Execute(ctx context.Context, token string, id int64) error {
	session, err := uc.validate.Execute(ctx, token)
	if err != nil {
		return err
	}
	if !session.CanDelete {
		return entity.ErrForbidden
	}

	p, err := uc.persons.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.recognition.DeletePerson(ctx, token, p.PartnerPersonID); err != nil {
		return err
	}
	if err := uc.persons.Delete(ctx, p.ID); err != nil {
		return err
	}
	if err := uc.images.Delete(context.WithoutCancel(ctx), p.ImagePath); err != nil {
		slog.Warn("delete person image", "path", p.ImagePath, "err", err)
	}
	return nil
}
