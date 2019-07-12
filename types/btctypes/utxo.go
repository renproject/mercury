package btctypes

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

type UTXO interface {
	Amount() Amount
	TxHash() TxHash
	ScriptPubKey() string
	Vout() uint32
	Confirmations() Confirmations

	SigHash(hashType txscript.SigHashType, tx *wire.MsgTx, idx int) ([]byte, error)
	AddData(builder *txscript.ScriptBuilder)
}

type UTXOs []UTXO

func (utxos UTXOs) Sum() Amount {
	total := Amount(0)
	for _, utxo := range utxos {
		total += utxo.Amount()
	}
	return total
}

func (utxos *UTXOs) Filter(confs Confirmations) UTXOs {
	newList := UTXOs{}
	for _, utxo := range *utxos {
		if utxo.Confirmations() >= confs {
			newList = append(newList, utxo)
		}
	}
	return newList
}

func NewStandardUTXO(txHash TxHash, amount Amount, scriptPubKey string, vout uint32, confirmations Confirmations) StandardUTXO {
	return StandardUTXO{
		txHash:        txHash,
		amount:        amount,
		scriptPubKey:  scriptPubKey,
		vout:          vout,
		confirmations: confirmations,
	}
}

type StandardUTXO struct {
	txHash        TxHash
	amount        Amount
	scriptPubKey  string
	vout          uint32
	confirmations Confirmations
}

func (u StandardUTXO) Confirmations() Confirmations {
	return u.confirmations
}

func (u StandardUTXO) Amount() Amount {
	return u.amount
}

func (u StandardUTXO) TxHash() TxHash {
	return u.txHash
}

func (u StandardUTXO) ScriptPubKey() string {
	return u.scriptPubKey
}

func (u StandardUTXO) Vout() uint32 {
	return u.vout
}

func (u StandardUTXO) SigHash(hashType txscript.SigHashType, tx *wire.MsgTx, idx int) ([]byte, error) {
	scriptPubKey, err := hex.DecodeString(u.scriptPubKey)
	if err != nil {
		return nil, err
	}
	return txscript.CalcSignatureHash(scriptPubKey, hashType, tx, idx)
}

func (StandardUTXO) AddData(*txscript.ScriptBuilder) {
}

type ScriptUTXO struct {
	utxo StandardUTXO

	Script          []byte
	UpdateSigScript func(builder *txscript.ScriptBuilder)
}

func (u ScriptUTXO) Amount() Amount {
	return u.utxo.Amount()
}

func (u ScriptUTXO) TxHash() TxHash {
	return u.utxo.TxHash()
}

func (u ScriptUTXO) ScriptPubKey() string {
	return u.utxo.ScriptPubKey()
}

func (u ScriptUTXO) Vout() uint32 {
	return u.utxo.Vout()
}

func (u ScriptUTXO) SigHash(hashType txscript.SigHashType, tx *wire.MsgTx, idx int) ([]byte, error) {
	return txscript.CalcSignatureHash(u.Script, hashType, tx, idx)
}

func (u ScriptUTXO) AddData(builder *txscript.ScriptBuilder) {
	u.UpdateSigScript(builder)
}
