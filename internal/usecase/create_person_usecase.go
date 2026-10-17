package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/gateway"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

type CreatePersonInput struct {
	Name  string
	Email string
	Code  string // opcional, padrão 0000
	Image []byte
}

type CreatePersonUseCase struct {
	persons     repository.PersonRepository
	recognition gateway.FaceRecognitionGateway
	images      gateway.ImageStorage
	hasher      gateway.CodeHasher
}

func NewCreatePersonUseCase(
	persons repository.PersonRepository,
	recognition gateway.FaceRecognitionGateway,
	images gateway.ImageStorage,
	hasher gateway.CodeHasher,
) *CreatePersonUseCase {
	return &CreatePersonUseCase{persons: persons, recognition: recognition, images: images, hasher: hasher}
}

func (uc *CreatePersonUseCase) Execute(ctx context.Context, in CreatePersonInput) (*entity.Person, error) {
	name, email, code, err := entity.NormalizePersonInput(in.Name, in.Email, in.Code)
	if err != nil {
		return nil, err
	}
	img, err := entity.NewImage(in.Image)
	if err != nil {
		return nil, err
	}

	if exists, err := uc.persons.EmailExists(ctx, email); err != nil {
		return nil, err
	} else if exists {
		return nil, entity.ErrEmailAlreadyExists
	}

	registered, err := uc.recognition.RegisterPerson(ctx, img)
	if err != nil {
		return nil, err
	}
	if _, err := uc.persons.FindByPartnerPersonID(ctx, registered.PartnerPersonID); err == nil {
		return nil, entity.ErrPersonAlreadyRegistered
	} else if !errors.Is(err, entity.ErrPersonNotFound) {
		return nil, err
	}

	codeHash, err := uc.hasher.Hash(code)
	if err != nil {
		return nil, err
	}

	path, err := uc.images.Save(ctx, fmt.Sprintf("%d_%s", registered.PartnerPersonID, rand.Text()[:10]), img)
	if err != nil {
		return nil, err
	}

	p := &entity.Person{
		Name:            name,
		Email:           email,
		PartnerPersonID: registered.PartnerPersonID,
		ImagePath:       path,
		CodeHash:        codeHash,
	}
	if err := uc.persons.Create(ctx, p); err != nil {
		_ = uc.images.Delete(context.WithoutCancel(ctx), path)
		return nil, err
	}
	return p, nil
}
