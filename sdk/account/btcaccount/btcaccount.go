package btcaccount

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

// ErrInsufficientBalance returns an error which returned when account doesn't have enough funds to make the tx.
func ErrInsufficientBalance(expect, have string) error {
	return fmt.Errorf("insufficient balance, got = %v, have = %v", expect, have)
}

// Account is a bitcoin wallet which can transfer funds and building tx.
type Account struct {
	Client *btcclient.Client

	logger logrus.FieldLogger
	key    *ecdsa.PrivateKey
}

// NewAccount returns a new Account from the given private key.
func NewAccount(logger logrus.FieldLogger, client *btcclient.Client, key *ecdsa.PrivateKey) *Account {
	return &Account{
		Client: client,
		logger: logger,
		key:    key,
	}
}

// NewAccountFromWIF returns a new Account from the given WIF
func NewAccountFromWIF(logger logrus.FieldLogger, client *btcclient.Client, wifStr string) (*Account, error) {
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, err
	}
	privKey := (*ecdsa.PrivateKey)(wif.PrivKey)
	return &Account{
		Client: client,
		logger: logger,
		key:    privKey,
	}, nil
}

// Address returns the Address of the account
func (acc *Account) Address() (btctypes.Address, error) {
	return btctypes.AddressFromPubKey(&acc.key.PublicKey, acc.Client.Network)
}

// Transfer transfer certain amount value to the target address.
func (acc *Account) Transfer(ctx context.Context, to btctypes.Address, value btctypes.Amount, fee int64) error {
	log.Print(1)
	// Get all utxos owned by the acc
	address, err := acc.Address()
	if err != nil {
		return err
	}
	utxos, err := acc.Client.UTXOs(ctx, address, 999999, 0)
	if err != nil {
		return err
	}

	log.Print(2)
	// Check if we have enough funds
	balance, err := acc.Client.Balance(ctx, address, btcclient.MaxUTXOLimit, 0)
	if err != nil {
		return err
	}
	if balance < value {
		return ErrInsufficientBalance(fmt.Sprintf("%v", value), fmt.Sprintf("%v", balance))
	}
	log.Print(3)

	// todo : select some utxos from all the utxos we have.
	tx, err := acc.Client.BuildUnsignedTx(utxos, btctypes.Recipient{to, value})

	if err != nil {
		return err
	}

	subScripts := tx.SignatureHashes()
	sigs := make([]*btcec.Signature, len(subScripts))

	for i, subScript := range subScripts {
		sigs[i], err = (*btcec.PrivateKey)(acc.key).Sign(subScript)
		if err != nil {
			return err
		}
	}
	serializedPK := btctypes.SerializePublicKey(&acc.key.PublicKey, acc.Client.Network)
	if err := tx.InjectSignatures(sigs, serializedPK); err != nil {
		if err != nil {
			return err
		}
	}

	log.Print("stx = ", hex.EncodeToString(tx.Serialize()))

	return nil

	// Submit the signed tx
	return acc.Client.SubmitSTX(ctx, tx.Serialize())
}
