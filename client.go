package jellyfin

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultTimeOut = 30 * time.Second
)

// Client is the root struct for all Jellyfin API calls
type Client struct {
	HTTPClient    *http.Client
	baseURL       *url.URL
	ClientName    string
	ClientVersion string

	loggedIn bool
	token    string
	serverID string
	username string
	userID   string
	deviceID string // needs to be unique for a user+device combo
}

// NewClient creates a jellyfin Client using the url provided.
func NewClient(urlStr, clientName, clientVersion string) (*Client, error) {
	// validate the baseurl
	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return &Client{
		HTTPClient: &http.Client{
			Timeout: DefaultTimeOut,
		},
		baseURL:       baseURL,
		ClientName:    clientName,
		ClientVersion: clientVersion,
	}, nil
}

// BaseURL return a copy of the baseURL.
func (c *Client) BaseURL() *url.URL {
	u := *c.baseURL
	return &u
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

// Login authenticates a user into the server provided in Client.
// If the login is successful, the access token is stored for future API calls.
func (c *Client) Login(ctx context.Context, username, password string) error {
	body := map[string]string{
		"Username": username,
		"PW":       password,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/Users/authenticatebyname", c.BaseURL()), io.NopCloser(bytes.NewBuffer(bodyBytes)))
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	req.Header.Set("X-Emby-Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		c.loggedIn = true
		dto := loginResponse{}
		err := json.NewDecoder(resp.Body).Decode(&dto)
		if err != nil {
			return fmt.Errorf("invalid login response: %w", err)
		}

		c.token = dto.Token
		c.serverID = dto.ServerId
		c.username = username
		c.userID = dto.User.UserId
		c.deviceID = "" // recalculate it next request, should be different per username
	case http.StatusBadRequest:
		reason, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		} else {
			return fmt.Errorf("login failed: %s", reason)
		}
	default:
		reason, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
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

	return res, nil
}

func (c *Client) LoggedInUser() string {
	return c.username
}

func (c *Client) authHeader() string {
	auth := fmt.Sprintf("MediaBrowser Client=\"%s\", Device=\"%s\", DeviceId=\"%s\", Version=\"%s\"",
		c.ClientName, deviceName(), c.ensureDeviceID(), c.ClientVersion)
	return auth
}

func (c *Client) ensureDeviceID() string {
	if c.deviceID == "" {
		mac, err := macaddress()
		if err != nil {
			mac = randomKey(16)
		}
		c.deviceID = fmt.Sprintf("%x", md5.Sum([]byte(mac+c.username)))
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

// adapted from https://gist.github.com/tsilvers/085c5f39430ced605d970094edf167ba
func macaddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", errors.New("failed to get net interfaces")
	}

	for _, i := range interfaces {
		if i.Flags&net.FlagUp != 0 && !bytes.Equal(i.HardwareAddr, nil) {
			// Skip locally administered addresses
			if i.HardwareAddr[0]&2 == 2 {
				continue
			}

			var mac uint64
			for j, b := range i.HardwareAddr {
				if j >= 8 {
					break
				}
				mac <<= 8
				mac += uint64(b)
			}

			return fmt.Sprintf("%16.16X", mac), nil
		}
	}

	return "", errors.New("no mac address found")
}
