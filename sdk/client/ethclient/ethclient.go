package ethclient

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/ethtypes"
)

// EthClient is a client which is used to talking with certain bitcoin network. It can interacting with the blockchain
// through Mercury server.
type EthClient struct {
	url    string
	client *ethclient.Client
}

// NewEthClient returns a new EthClient of given bitcoin network.
func NewEthClient(network ethtypes.EthNetwork) (*EthClient, error) {
	var url string
	switch network {
	case ethtypes.EthMainnet:
		url = "https://ren-mercury.herokuapp.com/eth"
	case ethtypes.EthKovan:
		url = "https://ren-mercury.herokuapp.com/eth-kovan"
	default:
		return &EthClient{}, types.ErrUnknownNetwork
	}
	client, err := ethclient.Dial(url)
	if err != nil {
		return &EthClient{}, err
	}
	return &EthClient{
		url:    url,
		client: client,
	}, nil
}

// Balance returns the balance of the given ethereum address.
func (client *EthClient) Balance(ctx context.Context, address ethtypes.EthAddr) (ethtypes.Amount, error) {
	value, err := client.client.BalanceAt(ctx, common.Address(address), nil)
	if err != nil {
		return ethtypes.Amount{}, err
	}
	fmt.Println(value)
	return ethtypes.WeiFromBig(value), nil
}
