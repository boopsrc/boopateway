package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/usecase"
)

type PersonHandler struct {
	createPerson *usecase.CreatePersonUseCase
	getPerson    *usecase.GetPersonUseCase
}

func NewPersonHandler(create *usecase.CreatePersonUseCase, get *usecase.GetPersonUseCase) *PersonHandler {
	return &PersonHandler{createPerson: create, getPerson: get}
}

// personResponse expõe apenas dados não sensíveis.
type personResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func toPersonResponse(p *entity.Person) personResponse {
	return personResponse{ID: p.ID, Name: p.Name, Email: p.MaskedEmail(), CreatedAt: p.CreatedAt}
}

// Create POST /persons (multipart: name, email, code?, image)
func (h *PersonHandler) Create(c *gin.Context) {
	img, err := readFormImage(c)
	if err != nil {
		respondError(c, err)
		return
	}
	p, err := h.createPerson.Execute(c.Request.Context(), usecase.CreatePersonInput{
		Name:  c.PostForm("name"),
		Email: c.PostForm("email"),
		Code:  c.PostForm("code"),
		Image: img,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toPersonResponse(p))
}

// Get GET /persons/:id
func (h *PersonHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, entity.ValidationError("id inválido"))
		return
	}
	p, err := h.getPerson.Execute(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPersonResponse(p))
}
