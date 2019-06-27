package ethaccount

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/types/ethtypes"
)

type Account interface {
	CreateUTX(ctx context.Context, toAddress ethtypes.EthAddr, value ethtypes.Amount, gasLimit uint64, gasPrice ethtypes.Amount, data []byte) (ethtypes.EthUnsignedTx, error)
	SignUTX(ctx context.Context, utx ethtypes.EthUnsignedTx) (ethtypes.EthSignedTx, error)
}

type account struct {
	client *ethclient.EthClient

	address ethtypes.EthAddr
	key     *ecdsa.PrivateKey
}

func NewEthAccount(client *ethclient.EthClient, key *ecdsa.PrivateKey) Account {
	addressString := crypto.PubkeyToAddress(key.PublicKey).Hex()
	address := ethtypes.HexStringToEthAddr(addressString)
	return &account{
		client:  client,
		address: address,
		key:     key,
	}
}

func (acc *account) CreateUTX(ctx context.Context, toAddress ethtypes.EthAddr, value ethtypes.Amount, gasLimit uint64, gasPrice ethtypes.Amount, data []byte) (ethtypes.EthUnsignedTx, error) {
	nonce, err := acc.client.PendingNonceAt(ctx, acc.address)
	if err != nil {
		return nil, err
	}
	return acc.client.CreateUTX(nonce, toAddress, value, gasLimit, gasPrice, data), nil
}

func (acc *account) SignUTX(ctx context.Context, utx ethtypes.EthUnsignedTx) (ethtypes.EthSignedTx, error) {
	return acc.client.SignUTX(ctx, utx, acc.key)
}
