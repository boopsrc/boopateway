package handler

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenIssuer gera um token admin válido no booprecognition.
type TokenIssuer func(ctx context.Context) (string, error)

// AuthHandler troca a senha de admin por um token do booprecognition gerado na hora,
// para o portal não precisar que o token seja colado manualmente.
type AuthHandler struct {
	password string
	issue    TokenIssuer
}

func NewAuthHandler(password string, issue TokenIssuer) *AuthHandler {
	return &AuthHandler{password: password, issue: issue}
}

type loginRequest struct {
	Password string `json:"password"`
}

// Login POST /admin/login
func (h *AuthHandler) Login(c *gin.Context) {
	if h.password == "" || h.issue == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "login por senha desabilitado (defina ADMIN_PASSWORD)"})
		return
	}
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "informe a senha"})
		return
	}
	// Compara os hashes para o tempo não depender do tamanho da senha.
	got, want := sha256.Sum256([]byte(req.Password)), sha256.Sum256([]byte(h.password))
	if subtle.ConstantTimeCompare(got[:], want[:]) != 1 {
		time.Sleep(time.Second) // atrasa tentativas de força bruta
		c.JSON(http.StatusUnauthorized, gin.H{"error": "senha inválida"})
		return
	}
	token, err := h.issue(c.Request.Context())
	if err != nil {
		slog.Error("admin login: falha ao gerar token", "err", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "não foi possível gerar o token no serviço de reconhecimento facial"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
