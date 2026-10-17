package booprecognition

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/gateway"
)

// FaceRecognitionGateway adapta o Client Boop para a interface do domínio.
type FaceRecognitionGateway struct {
	client *Client
}

var _ gateway.FaceRecognitionGateway = (*FaceRecognitionGateway)(nil)

func NewFaceRecognitionGateway(client *Client) *FaceRecognitionGateway {
	return &FaceRecognitionGateway{client: client}
}

func (g *FaceRecognitionGateway) RegisterPerson(ctx context.Context, img entity.Image) (*gateway.RegisteredPerson, error) {
	resp, err := g.client.RegisterPerson(ctx, img.Data, img.ContentType)
	if err != nil {
		return nil, mapError("boop register person", err)
	}
	return &gateway.RegisteredPerson{PartnerPersonID: resp.Identification, Created: resp.Created}, nil
}

func (g *FaceRecognitionGateway) SearchPerson(ctx context.Context, img entity.Image) (int64, bool, error) {
	resp, found, err := g.client.SearchPerson(ctx, img.Data, img.ContentType)
	if err != nil {
		return 0, false, mapError("boop search person", err)
	}
	if !found {
		return 0, false, nil
	}
	return resp.Identification, true, nil
}

// mapError separa foto recusada pelo Boop (erro do cliente, 400) de falha do parceiro (502).
func mapError(op string, err error) error {
	var rejected *ImageRejectedError
	if errors.As(err, &rejected) {
		slog.Warn(op, "err", err)
		return entity.ValidationError("imagem recusada pelo reconhecimento facial: " + rejected.Detail)
	}
	slog.Error(op, "err", err)
	return fmt.Errorf("%w: %v", entity.ErrPartnerUnavailable, err)
}
