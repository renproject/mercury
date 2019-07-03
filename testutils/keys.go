package testutils

import (
	"bytes"
	"crypto/ecdsa"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/pkg/errors"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/tyler-smith/go-bip39"
)

// ErrInvalidMnemonic is returned when the mnemonic is invalid.
var ErrInvalidMnemonic = errors.New("invalid mnemonic")

// HdKey is a hierarchical deterministic extended key.
type HdKey struct {
	*hdkeychain.ExtendedKey
}

// LoadHdWalletFromEnv loads the mnemonic and passphrase from environment variables and generate a HdKey from that.
func LoadHdWalletFromEnv(mnemonicEnv, passphraseEnv string) (HdKey, error) {
	mnemonic, passphrase := os.Getenv(mnemonicEnv), os.Getenv(passphraseEnv)
	if mnemonic == "" {
		return HdKey{}, ErrInvalidMnemonic
	}
	seed := bip39.NewSeed(mnemonic, passphrase)
	key, err := hdkeychain.NewMaster(seed, &chaincfg.TestNet3Params)
	return HdKey{
		ExtendedKey: key,
	}, err
}

// LoadHdWallet generates a HdKey from the given mnemonic and passphrase.
func LoadHdWallet(mnemonic, passphrase string) (HdKey, error) {
	seed := bip39.NewSeed(mnemonic, passphrase)
	key, err := hdkeychain.NewMaster(seed, &chaincfg.TestNet3Params)
	return HdKey{
		ExtendedKey: key,
	}, err
}

// EcdsaKey return the ECDSA key on the given path of the HD key.
func (hdkey HdKey) EcdsaKey(path ...uint32) (*ecdsa.PrivateKey, error) {
	var key *hdkeychain.ExtendedKey
	var err error
	for _, val := range path {
		key, err = hdkey.Child(val)
		if err != nil {
			return nil, err
		}
	}
	privKey, err := key.ECPrivKey()
	if err != nil {
		return nil, err
	}
	return privKey.ToECDSA(), nil
}

// EcdsaKey return the ECDSA key on the given path of the HD key.
func (hdkey HdKey) Address(network btctypes.Network, path ...uint32) (btctypes.Address, error) {
	var key *hdkeychain.ExtendedKey
	var err error
	for _, val := range path {
		key, err = hdkey.Child(val)
		if err != nil {
			return nil, err
		}
	}
	address, err := key.Address(network.Params())
	if err != nil {
		return nil, err
	}
	addressStr := address.String()
	return btctypes.AddressFromBase58(addressStr, network)
}

// TODO : need to be fixed, the stx generated from this tx is not valid at the moment.
func GenerateSignedTx(network btctypes.Network, key *ecdsa.PrivateKey, destination string, amount int64, txHash string) (btctypes.Tx, error) {
	wif, err := btcutil.NewWIF((*btcec.PrivateKey)(key), network.Params(), network == btctypes.Mainnet)
	if err != nil {
		return btctypes.Tx{}, err
	}

	addresspubkey, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeUncompressed(), &chaincfg.MainNetParams)
	if err != nil {
		return btctypes.Tx{}, err
	}
	sourceTx := wire.NewMsgTx(wire.TxVersion)
	sourceUtxoHash, _ := chainhash.NewHashFromStr(txHash)
	sourceUtxo := wire.NewOutPoint(sourceUtxoHash, 0)
	sourceTxIn := wire.NewTxIn(sourceUtxo, nil, nil)
	destinationAddress, err := btcutil.DecodeAddress(destination, &chaincfg.MainNetParams)
	sourceAddress, err := btcutil.DecodeAddress(addresspubkey.EncodeAddress(), &chaincfg.MainNetParams)
	if err != nil {
		return btctypes.Tx{}, err
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
		return btctypes.Tx{}, err
	}
	redeemTx.TxIn[0].SignatureScript = sigScript
	flags := txscript.StandardVerifyFlags
	vm, err := txscript.NewEngine(sourceTx.TxOut[0].PkScript, redeemTx, 0, flags, nil, nil, amount)
	if err != nil {
		return btctypes.Tx{}, err
	}
	if err := vm.Execute(); err != nil {
		return btctypes.Tx{}, err
	}
	var unsignedTx bytes.Buffer
	var signedTx bytes.Buffer
	sourceTx.Serialize(&unsignedTx)
	redeemTx.Serialize(&signedTx)

	return signedTx.Bytes(), nil
}
