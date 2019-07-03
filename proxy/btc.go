package proxy

import (
	"net/http"

	"github.com/renproject/mercury/rpc/btcrpc"
	"github.com/renproject/mercury/types"
)

// BtcProxy proxies the request to different Bitcoin clients.
type BtcProxy struct {
	Clients []btcrpc.Client
}

// NewBtcProxy returns a new BtcProxy for a given network.
func NewBtcProxy(clients ...btcrpc.Client) *BtcProxy {
	return &BtcProxy{
		Clients: clients,
	}
}

func (btc *BtcProxy) ProxyRequest(r *http.Request) (*http.Response, error) {
	errs := types.NewErrList(len(btc.Clients))
	for i, client := range btc.Clients {
		response, err := client.HandleRequest(r)
		if err != nil {
			errs[i] = err
			continue
		}
		return response, nil
	}
	return nil, errs
}
