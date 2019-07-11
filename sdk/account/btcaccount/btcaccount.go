package btcaccount

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

// ErrInsufficientBalance returns an error which returned when account doesn't have enough funds to make the tx.
func ErrInsufficientBalance(expect, have string) error {
	return fmt.Errorf("insufficient balance, got = %v, have = %v", expect, have)
}

// Account provides functions for interacting with a bitcoin account
type Account interface {
	Address() btctypes.Address
	PrivateKey() *ecdsa.PrivateKey
	Transfer(to btctypes.Address, value btctypes.Amount, speed types.TxSpeed) (btctypes.TxHash, error)
	UTXOs() (utxos btctypes.UTXOs, err error)
}

// account is a bitcoin wallet which can transfer funds and building tx.
type account struct {
	Client btcclient.Client

	address    btctypes.Address
	logger     logrus.FieldLogger
	key        *ecdsa.PrivateKey
	gasStation BtcGasStation
}

// New returns a new Account from the given private key.
func New(logger logrus.FieldLogger, client btcclient.Client, key *ecdsa.PrivateKey) (Account, error) {
	if key == nil {
		panic("cannot create account with nil key")
	}
	address, err := btctypes.AddressFromPubKey(&key.PublicKey, client.Network())
	if err != nil {
		return &account{}, err
	}
	return &account{
		Client:     client,
		address:    address,
		logger:     logger,
		key:        key,
		gasStation: NewBtcGasStation(logger, 10*time.Minute),
	}, nil
}

// NewAccountFromWIF returns a new Account from the given WIF
func NewAccountFromWIF(logger logrus.FieldLogger, client btcclient.Client, wifStr string) (Account, error) {
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, err
	}
	privKey := (*ecdsa.PrivateKey)(wif.PrivKey)
	return New(logger, client, privKey)
}

// RandomAccount returns a new Account using a random private key.
func RandomAccount(logger logrus.FieldLogger, client btcclient.Client) (Account, error) {
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return New(logger, client, key)
}

// Address returns the Address of the account
func (acc *account) Address() btctypes.Address {
	return acc.address
}

// Address returns the Address of the account
func (acc *account) PrivateKey() *ecdsa.PrivateKey {
	return acc.key
}

// UTXOs returns the UTXOs for an imported account.
func (acc *account) UTXOs() (utxos btctypes.UTXOs, err error) {
	return acc.Client.UTXOsFromAddress(acc.address)
}

// Transfer transfer certain amount value to the target address. Important: this only works for accounts that have been
// imported into the Bitcoin node.
func (acc *account) Transfer(to btctypes.Address, value btctypes.Amount, speed types.TxSpeed) (btctypes.TxHash, error) {
	utxos, err := acc.UTXOs()
	if err != nil {
		return "", err
	}

	fee := acc.gasStation.CalculateGasAmount(context.Background(), speed, acc.Client.EstimateTxSize(len(utxos), 2))
	// Check if we have enough funds
	balance := utxos.Sum()
	if balance < value+fee {
		return "", ErrInsufficientBalance(fmt.Sprintf("%v", value), fmt.Sprintf("%v", balance))
	}

	// todo : select some utxos from all the utxos we have.
	tx, err := acc.Client.BuildUnsignedTx(utxos, btctypes.Recipients{{Address: to, Amount: value}}, acc.Address(), fee)

	if err != nil {
		return "", err
	}

	if err := tx.Sign(acc.key); err != nil {
		return "", err
	}

	// Submit the signed tx
	return acc.Client.SubmitSignedTx(tx)
}
