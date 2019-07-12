package btctypes

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

type SignatureHash []byte

type SerializedPubKey []byte

type ScriptData []byte

type Signature btcec.Signature

type TxHash string

type Tx interface {
	Hash() TxHash
	IsSigned() bool
	SignatureHashes() []SignatureHash
	InjectSignatures(sigs []*btcec.Signature, serializedPubKey SerializedPubKey) error
	Serialize() ([]byte, error)
	Tx() *wire.MsgTx
	UTXOs() UTXOs
	Sign(key *ecdsa.PrivateKey) (err error)
}

type tx struct {
	network   Network
	tx        *wire.MsgTx
	sigHashes []SignatureHash
	utxos     UTXOs
	signed    bool
}

func NewUnsignedTx(network Network, utxos UTXOs, msgtx *wire.MsgTx) (Tx, error) {
	t := tx{
		network:   network,
		tx:        msgtx,
		sigHashes: []SignatureHash{},
		utxos:     utxos,
		signed:    false,
	}

	for i, utxo := range utxos {
		sigHash, err := utxo.SigHash(txscript.SigHashAll, msgtx, i)
		if err != nil {
			return nil, err
		}
		t.sigHashes = append(t.sigHashes, sigHash)
	}
	return &t, nil
}

func (t *tx) Tx() *wire.MsgTx {
	return t.tx
}

func (t *tx) Hash() TxHash {
	return TxHash(t.tx.TxHash().String())
}

func (t *tx) IsSigned() bool {
	return t.signed
}

func (t *tx) UTXOs() UTXOs {
	return t.utxos
}

func (t *tx) Sign(key *ecdsa.PrivateKey) (err error) {
	subScripts := t.SignatureHashes()
	sigs := make([]*btcec.Signature, len(subScripts))
	for i, subScript := range subScripts {
		sigs[i], err = (*btcec.PrivateKey)(key).Sign(subScript)
		if err != nil {
			return err
		}
	}
	serializedPK := SerializePublicKey(&key.PublicKey, t.network)
	return t.InjectSignatures(sigs, serializedPK)
}

// InjectSignaturesWithData injects the signed signatureHashes into the Tx. You cannot use the USTX anymore.
// scriptData is additional data to be appended to the signature script, it can be nil or an empty byte array.
func (t *tx) InjectSignatures(sigs []*btcec.Signature, serializedPubKey SerializedPubKey) error {
	// Pre-condition checks
	if t.IsSigned() {
		panic("pre-condition violation: cannot inject signatures into signed transaction")
	}
	if t.tx == nil {
		panic("pre-condition violation: cannot inject signatures into nil transaction")
	}
	if len(sigs) <= 0 {
		panic("pre-condition violation: cannot inject empty signatures")
	}
	if len(sigs) != len(t.tx.TxIn) {
		panic(fmt.Errorf("pre-condition violation: expected signature len=%v to equal transaction input len=%v", len(sigs), len(t.tx.TxIn)))
	}
	if len(serializedPubKey) <= 0 {
		panic("pre-condition violation: cannot inject signatures with empty pubkey")
	}

	for i, sig := range sigs {
		builder := txscript.NewScriptBuilder()
		builder.AddData(append(sig.Serialize(), byte(txscript.SigHashAll)))
		builder.AddData(serializedPubKey)
		t.utxos[i].AddData(builder)
		sigScript, err := builder.Script()
		if err != nil {
			return err
		}
		t.tx.TxIn[i].SignatureScript = sigScript
	}
	t.signed = true
	return nil
}

func (t *tx) ReplaceSignatureHash(script []byte, mode txscript.SigHashType, i int) (err error) {
	if i < 0 || i >= len(t.sigHashes) {
		panic(fmt.Errorf("pre-condition violation: signature hash index=%v is out of range", i))
	}
	t.sigHashes[i], err = txscript.CalcSignatureHash(script, mode, t.tx, i)
	return
}

// SignatureHashes returns a list of signature hashes need to be signed.
func (t *tx) SignatureHashes() []SignatureHash {
	return t.sigHashes
}

// Serialize returns the serialized tx in bytes.
func (t *tx) Serialize() ([]byte, error) {
	// Pre-condition checks
	if t.tx == nil {
		panic("pre-condition violation: cannot serialize nil transaction")
	}

	buf := bytes.NewBuffer([]byte{})
	if err := t.tx.Serialize(buf); err != nil {
		return nil, err
	}
	bufBytes := buf.Bytes()

	// Post-condition checks
	if len(bufBytes) <= 0 {
		panic(fmt.Errorf("post-condition violation: serialized transaction len=%v is invalid", len(bufBytes)))
	}
	return bufBytes, nil
}
