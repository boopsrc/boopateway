package booprecognition

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/gateway"
)

// AdminRecognitionGateway adapta o Client Boop para as operações administrativas do domínio.
type AdminRecognitionGateway struct {
	client *Client
}

var _ gateway.AdminRecognitionGateway = (*AdminRecognitionGateway)(nil)

func NewAdminRecognitionGateway(client *Client) *AdminRecognitionGateway {
	return &AdminRecognitionGateway{client: client}
}

func (g *AdminRecognitionGateway) ValidateToken(ctx context.Context, token string) ([]string, error) {
	scopes, err := g.client.ValidateToken(ctx, token)
	return scopes, mapAdminError("boop validate token", err)
}

func (g *AdminRecognitionGateway) DeletePerson(ctx context.Context, token string, partnerPersonID int64) error {
	return mapAdminError("boop delete person", g.client.DeletePerson(ctx, token, partnerPersonID))
}

func mapAdminError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errUnauthorized):
		return entity.ErrInvalidToken
	case errors.Is(err, errForbidden):
		return entity.ErrForbidden
	default:
		slog.Error(op, "err", err)
		return fmt.Errorf("%w: %v", entity.ErrPartnerUnavailable, err)
	}
}
