package jellyfin

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
