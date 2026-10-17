package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/boopateway/boopateway/internal/domain/repository"
)

const (
	IdempotencyHeader = "Idempotency-Key"
	MaxBodyBytes      = 10 << 20 // 10 MiB
)

// Idempotency garante que a mesma Idempotency-Key + rota produza um único efeito.
//   - primeira chamada: processa e grava a resposta (2xx/4xx);
//   - repetição com mesmo payload: devolve a resposta gravada (Idempotent-Replayed: true);
//   - repetição com payload diferente: 422;
//   - repetição enquanto a primeira ainda processa: 409;
//   - respostas 5xx liberam a chave para nova tentativa.
func Idempotency(repo repository.IdempotencyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader(IdempotencyHeader)
		if key == "" || len(key) > 255 {
			abort(c, http.StatusBadRequest, "header Idempotency-Key é obrigatório (máx. 255 caracteres)")
			return
		}

		body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes))
		if err != nil {
			abort(c, http.StatusRequestEntityTooLarge, "requisição muito grande")
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		hash, err := requestHash(c.Request, body)
		if err != nil {
			abort(c, http.StatusBadRequest, "corpo da requisição inválido")
			return
		}

		ctx := c.Request.Context()
		scope := c.Request.Method + " " + c.Request.URL.Path
		rec, started, err := repo.Begin(ctx, key, scope, hash)
		if err != nil {
			slog.Error("idempotency begin", "err", err)
			abort(c, http.StatusInternalServerError, "erro interno")
			return
		}
		if !started {
			switch {
			case rec.RequestHash != hash:
				abort(c, http.StatusUnprocessableEntity, "Idempotency-Key já usada com outro payload")
			case rec.StatusCode == 0:
				abort(c, http.StatusConflict, "requisição com esta Idempotency-Key ainda em processamento")
			default:
				c.Header("Idempotent-Replayed", "true")
				c.Data(rec.StatusCode, "application/json; charset=utf-8", rec.Response)
				c.Abort()
			}
			return
		}

		rw := &bodyRecorder{ResponseWriter: c.Writer}
		c.Writer = rw
		c.Next()

		ctx = context.WithoutCancel(ctx)
		if status := rw.Status(); status >= 500 {
			err = repo.Release(ctx, key, scope)
		} else {
			err = repo.Complete(ctx, key, scope, status, rw.buf.Bytes())
		}
		if err != nil {
			slog.Error("idempotency finish", "err", err)
		}
	}
}

func abort(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"error": msg})
}

// requestHash calcula um hash estável do payload. Em multipart o boundary muda a cada
// envio, então o hash é feito sobre os campos (nome + conteúdo), não sobre os bytes brutos.
func requestHash(r *http.Request, body []byte) (string, error) {
	h := sha256.New()
	h.Write([]byte(r.Method + " " + r.URL.Path + "\n"))

	mediaType, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaType != "multipart/form-data" {
		h.Write(body)
		return hex.EncodeToString(h.Sum(nil)), nil
	}

	type field struct{ name, sum string }
	var fields []field
	mr := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	for {
		p, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		ph := sha256.New()
		if _, err := io.Copy(ph, p); err != nil {
			return "", err
		}
		fields = append(fields, field{p.FormName(), hex.EncodeToString(ph.Sum(nil))})
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].name < fields[j].name })
	for _, f := range fields {
		h.Write([]byte(f.name + "=" + f.sum + "\n"))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

type bodyRecorder struct {
	gin.ResponseWriter
	buf bytes.Buffer
}

func (w *bodyRecorder) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyRecorder) WriteString(s string) (int, error) {
	w.buf.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
