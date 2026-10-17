package repository

import "context"

// RequestLog é o registro de uma chamada feita à API.
type RequestLog struct {
	ID         int64
	CreatedAt  string
	Method     string
	Path       string // rota com os valores reais (ex.: /persons/12)
	Route      string // padrão da rota (ex.: /persons/:id); vazio se não houve rota
	Query      string
	Status     int
	DurationMs int64
	ClientIP   string
	UserAgent  string
	Error      string // mensagem de erro da resposta, quando status >= 400
}

// RequestLogFilter filtra a busca de logs. Campos vazios/zero são ignorados.
type RequestLogFilter struct {
	Method      string
	Path        string // contém
	StatusClass int    // 2, 3, 4 ou 5 (2xx, 3xx...)
	From, To    string // RFC 3339
}

type RequestLogRepository interface {
	Insert(ctx context.Context, l *RequestLog) error
	Search(ctx context.Context, f RequestLogFilter, limit, offset int) ([]RequestLog, int, error)
}
