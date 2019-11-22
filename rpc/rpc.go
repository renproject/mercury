package rpc

import (
	"bytes"
	"fmt"
	"net/http"
)

// Client is a RPC client which can send and retrieve information from a blockchain through JSON-RPC. `data` is the
// request data we want to send to the ZCash node, and `r` is the original request in case we need to access any query
// parameters or other fields.
type Client interface {
	HandleRequest(r *http.Request, data []byte) (*http.Response, error)
}

// client implements the `Client` interface.
type client struct {
	host     string
	username string
	password string
}

// NewClient returns a new client.
func NewClient(host, username, password string) Client {
	return &client{
		host:     host,
		username: username,
		password: password,
	}
}

// HandleRequest implements the `Client` interface.
func (node *client) HandleRequest(r *http.Request, data []byte) (*http.Response, error) {
	client := http.Client{}
	req, err := http.NewRequest("POST", node.host, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	if node.username != "" || node.password != "" {
		req.SetBasicAuth(node.username, node.password)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot construct post request: %v", err)
	}
	return client.Do(req)
}
