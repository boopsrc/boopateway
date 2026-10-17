package booprecognition

import "context"

const personsPath = "/persons"

// RegisterPerson chama POST /persons. Se a pessoa já existir no Boop, Created vem false.
func (c *Client) RegisterPerson(ctx context.Context, img []byte, contentType string) (*personResponse, error) {
	var out personResponse
	if err := c.postImage(ctx, personsPath, img, contentType, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
