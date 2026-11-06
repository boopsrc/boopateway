package booprecognition

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const oauthTokenPath = "/oauth/token"

// IssueToken gera um token novo no Boop via OAuth2 client_credentials (POST /oauth/token).
// scope vazio = scopes padrão do cliente.
func (c *Client) IssueToken(ctx context.Context, clientID, clientSecret, scope string) (string, error) {
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}
	if scope != "" {
		form.Set("scope", scope)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+oauthTokenPath, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("booprecognition: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("booprecognition: /oauth/token status %d: %s", resp.StatusCode, msg)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || out.AccessToken == "" {
		return "", fmt.Errorf("booprecognition: resposta de /oauth/token sem access_token")
	}
	return out.AccessToken, nil
}
