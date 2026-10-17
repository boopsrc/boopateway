// Package portal serve o portal web administrativo e repassa /api/* para a API do Boopateway.
package portal

import (
	_ "embed"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed web/index.html
var page []byte

// NewRouter cria o servidor do portal. apiURL é a API do Boopateway (ex.: http://api:8080).
func NewRouter(apiURL *url.URL) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(apiURL)
			pr.Out.URL.Path = strings.TrimPrefix(pr.In.URL.Path, "/api")
			pr.Out.URL.RawPath = ""
		},
	}

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.Any("/api/*path", gin.WrapH(proxy))
	r.GET("/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Data(http.StatusOK, "text/html; charset=utf-8", page)
	})
	return r
}
