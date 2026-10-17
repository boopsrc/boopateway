package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const adminTokenKey = "admin_token"

// AdminToken exige "Authorization: Bearer <token>". A validade do token é conferida
// no serviço de reconhecimento facial pelos use cases.
func AdminToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		token = strings.TrimSpace(token)
		if !ok || token == "" {
			abort(c, http.StatusUnauthorized, "informe o token no header Authorization: Bearer <token>")
			return
		}
		c.Set(adminTokenKey, token)
		c.Next()
	}
}

func AdminTokenFrom(c *gin.Context) string {
	return c.GetString(adminTokenKey)
}
