package jellyfin

import (
	"fmt"
	"io"
)

type setFavoriteBody struct{}

func (c *Client) SetFavorite(id string, favorite bool) error {
	endpoint := fmt.Sprintf("/Users/%s/FavoriteItems/%s", c.userID, id)
	var resp io.ReadCloser
	var err error
	if favorite {
		resp, err = c.post(endpoint, c.defaultParams(), setFavoriteBody{})
	} else {
		resp, err = c.delete(endpoint, c.defaultParams())
	}
	if err != nil {
		return err
	}
	resp.Close()
	return nil
}

type refreshLibraryBody struct{}

func (c *Client) RefreshLibrary() error {
	resp, err := c.post("/Library/Refresh", c.defaultParams(), refreshLibraryBody{})
	if err != nil {
		return err
	}
	resp.Close()
	return nil
}

type PlayEvent string

const (
	Start      PlayEvent = "start"
	Stop       PlayEvent = "stop"
	Pause      PlayEvent = "pause"
	Unpause    PlayEvent = "unpause"
	TimeUpdate PlayEvent = "timeupdate"
)

type playStatusBody struct {
	ItemId        string `json:"ItemId"`
	PositionTicks int64  `json:"PositionTicks"`
	EventName     string `json:"EventName,omitempty"`
	IsPaused      *bool  `json:"IsPaused,omitempty"`
}

func (c *Client) UpdatePlayStatus(songID string, event PlayEvent, positionTicks int64) error {
	body := playStatusBody{ItemId: songID, PositionTicks: positionTicks}
	path := "/Sessions/Playing/Progress"
	isPaused := new(bool)
	switch event {
	case Start:
		path = "/Sessions/Playing"
	case Stop:
		path = "/Sessions/Playing/Stopped"
		*isPaused = true
		body.IsPaused = isPaused
	case Pause:
		*isPaused = true
		body.IsPaused = isPaused
		body.EventName = string(event)
	case Unpause:
		*isPaused = false
		body.IsPaused = isPaused
		body.EventName = string(event)
	case TimeUpdate:
		body.EventName = string(event)
	}

	resp, err := c.post(path, c.defaultParams(), body)
	if err != nil {
		return err
	}
	resp.Close()
	return nil
}
