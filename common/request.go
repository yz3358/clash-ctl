package common

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
)

type HTTPError struct {
	Message string `json:"message"`
}

// MakeRequest compose a resty.Client
// that send request to given Server
func MakeRequest(s Server) *resty.Client {
	u := s.URL()
	client := resty.New().SetBaseURL(u.String())

	if s.Secret != "" {
		client.SetHeader("Authorization", fmt.Sprintf("Bearer %s", s.Secret))
	}

	return client
}

func MakeWebsocket(s Server, path string) (*websocket.Conn, error) {
	u := s.WebsocketURL()
	u.Path = path

	header := http.Header{}
	if s.Secret != "" {
		header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Secret))
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	c, _, err := dialer.Dial(u.String(), header)
	return c, err
}
