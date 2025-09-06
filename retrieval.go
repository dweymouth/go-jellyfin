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

type TranscodeOptions struct {
	// Audio codec to request, e.g. "mp3"
	AudioCodec string

	// Requested audio bit rate, e.g. 192000
	// If 0, use encoder default.
	AudioBitRate uint32

	// Requested container for the transcoding.
	// Required when requesting transcoding.
	Container string
}

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

func (c *Client) GetStreamURL(id string, transcodeOptions *TranscodeOptions) (string, error) {
	path := fmt.Sprintf("/audio/%s/stream", id)
	params := c.defaultParams()
	params["playSessionId"] = randomKey(32)
	params["api_key"] = c.token
	if transcodeOptions != nil {
		params["container"] = transcodeOptions.Container
		params["audioCodec"] = transcodeOptions.AudioCodec
		if br := transcodeOptions.AudioBitRate; br > 0 {
			params["audioBitRate"] = strconv.Itoa(int(br))
		}
	} else {
		params["static"] = "true"
	}
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
