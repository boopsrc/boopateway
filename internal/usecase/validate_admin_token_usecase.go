package usecase

import (
	"context"
	"slices"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/gateway"
)

type AdminSession struct {
	Scopes    []string
	CanDelete bool
}

// ValidateAdminTokenUseCase confere no serviço de reconhecimento facial se o token é válido.
type ValidateAdminTokenUseCase struct {
	recognition gateway.AdminRecognitionGateway
}

func NewValidateAdminTokenUseCase(recognition gateway.AdminRecognitionGateway) *ValidateAdminTokenUseCase {
	return &ValidateAdminTokenUseCase{recognition: recognition}
}

func (uc *ValidateAdminTokenUseCase) Execute(ctx context.Context, token string) (*AdminSession, error) {
	if token == "" {
		return nil, entity.ErrInvalidToken
	}
	scopes, err := uc.recognition.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return &AdminSession{Scopes: scopes, CanDelete: slices.Contains(scopes, gateway.ScopePersonsDelete)}, nil
}
