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
	Admin   *handler.AdminHandler
}

func NewRouter(h Handlers, idempotency repository.IdempotencyRepository, logs repository.RequestLogRepository) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), middleware.RequestLog(logs), gin.Recovery())
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

	// Rotas do portal administrativo: token validado no serviço de reconhecimento facial.
	admin := r.Group("/admin", middleware.AdminToken())
	admin.GET("/session", h.Admin.Session)
	admin.GET("/persons", h.Admin.ListPersons)
	admin.GET("/persons/:id/image", h.Admin.PersonImage)
	admin.GET("/persons/:id/transactions", h.Admin.PersonTransactions)
	admin.DELETE("/persons/:id", h.Admin.DeletePerson)
	admin.GET("/transactions", h.Admin.SearchTransactions)
	admin.GET("/logs", h.Admin.RequestLogs)

	return r
}
