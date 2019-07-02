package proxy

import (
	"fmt"
	"net/http"

	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/ethtypes"
)

// EthProxy proxies the request to different bitcoin clients.
type EthProxy interface {
	Network() ethtypes.Network
	HandleRequest(r *http.Request) (*http.Response, error)
}

type infuraProxy struct {
	network    ethtypes.Network
	url        string
	apiKey     string
	taggedKeys map[string]string
}

// NewInfuraProxy returns a new infuraProxy which implements the EthProxy interface
func NewInfuraProxy(network ethtypes.Network, apiKey string, taggedKeys map[string]string) (EthProxy, error) {
	var infuraNetwork string
	switch network {
	case ethtypes.Mainnet:
		infuraNetwork = "mainnet"
	case ethtypes.Kovan:
		infuraNetwork = "kovan"
	default:
		return &infuraProxy{}, types.ErrUnknownNetwork
	}
	return &infuraProxy{
		network:    network,
		url:        fmt.Sprintf("https://%s.infura.io/v3", infuraNetwork),
		apiKey:     apiKey,
		taggedKeys: taggedKeys,
	}, nil
}

func (eth *infuraProxy) Network() ethtypes.Network {
	return eth.network
}

func (eth *infuraProxy) HandleRequest(r *http.Request) (*http.Response, error) {
	tag := r.URL.Query().Get("tag")
	apiKey := eth.taggedKeys[tag]
	if apiKey == "" {
		apiKey = eth.apiKey
	}
	return http.Post(fmt.Sprintf("%s/%s", eth.url, apiKey), "application/json", r.Body)
}
