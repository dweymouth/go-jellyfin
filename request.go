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
	resp, err := c.makeDo(http.MethodGet, url, nil, params, nil)
	if resp != nil {
		return resp.Body, err
	}
	return nil, err
}

func (c *Client) delete(url string, params params) (io.ReadCloser, error) {
	resp, err := c.makeDo(http.MethodDelete, url, nil, params, nil)
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
	resp, err := c.makeDo(http.MethodPost, url, bodyEnc, params, nil)
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

// makeDo constructs request and performs Do.
// Set authorization header and build url query.
// Make request, parse response code and raise error if needed. Else return response body
func (c *Client) makeDo(method, url string, body []byte, params params, headers map[string]string) (*http.Response, error) {
	var req *http.Request
	var err error

	u := fmt.Sprintf("%s%s", c.BaseURL(), url)

	// generate http.Request
	if body != nil {
		req, err = http.NewRequest(method, u, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, u, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// set headers
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Emby-Token", c.token)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// set params
	if params != nil {
		q := req.URL.Query()
		for i, v := range params {
			q.Add(i, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	// DO
	//start := time.Now()
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	//took := time.Since(start)
	//logrus.Debugf("%s %s: %d (%d ms)", req.Method, req.URL.Path, resp.StatusCode, took.Milliseconds())

	// check response for errors and return the response
	return checkResponse(resp)
}

// checkResponse determines if there is was an error returned by jellyfin.
func checkResponse(resp *http.Response) (*http.Response, error) {
	// 200 or 204 is all good
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return resp, nil
	}

	// read in the body and look for an error message
	bytes, _ := io.ReadAll(resp.Body)
	msg := "no body"
	if len(bytes) > 0 {
		msg = string(bytes)
	}

	errMsg := errUnexpectedStatusCode

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
	}

	return resp, fmt.Errorf("%s, code: %s, msg: %s", errMsg, resp.Status, msg)
}
