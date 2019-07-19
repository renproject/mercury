package ethrpc

import (
	"bytes"
	"fmt"
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
	req, err := http.NewRequest("POST", client.url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("cannot construct post request for infura: %v", err)
	}
	return http.DefaultClient.Do(req)
}
