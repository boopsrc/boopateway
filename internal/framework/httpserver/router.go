package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/boopateway/boopateway/internal/domain/repository"
	"github.com/boopateway/boopateway/internal/framework/httpserver/docs"
	"github.com/boopateway/boopateway/internal/framework/httpserver/handler"
	"github.com/boopateway/boopateway/internal/framework/httpserver/middleware"
)

type Handlers struct {
	Person  *handler.PersonHandler
	Payment *handler.PaymentHandler
}

func NewRouter(h Handlers, idempotency repository.IdempotencyRepository) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.MaxMultipartMemory = middleware.MaxBodyBytes

	idem := middleware.Idempotency(idempotency)

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/docs", docs.Handler)

	persons := r.Group("/persons")
	persons.POST("", idem, h.Person.Create)
	persons.GET("/:id", h.Person.Get)

	payments := r.Group("/payments")
	payments.POST("", idem, h.Payment.Create)
	payments.GET("/:id", h.Payment.Get)
	payments.POST("/:id/confirm", idem, h.Payment.Confirm)

	return r
}
