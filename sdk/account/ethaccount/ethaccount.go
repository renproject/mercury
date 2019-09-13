package ethaccount

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/ethtypes"
)

type Account interface {
	Client() ethclient.Client
	Address() ethtypes.Address
	Balance(ctx context.Context) (ethtypes.Amount, error)
	Transfer(ctx context.Context, toAddress ethtypes.Address, value ethtypes.Amount, gasPrice ethtypes.Amount) (types.TxHash, error)
	Transact(ctx context.Context, toAddress ethtypes.Address, value ethtypes.Amount, gasPrice ethtypes.Amount, gasLimit uint64, data []byte) (types.TxHash, error)
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

func (acc *account) Transfer(ctx context.Context, toAddress ethtypes.Address, value ethtypes.Amount, gasPrice ethtypes.Amount) (types.TxHash, error) {
	return acc.Transact(ctx, toAddress, value, gasPrice, 21000, nil)
}

func (acc *account) Transact(ctx context.Context, toAddress ethtypes.Address, value ethtypes.Amount, gasPrice ethtypes.Amount, gasLimit uint64, data []byte) (types.TxHash, error) {
	nonce, err := acc.client.PendingNonceAt(ctx, acc.address)
	// fmt.Printf("nonce fetched back from infura: %v", nonce)
	if err != nil {
		return types.TxHash(""), fmt.Errorf("failed to get pending nonce: %v", err)
	}
	tx, err := acc.client.BuildUnsignedTx(ctx, nonce, toAddress, value, gasLimit, gasPrice, data)
	if err != nil {
		return types.TxHash(""), err
	}
	if err := tx.Sign(acc.key); err != nil {
		return types.TxHash(""), err
	}
	return acc.client.PublishSignedTx(ctx, tx)
}

func (acc *account) BuildUnsignedTx(ctx context.Context, toAddress ethtypes.Address, value ethtypes.Amount, gasLimit uint64, gasPrice ethtypes.Amount, data []byte) (ethtypes.Tx, error) {
	nonce, err := acc.client.PendingNonceAt(ctx, acc.address)
	// fmt.Printf("nonce fetched back from infura: %v", nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending nonce: %v", err)
	}
	return acc.client.BuildUnsignedTx(ctx, nonce, toAddress, value, gasLimit, gasPrice, data)
}

func (acc *account) Balance(ctx context.Context) (ethtypes.Amount, error) {
	return acc.client.Balance(ctx, acc.Address())
}

func (acc *account) Client() ethclient.Client {
	return acc.client
}

func (acc *account) SignUnsignedTx(ctx context.Context, utx ethtypes.Tx) error {
	return utx.Sign(acc.key)
}

func (acc *account) Address() ethtypes.Address {
	return acc.address
}

func (acc *account) PrivateKey() *ecdsa.PrivateKey {
	return acc.key
}
