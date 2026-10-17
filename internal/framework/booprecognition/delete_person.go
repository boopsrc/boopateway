package booprecognition

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

var errForbidden = fmt.Errorf("booprecognition: token sem permissão")

// DeletePerson chama DELETE /persons/{id} (scope persons:delete). 404 = já não existe no Boop.
func (c *Client) DeletePerson(ctx context.Context, token string, personID int64) error {
	resp, err := c.do(ctx, http.MethodDelete, personsPath+"/"+strconv.FormatInt(personID, 10), token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode < 300, resp.StatusCode == http.StatusNotFound:
		return nil
	case resp.StatusCode == http.StatusUnauthorized:
		return errUnauthorized
	case resp.StatusCode == http.StatusForbidden:
		return errForbidden
	default:
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("booprecognition: delete status %d: %s", resp.StatusCode, msg)
	}
}
