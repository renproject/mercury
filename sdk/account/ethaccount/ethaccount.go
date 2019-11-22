package ethaccount

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/types/ethtypes"
)

type Account interface {
	Client() ethclient.Client
	Address() ethtypes.Address
	Balance(ctx context.Context) (ethtypes.Amount, error)
	Transact(ctx context.Context, unsignedTx ethtypes.Tx) (ethtypes.TxHash, error)
	Transfer(ctx context.Context, toAddress ethtypes.Address, value ethtypes.Amount, gasPrice ethtypes.Amount) (ethtypes.TxHash, error)
	PrivateKey() *ecdsa.PrivateKey
}

type account struct {
	client ethclient.Client

	address ethtypes.Address
	key     *ecdsa.PrivateKey
}

func NewAccountFromPrivateKey(client ethclient.Client, key *ecdsa.PrivateKey) (Account, error) {
	addressString := crypto.PubkeyToAddress(key.PublicKey).Hex()
	address := ethtypes.AddressFromHex(addressString)
	return &account{
		client:  client,
		address: address,
		key:     key,
	}, nil
}

func NewAccountFromMnemonic(client ethclient.Client, mnemonic, derivationPath string) (Account, error) {
	// Get the wallet
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return &account{}, err
	}

	// Get the account
	path := hdwallet.MustParseDerivationPath(derivationPath)
	acc, err := wallet.Derive(path, false)
	if err != nil {
		return &account{}, err
	}

	// Get the key
	key, err := wallet.PrivateKey(acc)
	return NewAccountFromPrivateKey(client, key)
}

func RandomAccount(client ethclient.Client) (Account, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return &account{}, err
	}
	return NewAccountFromPrivateKey(client, privateKey)
}

func (acc *account) Transfer(ctx context.Context, toAddress ethtypes.Address, value ethtypes.Amount, gasPrice ethtypes.Amount) (ethtypes.TxHash, error) {
	nonce, err := acc.client.PendingNonceAt(ctx, acc.address)
	// fmt.Printf("nonce fetched back from infura: %v", nonce)
	if err != nil {
		return ethtypes.TxHash{}, fmt.Errorf("failed to get pending nonce: %v", err)
	}
	tx, err := acc.client.BuildUnsignedTx(ctx, nonce, &toAddress, value, 21000, gasPrice, nil)
	if err != nil {
		return ethtypes.TxHash{}, err
	}
	if err := tx.Sign(acc.key); err != nil {
		return ethtypes.TxHash{}, err
	}
	return acc.client.PublishSignedTx(ctx, tx)
}

func (acc *account) Balance(ctx context.Context) (ethtypes.Amount, error) {
	return acc.client.Balance(ctx, acc.Address())
}

func (acc *account) Client() ethclient.Client {
	return acc.client
}

func (acc *account) Transact(ctx context.Context, utx ethtypes.Tx) (ethtypes.TxHash, error) {
	if err := utx.Sign(acc.key); err != nil {
		return ethtypes.TxHash{}, err
	}
	txHash, err := acc.client.PublishSignedTx(ctx, utx)
	if err != nil {
		for {
			if isNonceError(err) {
				txHash, err = acc.retryTx(ctx, utx)
				if err == nil {
					return txHash, nil
				}
			} else {
				return ethtypes.TxHash{}, err
			}
		}
	}
	return txHash, nil
}

func (acc *account) Address() ethtypes.Address {
	return acc.address
}

func (acc *account) PrivateKey() *ecdsa.PrivateKey {
	return acc.key
}

func (acc *account) retryTx(ctx context.Context, utx ethtypes.Tx) (ethtypes.TxHash, error) {
	updatedNonce, err := acc.client.PendingNonceAt(ctx, acc.address)
	if err != nil {
		return ethtypes.TxHash{}, err
	}
	utx.SetNonce(updatedNonce)
	if err := utx.Sign(acc.key); err != nil {
		return ethtypes.TxHash{}, err
	}
	return acc.client.PublishSignedTx(ctx, utx)
}

func isNonceError(err error) bool {
	return (strings.Compare(err.Error(), core.ErrReplaceUnderpriced.Error()) == 0 ||
		strings.Compare(err.Error(), core.ErrNonceTooHigh.Error()) == 0 ||
		strings.Compare(err.Error(), core.ErrNonceTooLow.Error()) == 0)
}
