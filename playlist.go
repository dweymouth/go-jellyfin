package jellyfin

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetPlaylistSongs(playlistID string) ([]*Song, error) {
	params := c.defaultParams()
	params.setIncludeFields(songIncludeFields...)

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

func (c *Client) DeletePlaylist(playlistID string) error {
	resp, err := c.delete(fmt.Sprintf("/Items/%s", playlistID), c.defaultParams())
	if err != nil {
		return fmt.Errorf("delete playlist: %v", err)
	}
	defer resp.Close()
	return nil
}
