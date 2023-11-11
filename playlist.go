package jellyfin

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetPlaylistSongs(playlistID string) ([]*Song, error) {
	params := c.defaultParams()
	params.setIncludeFields("Genres", "DateCreated", "MediaSources", "UserData", "ParentId")

	resp, err := c.get(fmt.Sprintf("/Playlists/%s/Items", playlistID), params)
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
