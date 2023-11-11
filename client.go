package jellyfin

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
)

// Client is the root struct for all Jellyfin API calls
type Client struct {
	HTTPClient    *http.Client
	BaseURL       string
	ClientName    string
	ClientVersion string

	loggedIn bool
	token    string
	serverID string
	userID   string
	deviceID string
}

func NewClient(baseURL, clientName, clientVersion string) *Client {
	return &Client{
		HTTPClient:    &http.Client{},
		BaseURL:       baseURL,
		ClientName:    clientName,
		ClientVersion: clientVersion,
	}
}

type loginResponse struct {
	User     userResponse `json:"User"`
	Token    string       `json:"AccessToken"`
	ServerId string       `json:"ServerId"`
}

type userResponse struct {
	Name     string `json:"Name"`
	ServerId string `json:"ServerId"`
	UserId   string `json:"Id"`
}

func (c *Client) Login(username, password string) error {
	body := map[string]string{}
	body["Username"] = username
	body["PW"] = password

	b := &bytes.Buffer{}
	_ = json.NewEncoder(b).Encode(body)

	auth := c.authHeader()
	req, err := http.NewRequest("POST", c.BaseURL+"/Users/authenticatebyname", b)
	if err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	req.Header.Set("X-Emby-Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		c.loggedIn = true
		dto := loginResponse{}
		err := json.NewDecoder(resp.Body).Decode(&dto)
		if err != nil {
			return fmt.Errorf("invalid login response: %v", err)
		}

		c.token = dto.Token
		c.serverID = dto.ServerId
		c.userID = dto.User.UserId
		c.loggedIn = true
	case http.StatusBadRequest:
		reason, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("login failed: %v", err)
		} else {
			return fmt.Errorf("login failed: %s", reason)
		}
	default:
		reason, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("login failed: %v", err)
		} else {
			return fmt.Errorf("login failed: %s", reason)
		}
	}
	return nil
}

type PingResponse struct {
	LocalAddress    string
	ServerName      string
	Version         string
	ProductName     string
	OperatingSystem string
	Id              string
}

func (c *Client) Ping() (*PingResponse, error) {
	body, err := c.get("/System/Info/Public", nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	res := &PingResponse{}
	err = json.NewDecoder(body).Decode(res)
	if err != nil {
		return nil, fmt.Errorf("invalid json response: %v", err)
	}

	//logrus.Debugf("Connect to server %s, (id %s)", res.ServerName, res.Id)
	return res, nil
}

func (c *Client) authHeader() string {
	//id, err := machineid.ProtectedID(config.AppName)
	//if err != nil {
	//logrus.Errorf("get unique host id: %v", err)
	//	id = util.RandomKey(30)
	//}

	auth := fmt.Sprintf("MediaBrowser Client=\"%s\", Device=\"%s\", DeviceId=\"%s\", Version=\"%s\"",
		c.ClientName, deviceName(), randomKey(30), c.ClientVersion)
	return auth
}

func (c *Client) ensureDeviceID() string {
	if c.deviceID == "" {
		c.deviceID = deviceName()
	}
	return c.deviceID
}

func deviceName() string {
	hostname, err := os.Hostname()
	if err != nil {
		switch runtime.GOOS {
		case "darwin":
			hostname = "mac"
		default:
			hostname = runtime.GOOS
		}
	}
	return hostname
}

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

func randomKey(length int) string {
	r := rand.Reader
	data := make([]byte, length)
	r.Read(data)

	for i, b := range data {
		data[i] = letters[b%byte(len(letters))]
	}
	return string(data)
}
