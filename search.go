package jellyfin

import (
	"encoding/json"
	"fmt"
	"io"
)

func searchDtoToItems(rc io.ReadCloser, itemType mediaItemType) (*SearchResult, error) {
	var result interface{}
	switch itemType {
	case mediaTypeAudio:
		result = &songs{}
	case mediaTypeAlbum:
		result = &albums{}
	case mediaTypeArtist:
		result = &artists{}
	case mediaTypePlaylist:
		result = &playlists{}
	default:
		return nil, fmt.Errorf("unknown item type: %s", itemType)
	}

	err := json.NewDecoder(rc).Decode(result)
	if err != nil {
		return nil, fmt.Errorf("decode item %s: %v", itemType, err)
	}

	searchResult := &SearchResult{}
	switch itemType {
	case mediaTypeAudio:
		searchResult.Songs = result.(*songs).Songs
	case mediaTypeAlbum:
		searchResult.Albums = result.(*albums).Albums
	case mediaTypeArtist:
		searchResult.Artists = result.(*artists).Artists
	case mediaTypePlaylist:
		searchResult.Playlists = result.(*playlists).Playlists
	}

	return searchResult, nil
}

// Search searches audio items
func (jf *Client) Search(query string, itemType ItemType, limit int) (*SearchResult, error) {
	if limit == 0 {
		limit = 40
	}
	params := jf.defaultParams()
	params.enableRecursive()
	params["SearchTerm"] = query
	params["Limit"] = fmt.Sprint(limit)
	params["IncludePeople"] = "false"
	params["IncludeMedia"] = "true"

	// default search URL
	url := fmt.Sprintf("/Users/%s/Items", jf.userID)

	var mediaType mediaItemType
	switch itemType {
	case TypeArtist:
		mediaType = mediaTypeArtist
		params["IncludeArtists"] = "true"
		params["IncludeMedia"] = "false"
		url = "/Artists"
	case TypeAlbum:
		mediaType = mediaTypeAlbum
		params.setIncludeTypes(mediaTypeAlbum)
	case TypeSong:
		mediaType = mediaTypeAudio
		params.setIncludeTypes(mediaTypeAudio)
	case TypePlaylist:
		mediaType = mediaTypePlaylist
		params.setIncludeTypes(mediaTypePlaylist)
	default:
		return nil, fmt.Errorf("itemType %s not supported", itemType)
	}

	body, err := jf.get(url, params)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %s", err.Error())
	}

	return searchDtoToItems(body, mediaType)
}
