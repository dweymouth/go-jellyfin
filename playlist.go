package jellyfin

import (
	"fmt"
	"strings"
)

type createPlaylistBody struct {
	Name      string   `json:"Name"`
	Overview  string   `json:"Overview,omitempty"`
	IsPublic  bool     `json:"IsPublic,omitempty"`
	Ids       []string `json:"Ids,omitempty"`
	UserID    string   `json:"UserId"`
	MediaType string   `json:"MediaType"`
}

func (c *Client) CreatePlaylist(name, description string, public bool, trackIDs []string) error {
	body := createPlaylistBody{
		Name:      name,
		Overview:  description,
		IsPublic:  public,
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

	return c.parseSongs(resp)
}

type updatePlaylistBody struct {
	Name         string            `json:"Name"`
	Overview     string            `json:"Overview"`
	IsPublic     bool              `json:"IsPublic"`
	DateCreated  string            `json:"DateCreated"`
	Genres       []string          `json:"Genres"`
	PremiereDate string            `json:"PremiereDate"`
	ProviderIds  map[string]string `json:"ProviderIds"`
	Tags         []string          `json:"Tags"`
}

func (c *Client) UpdatePlaylistMetadata(playlistID, name, overview string, public bool) error {
	pl, err := c.GetPlaylist(playlistID)
	if err != nil {
		return err
	}

	params := c.defaultParams()
	body := updatePlaylistBody{
		Name:         name,
		Overview:     overview,
		IsPublic:     public,
		DateCreated:  pl.DateCreated,  // Required
		Genres:       pl.Genres,       // Required
		PremiereDate: pl.PremiereDate, // Required
		Tags:         pl.Tags,         // Required
		ProviderIds:  pl.ProviderIds,  // Required
	}
	resp, err := c.post(fmt.Sprintf("/Items/%s", playlistID), params, body)
	if err != nil {
		return fmt.Errorf("update playlist metadata: %v", err)
	}
	resp.Close()
	return nil
}

func (c *Client) AddSongsToPlaylist(playlistID string, trackIDs []string) error {
	params := c.defaultParams()
	params["ids"] = strings.Join(trackIDs, ",")
	resp, err := c.post(fmt.Sprintf("/Playlists/%s/Items", playlistID), params, struct{}{})
	if err != nil {
		return fmt.Errorf("add songs to playlist: %v", err)
	}
	resp.Close()
	return nil
}

func (c *Client) RemoveSongsFromPlaylist(playlistID string, removeIndexes []int) error {
	songs, err := c.GetPlaylistSongs(playlistID)
	if err != nil {
		return err
	}
	removeItemIds := make([]string, 0, len(removeIndexes))
	for _, idx := range removeIndexes {
		if idx < len(songs) {
			removeItemIds = append(removeItemIds, songs[idx].PlaylistItemId)
		}
	}

	params := c.defaultParams()
	params["entryIds"] = strings.Join(removeItemIds, ",")
	resp, err := c.delete(fmt.Sprintf("/Playlists/%s/Items", playlistID), params)
	if err != nil {
		return fmt.Errorf("remove songs from playlist: %v", err)
	}
	resp.Close()
	return nil
}

func (c *Client) MovePlaylistSong(playlistID string, trackID string, newIdx int) error {
	endpoint := fmt.Sprintf("/Playlists/%s/Items/%s/Move/%d", playlistID, trackID, newIdx)
	resp, err := c.post(endpoint, c.defaultParams(), struct{}{})
	if err != nil {
		return fmt.Errorf("move playlist song: %v", err)
	}
	resp.Close()
	return nil

}

func (c *Client) DeletePlaylist(playlistID string) error {
	resp, err := c.delete(fmt.Sprintf("/Items/%s", playlistID), c.defaultParams())
	if err != nil {
		return fmt.Errorf("delete playlist: %v", err)
	}
	defer resp.Close()
	return nil
}
