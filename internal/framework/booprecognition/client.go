// Package booprecognition contém toda a integração com a API de reconhecimento facial Boop.
//
//	client.go                    -> HTTP, autenticação e envio multipart
//	dto.go                       -> contratos (JSON) da API Boop
//	register_person.go           -> POST /persons
//	search_person.go             -> POST /identify (busca de pessoa pela foto)
//	face_recognition_gateway.go  -> adapter que implementa gateway.FaceRecognitionGateway
package booprecognition

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

var errNotFound = errors.New("booprecognition: pessoa não encontrada")

type Config struct {
	BaseURL    string
	Token      string
	SearchPath string // padrão /identify
	Timeout    time.Duration
}

type Client struct {
	baseURL    string
	token      string
	searchPath string
	http       *http.Client
}

func NewClient(cfg Config) *Client {
	if cfg.SearchPath == "" {
		cfg.SearchPath = identifyPath
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 20 * time.Second
	}
	return &Client{
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		token:      cfg.Token,
		searchPath: cfg.SearchPath,
		http:       &http.Client{Timeout: cfg.Timeout},
	}
}

// do executa uma requisição autenticada com o token informado (operações administrativas).
func (c *Client) do(ctx context.Context, method, path, token string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("booprecognition: %w", err)
	}
	return resp, nil
}

// postImage envia a foto como multipart (campo "image") e decodifica o JSON de resposta em out.
func (c *Client) postImage(ctx context.Context, path string, img []byte, contentType string, out any) error {
	body, formContentType, err := buildImageForm(img, contentType)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", formContentType)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("booprecognition: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return errNotFound
	}
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity {
		return &ImageRejectedError{Detail: readDetail(resp.Body)}
	}
	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("booprecognition: status %d: %s", resp.StatusCode, msg)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("booprecognition: resposta inválida: %w", err)
	}
	return nil
}

func buildImageForm(img []byte, contentType string) (*bytes.Buffer, string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	filename := "foto.jpg"
	if contentType == "image/png" {
		filename = "foto.png"
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, filename))
	h.Set("Content-Type", contentType)

	part, err := mw.CreatePart(h)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(img); err != nil {
		return nil, "", err
	}
	if err := mw.Close(); err != nil {
		return nil, "", err
	}
	return &buf, mw.FormDataContentType(), nil
}
