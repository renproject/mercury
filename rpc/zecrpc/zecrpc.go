package zecrpc

import (
	"fmt"
	"net/http"

	"github.com/renproject/mercury/types/zectypes"
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
	network  zectypes.Network
}

// NewNodeClient returns a new nodeClient.
func NewNodeClient(network zectypes.Network, host, username, password string) (Client, error) {
	return &nodeClient{
		host:     host,
		username: username,
		password: password,
		network:  network,
	}, nil
}

// BlockInfo implements the `Client` interface.
func (node *nodeClient) HandleRequest(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(node.username, node.password)
	return http.Post(fmt.Sprintf("%s", node.host), "application/json", r.Body)
}
