package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/boopateway/boopateway/internal/domain/entity"
)

// respondError traduz erros do domínio para respostas HTTP.
func respondError(c *gin.Context, err error) {
	var (
		validation  entity.ValidationError
		invalidCode *entity.InvalidCodeError
	)
	switch {
	case errors.As(err, &validation):
		c.JSON(http.StatusBadRequest, gin.H{"error": validation.Error()})
	case errors.As(err, &invalidCode):
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":              "código inválido",
			"status":             invalidCode.Status,
			"remaining_attempts": invalidCode.RemainingAttempts,
		})
	case errors.Is(err, entity.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, entity.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, entity.ErrPersonNotFound), errors.Is(err, entity.ErrTransactionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, entity.ErrEmailAlreadyExists),
		errors.Is(err, entity.ErrPersonAlreadyRegistered),
		errors.Is(err, entity.ErrTransactionNotPending):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, entity.ErrPartnerUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"error": entity.ErrPartnerUnavailable.Error()})
	default:
		slog.Error("unhandled error", "path", c.FullPath(), "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno"})
	}
}
