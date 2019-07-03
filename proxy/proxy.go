package proxy

import (
	"net/http"

	"github.com/renproject/mercury/rpc"
	"github.com/renproject/mercury/types"
)

// Proxy proxies the request to different clients.
type Proxy struct {
	Clients []rpc.Client
}

// NewProxy returns a new Proxy for a given network.
func NewProxy(clients ...rpc.Client) *Proxy {
	return &Proxy{
		Clients: clients,
	}
}

func (proxy *Proxy) ProxyRequest(r *http.Request) (*http.Response, error) {
	errs := types.NewErrList(len(proxy.Clients))
	for i, client := range proxy.Clients {
		response, err := client.HandleRequest(r)
		if err != nil {
			errs[i] = err
			continue
		}
		return response, nil
	}
	return nil, errs
}
