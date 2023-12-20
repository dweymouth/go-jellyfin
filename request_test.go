package jellyfin

import (
	"net/url"
	"reflect"
	"testing"
)

func TestClient_encodeGETUrl(t *testing.T) {
	tests := []struct {
		name     string
		c        *Client
		endpoint string
		params   params
		want     string
		wantErr  bool
	}{

		{
			name: "POSITIVE - encodes the url",
			c: &Client{
				baseURL: &url.URL{
					Scheme: "https",
					Host:   "jellyfin.example.com",
				},
			},
			endpoint: "/audio/1234/stream",
			params: params{
				"UserId":        "userID",
				"DeviceId":      "deviceID",
				"playSessionId": "deviceID",
				"static":        "true",
				"api_key":       "5678",
			},
			want:    "https://jellyfin.example.com/audio/1234/stream?UserId=userID&DeviceId=deviceID&playSessionId=deviceID&static=true&api_key=5678",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.encodeGETUrl(tt.endpoint, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.encodeGETUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// since params was a map, we can't do a direct string comparison. The map values can get added in any order.
			// gotta parse out the 'want' url and the 'got' url and determine if the query values are the same.

			wantParsed, err := url.Parse(tt.want)
			if err != nil {
				t.Errorf("unable to parse want url: %v", err)
			}

			gotParsed, err := url.Parse(got)
			if err != nil {
				t.Errorf("unable to parse got url: %v", err)
			}

			if wantParsed.Scheme != gotParsed.Scheme {
				t.Errorf("schemes did not match, got %v, want %v", got, tt.want)
			}

			if wantParsed.Host != gotParsed.Host {
				t.Errorf("hosts did not match, got %v, want %v", got, tt.want)
			}

			// compare the queries
			for k, v := range wantParsed.Query() {
				gotValue, contained := gotParsed.Query()[k]
				if !contained {
					t.Errorf("param [%s] not contained in got", k)
				}

				// contained but now check the value
				if !reflect.DeepEqual(v, gotValue) {
					t.Errorf("value of param [%s] %v was not equal to got %v", k, v, gotValue)
				}
			}
		})
	}
}
