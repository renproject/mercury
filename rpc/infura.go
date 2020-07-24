package rpc

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/renproject/mercury/types/ethtypes"
)

// infuraClient implements the `Client` interface.
type infuraClient struct {
	client  *http.Client
	network ethtypes.Network
	url     string
}

// NewInfuraClient returns a new infuraClient.
func NewInfuraClient(client *http.Client, network ethtypes.Network, key string) Client {
	return &infuraClient{
		client:  client,
		network: network,
		url:     fmt.Sprintf("https://%s.infura.io/v3/%v", network.String(), key),
	}
}

// HandleRequest implements the `Client` interface.
func (infura *infuraClient) HandleRequest(r *http.Request, data []byte) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequest("POST", infura.url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("cannot construct post request for infura: %v", err)
	}
	req.WithContext(ctx)

	return infura.client.Do(r)
}
