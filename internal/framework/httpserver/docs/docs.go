// Package docs serve a página de referência da API.
package docs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed index.html
var page []byte

func Handler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", page)
}
