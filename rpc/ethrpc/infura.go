package ethrpc

import (
	"fmt"
	"net/http"

	"github.com/renproject/mercury/rpc"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/ethtypes"
)

// infuraClient implements the Client interface.
type infuraClient struct {
	network ethtypes.Network
	url     string
	apiKey  string
}

// NewNodeClient returns a new infuraClient.
func NewInfuraClient(network ethtypes.Network, apiKey string) (rpc.Client, error) {
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
		network: network,
		url:     fmt.Sprintf("https://%s.infura.io/v3", infuraNetwork),
		apiKey:  apiKey,
	}, nil
}

// BlockInfo implements the `Client` interface.
func (infura *infuraClient) HandleRequest(r *http.Request) (*http.Response, error) {
	return http.Post(fmt.Sprintf("%s/%s", infura.url, infura.apiKey), "application/json", r.Body)
}
