package jellyfin

import (
	"encoding/json"
	"errors"
	"fmt"
)

type createPlaylistBody struct {
	Name      string   `json:"Name"`
	Ids       []string `json:"Ids,omitempty"`
	UserID    string   `json:"UserId"`
	MediaType string   `json:"MediaType"`
}

func (c *Client) CreatePlaylist(name string, trackIDs []string) error {
	body := createPlaylistBody{
		Name:      name,
		UserID:    c.userID,
		MediaType: "Audio",
		Ids:       trackIDs,
	}
	resp, err := c.post("/Playlists", c.defaultParams(), body)
	if err != nil {
		return fmt.Errorf("create playlist: %v", err)
	}
	resp.Close()
	return nil
}

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

func (c *Client) AddPlaylistTracks(playlistID string, trackIDs []string) error {
	return errors.New("not implemented")
}

func (c *Client) DeletePlaylist(playlistID string) error {
	resp, err := c.delete(fmt.Sprintf("/Items/%s", playlistID), c.defaultParams())
	if err != nil {
		return fmt.Errorf("delete playlist: %v", err)
	}
	defer resp.Close()
	return nil
}
