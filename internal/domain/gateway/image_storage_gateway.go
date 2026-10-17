package gateway

import (
	"context"

	"github.com/boopateway/boopateway/internal/domain/entity"
)

type ImageStorage interface {
	Save(ctx context.Context, name string, img entity.Image) (path string, err error)
	Delete(ctx context.Context, path string) error
}
