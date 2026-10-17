package storage

import (
	"context"
	"os"
	"path/filepath"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/gateway"
)

// LocalImageStorage grava as fotos das pessoas em disco.
type LocalImageStorage struct{ dir string }

var _ gateway.ImageStorage = (*LocalImageStorage)(nil)

func NewLocalImageStorage(dir string) (*LocalImageStorage, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	return &LocalImageStorage{dir: dir}, nil
}

func (s *LocalImageStorage) Save(_ context.Context, name string, img entity.Image) (string, error) {
	path := filepath.Join(s.dir, name+img.Extension())
	if err := os.WriteFile(path, img.Data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func (s *LocalImageStorage) Delete(_ context.Context, path string) error {
	return os.Remove(path)
}
