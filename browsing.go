package jellyfin

import (
	"encoding/json"
	"fmt"
	"io"
)

var (
	songIncludeFields     = []string{"Genres", "DateCreated", "MediaSources", "UserData", "ParentId"}
	albumIncludeFields    = []string{"Genres", "DateCreated", "ChildCount", "UserData", "ParentId"}
	playlistIncludeFields = []string{"Genres", "DateCreated", "MediaSources", "ChildCount", "Parent", "Overview"}
	artistIncludeFields   = []string{"ChildCount", "UserData"}
)

// GetUserViews returns top level collections that the
// logged-in user can access.
func (c *Client) GetUserViews() ([]*BaseItem, error) {
	params := c.defaultParams()
	resp, err := c.get(fmt.Sprintf("/Users/%s/Views", c.userID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	items := items{}
	err = json.NewDecoder(resp).Decode(&items)
	if err != nil {
		return nil, fmt.Errorf("decode json: %v", err)
	}
	return items.Items, nil
}

// GetAlbums returns albums with given sort, filter, and paging options.
// - Can be used to get an artist's discography with ArtistID filter.
func (c *Client) GetAlbums(opts QueryOpts) ([]*Album, error) {
	params := c.defaultParams()
	params.enableRecursive()
	params.setPaging(opts.Paging)
	params.setSorting(opts.Sort)
	params.setFilter(mediaTypeAlbum, opts.Filter)
	params.setIncludeTypes(mediaTypeAlbum)
	params.setIncludeFields(albumIncludeFields...)
	resp, err := c.get(fmt.Sprintf("/Users/%s/Items", c.userID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	albums := albums{}
	err = json.NewDecoder(resp).Decode(&albums)
	if err != nil {
		return nil, fmt.Errorf("decode json: %v", err)
	}
	return albums.Albums, nil
}

func (c *Client) GetAlbumArtists(opts QueryOpts) ([]*Artist, error) {
	params := c.defaultParams()
	params.enableRecursive()
	params.setFilter(mediaTypeArtist, opts.Filter)
	params.setPaging(opts.Paging)
	params.setSorting(opts.Sort)
	params.setIncludeTypes(mediaTypeAlbum)
	params.setIncludeFields(artistIncludeFields...)
	resp, err := c.get("/Artists/AlbumArtists", params)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	return c.parseArtists(resp)
}

func (c *Client) GetArtist(artistID string) (*Artist, error) {
	artist := &Artist{}
	includeFields := append(artistIncludeFields, "Overview")
	err := c.getItemByID(artistID, artist, includeFields...)
	if err != nil {
		return nil, err
	}
	return artist, nil
}

func (c *Client) GetAlbum(albumID string) (*Album, error) {
	album := &Album{}
	includeFields := append(albumIncludeFields, "Overview")
	err := c.getItemByID(albumID, album, includeFields...)
	if err != nil {
		return nil, err
	}
	return album, nil
}

func (c *Client) GetSong(songID string) (*Song, error) {
	song := &Song{}
	err := c.getItemByID(songID, song, songIncludeFields...)
	if err != nil {
		return nil, err
	}
	return song, nil
}

func (c *Client) GetSimilarArtists(artistID string) ([]*Artist, error) {
	params := c.defaultParams()
	params.enableRecursive()
	params.setIncludeTypes(mediaTypeArtist)
	params.setLimit(15)
	resp, err := c.get(fmt.Sprintf("/Items/%s/Similar", artistID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	return c.parseArtists(resp)
}

func (c *Client) GetGenres(paging Paging, parentID string) ([]NameID, error) {
	params := c.defaultParams()
	params.enableRecursive()
	params.setSorting(Sort{Field: SortByName, Mode: SortAsc})
	params.setPaging(paging)
	if parentID != "" {
		params.setFilter("Genre", Filter{ParentID: parentID})
	}

	resp, err := c.get("/MusicGenres", params)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	body := struct {
		Items []NameID
		Count int `json:"TotalRecordCount"`
	}{}

	err = json.NewDecoder(resp).Decode(&body)
	if err != nil {
		return nil, fmt.Errorf("decode json: %v", err)
	}

	return body.Items, nil
}

// Get songs matching the given filter criteria with given sorting and paging.
//   - Can be used to get an album track list with the ParentID filter.
//   - Can be used to get top songs for an artist with the ArtistId filter
//     and sorting by CommunityRating descending
func (c *Client) GetSongs(opts QueryOpts) ([]*Song, error) {
	params := c.defaultParams()
	params.setIncludeTypes(mediaTypeAudio)
	params.setPaging(opts.Paging)
	params.setSorting(opts.Sort)
	params.setFilter(mediaTypeAudio, opts.Filter)
	params.enableRecursive()
	params.setIncludeFields(songIncludeFields...)

	resp, err := c.get(fmt.Sprintf("/Users/%s/Items", c.userID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	return c.parseSongs(resp)
}

// GetPlaylists retrieves all playlists. Each playlists song count is known, but songs must be
// retrieved separately
func (c *Client) GetPlaylists() ([]*Playlist, error) {
	params := c.defaultParams()
	params.setIncludeTypes(mediaTypePlaylist)
	params.enableRecursive()
	params.setIncludeFields(playlistIncludeFields...)

	resp, err := c.get(fmt.Sprintf("/Users/%s/Items", c.userID), params)
	if err != nil {
		return nil, fmt.Errorf("get playlists: %v", err)
	}
	defer resp.Close()

	dto := playlists{}
	if err = json.NewDecoder(resp).Decode(&dto); err != nil {
		return nil, fmt.Errorf("parse playlists: %v", err)
	}

	// filter MediaTypes:
	//   - "Audio"   for music playlists created by Jellyfin UI
	//   - "Unknown" for .m3u files discovered in music libraries
	musicPlaylists := make([]*Playlist, 0)
	for _, pl := range dto.Playlists {
		if pl.MediaType == string(mediaTypeAudio) || pl.MediaType == string(mediaTypeUnknown) {
			musicPlaylists = append(musicPlaylists, pl)
		}
	}

	return musicPlaylists, nil
}

func (c *Client) GetPlaylist(playlistID string) (*Playlist, error) {
	playlist := &Playlist{}
	includeFields := append(playlistIncludeFields, "PremiereDate", "Tags", "ProviderIds")
	err := c.getItemByID(playlistID, playlist, includeFields...)
	if err != nil {
		return nil, err
	}
	return playlist, nil
}

func (c *Client) GetInstantMix(id string, idType ItemType, limit int) ([]*Song, error) {
	path := "/Items/%s/InstantMix"
	switch idType {
	case TypeArtist:
		path = "/Artists/%s/InstantMix"
	case TypeAlbum:
		path = "/Albums/%s/InstantMix"
	case TypeSong:
		path = "/Songs/%s/InstantMix"
	}

	params := c.defaultParams()
	params.setIncludeFields(songIncludeFields...)
	params.setLimit(limit)
	resp, err := c.get(fmt.Sprintf(path, id), params)
	if err != nil {
		return nil, fmt.Errorf("get instant mix: %v", err)
	}
	defer resp.Close()

	return c.parseSongs(resp)
}

func (c *Client) getItemByID(itemID string, dto interface{}, includeFields ...string) error {
	params := c.defaultParams()
	if len(includeFields) > 0 {
		params.setIncludeFields(includeFields...)
	}
	resp, err := c.get(fmt.Sprintf("/Users/%s/Items/%s", c.userID, itemID), params)
	if err != nil {
		return err
	}
	defer resp.Close()
	if err := json.NewDecoder(resp).Decode(dto); err != nil {
		return fmt.Errorf("parse item: %v", err)
	}
	return nil
}

func (c *Client) parseArtists(resp io.Reader) ([]*Artist, error) {
	artists := &artists{}
	if err := json.NewDecoder(resp).Decode(&artists); err != nil {
		return nil, fmt.Errorf("decode json: %v", err)
	}
	return artists.Artists, nil
}

func (c *Client) parseSongs(resp io.Reader) ([]*Song, error) {
	songs := songs{}
	if err := json.NewDecoder(resp).Decode(&songs); err != nil {
		return nil, fmt.Errorf("parse songs: %v", err)
	}
	return songs.Songs, nil
}
