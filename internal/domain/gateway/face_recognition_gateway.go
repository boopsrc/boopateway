package gateway

import (
	"context"

	"github.com/boopateway/boopateway/internal/domain/entity"
)

type RegisteredPerson struct {
	PartnerPersonID int64
	Created         bool // false = o parceiro já conhecia esta pessoa
}

// FaceRecognitionGateway abstrai o serviço de reconhecimento facial de terceiros.
// Implementação: internal/framework/booprecognition.
type FaceRecognitionGateway interface {
	RegisterPerson(ctx context.Context, img entity.Image) (*RegisteredPerson, error)
	// SearchPerson retorna found=false quando o parceiro não reconhece a foto.
	SearchPerson(ctx context.Context, img entity.Image) (partnerPersonID int64, found bool, err error)
}
