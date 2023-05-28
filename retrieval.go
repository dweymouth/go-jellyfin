package jellyfin

import (
	"encoding/json"
	"fmt"
	"io"
)

// GetAlbums returns albums with given sort, filter, and paging options.
func (c *Client) GetAlbums(opts QueryOpts) ([]*Album, error) {
	params := c.defaultParams()
	params.enableRecursive()
	params.setPaging(opts.Paging)
	params.setSorting(opts.Sort)
	params.setFilter(mediaTypeAlbum, opts.Filter)
	params.setIncludeTypes(mediaTypeAlbum)
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
	resp, err := c.get("/Artists/AlbumArtists", params)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		return nil, err
	}
	return c.parseArtists(resp)
}

func (c *Client) GetSimilarArtists(artistID string) ([]*Artist, error) {
	params := c.defaultParams()
	params.enableRecursive()
	params.setSorting(Sort{Field: SortByName, Mode: SortAsc})
	params.setLimit(50)
	resp, err := c.get(fmt.Sprintf("/Items/%s/Similar", artistID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	return c.parseArtists(resp)
}

func (c *Client) GetGenres(paging Paging) ([]NameID, error) {
	params := c.defaultParams()
	params.enableRecursive()
	params.setSorting(Sort{Field: SortByName, Mode: SortAsc})
	params.setPaging(paging)
	// TODO
	//params.setParentId(c.musicView)

	resp, err := c.get("/Genres", params)
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

func (c *Client) parseArtists(resp io.Reader) ([]*Artist, error) {
	artists := &artists{}
	err := json.NewDecoder(resp).Decode(&artists)
	if err != nil {
		return nil, fmt.Errorf("decode json: %v", err)
	}
	return artists.Artists, nil
}
