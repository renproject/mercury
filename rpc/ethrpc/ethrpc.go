package ethrpc

import (
	"bytes"
	"net/http"

	"github.com/renproject/mercury/rpc"
)

// client implements the `Client` interface.
type client struct {
	url string
}

// New returns a new rpc Client.
func New(url string) (rpc.Client, error) {
	return &client{
		url: url,
	}, nil
}

// HandleRequest implements the `Client` interface.
func (client *client) HandleRequest(r *http.Request, data []byte) (*http.Response, error) {
	return http.Post(client.url, "application/json", bytes.NewBuffer(data))
}
