package testutil

import (
	"crypto/ecdsa"
	"crypto/rand"
	"os"

	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/pkg/errors"
	"github.com/renproject/mercury/testutil/hdutil"
	"github.com/renproject/mercury/types/btctypes"
)

// ErrInvalidMnemonic is returned when the mnemonic is invalid.
var ErrInvalidMnemonic = errors.New("invalid mnemonic")

// HdKey is a hierarchical deterministic extended key.
type HdKey struct {
	ExtendedKey *hdkeychain.ExtendedKey
	network     btctypes.Network
}

// LoadHdWalletFromEnv loads the mnemonic and passphrase from environment variables and generate a HdKey from that.
func LoadHdWalletFromEnv(mnemonicEnv, passphraseEnv string, network btctypes.Network) (HdKey, error) {
	mnemonic, passphrase := os.Getenv(mnemonicEnv), os.Getenv(passphraseEnv)
	if mnemonic == "" {
		return HdKey{}, ErrInvalidMnemonic
	}
	return LoadHdWallet(mnemonic, passphrase, network)
}

// LoadHdWallet generates a HdKey from the given mnemonic and passphrase.
func LoadHdWallet(mnemonic, passphrase string, network btctypes.Network) (HdKey, error) {
	key, err := hdutil.DeriveExtendedPrivKey(mnemonic, passphrase, network)
	if err != nil {
		return HdKey{}, err
	}
	return HdKey{
		ExtendedKey: key,
		network:     network,
	}, err
}

// EcdsaKey return the ECDSA key on the given path of the HD key.
func (hdkey HdKey) EcdsaKey(path ...uint32) (*ecdsa.PrivateKey, error) {
	return hdutil.DerivePrivKey(hdkey.ExtendedKey, path...)
}

// Address return the Address of the HD key with the current path.
func (hdkey HdKey) Address(path ...uint32) (btctypes.Address, error) {
	key, err := hdkey.EcdsaKey(path...)
	if err != nil {
		return nil, err
	}
	return btctypes.AddressFromPubKey(key.PublicKey, hdkey.network)
}

// SegWitAddress return the SegWitAddress of the HD key with the current path.
func (hdkey HdKey) SegWitAddress(path ...uint32) (btctypes.Address, error) {
	key, err := hdkey.EcdsaKey(path...)
	if err != nil {
		return nil, err
	}
	return btctypes.SegWitAddressFromPubKey(key.PublicKey, hdkey.network)
}

func RandomAddress(network btctypes.Network) (btctypes.Address, error) {
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return btctypes.AddressFromPubKey(key.PublicKey, network)
}

func RandomSegWitAddress(network btctypes.Network) (btctypes.Address, error) {
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return btctypes.SegWitAddressFromPubKey(key.PublicKey, network)
}
