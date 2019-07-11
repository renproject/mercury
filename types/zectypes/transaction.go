package zectypes

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/iqoption/zecutil"
	"github.com/renproject/mercury/types/btctypes"
)

type SignatureHash = btctypes.SignatureHash

type SerializedPubKey = btctypes.SerializedPubKey

type Signature = btctypes.Signature

type TxHash = btctypes.TxHash

type ScriptData []byte

type Tx struct {
	network   Network
	tx        *zecutil.MsgTx
	sigHashes []SignatureHash
	utxos     UTXOs
	signed    bool
}

func NewUnsignedTx(network Network, utxos UTXOs, tx *zecutil.MsgTx) Tx {
	return Tx{
		network:   network,
		tx:        tx,
		sigHashes: []SignatureHash{},
		utxos:     utxos,
		signed:    false,
	}
}

func (tx *Tx) Tx() *zecutil.MsgTx {
	return tx.tx
}

func (tx *Tx) Hash() TxHash {
	return TxHash(tx.tx.TxHash().String())
}

func (tx *Tx) IsSigned() bool {
	return tx.signed
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
	serializedPK := SerializePublicKey(&key.PublicKey, tx.network)
	scriptData := make([]ScriptData, len(sigs))
	return tx.InjectSignaturesWithData(sigs, serializedPK, scriptData)
}

// InjectSignaturesWithData injects the signed signatureHashes into the Tx. You cannot use the USTX anymore.
// scriptData is additional data to be appended to the signature script, it can be nil or an empty byte array.
func (tx *Tx) InjectSignaturesWithData(sigs []*btcec.Signature, serializedPubKey SerializedPubKey, scriptData []ScriptData) error {
	// Pre-condition checks
	if tx.IsSigned() {
		panic("pre-condition violation: cannot inject signatures into signed transaction")
	}
	if tx.tx == nil {
		panic("pre-condition violation: cannot inject signatures into nil transaction")
	}
	if len(sigs) <= 0 {
		panic("pre-condition violation: cannot inject empty signatures")
	}
	if len(sigs) != len(tx.tx.TxIn) {
		panic(fmt.Errorf("pre-condition violation: expected signature len=%v to equal transaction input len=%v", len(sigs), len(tx.tx.TxIn)))
	}
	if len(sigs) != len(scriptData) {
		panic(fmt.Errorf("pre-condition violation: expected scriptData len=%v to equal signature len=%v", len(scriptData), len(sigs)))
	}
	if len(serializedPubKey) <= 0 {
		panic("pre-condition violation: cannot inject signatures with empty pubkey")
	}

	for i, sig := range sigs {
		builder := txscript.NewScriptBuilder()
		builder.AddData(append(sig.Serialize(), byte(txscript.SigHashAll)))
		builder.AddData(serializedPubKey)
		if scriptData[i] != nil && len(scriptData[i]) > 0 {
			builder.AddData(scriptData[i])
		}
		sigScript, err := builder.Script()
		if err != nil {
			return err
		}
		tx.tx.TxIn[i].SignatureScript = sigScript
	}
	tx.signed = true
	return nil
}

func (tx *Tx) AppendSignatureHash(script []byte, mode txscript.SigHashType, amt int64) error {
	sigHash, err := CalcSignatureHash(script, mode, tx.tx, len(tx.sigHashes), amt)
	if err != nil {
		return err
	}
	tx.sigHashes = append(tx.sigHashes, sigHash)
	return nil
}

func (tx *Tx) ReplaceSignatureHash(script []byte, mode txscript.SigHashType, i int, amt int64) (err error) {
	if i < 0 || i >= len(tx.sigHashes) {
		panic(fmt.Errorf("pre-condition violation: signature hash index=%v is out of range", i))
	}
	tx.sigHashes[i], err = CalcSignatureHash(script, mode, tx.tx, i, amt)
	return
}

// SignatureHashes returns a list of signature hashes need to be signed.
func (tx *Tx) SignatureHashes() []SignatureHash {
	return tx.sigHashes
}

// Serialize returns the serialized tx in bytes.
func (tx *Tx) Serialize() []byte {
	// Pre-condition checks
	if tx.tx == nil {
		panic("pre-condition violation: cannot serialize nil transaction")
	}

	buf := bytes.NewBuffer([]byte{})
	if err := tx.tx.Serialize(buf); err != nil {
		return nil
	}
	bufBytes := buf.Bytes()

	// Post-condition checks
	if len(bufBytes) <= 0 {
		panic(fmt.Errorf("post-condition violation: serialized transaction len=%v is invalid", len(bufBytes)))
	}
	return bufBytes
}

func (tx *Tx) Verify() error {
	for i, utxo := range tx.utxos {
		scriptPubKey, err := hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			return err
		}

		engine, err := txscript.NewEngine(scriptPubKey, tx.tx.MsgTx, i,
			txscript.StandardVerifyFlags, txscript.NewSigCache(10),
			txscript.NewTxSigHashes(tx.tx.MsgTx), int64(utxo.Amount))
		if err != nil {
			return err
		}
		if err := engine.Execute(); err != nil {
			return err
		}
	}
	return nil
}