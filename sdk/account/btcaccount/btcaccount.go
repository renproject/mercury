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

type Account interface {
	Address() btctypes.Address
	Transfer(ctx context.Context, to btctypes.Address, value btctypes.Amount, fee int64) error
}

// account is a bitcoin wallet which can transfer funds and building tx.
type account struct {
	Client *btcclient.Client

	address btctypes.Address
	logger  logrus.FieldLogger
	key     *ecdsa.PrivateKey
}

// New returns a new Account from the given private key.
func New(logger logrus.FieldLogger, client *btcclient.Client, key *ecdsa.PrivateKey) (Account, error) {
	if key == nil {
		panic("cannot create account with nil key")
	}
	address, err := btctypes.AddressFromPubKey(&key.PublicKey, client.Network)
	if err != nil {
		return &account{}, err
	}
	return &account{
		Client:  client,
		address: address,
		logger:  logger,
		key:     key,
	}, nil
}

// NewAccountFromWIF returns a new Account from the given WIF
func NewAccountFromWIF(logger logrus.FieldLogger, client *btcclient.Client, wifStr string) (Account, error) {
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, err
	}
	privKey := (*ecdsa.PrivateKey)(wif.PrivKey)
	return &account{
		Client: client,
		logger: logger,
		key:    privKey,
	}, nil
}

// Address returns the Address of the account
func (acc *account) Address() btctypes.Address {
	return acc.address
}

// Transfer transfer certain amount value to the target address.
func (acc *account) Transfer(ctx context.Context, to btctypes.Address, value btctypes.Amount, fee int64) error {
	utxos, err := acc.Client.UTXOs(ctx, acc.address, btcclient.MaxUTXOLimit, btcclient.MinConfirmations)
	if err != nil {
		return err
	}

	// Check if we have enough funds
	balance, err := acc.Client.Balance(ctx, acc.address, btcclient.MaxUTXOLimit, btcclient.MinConfirmations)
	if err != nil {
		return err
	}
	if balance < value {
		return ErrInsufficientBalance(fmt.Sprintf("%v", value), fmt.Sprintf("%v", balance))
	}

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

	// Submit the signed tx
	return acc.Client.SubmitSignedTx(ctx, tx)
}
