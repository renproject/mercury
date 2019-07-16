package ethclient

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/sirupsen/logrus"
)

// Client is a client which is used to interact with the Ethereum network using the Mercury server.
type Client interface {
	EthClient() *ethclient.Client
	Balance(context.Context, ethtypes.Address) (ethtypes.Amount, error)
	BlockNumber(context.Context) (*big.Int, error)
	SuggestGasPrice(context.Context, types.TxSpeed) ethtypes.Amount
	PendingNonceAt(context.Context, ethtypes.Address) (uint64, error)
	BuildUnsignedTx(context.Context, uint64, ethtypes.Address, ethtypes.Amount, uint64, ethtypes.Amount, []byte) (ethtypes.Tx, error)
	PublishSignedTx(context.Context, ethtypes.Tx) (ethtypes.TxHash, error)
	GasLimit(context.Context) (uint64, error)
	Confirmations(ctx context.Context, hash ethtypes.TxHash) (*big.Int, error)
}

type client struct {
	url        string
	client     *ethclient.Client
	logger     logrus.FieldLogger
	gasStation EthGasStation
}

// New returns a new Client of given ethereum network.
func New(logger logrus.FieldLogger, network ethtypes.Network) (Client, error) {
	var url string
	switch network {

	case ethtypes.Mainnet:
		url = "http://206.189.83.88:5000/eth/mainnet"
	case ethtypes.Kovan:
		url = "http://206.189.83.88:5000/eth/testnet"
	default:
		return nil, types.ErrUnknownNetwork
	}
	return NewCustomClient(logger, url)
}

// NewCustomClient returns an Client for a specific RPC url
func NewCustomClient(logger logrus.FieldLogger, url string) (Client, error) {
	ec, err := ethclient.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("error creating EthClient at url=%v. %v", url, err)
	}
	return &client{
		url:        url,
		client:     ec,
		logger:     logger,
		gasStation: NewEthGasStation(logger, 30*time.Minute),
	}, nil
}

// EthClient returns the eth client of the given ethereum address.
func (c *client) EthClient() *ethclient.Client {
	return c.client
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

func (c *client) SuggestGasPrice(ctx context.Context, speed types.TxSpeed) ethtypes.Amount {
	gasStationPrice, err := c.gasStation.GasRequired(ctx, speed)
	if err == nil {
		return gasStationPrice
	}
	c.logger.Errorf("error getting gas from EthGasStation: %v", err)
	c.logger.Infof("trying gas price from EthClient")
	ethClientPrice, err := c.client.SuggestGasPrice(ctx)
	if err == nil {
		return ethtypes.WeiFromBig(ethClientPrice)
	}
	c.logger.Errorf("error getting gas from EthClient: %v", err)
	c.logger.Infof("using 21 Gwei as gas price")
	return ethtypes.Gwei(21)
}

func (c *client) PendingNonceAt(ctx context.Context, fromAddress ethtypes.Address) (uint64, error) {
	return c.client.PendingNonceAt(ctx, common.Address(fromAddress))
}

func (c *client) BuildUnsignedTx(ctx context.Context, nonce uint64, toAddress ethtypes.Address, value ethtypes.Amount, gasLimit uint64, gasPrice ethtypes.Amount, data []byte) (ethtypes.Tx, error) {
	chainID, err := c.client.NetworkID(ctx)
	if err != nil {
		return ethtypes.Tx{}, err
	}
	return ethtypes.NewUnsignedTx(chainID, nonce, toAddress, value, gasLimit, gasPrice, data), nil
}

// PublishSTX publishes a signed transaction
func (c *client) PublishSignedTx(ctx context.Context, tx ethtypes.Tx) (ethtypes.TxHash, error) {
	// Pre-condition checks
	if !tx.IsSigned() {
		panic("pre-condition violation: cannot publish unsigned transaction")
	}
	return tx.Hash(), c.client.SendTransaction(ctx, tx.ToTransaction())
}

func (c *client) Confirmations(ctx context.Context, hash ethtypes.TxHash) (*big.Int, error) {
	currentBlockNumber, err := c.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching current block number: %v", err)
	}
	receipt, err := c.client.TransactionReceipt(ctx, common.Hash(hash))
	if err != nil {
		return nil, fmt.Errorf("error fetching tx hash=%v receipt: %v", hash, err)
	}
	confs := big.NewInt(0).Sub(currentBlockNumber, receipt.BlockNumber)
	return confs, nil
}

// BlockNumber returns the gas limit of the latest block.
func (c *client) GasLimit(ctx context.Context) (uint64, error) {
	value, err := c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}
	return value.GasLimit, nil
}
