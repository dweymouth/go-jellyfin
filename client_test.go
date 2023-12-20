package jellyfin

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name          string
		urlStr        string
		clientName    string
		clientVersion string
		want          *Client
		wantErr       bool
	}{
		{
			name:          "POSITIVE - makes new client and parses url",
			urlStr:        "https://jellyfin.example.com/",
			clientName:    "supersonic",
			clientVersion: "1",
			want: &Client{
				baseURL: &url.URL{
					Scheme: "https",
					Host:   "jellyfin.example.com",
					Path:   "/",
				},
				ClientName:    "supersonic",
				ClientVersion: "1",
				HTTPClient: &http.Client{
					Timeout: DefaultTimeOut,
				},
			},
			wantErr: false,
		},
		{
			name:          "POSITIVE - makes new client and parses ip",
			urlStr:        "http://10.0.0.10:8096",
			clientName:    "supersonic",
			clientVersion: "1",
			want: &Client{
				baseURL: &url.URL{
					Scheme: "http",
					Host:   "10.0.0.10:8096",
					Path:   "/",
				},
				ClientName:    "supersonic",
				ClientVersion: "1",
				HTTPClient: &http.Client{
					Timeout: DefaultTimeOut,
				},
			},
			wantErr: false,
		},
		{
			name:          "NEGATIVE - empty url",
			urlStr:        "",
			clientName:    "supersonic",
			clientVersion: "1",
			want:          nil,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.urlStr, tt.clientName, tt.clientVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
