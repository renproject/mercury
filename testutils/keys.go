package testutils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/pkg/errors"
	"github.com/renproject/mercury/testutils/hdutil"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/mercury/types/zectypes"
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

// EcdsaKey return the ECDSA key on the given path of the HD key.
func (hdkey HdKey) BTCAddress(path ...uint32) (btctypes.Address, error) {
	key, err := hdkey.EcdsaKey(path...)
	if err != nil {
		return nil, err
	}
	return btctypes.AddressFromPubKey(&key.PublicKey, hdkey.network)
}

// EcdsaKey return the ECDSA key on the given path of the HD key.
func (hdkey HdKey) ZECAddress(path ...uint32) (zectypes.Address, error) {
	key, err := hdkey.EcdsaKey(path...)
	if err != nil {
		return nil, err
	}
	return zectypes.AddressFromPubKey(&key.PublicKey, hdkey.network)
}

func RandomBTCAddress(network btctypes.Network) (btctypes.Address, error) {
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return btctypes.AddressFromPubKey(&key.PublicKey, network)
}

func RandomZECAddress(network btctypes.Network) (btctypes.Address, error) {
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return zectypes.AddressFromPubKey(&key.PublicKey, network)
}

// TODO : need to be fixed, the stx generated from this tx is not valid at the moment.
func GenerateSignedTx(network btctypes.Network, key *ecdsa.PrivateKey, destination string, amount int64, txHash string) ([]byte, error) {
	wif, err := btcutil.NewWIF((*btcec.PrivateKey)(key), network.Params(), network == btctypes.Mainnet)
	if err != nil {
		return nil, err
	}

	addresspubkey, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeUncompressed(), &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	sourceTx := wire.NewMsgTx(wire.TxVersion)
	sourceUtxoHash, _ := chainhash.NewHashFromStr(txHash)
	sourceUtxo := wire.NewOutPoint(sourceUtxoHash, 0)
	sourceTxIn := wire.NewTxIn(sourceUtxo, nil, nil)
	destinationAddress, err := btcutil.DecodeAddress(destination, &chaincfg.MainNetParams)
	sourceAddress, err := btcutil.DecodeAddress(addresspubkey.EncodeAddress(), &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	destinationPkScript, _ := txscript.PayToAddrScript(destinationAddress)
	sourcePkScript, _ := txscript.PayToAddrScript(sourceAddress)
	sourceTxOut := wire.NewTxOut(amount, sourcePkScript)
	sourceTx.AddTxIn(sourceTxIn)
	sourceTx.AddTxOut(sourceTxOut)
	sourceTxHash := sourceTx.TxHash()
	redeemTx := wire.NewMsgTx(wire.TxVersion)
	prevOut := wire.NewOutPoint(&sourceTxHash, 0)
	redeemTxIn := wire.NewTxIn(prevOut, nil, nil)
	redeemTx.AddTxIn(redeemTxIn)
	redeemTxOut := wire.NewTxOut(amount, destinationPkScript)
	redeemTx.AddTxOut(redeemTxOut)
	sigScript, err := txscript.SignatureScript(redeemTx, 0, sourceTx.TxOut[0].PkScript, txscript.SigHashAll, wif.PrivKey, false)
	if err != nil {
		return nil, err
	}
	redeemTx.TxIn[0].SignatureScript = sigScript
	flags := txscript.StandardVerifyFlags
	vm, err := txscript.NewEngine(sourceTx.TxOut[0].PkScript, redeemTx, 0, flags, nil, nil, amount)
	if err != nil {
		return nil, err
	}
	if err := vm.Execute(); err != nil {
		return nil, err
	}
	var unsignedTx bytes.Buffer
	var signedTx bytes.Buffer
	sourceTx.Serialize(&unsignedTx)
	redeemTx.Serialize(&signedTx)

	return signedTx.Bytes(), nil
}
