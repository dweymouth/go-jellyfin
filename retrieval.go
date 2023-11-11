package jellyfin

import (
	"fmt"
	"io"
	"strconv"
)

func (c *Client) GetItemImage(itemID, imageTag string, size, quality int) (io.ReadCloser, error) {
	path := fmt.Sprintf("/Items/%s/Images/%s", itemID, imageTag)
	params := c.defaultParams()
	params["width"] = strconv.Itoa(size)
	params["quality"] = strconv.Itoa(quality)
	return c.get(path, params)
}

func (c *Client) GetStreamURL(id string) (string, error) {
	path := fmt.Sprintf("/audio/%s/stream", id)
	params := c.defaultParams()
	params["playSessionId"] = c.deviceID
	params["static"] = "true"
	params["api_key"] = c.token
	return c.encodeGETUrl(path, params)
}
