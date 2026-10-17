package booprecognition

import (
	"context"
	"errors"
)

// SearchPerson identifica a pessoa pela foto usando o endpoint configurado em SearchPath
// (padrão POST /identify, que não cadastra rostos desconhecidos).
// O /identify devolve {"identification", "precision", "history"} ou null quando não reconhece;
// null e 404 são tratados como "não encontrada" (found=false).
func (c *Client) SearchPerson(ctx context.Context, img []byte, contentType string) (resp *personResponse, found bool, err error) {
	var out *personResponse
	err = c.postImage(ctx, c.searchPath, img, contentType, &out)
	if errors.Is(err, errNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if out == nil {
		return nil, false, nil
	}
	return out, true, nil
}
