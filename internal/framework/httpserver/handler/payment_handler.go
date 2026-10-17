package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/usecase"
)

type PaymentHandler struct {
	createPayment  *usecase.CreatePaymentUseCase
	getPayment     *usecase.GetPaymentUseCase
	confirmPayment *usecase.ConfirmPaymentUseCase
}

func NewPaymentHandler(
	create *usecase.CreatePaymentUseCase,
	get *usecase.GetPaymentUseCase,
	confirm *usecase.ConfirmPaymentUseCase,
) *PaymentHandler {
	return &PaymentHandler{createPayment: create, getPayment: get, confirmPayment: confirm}
}

type paymentResponse struct {
	TransactionID string                   `json:"transaction_id"`
	AmountCents   int64                    `json:"amount_cents"`
	Status        entity.TransactionStatus `json:"status"`
	PersonName    string                   `json:"person_name,omitempty"`
}

func toPaymentResponse(t *entity.Transaction, p *entity.Person) paymentResponse {
	resp := paymentResponse{TransactionID: t.ID, AmountCents: t.AmountCents, Status: t.Status}
	if p != nil {
		resp.PersonName = p.Name
	}
	return resp
}

// Create POST /payments (multipart: amount_cents, image)
func (h *PaymentHandler) Create(c *gin.Context) {
	amount, err := strconv.ParseInt(c.PostForm("amount_cents"), 10, 64)
	if err != nil {
		respondError(c, entity.ValidationError("amount_cents deve ser um inteiro positivo (em centavos)"))
		return
	}
	img, err := readFormImage(c)
	if err != nil {
		respondError(c, err)
		return
	}
	out, err := h.createPayment.Execute(c.Request.Context(), usecase.CreatePaymentInput{AmountCents: amount, Image: img})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toPaymentResponse(out.Transaction, out.Person))
}

// Get GET /payments/:id
func (h *PaymentHandler) Get(c *gin.Context) {
	t, err := h.getPayment.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPaymentResponse(t, nil))
}

type confirmPaymentRequest struct {
	Code string `json:"code"`
}

// Confirm POST /payments/:id/confirm (JSON: {"code":"1234"})
func (h *PaymentHandler) Confirm(c *gin.Context) {
	var req confirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, entity.ValidationError(`envie {"code": "<4 dígitos>"}`))
		return
	}
	out, err := h.confirmPayment.Execute(c.Request.Context(), usecase.ConfirmPaymentInput{
		TransactionID: c.Param("id"),
		Code:          req.Code,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPaymentResponse(out.Transaction, out.Person))
}
