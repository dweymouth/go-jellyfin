package jellyfin

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type params map[string]string

func (c *Client) defaultParams() params {
	params := params{}
	params["UserId"] = c.userID
	params["DeviceId"] = c.ensureDeviceID()
	return params
}

func (p params) setSorting(sort Sort) {

	field := "SortName"
	order := "Ascending"

	if sort.Mode == SortAsc {
		order = "Ascending"
	} else if sort.Mode == SortDesc {
		order = "Descending"
	}

	switch sort.Field {
	case SortByDate:
		field = "ProductionYear,ProductionYear,SortName"
	case SortByName:
		field = "SortName"
		// Todo: following depend on item type
	case SortByAlbum:
		field = "Album,SortName"
	case SortByArtist:
		field = "Artist,SortName"
	case SortByPlayCount:
		field = "PlayCount,SortName"
	case SortByRandom:
		field = "Random,SortName"
	case SortByLatest:
		field = "DateCreated,SortName"
	case SortByLastPlayed:
		field = "DatePlayed,SortName"
	}

	p["SortBy"] = field
	p["SortOrder"] = order
}

func (p params) setPaging(paging Paging) {
	p["Limit"] = strconv.Itoa(paging.Limit)
	p["StartIndex"] = strconv.Itoa(paging.StartIndex)
}

func (p params) setLimit(n int) {
	p["Limit"] = strconv.Itoa(n)
}

func (p params) setIncludeTypes(itemType mediaItemType) {
	p["IncludeItemTypes"] = string(itemType)
}

func (p params) enableRecursive() {
	p["Recursive"] = "true"
}

func (p params) setParentID(id string) {
	p["ParentId"] = id
}

func (p params) setFilter(tItem mediaItemType, filter Filter) {
	f := ""
	if filter.Favorite {
		f = appendFilter(f, "IsFavorite", ",")
	}

	// jellyfin server does not seem to like sorting artists by play status.
	// https://github.com/jellyfin/jellyfin/issues/2672
	if tItem != mediaTypeArtist {
		if filter.FilterPlayed == FilterIsPlayed {
			f = appendFilter(f, "IsPlayed", ",")
		} else if filter.FilterPlayed == FilterIsNotPlayed {
			f = appendFilter(f, "IsUnPlayed", ",")
		}
	}

	if tItem != mediaTypeArtist {
		if filter.yearRangeValid() && filter.YearRange[0] > 0 {
			years := ""
			totalYears := filter.YearRange[1] - filter.YearRange[0]
			if totalYears == 0 {
				years = strconv.Itoa(filter.YearRange[0])
			} else {
				for i := 0; i < totalYears+1; i++ {
					year := filter.YearRange[0] + i
					years = appendFilter(years, strconv.Itoa(year), ",")
				}
			}
			p["Years"] = years
		}
	}

	if len(filter.Genres) > 0 {
		genres := ""
		for _, v := range filter.Genres {
			genres = appendFilter(genres, v.Name, "|")
		}
		p["Genres"] = genres
	}

	if f != "" {
		p["Filters"] = f
	}
}

func appendFilter(old, new string, separator string) string {
	if old == "" {
		return new
	}
	return old + separator + new
}

func (c *Client) get(url string, params params) (io.ReadCloser, error) {
	resp, err := c.makeRequest("GET", url, nil, params, nil)
	if resp != nil {
		return resp.Body, err
	}
	return nil, err
}

// Construct request
// Set authorization header and build url query
// Make request, parse response code and raise error if needed. Else return response body
func (c *Client) makeRequest(method, url string, body *[]byte, params params,
	headers map[string]string) (*http.Response, error) {
	var reader *bytes.Buffer
	var req *http.Request
	var err error
	if body != nil {
		reader = bytes.NewBuffer(*body)
		req, err = http.NewRequest(method, c.BaseURL+url, reader)
	} else {
		req, err = http.NewRequest(method, c.BaseURL+url, nil)
	}

	if err != nil {
		return &http.Response{}, fmt.Errorf("failed to make request: %v", err)
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Emby-Token", c.token)

	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	if params != nil {
		q := req.URL.Query()
		for i, v := range params {
			q.Add(i, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	//start := time.Now()
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed make request: %v", err)
	}
	//took := time.Since(start)
	//logrus.Debugf("%s %s: %d (%d ms)", req.Method, req.URL.Path, resp.StatusCode, took.Milliseconds())

	if resp.StatusCode == 200 || resp.StatusCode == 204 {
		return resp, nil
	}
	bytes, _ := io.ReadAll(resp.Body)
	var msg string = "no body"
	if len(bytes) > 0 {
		msg = string(bytes)
	}
	var errMsg string
	switch resp.StatusCode {
	case 400:
		errMsg = errInvalidRequest
	case 401:
		errMsg = errUnauthorized
	case 403:
		errMsg = errForbidden
	case 404:
		errMsg = errNotFound
	case 500:
		errMsg = errServerError
	default:
		errMsg = errUnexpectedStatusCode
	}
	return resp, fmt.Errorf("%s, code: %d, msg: %s", errMsg, resp.StatusCode, msg)
}
