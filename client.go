package jellyfin

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
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
func NewClient(urlStr, clientName, clientVersion string, opts ...ClientOptionFunc) (*Client, error) {
	// validate the baseurl
	if urlStr == "" {
		return nil, errors.New("url must be provided")
	}
	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	cli := &Client{
		HTTPClient: &http.Client{
			Timeout: DefaultTimeOut,
		},
		baseURL:       baseURL,
		ClientName:    clientName,
		ClientVersion: clientVersion,
	}

	// perform any options provided
	for _, option := range opts {
		option(cli)
	}

	return cli, nil
}

// ClientOptionFunc can be used to customize a new jellyfin API client.
type ClientOptionFunc func(*Client)

// Http timeout override.
func WithTimeout(timeout time.Duration) ClientOptionFunc {
	return func(c *Client) {
		c.HTTPClient.Timeout = timeout
	}
}

// Http client override.
func WithHTTPClient(httpClient *http.Client) ClientOptionFunc {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
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
func (c *Client) Login(username, password string) error {
	body := map[string]string{
		"Username": username,
		"PW":       password,
	}

	u, err := url.JoinPath(c.BaseURL().String(), "/Users/authenticatebyname")
	if err != nil {
		return fmt.Errorf("unable to parse url path: %w", err)
	}

	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(body); err != nil {
		return fmt.Errorf("unable to encode body: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, u, b)
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

// Ping queries the jellyfin server for a response.
// Return is some basic information about the jellyfin server.
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

// LoggedInUser returns the user associated with this Client.
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

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// randomKey returns a string at the length desired of random mixed-case letters.
func randomKey(length int) string {
	data := make([]byte, length)

	for i := range data {
		data[i] = letters[rand.Intn(len(letters))]
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
