package zecrpc

import (
	"fmt"
	"net/http"

	"github.com/renproject/mercury/rpc"
)

// nodeClient implements the Client interface.
type nodeClient struct {
	host     string
	username string
	password string
}

// NewNodeClient returns a new nodeClient.
func NewNodeClient(host, username, password string) (rpc.Client, error) {
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
