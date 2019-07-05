package ethclient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/ethtypes"
)

// Client is a client which is used to talking with certain bitcoin network. It can interacting with the blockchain
// through Mercury server.
type Client interface {
	Balance(context.Context, ethtypes.Address) (ethtypes.Amount, error)
	BlockNumber(context.Context) (*big.Int, error)
	SuggestGasPrice(context.Context) (ethtypes.Amount, error)
	PendingNonceAt(context.Context, ethtypes.Address) (uint64, error)
	BuildUnsignedTx(context.Context, uint64, ethtypes.Address, ethtypes.Amount, uint64, ethtypes.Amount, []byte) (ethtypes.Tx, error)
	PublishSignedTx(context.Context, ethtypes.Tx) error
	GasLimit(context.Context) (uint64, error)
}

type client struct {
	url    string
	client *ethclient.Client
}

// NewClient returns a new Client of given ethereum network.
func New(network ethtypes.Network) (Client, error) {
	var url string
	switch network {

	case ethtypes.Mainnet:
		url = "http://206.189.83.88:5000/eth/mainnet"
	case ethtypes.Kovan:
		url = "http://206.189.83.88:5000/eth/testnet"
	default:
		return &client{}, types.ErrUnknownNetwork
	}
	return NewCustomClient(url)
}

// NewCustomClient returns an Client for a specific RPC url
func NewCustomClient(url string) (Client, error) {
	ec, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	return &client{
		url:    url,
		client: ec,
	}, nil
}

// Balance returns the balance of the given ethereum address.
func (c *client) Balance(ctx context.Context, address ethtypes.Address) (ethtypes.Amount, error) {
	value, err := c.client.BalanceAt(ctx, common.Address(address), nil)
	if err != nil {
		return ethtypes.Amount{}, err
	}
	return ethtypes.WeiFromBig(value), nil
}

// BlockNumber returns the current highest block number.
func (c *client) BlockNumber(ctx context.Context) (*big.Int, error) {
	value, err := c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	return value.Number, nil
}

func (c *client) SuggestGasPrice(ctx context.Context) (ethtypes.Amount, error) {
	price, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return ethtypes.Amount{}, err
	}
	return ethtypes.WeiFromBig(price), err
}

func (client *client) PendingNonceAt(ctx context.Context, fromAddress ethtypes.Address) (uint64, error) {
	return client.client.PendingNonceAt(ctx, common.Address(fromAddress))
}

func (c *client) BuildUnsignedTx(ctx context.Context, nonce uint64, toAddress ethtypes.Address, value ethtypes.Amount, gasLimit uint64, gasPrice ethtypes.Amount, data []byte) (ethtypes.Tx, error) {
	chainID, err := c.client.NetworkID(ctx)
	if err != nil {
		return ethtypes.Tx{}, err
	}
	return ethtypes.NewUnsignedTx(chainID, nonce, toAddress, value, gasLimit, gasPrice, data), nil
}

// PublishSTX publishes a signed transaction
func (c *client) PublishSignedTx(ctx context.Context, tx ethtypes.Tx) error {
	// Pre-condition checks
	if !tx.IsSigned() {
		panic("pre-condition violation: cannot publish unsigned transaction")
	}
	return c.client.SendTransaction(ctx, tx.ToTransaction())
}

// BlockNumber returns the gas limit of the latest block.
func (c *client) GasLimit(ctx context.Context) (uint64, error) {
	value, err := c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}
	return value.GasLimit, nil
}
