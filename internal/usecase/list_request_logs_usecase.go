package usecase

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/boopateway/boopateway/internal/domain/entity"
	"github.com/boopateway/boopateway/internal/domain/repository"
)

type ListRequestLogsInput struct {
	Token  string
	Filter repository.RequestLogFilter
	Page   int
	Limit  int
}

type ListRequestLogsOutput struct {
	Items []repository.RequestLog
	Total int
	Page  int
	Limit int
}

// ListRequestLogsUseCase lista as chamadas feitas à API, das mais recentes para as mais antigas.
type ListRequestLogsUseCase struct {
	validate *ValidateAdminTokenUseCase
	logs     repository.RequestLogRepository
}

func NewListRequestLogsUseCase(validate *ValidateAdminTokenUseCase, logs repository.RequestLogRepository) *ListRequestLogsUseCase {
	return &ListRequestLogsUseCase{validate: validate, logs: logs}
}

var validMethods = map[string]bool{
	http.MethodGet: true, http.MethodPost: true, http.MethodPut: true,
	http.MethodPatch: true, http.MethodDelete: true, http.MethodHead: true, http.MethodOptions: true,
}

func (uc *ListRequestLogsUseCase) Execute(ctx context.Context, in ListRequestLogsInput) (*ListRequestLogsOutput, error) {
	if _, err := uc.validate.Execute(ctx, in.Token); err != nil {
		return nil, err
	}
	f := in.Filter
	f.Method = strings.ToUpper(strings.TrimSpace(f.Method))
	f.Path = strings.TrimSpace(f.Path)
	if f.Method != "" && !validMethods[f.Method] {
		return nil, entity.ValidationError("method inválido")
	}
	if f.StatusClass != 0 && (f.StatusClass < 1 || f.StatusClass > 5) {
		return nil, entity.ValidationError("status deve ser 1xx, 2xx, 3xx, 4xx ou 5xx")
	}
	var err error
	if f.From, err = normalizeTime(f.From, "from"); err != nil {
		return nil, err
	}
	if f.To, err = normalizeTime(f.To, "to"); err != nil {
		return nil, err
	}

	page := max(in.Page, 1)
	limit := in.Limit
	if limit <= 0 {
		limit = DefaultPageSize
	}
	limit = min(limit, MaxPageSize)

	items, total, err := uc.logs.Search(ctx, f, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	return &ListRequestLogsOutput{Items: items, Total: total, Page: page, Limit: limit}, nil
}

// normalizeTime converte para o formato gravado no banco (UTC, milissegundos) para comparar como texto.
func normalizeTime(v, field string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return "", entity.ValidationError(field + " deve estar no formato RFC 3339 (ex.: 2026-10-17T10:00:00Z)")
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z"), nil
}
