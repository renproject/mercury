package btcaccount

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
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
	PrivateKey() *ecdsa.PrivateKey
	Transfer(ctx context.Context, to btctypes.Address, value btctypes.Amount, fee btctypes.Amount) error
	Balance(ctx context.Context) (value btctypes.Amount, err error)
	UTXOs(ctx context.Context) (utxos []btctypes.UTXO, err error)
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
	return New(logger, client, privKey)
}

// RandomAccount returns a new Account using a random private key.
func RandomAccount(logger logrus.FieldLogger, client *btcclient.Client) (Account, error) {
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

// Transfer transfer certain amount value to the target address.
func (acc *account) Balance(ctx context.Context) (value btctypes.Amount, err error) {
	return acc.Client.Balance(ctx, acc.address, btcclient.MaxUTXOLimit, btcclient.MinConfirmations)
}

func (acc *account) UTXOs(ctx context.Context) (utxos []btctypes.UTXO, err error) {
	return acc.Client.UTXOs(ctx, acc.address, btcclient.MaxUTXOLimit, btcclient.MinConfirmations)
}

// Transfer transfer certain amount value to the target address.
func (acc *account) Transfer(ctx context.Context, to btctypes.Address, value btctypes.Amount, fee btctypes.Amount) error {
	utxos, err := acc.UTXOs(ctx)
	if err != nil {
		return err
	}

	// Check if we have enough funds
	balance, err := acc.Balance(ctx)
	if err != nil {
		return err
	}
	if balance < value+fee {
		return ErrInsufficientBalance(fmt.Sprintf("%v", value), fmt.Sprintf("%v", balance))
	}

	// todo : select some utxos from all the utxos we have.
	tx, err := acc.Client.BuildUnsignedTx(utxos, btctypes.Recipient{to, value}, btctypes.Recipient{Address: acc.Address(), Amount: balance - value - fee})

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
