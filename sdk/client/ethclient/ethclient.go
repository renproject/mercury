package ethclient

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/ethtypes"
)

// EthClient is a client which is used to talking with certain bitcoin network. It can interacting with the blockchain
// through Mercury server.
type EthClient interface {
	Balance(context.Context, ethtypes.EthAddr) (ethtypes.Amount, error)
	BlockNumber(context.Context) (*big.Int, error)
	SuggestGasPrice(context.Context) (ethtypes.Amount, error)
	PendingNonceAt(context.Context, ethtypes.EthAddr) (uint64, error)
	CreateUTX(uint64, ethtypes.EthAddr, ethtypes.Amount, uint64, ethtypes.Amount, []byte) ethtypes.EthUnsignedTx
	SignUTX(context.Context, ethtypes.EthUnsignedTx, *ecdsa.PrivateKey) (ethtypes.EthSignedTx, error)
	PublishSTX(context.Context, ethtypes.EthSignedTx) error
	GasLimit(context.Context) (uint64, error)
}

type ethClient struct {
	url    string
	client *ethclient.Client
}

// NewEthClient returns a new EthClient of given ethereum network.
func NewEthClient(network ethtypes.EthNetwork) (EthClient, error) {
	var url string
	switch network {

	case ethtypes.EthMainnet:
		url = "https://ren-mercury.herokuapp.com/eth"
	case ethtypes.EthKovan:
		url = "https://ren-mercury.herokuapp.com/eth-kovan"
	default:
		return &ethClient{}, types.ErrUnknownNetwork
	}
	return NewCustomEthClient(url)
}

// NewCustomEthClient returns an EthClient for a specific RPC url
func NewCustomEthClient(url string) (EthClient, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return &ethClient{}, err
	}
	return &ethClient{
		url:    url,
		client: client,
	}, nil
}

// Balance returns the balance of the given ethereum address.
func (client *ethClient) Balance(ctx context.Context, address ethtypes.EthAddr) (ethtypes.Amount, error) {
	value, err := client.client.BalanceAt(ctx, common.Address(address), nil)
	if err != nil {
		return ethtypes.Amount{}, err
	}
	return ethtypes.WeiFromBig(value), nil
}

// BlockNumber returns the current highest block number.
func (client *ethClient) BlockNumber(ctx context.Context) (*big.Int, error) {
	value, err := client.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	return value.Number, nil
}

func (client *ethClient) SuggestGasPrice(ctx context.Context) (ethtypes.Amount, error) {
	price, err := client.client.SuggestGasPrice(ctx)
	if err != nil {
		return ethtypes.Amount{}, err
	}
	return ethtypes.WeiFromBig(price), err
}

func (client *ethClient) PendingNonceAt(ctx context.Context, fromAddress ethtypes.EthAddr) (uint64, error) {
	return client.client.PendingNonceAt(ctx, common.Address(fromAddress))
}

func (client *ethClient) CreateUTX(nonce uint64, toAddress ethtypes.EthAddr, value ethtypes.Amount, gasLimit uint64, gasPrice ethtypes.Amount, data []byte) ethtypes.EthUnsignedTx {
	return ethtypes.EthUnsignedTx(coretypes.NewTransaction(nonce, common.Address(toAddress), value.ToBig(), gasLimit, gasPrice.ToBig(), data))
}

func (client *ethClient) SignUTX(ctx context.Context, utx ethtypes.EthUnsignedTx, key *ecdsa.PrivateKey) (ethtypes.EthSignedTx, error) {
	chainID, err := client.client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	signedTx, err := coretypes.SignTx((*coretypes.Transaction)(utx), coretypes.NewEIP155Signer(chainID), key)
	if err != nil {
		return nil, err
	}
	return ethtypes.EthSignedTx(signedTx), nil
}

// PublishSTX publishes a signed transaction
func (client *ethClient) PublishSTX(ctx context.Context, stx ethtypes.EthSignedTx) error {
	return client.client.SendTransaction(ctx, (*coretypes.Transaction)(stx))
}

// BlockNumber returns the gas limit of the latest block.
func (client *ethClient) GasLimit(ctx context.Context) (uint64, error) {
	value, err := client.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}
	return value.GasLimit, nil
}
