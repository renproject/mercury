package btcrpc

import (
	"fmt"
	"net/http"
)

// Client is a RPC client which can send and retrieve information from the Bitcoin blockchain through JSON-RPC.
type Client interface {
	HandleRequest(r *http.Request) (*http.Response, error)
}

// nodeClient implements the Client interface.
type nodeClient struct {
	host     string
	username string
	password string
}

// NewNodeClient returns a new nodeClient.
func NewNodeClient(host, username, password string) (Client, error) {
	return &nodeClient{
		host:     host,
		username: username,
		password: password,
	}, nil
}

// BlockInfo implements the `Client` interface.
func (node *nodeClient) HandleRequest(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(node.username, node.password)
	return http.Post(fmt.Sprintf("%s", node.host), "application/json", r.Body)
}
