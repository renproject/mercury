package btcaccount

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/renproject/mercury/types/btctypes"
)

// Recipient represents a receiver in the bitcoin transaction. It can be a publicKey address or a srciptHash address.
type Recipient struct {
	Addr   btctypes.Addr
	Amount btctypes.Amount
}

// Tx is a transaction in bitcoin blockchain which transfer values from addresses to addresses.
type Tx struct {
	network btctypes.Network
	tx *wire.MsgTx
	sigHashes    [][]byte
}

// SignatureHashes returns a list of signature hashes need to be signed.
func (tx *Tx) SignatureHashes() [][]byte {
	return tx.sigHashes
}

// InjectSignature injects the signed signatureHashes into the Tx.
func (tx *Tx) InjectSignature(sigs []*btcec.Signature, serializedPubKey []byte) error {
	for i, sig := range sigs {
		builder := txscript.NewScriptBuilder()
		builder.AddData(append(sig.Serialize(), byte(txscript.SigHashAll)))
		builder.AddData(serializedPubKey)
		sigScript, err := builder.Script()
		if err != nil {
			return err
		}
		tx.tx.TxIn[i].SignatureScript = sigScript
	}
	return nil
}

// Serialize returns the serialized tx in bytes.
func (tx *Tx) Serialize() []byte {
	if tx.tx == nil {
		return nil
	}

	buf := bytes.NewBuffer([]byte{})
	if err := tx.tx.Serialize(buf); err != nil {
		return nil
	}
	return buf.Bytes()
}

// TxBuilder is something can build up a Tx.
type TxBuilder interface {

	// Build build a transaction on given networks which consumes the given utxos and transfer to transfer to given
	// recipients.
	Build (network btctypes.Network, utxos []btctypes.UTXO, recipients ...Recipient) (Tx, error)
}

type builder struct {}

func (builder *builder) Build(network btctypes.Network, utxos []btctypes.UTXO, recipients ...Recipient) (Tx, error) {
	newTx := wire.NewMsgTx(wire.TxVersion)

	// Fill the utxos we want to use as newTx inputs.
	for _ , utxo := range utxos{
		hash, err := chainhash.NewHashFromStr(utxo.TxHash)
		if err != nil {
			return Tx{}, err
		}

		sourceUtxo := wire.NewOutPoint(hash, utxo.Vout)
		sourceTxIn := wire.NewTxIn(sourceUtxo,nil,nil)
		newTx.AddTxIn(sourceTxIn)
	}

	// Fill newTx output with the target address we want to receive the funds.
	for _, recipient := range recipients {
		outputPkScript, err := txscript.PayToAddrScript(recipient.Addr)
		if err != nil {
			return Tx{}, err
		}
		sourceTxOut := wire.NewTxOut(int64(recipient.Amount), outputPkScript)
		newTx.AddTxOut(sourceTxOut)
	}

	// Get the signature hashes we need to sign
	sigHashes := make([][]byte, len(utxos))
	for i, utxo := range utxos{
		script, err:= hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			return Tx{}, err
		}
		sigHashes[i], err = txscript.CalcSignatureHash(script, txscript.SigHashAll, newTx, i)
		if err != nil {
			return Tx{}, err
		}
	}

	return Tx{
		network:   network,
		tx:        newTx,
		sigHashes: sigHashes,
	}, nil
}
