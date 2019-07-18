package ethrpc

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/renproject/mercury/rpc"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/ethtypes"
)

// infuraClient implements the `Client` interface.
type infuraClient struct {
	network    ethtypes.Network
	url        string
	apiKey     string
	taggedKeys map[string]string
}

// NewInfuraClient returns a new infuraClient.
func NewInfuraClient(network ethtypes.Network, apiKey string, taggedKeys map[string]string) (rpc.Client, error) {
	var infuraNetwork string
	switch network {
	case ethtypes.Mainnet:
		infuraNetwork = "mainnet"
	case ethtypes.Kovan:
		infuraNetwork = "kovan"
	default:
		return &infuraClient{}, types.ErrUnknownNetwork
	}
	return &infuraClient{
		network:    network,
		url:        fmt.Sprintf("https://%s.infura.io/v3", infuraNetwork),
		apiKey:     apiKey,
		taggedKeys: taggedKeys,
	}, nil
}

// HandleRequest implements the `Client` interface.
func (infura *infuraClient) HandleRequest(r *http.Request, data []byte) (*http.Response, error) {
	tag := r.URL.Query().Get("tag")
	apiKey := infura.taggedKeys[tag]
	if apiKey == "" {
		apiKey = infura.apiKey
	}
	client := http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", infura.url, apiKey), bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("cannot construct post request for infura: %v", err)
	}
	return client.Do(req)
}
