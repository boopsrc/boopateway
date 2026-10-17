package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/boopateway/boopateway/internal/domain/repository"
)

// maxErrorBody limita quanto do corpo de uma resposta de erro é lido para extrair a mensagem.
const maxErrorBody = 4 << 10

// skipLogPaths não são registradas: health checks e a própria consulta de logs geram só ruído.
var skipLogPaths = map[string]bool{"/healthz": true, "/docs": true, "/admin/logs": true}

// RequestLog grava cada chamada à API (método, rota, status, duração, IP e mensagem de erro).
// Corpos de requisição e headers não são gravados: contêm imagens, códigos e tokens.
func RequestLog(logs repository.RequestLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipLogPaths[c.Request.URL.Path] {
			c.Next()
			return
		}
		start := time.Now()
		w := &errorCaptureWriter{ResponseWriter: c.Writer}
		c.Writer = w
		c.Next()

		entry := &repository.RequestLog{
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			Route:      c.FullPath(),
			Query:      c.Request.URL.RawQuery,
			Status:     w.Status(),
			DurationMs: time.Since(start).Milliseconds(),
			ClientIP:   c.ClientIP(),
			UserAgent:  truncate(c.Request.UserAgent(), 256),
		}
		if entry.Status >= 400 {
			entry.Error = errorMessage(w.body.Bytes())
		}
		// O 404 de rota inexistente é escrito pelo gin sem passar pelo writer acima.
		if entry.Route == "" && entry.Error == "" {
			entry.Error = "rota não encontrada"
		}
		// O contexto da requisição pode já estar cancelado; o log é gravado mesmo assim.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), 2*time.Second)
		defer cancel()
		if err := logs.Insert(ctx, entry); err != nil {
			slog.Error("request log", "err", err)
		}
	}
}

// errorCaptureWriter guarda o início do corpo das respostas de erro.
type errorCaptureWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *errorCaptureWriter) Write(b []byte) (int, error) {
	if w.Status() >= 400 && w.body.Len() < maxErrorBody {
		w.body.Write(b[:min(len(b), maxErrorBody-w.body.Len())])
	}
	return w.ResponseWriter.Write(b)
}

func (w *errorCaptureWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func errorMessage(body []byte) string {
	var resp struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &resp) == nil && resp.Error != "" {
		return truncate(resp.Error, 500)
	}
	return truncate(strings.TrimSpace(string(body)), 500)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
