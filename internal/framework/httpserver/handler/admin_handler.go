package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/repository"
	"github.com/boopateway/boopateway/internal/framework/httpserver/middleware"
	"github.com/boopateway/boopateway/internal/usecase"
)

type AdminHandler struct {
	validate     *usecase.ValidateAdminTokenUseCase
	listPersons  *usecase.ListPersonsUseCase
	personImage  *usecase.GetPersonImageUseCase
	deletePerson *usecase.DeletePersonUseCase
	transactions *usecase.ListPersonTransactionsUseCase
	search       *usecase.SearchTransactionsUseCase
	logs         *usecase.ListRequestLogsUseCase
}

func NewAdminHandler(
	validate *usecase.ValidateAdminTokenUseCase,
	list *usecase.ListPersonsUseCase,
	image *usecase.GetPersonImageUseCase,
	del *usecase.DeletePersonUseCase,
	transactions *usecase.ListPersonTransactionsUseCase,
	search *usecase.SearchTransactionsUseCase,
	logs *usecase.ListRequestLogsUseCase,
) *AdminHandler {
	return &AdminHandler{validate: validate, listPersons: list, personImage: image, deletePerson: del, transactions: transactions, search: search, logs: logs}
}

type adminPersonResponse struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	PartnerPersonID int64  `json:"partner_person_id"`
	CreatedAt       string `json:"created_at"`
}

// Session GET /admin/session
func (h *AdminHandler) Session(c *gin.Context) {
	s, err := h.validate.Execute(c.Request.Context(), middleware.AdminTokenFrom(c))
	if err != nil {
		respondError(c, err)
		return
	}
	scopes := s.Scopes
	if scopes == nil {
		scopes = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"scopes": scopes, "can_delete": s.CanDelete})
}

// ListPersons GET /admin/persons?page=1&limit=20
func (h *AdminHandler) ListPersons(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	out, err := h.listPersons.Execute(c.Request.Context(), usecase.ListPersonsInput{
		Token: middleware.AdminTokenFrom(c), Page: page, Limit: limit,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	items := make([]adminPersonResponse, 0, len(out.Items))
	for _, p := range out.Items {
		items = append(items, adminPersonResponse{
			ID: p.ID, Name: p.Name, Email: p.Email, PartnerPersonID: p.PartnerPersonID, CreatedAt: p.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": out.Total, "page": out.Page, "limit": out.Limit})
}

// PersonImage GET /admin/persons/:id/image
func (h *AdminHandler) PersonImage(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	path, err := h.personImage.Execute(c.Request.Context(), middleware.AdminTokenFrom(c), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.Header("Cache-Control", "private, max-age=300")
	c.File(path)
}

// DeletePerson DELETE /admin/persons/:id
func (h *AdminHandler) DeletePerson(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.deletePerson.Execute(c.Request.Context(), middleware.AdminTokenFrom(c), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, entity.ValidationError("id inválido"))
		return 0, false
	}
	return id, true
}

type adminTransactionResponse struct {
	TransactionID string                   `json:"transaction_id"`
	AmountCents   int64                    `json:"amount_cents"`
	Status        entity.TransactionStatus `json:"status"`
	Attempts      int                      `json:"attempts"`
	CreatedAt     string                   `json:"created_at"`
	UpdatedAt     string                   `json:"updated_at"`
}

// PersonTransactions GET /admin/persons/:id/transactions?page=1&limit=20
func (h *AdminHandler) PersonTransactions(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	out, err := h.transactions.Execute(c.Request.Context(), usecase.ListPersonTransactionsInput{
		Token: middleware.AdminTokenFrom(c), PersonID: id, Page: page, Limit: limit,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	items := make([]adminTransactionResponse, 0, len(out.Items))
	for _, t := range out.Items {
		items = append(items, adminTransactionResponse{
			TransactionID: t.ID, AmountCents: t.AmountCents, Status: t.Status,
			Attempts: t.Attempts, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": out.Total, "page": out.Page, "limit": out.Limit})
}

type adminTransactionSearchResponse struct {
	adminTransactionResponse
	PersonID        *int64  `json:"person_id"`
	PersonName      *string `json:"person_name"`
	PersonEmail     *string `json:"person_email"`
	PartnerPersonID *int64  `json:"partner_person_id"`
}

// SearchTransactions GET /admin/transactions?name=&email=&amount_cents=&status=&page=&limit=
func (h *AdminHandler) SearchTransactions(c *gin.Context) {
	filter := repository.TransactionFilter{
		Name:   c.Query("name"),
		Email:  c.Query("email"),
		Status: entity.TransactionStatus(c.Query("status")),
	}
	if v := c.Query("amount_cents"); v != "" {
		amount, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			respondError(c, entity.ValidationError("amount_cents deve ser um inteiro positivo (em centavos)"))
			return
		}
		filter.AmountCents = &amount
	}
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	out, err := h.search.Execute(c.Request.Context(), usecase.SearchTransactionsInput{
		Token: middleware.AdminTokenFrom(c), Filter: filter, Page: page, Limit: limit,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	items := make([]adminTransactionSearchResponse, 0, len(out.Items))
	for _, t := range out.Items {
		items = append(items, adminTransactionSearchResponse{
			adminTransactionResponse: adminTransactionResponse{
				TransactionID: t.ID, AmountCents: t.AmountCents, Status: t.Status,
				Attempts: t.Attempts, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
			},
			PersonID: t.PersonID, PersonName: t.PersonName, PersonEmail: t.PersonEmail, PartnerPersonID: t.PartnerPersonID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": out.Total, "page": out.Page, "limit": out.Limit})
}

type adminRequestLogResponse struct {
	ID         int64  `json:"id"`
	CreatedAt  string `json:"created_at"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Route      string `json:"route"`
	Query      string `json:"query"`
	Status     int    `json:"status"`
	DurationMs int64  `json:"duration_ms"`
	ClientIP   string `json:"client_ip"`
	UserAgent  string `json:"user_agent"`
	Error      string `json:"error"`
}

// RequestLogs GET /admin/logs?method=&path=&status=4xx&from=&to=&page=&limit=
func (h *AdminHandler) RequestLogs(c *gin.Context) {
	filter := repository.RequestLogFilter{
		Method: c.Query("method"),
		Path:   c.Query("path"),
		From:   c.Query("from"),
		To:     c.Query("to"),
	}
	if v := c.Query("status"); v != "" {
		class, err := strconv.Atoi(strings.TrimSuffix(strings.ToLower(v), "xx"))
		if err != nil {
			respondError(c, entity.ValidationError("status deve ser 1xx, 2xx, 3xx, 4xx ou 5xx"))
			return
		}
		filter.StatusClass = class
	}
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	out, err := h.logs.Execute(c.Request.Context(), usecase.ListRequestLogsInput{
		Token: middleware.AdminTokenFrom(c), Filter: filter, Page: page, Limit: limit,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	items := make([]adminRequestLogResponse, 0, len(out.Items))
	for _, l := range out.Items {
		items = append(items, adminRequestLogResponse{
			ID: l.ID, CreatedAt: l.CreatedAt, Method: l.Method, Path: l.Path, Route: l.Route, Query: l.Query,
			Status: l.Status, DurationMs: l.DurationMs, ClientIP: l.ClientIP, UserAgent: l.UserAgent, Error: l.Error,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": out.Total, "page": out.Page, "limit": out.Limit})
}
