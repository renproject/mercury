package btctypes

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/renproject/mercury/types"
)

// Amount represents bitcoin value in Satoshi (1e-8 Bitcoin).
type Amount int64

const (
	SAT = Amount(1)
	BTC = Amount(1e8 * SAT)
)

// Network of Bitcoin blockchain.
type Network uint8

const (
	Localnet Network = 0
	Mainnet  Network = 1
	Testnet  Network = 2
)

// NewNetwork parse the network from a string.
func NewNetwork(network string) Network {
	network = strings.ToLower(strings.TrimSpace(network))
	switch network {
	case "mainnet":
		return Mainnet
	case "testnet", "testnet3":
		return Testnet
	case "localnet", "localhost":
		return Localnet
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// Params returns the params config for the network
func (network Network) Params() *chaincfg.Params {
	switch network {
	case Mainnet:
		return &chaincfg.MainNetParams
	case Testnet, Localnet:
		return &chaincfg.TestNet3Params
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// String implements the `Stringer` interface.
func (network Network) String() string {
	switch network {
	case Mainnet:
		return "mainnet"
	case Testnet:
		return "testnet"
	case Localnet:
		return "localnet"
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// Addr is an interface type for any type of destination a transaction output may spend to. This includes pay-to-pubkey
// (P2PK), pay-to-pubkey-hash (P2PKH), and pay-to-script-hash (P2SH). Address is designed to be generic enough that
// other kinds of addresses may be added in the future without changing the decoding and encoding API.
type Address btcutil.Address

// AddressFromBase58 decodes the base58 encoding bitcoin address to a `Addr`.
func AddressFromBase58(addr string, network Network) (Address, error) {
	return btcutil.DecodeAddress(addr, network.Params())
}

// AddressFromPubKey gets the `Addr` from a public key.
func AddressFromPubKey(pubkey *ecdsa.PublicKey, network Network) (Address, error) {
	addr, err := btcutil.NewAddressPubKey(SerializePublicKey(pubkey, network), network.Params())
	if err != nil {
		return nil, fmt.Errorf("cannot decode address from public key: %v", err)
	}

	return btcutil.DecodeAddress(addr.EncodeAddress(), network.Params())
}

// SerializePublicKey serializes the public key to bytes.
func SerializePublicKey(pubkey *ecdsa.PublicKey, network Network) []byte {
	switch network {
	case Mainnet:
		return (*btcec.PublicKey)(pubkey).SerializeCompressed()
	case Testnet, Localnet:
		return (*btcec.PublicKey)(pubkey).SerializeUncompressed()
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// RandomAddress returns a random Addr on given network.
func RandomAddress(network Network) (Address, error) {
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return AddressFromPubKey(&key.PublicKey, network)
}

type UTXO struct {
	TxHash       string `json:"txHash"`
	Amount       Amount `json:"amount"`
	ScriptPubKey string `json:"scriptPubKey"`
	Vout         uint32 `json:"vout"`
}

type UTXOs []UTXO

func (utxos *UTXOs) Sum() Amount {
	total := Amount(0)
	for _, utxo := range *utxos {
		total += Amount(utxo.Amount)
	}
	return total
}

type SignatureHash []byte

type SerializedPubKey []byte

type Signature btcec.Signature

type Tx struct {
	Network   Network
	Tx        *wire.MsgTx
	SigHashes []SignatureHash
	Signed    bool
}

func NewUnsignedTx(network Network, tx *wire.MsgTx) Tx {
	return Tx{
		Network:   network,
		Tx:        tx,
		SigHashes: []SignatureHash{},
		Signed:    false,
	}
}

func (tx *Tx) IsSigned() bool {
	return tx.Signed
}

func (tx *Tx) Sign(key *ecdsa.PrivateKey) (err error) {
	subScripts := tx.SignatureHashes()
	sigs := make([]*btcec.Signature, len(subScripts))
	for i, subScript := range subScripts {
		sigs[i], err = (*btcec.PrivateKey)(key).Sign(subScript)
		if err != nil {
			return err
		}
	}
	serializedPK := SerializePublicKey(&key.PublicKey, tx.Network)
	return tx.InjectSignatures(sigs, serializedPK)
}

// InjectSignatures injects the signed signatureHashes into the Tx. You cannot use the USTX anymore.
func (tx *Tx) InjectSignatures(sigs []*btcec.Signature, serializedPubKey SerializedPubKey) error {
	// Pre-condition checks
	if tx.IsSigned() {
		panic("pre-condition violation: cannot inject signatures into signed transaction")
	}
	if tx.Tx == nil {
		panic("pre-condition violation: cannot inject signatures into nil transaction")
	}
	if len(sigs) <= 0 {
		panic("pre-condition violation: cannot inject empty signatures")
	}
	if len(sigs) != len(tx.Tx.TxIn) {
		panic(fmt.Errorf("pre-condition violation: expected signature len=%v to equal transaction input len=%v", len(sigs), len(tx.Tx.TxIn)))
	}
	if len(serializedPubKey) <= 0 {
		panic("pre-condition violation: cannot inject signatures with empty pubkey")
	}

	for i, sig := range sigs {
		builder := txscript.NewScriptBuilder()
		builder.AddData(append(sig.Serialize(), byte(txscript.SigHashAll)))
		builder.AddData(serializedPubKey)
		sigScript, err := builder.Script()
		if err != nil {
			return err
		}
		tx.Tx.TxIn[i].SignatureScript = sigScript
	}
	tx.Signed = true
	return nil
}

func (tx *Tx) AppendSignatureHash(script []byte, mode txscript.SigHashType) error {
	sigHash, err := txscript.CalcSignatureHash(script, mode, tx.Tx, len(tx.SigHashes))
	if err != nil {
		return err
	}
	tx.SigHashes = append(tx.SigHashes, sigHash)
	return nil
}

func (tx *Tx) ReplaceSignatureHash(script []byte, mode txscript.SigHashType, i int) (err error) {
	if i < 0 || i >= len(tx.SigHashes) {
		panic(fmt.Errorf("pre-condition violation: signature hash index=%v is out of range", i))
	}
	tx.SigHashes[i], err = txscript.CalcSignatureHash(script, mode, tx.Tx, i)
	return
}

// SignatureHashes returns a list of signature hashes need to be signed.
func (tx *Tx) SignatureHashes() []SignatureHash {
	return tx.SigHashes
}

// Serialize returns the serialized tx in bytes.
func (tx *Tx) Serialize() []byte {
	// Pre-condition checks
	if tx.Tx == nil {
		panic("pre-condition violation: cannot serialize nil transaction")
	}

	buf := bytes.NewBuffer([]byte{})
	if err := tx.Tx.Serialize(buf); err != nil {
		return nil
	}
	bufBytes := buf.Bytes()

	// Post-condition checks
	if len(bufBytes) <= 0 {
		panic(fmt.Errorf("post-condition violation: serialized transaction len=%v is invalid", len(bufBytes)))
	}
	return bufBytes
}

type Recipient struct {
	Address Address
	Amount  Amount
}

type Recipients []Recipient

type TxHash string

type Confirmations int64
