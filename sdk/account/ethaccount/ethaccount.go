package ethaccount

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/types/ethtypes"
)

type Account struct {
	Client *ethclient.EthClient

	Address ethtypes.EthAddr
	key     *ecdsa.PrivateKey
}

func NewEthAccount(client *ethclient.EthClient, key *ecdsa.PrivateKey) *Account {
	addressString := crypto.PubkeyToAddress(key.PublicKey).Hex()
	address := ethtypes.HexStringToEthAddr(addressString)
	return &Account{
		Client:  client,
		Address: address,
		key:     key,
	}
}

func (account Account) CreateUTX(ctx context.Context, toAddress ethtypes.EthAddr, value ethtypes.Amount, gasLimit uint64, gasPrice ethtypes.Amount, data []byte) (ethtypes.EthUnsignedTx, error) {
	nonce, err := account.Client.PendingNonceAt(context.Background(), account.Address)
	if err != nil {
		return nil, err
	}
	return account.Client.CreateUTX(nonce, toAddress, value, gasLimit, gasPrice, data), nil
}

func (account Account) SignUTX(ctx context.Context, utx ethtypes.EthUnsignedTx) (ethtypes.EthSignedTx, error) {
	return account.Client.SignUTX(ctx, utx, account.key)
}
