package jellyfin

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetPlaylistSongs(playlistID string) ([]*Song, error) {
	params := c.defaultParams()
	params.setParentID(playlistID)

	resp, err := c.get(fmt.Sprintf("/Users/%s/Items", c.userID), params)
	if err != nil {
		return nil, fmt.Errorf("get playlist songs: %v", err)
	}
	defer resp.Close()

	dto := songs{}
	err = json.NewDecoder(resp).Decode(&dto)
	if err != nil {
		return nil, fmt.Errorf("decode playlist songs: %v", err)
	}

	return c.parseSongs(resp)
}
