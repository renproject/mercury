package zecrpc

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/renproject/mercury/rpc"
)

// nodeClient implements the `Client` interface.
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

// HandleRequest implements the `Client` interface.
func (node *nodeClient) HandleRequest(r *http.Request, data []byte) (*http.Response, error) {
	client := http.Client{}
	req, err := http.NewRequest("POST", node.host, bytes.NewBuffer(data))
	req.SetBasicAuth(node.username, node.password)
	if err != nil {
		return nil, fmt.Errorf("cannot construct post request for zcash node: %v", err)
	}
	return client.Do(req)
}
