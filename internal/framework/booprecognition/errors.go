package booprecognition

import (
	"encoding/json"
	"io"
)

// ImageRejectedError: o Boop recusou a foto (400/422), ex.: "Nenhum rosto detectado na imagem".
// É problema da entrada do cliente, não indisponibilidade do parceiro.
type ImageRejectedError struct {
	Detail string
}

func (e *ImageRejectedError) Error() string {
	return "booprecognition: imagem recusada: " + e.Detail
}

// readDetail extrai o campo "detail" do erro do Boop.
func readDetail(body io.Reader) string {
	raw, _ := io.ReadAll(io.LimitReader(body, 2048))
	var payload struct {
		Detail any `json:"detail"`
	}
	if json.Unmarshal(raw, &payload) == nil {
		if s, ok := payload.Detail.(string); ok && s != "" {
			return s
		}
	}
	return "imagem inválida"
}
