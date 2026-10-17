package booprecognition

import (
	"context"
	"errors"
)

// SearchPerson identifica a pessoa pela foto usando o endpoint configurado em SearchPath.
// Um 404 do Boop é tratado como "não encontrada" (found=false).
func (c *Client) SearchPerson(ctx context.Context, img []byte, contentType string) (resp *personResponse, found bool, err error) {
	var out personResponse
	err = c.postImage(ctx, c.searchPath, img, contentType, &out)
	if errors.Is(err, errNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &out, true, nil
}
