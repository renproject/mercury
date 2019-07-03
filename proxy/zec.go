package proxy

import (
	"net/http"

	"github.com/renproject/mercury/rpc/zecrpc"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/zectypes"
)

// ZecProxy proxies the request to different Bitcoin clients.
type ZecProxy struct {
	Clients []zecrpc.Client
	Network zectypes.Network
}

// NewZecProxy returns a new ZecProxy for a given network.
func NewZecProxy(network zectypes.Network, clients ...zecrpc.Client) *ZecProxy {
	return &ZecProxy{
		Clients: clients,
		Network: network,
	}
}

func (zec *ZecProxy) ProxyRequest(r *http.Request) (*http.Response, error) {
	errs := types.NewErrList(len(zec.Clients))
	for i, client := range zec.Clients {
		response, err := client.HandleRequest(r)
		if err != nil {
			errs[i] = err
			continue
		}
		return response, nil
	}
	return nil, errs
}
