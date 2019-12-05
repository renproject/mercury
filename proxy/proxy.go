// Package proxy proxies requests to given clients. If a client returns an error for a given request, the next client is
// used. If all clients return errors, it returns each of the errors concatenated.
package proxy

import (
	"context"
	"net/http"

	"github.com/renproject/mercury/rpc"
	"github.com/renproject/mercury/types"
	"github.com/sirupsen/logrus"
)

// Proxy proxies the request to different clients.
type Proxy struct {
	Logger  logrus.FieldLogger
	Clients []rpc.Client
}

// NewProxy returns a new Proxy.
func NewProxy(logger logrus.FieldLogger, clients ...rpc.Client) *Proxy {
	return &Proxy{
		Logger:  logger,
		Clients: clients,
	}
}

func (proxy *Proxy) ProxyRequest(ctx context.Context, r *http.Request, data []byte) (*http.Response, error) {
	errs := types.NewErrList(len(proxy.Clients))
	for {
		for i, client := range proxy.Clients {
			select {
			case <-ctx.Done():
				return nil, errs
			default:
				response, err := client.HandleRequest(r, data)
				if err != nil {
					proxy.Logger.Errorf("failed to handle request on proxy %d: %v", i, err)
					errs[i] = err
					continue
				}
				return response, nil
			}
		}
	}
}
