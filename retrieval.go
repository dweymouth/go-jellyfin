package jellyfin

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"strconv"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

func (c *Client) GetItemImageBinary(itemID, imageTag string, size, quality int) (io.ReadCloser, error) {
	path := fmt.Sprintf("/Items/%s/Images/%s", itemID, imageTag)
	params := c.defaultParams()
	params["width"] = strconv.Itoa(size)
	params["quality"] = strconv.Itoa(quality)
	return c.get(path, params)
}

func (c *Client) GetItemImage(itemID, imageTag string, size, quality int) (image.Image, error) {
	body, err := c.GetItemImageBinary(itemID, imageTag, size, quality)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	image, _, err := image.Decode(body)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (c *Client) GetStreamURL(id string) (string, error) {
	path := fmt.Sprintf("/audio/%s/stream", id)
	params := c.defaultParams()
	params["playSessionId"] = c.deviceID
	params["static"] = "true"
	params["api_key"] = c.token
	return c.encodeGETUrl(path, params)
}

func (c *Client) GetLyrics(itemID string) (*Lyrics, error) {
	path := fmt.Sprintf("/Audio/%s/Lyrics", itemID)
	resp, err := c.get(path, c.defaultParams())
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	lyrics := &Lyrics{}
	err = json.NewDecoder(resp).Decode(lyrics)
	if err != nil {
		return nil, fmt.Errorf("decode lyric json: %v", err)
	}

	return lyrics, nil
}
