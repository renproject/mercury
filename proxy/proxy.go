// Package proxy proxies requests to given clients. If a client returns an error for a given request, the next client is
// used. If all clients return errors, it returns each of the errors concatenated.
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

// NewProxy returns a new Proxy.
func NewProxy(clients ...rpc.Client) *Proxy {
	return &Proxy{
		Clients: clients,
	}
}

func (proxy *Proxy) ProxyRequest(r *http.Request, data []byte) (*http.Response, error) {
	errs := types.NewErrList(len(proxy.Clients))
	for i, client := range proxy.Clients {
		response, err := client.HandleRequest(r, data)
		if err != nil {
			errs[i] = err
			continue
		}
		return response, nil
	}
	return nil, errs
}
