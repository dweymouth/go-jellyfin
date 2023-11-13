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
func (jf *Client) Search(query string, itemType ItemType, paging Paging) (*SearchResult, error) {
	params := jf.defaultParams()
	params.enableRecursive()
	params.setPaging(paging)
	params["SearchTerm"] = query

	var mediaType mediaItemType
	switch itemType {
	case TypeArtist:
		mediaType = mediaTypeArtist
		params.setIncludeFields(artistIncludeFields...)
	case TypeAlbum:
		mediaType = mediaTypeAlbum
		params.setIncludeFields(albumIncludeFields...)
	case TypeSong:
		mediaType = mediaTypeAudio
		params.setIncludeFields(songIncludeFields...)
	case TypePlaylist:
		mediaType = mediaTypePlaylist
		params.setIncludeFields(playlistIncludeFields...)
	default:
		return nil, fmt.Errorf("itemType %s not supported", itemType)
	}
	params.setIncludeTypes(mediaType)

	body, err := jf.get(fmt.Sprintf("/Users/%s/Items", jf.userID), params)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %s", err.Error())
	}

	return searchDtoToItems(body, mediaType)
}
