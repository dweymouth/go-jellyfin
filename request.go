package jellyfin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
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

	if sort.Field != "" {
		field = string(sort.Field)
	}

	p["SortBy"] = field
	p["SortOrder"] = order
}

func (p params) setPaging(paging Paging) {
	if paging.Limit > 0 {
		p["Limit"] = strconv.Itoa(paging.Limit)
	}
	p["StartIndex"] = strconv.Itoa(paging.StartIndex)
}

func (p params) setLimit(n int) {
	p["Limit"] = strconv.Itoa(n)
}

func (p params) setIncludeTypes(itemType mediaItemType) {
	p["IncludeItemTypes"] = string(itemType)
}

func (p params) setIncludeFields(fields ...string) {
	p["Fields"] = strings.Join(fields, ",")
}

func (p params) enableRecursive() {
	p["Recursive"] = "true"
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
				var sb strings.Builder
				for i := 0; i < totalYears+1; i++ {
					if i > 0 {
						sb.WriteString(",")
					}
					year := filter.YearRange[0] + i
					sb.WriteString(strconv.Itoa(year))
				}
				years = sb.String()
			}
			p["Years"] = years
		}
	}

	if len(filter.Genres) > 0 {
		p["Genres"] = strings.Join(filter.Genres, "|")
	}

	if f != "" {
		p["Filters"] = f
	}

	if filter.ArtistID != "" {
		p["ArtistIds"] = filter.ArtistID
	}

	if filter.ParentID != "" {
		p["ParentId"] = filter.ParentID
	}
}

func appendFilter(old, new string, separator string) string {
	if old == "" {
		return new
	}
	return old + separator + new
}

func (c *Client) get(url string, params params) (io.ReadCloser, error) {
	resp, err := c.makeRequest(http.MethodGet, url, nil, params, nil)
	if resp != nil {
		return resp.Body, err
	}
	return nil, err
}

func (c *Client) delete(url string, params params) (io.ReadCloser, error) {
	resp, err := c.makeRequest(http.MethodDelete, url, nil, params, nil)
	if resp != nil {
		return resp.Body, err
	}
	return nil, err
}

func (c *Client) post(url string, params params, body interface{}) (io.ReadCloser, error) {
	bodyEnc, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal POST body: %v", err)
	}
	resp, err := c.makeRequest(http.MethodPost, url, bodyEnc, params, nil)
	if resp != nil {
		return resp.Body, err
	}
	return nil, err
}

func (c *Client) encodeGETUrl(endpoint string, params params) (string, error) {
	baseUrl := c.BaseURL()
	baseUrl.Path = path.Join(baseUrl.Path, endpoint)
	req, err := http.NewRequest(http.MethodGet, baseUrl.String(), nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	for key, val := range params {
		q.Add(key, val)
	}
	req.URL.RawQuery = q.Encode()
	return req.URL.String(), nil
}

// Construct request
// Set authorization header and build url query
// Make request, parse response code and raise error if needed. Else return response body
func (c *Client) makeRequest(method, url string, body []byte, params params, headers map[string]string) (*http.Response, error) {
	var reader *bytes.Buffer
	var req *http.Request
	var err error
	if body != nil {
		reader = bytes.NewBuffer(body)
		req, err = http.NewRequest(method, fmt.Sprintf("%s%s", c.BaseURL(), url), reader)
	} else {
		req, err = http.NewRequest(method, fmt.Sprintf("%s%s", c.BaseURL(), url), nil)
	}

	if err != nil {
		return &http.Response{}, fmt.Errorf("failed to make request: %v", err)
	}
	if method == http.MethodPost {
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

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return resp, nil
	}
	bytes, _ := io.ReadAll(resp.Body)
	var msg string = "no body"
	if len(bytes) > 0 {
		msg = string(bytes)
	}
	var errMsg string
	switch resp.StatusCode {
	case http.StatusBadRequest:
		errMsg = errInvalidRequest
	case http.StatusUnauthorized:
		errMsg = errUnauthorized
	case http.StatusForbidden:
		errMsg = errForbidden
	case http.StatusNotFound:
		errMsg = errNotFound
	case http.StatusInternalServerError:
		errMsg = errServerError
	default:
		errMsg = errUnexpectedStatusCode
	}
	return resp, fmt.Errorf("%s, code: %d, msg: %s", errMsg, resp.StatusCode, msg)
}
