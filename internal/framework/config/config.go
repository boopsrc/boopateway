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
	AdminPassword  string // ADMIN_PASSWORD: senha do login do portal (vazio = login desabilitado)
	AdminBoopToken string // ADMIN_BOOP_TOKEN: token fixo entregue ao portal sem OAuth (padrão: BOOP_TOKEN)
	// Credenciais OAuth (client_credentials) para gerar um token novo a cada login do portal.
	BoopClientID     string
	BoopClientSecret string
	BoopAdminScope   string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:             env("ADDR", ":8080"),
		DBPath:           env("DB_PATH", filepath.Join("data", "boopateway.db")),
		ImageDir:         env("IMAGE_DIR", filepath.Join("data", "images")),
		BoopBaseURL:      env("BOOP_BASE_URL", "http://localhost:8000"),
		BoopToken:        os.Getenv("BOOP_TOKEN"),
		BoopSearchPath:   env("BOOP_SEARCH_PATH", "/identify"),
		AdminPassword:    os.Getenv("ADMIN_PASSWORD"),
		BoopClientID:     os.Getenv("BOOP_CLIENT_ID"),
		BoopClientSecret: os.Getenv("BOOP_CLIENT_SECRET"),
		BoopAdminScope:   os.Getenv("BOOP_ADMIN_SCOPE"),
	}
	cfg.AdminBoopToken = env("ADMIN_BOOP_TOKEN", cfg.BoopToken)
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
