package jellyfin

import (
	"fmt"
	"io"
)

type setFavoriteBody struct{}

func (c *Client) SetFavorite(id string, favorite bool) error {
	endpoint := fmt.Sprintf("/Users/%s/FavoriteItems/%s", c.userID, id)
	var resp io.ReadCloser
	var err error
	if favorite {
		resp, err = c.post(endpoint, c.defaultParams(), setFavoriteBody{})
	} else {
		resp, err = c.delete(endpoint, c.defaultParams())
	}
	if err != nil {
		return err
	}
	resp.Close()
	return nil
}
