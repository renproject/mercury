package proxy

import (
	"fmt"
	"net/http"

	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/ethtypes"
)

// EthProxy proxies the request to different bitcoin clients.
type EthProxy struct {
	Network ethtypes.EthNetwork
}

// NewEthProxy returns a new ethProxy for given network.
func NewEthProxy(network ethtypes.EthNetwork) (*EthProxy, error) {
	client := &EthProxy{
		Network: network,
	}
	switch network {
	case ethtypes.EthMainnet:
		return client, nil
	case ethtypes.EthKovan:
		return client, nil
	default:
		return &EthProxy{}, types.ErrUnknownNetwork
	}
}

func (eth *EthProxy) ForwardRequest(r *http.Request, apiKey string) (*http.Response, error) {
	network := eth.Network.String()
	return http.Post(fmt.Sprintf("https://%s.infura.io/v3/%s", network, apiKey), "application/json", r.Body)
}
