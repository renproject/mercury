package rpc

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/renproject/mercury/types/ethtypes"
)

// infuraClient implements the `Client` interface.
type infuraClient struct {
	network    ethtypes.Network
	url        string
	taggedKeys map[string]string
}

// NewInfuraClient returns a new infuraClient.
func NewInfuraClient(network ethtypes.Network, taggedKeys map[string]string) Client {
	return &infuraClient{
		network:    network,
		url:        fmt.Sprintf("https://%s.infura.io/v3", network.String()),
		taggedKeys: taggedKeys,
	}
}

// HandleRequest implements the `Client` interface.
func (infura *infuraClient) HandleRequest(r *http.Request, data []byte) (*http.Response, error) {
	tag := r.URL.Query().Get("tag")
	apiKey := infura.taggedKeys[tag]
	if apiKey == "" {
		apiKey = infura.taggedKeys[""]
	}
	client := http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", infura.url, apiKey), bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("cannot construct post request for infura: %v", err)
	}
	return client.Do(req)
}
