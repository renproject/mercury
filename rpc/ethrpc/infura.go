package ethrpc

import (
	"bytes"
	"fmt"
	"log"
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
	log.Println(fmt.Sprintf("accessing %s/%s", infura.url, apiKey))
	return http.Post(fmt.Sprintf("%s/%s", infura.url, apiKey), "application/json", bytes.NewBuffer(data))
}
