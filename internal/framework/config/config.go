package config

import (
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	Addr           string
	DBPath         string
	ImageDir       string
	BoopBaseURL    string
	BoopToken      string
	BoopSearchPath string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:           env("ADDR", ":8080"),
		DBPath:         env("DB_PATH", filepath.Join("data", "boopateway.db")),
		ImageDir:       env("IMAGE_DIR", filepath.Join("data", "images")),
		BoopBaseURL:    env("BOOP_BASE_URL", "http://localhost:8000"),
		BoopToken:      os.Getenv("BOOP_TOKEN"),
		BoopSearchPath: env("BOOP_SEARCH_PATH", "/identify"),
	}
	if cfg.BoopToken == "" {
		return cfg, errors.New("BOOP_TOKEN é obrigatório")
	}
	return cfg, nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
