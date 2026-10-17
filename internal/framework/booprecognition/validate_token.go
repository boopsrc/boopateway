package booprecognition

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const identifyPath = "/identify"

var errUnauthorized = fmt.Errorf("booprecognition: token inválido")

// ValidateToken confere o token no Boop sem efeito colateral: POST /identify sem body.
// O Boop autentica antes de validar o body, então 401 = token inválido e 422 (body
// faltando) ou 403 (sem o scope identify) = token válido.
func (c *Client) ValidateToken(ctx context.Context, token string) (scopes []string, err error) {
	resp, err := c.do(ctx, http.MethodPost, identifyPath, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, errUnauthorized
	case http.StatusUnprocessableEntity, http.StatusForbidden, http.StatusOK:
		return tokenScopes(token), nil
	default:
		return nil, fmt.Errorf("booprecognition: validação do token retornou status %d", resp.StatusCode)
	}
}

// tokenScopes lê o claim "scope" do JWT. Só decodifica (o Boop já validou a assinatura);
// serve para a interface saber o que exibir. A permissão real é aplicada pelo Boop.
func tokenScopes(token string) []string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}
	var claims struct {
		Scope string `json:"scope"`
	}
	if json.Unmarshal(raw, &claims) != nil {
		return nil
	}
	return strings.Fields(claims.Scope)
}
