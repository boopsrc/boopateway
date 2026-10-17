package handler

import (
	"io"

	"github.com/gin-gonic/gin"

	"github.com/boopateway/boopateway/internal/domain/entity"
)

// readFormImage lê o campo multipart "image". Validação de tipo/tamanho fica no domínio.
func readFormImage(c *gin.Context) ([]byte, error) {
	fh, err := c.FormFile("image")
	if err != nil {
		return nil, entity.ValidationError("campo image é obrigatório")
	}
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, entity.MaxImageBytes+1))
}
