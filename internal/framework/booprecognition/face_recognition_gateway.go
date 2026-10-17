package booprecognition

import (
	"context"
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
		slog.Error("boop register person", "err", err)
		return nil, fmt.Errorf("%w: %v", entity.ErrPartnerUnavailable, err)
	}
	return &gateway.RegisteredPerson{PartnerPersonID: resp.Identification, Created: resp.Created}, nil
}

func (g *FaceRecognitionGateway) SearchPerson(ctx context.Context, img entity.Image) (int64, bool, error) {
	resp, found, err := g.client.SearchPerson(ctx, img.Data, img.ContentType)
	if err != nil {
		slog.Error("boop search person", "err", err)
		return 0, false, fmt.Errorf("%w: %v", entity.ErrPartnerUnavailable, err)
	}
	if !found {
		return 0, false, nil
	}
	return resp.Identification, true, nil
}
