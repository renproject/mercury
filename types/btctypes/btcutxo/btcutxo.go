package btcutxo

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

type StandardBtcUTXO struct {
	txHash        types.TxHash
	amount        btctypes.Amount
	scriptPubKey  string
	vout          uint32
	confirmations types.Confirmations
}

func NewStandardBtcUTXO(txHash types.TxHash, amount btctypes.Amount, scriptPubKey string, vout uint32, confirmations types.Confirmations) StandardBtcUTXO {
	return StandardBtcUTXO{
		txHash:        txHash,
		amount:        amount,
		scriptPubKey:  scriptPubKey,
		vout:          vout,
		confirmations: confirmations,
	}
}

func (u StandardBtcUTXO) Confirmations() types.Confirmations {
	return u.confirmations
}

func (u StandardBtcUTXO) Amount() btctypes.Amount {
	return u.amount
}

func (u StandardBtcUTXO) TxHash() types.TxHash {
	return u.txHash
}

func (u StandardBtcUTXO) ScriptPubKey() string {
	return u.scriptPubKey
}

func (u StandardBtcUTXO) Vout() uint32 {
	return u.vout
}

func (u StandardBtcUTXO) SigHash(hashType txscript.SigHashType, txBytes []byte, idx int) ([]byte, error) {
	tx := new(wire.MsgTx)
	if err := tx.Deserialize(bytes.NewBuffer(txBytes)); err != nil {
		return nil, err
	}
	scriptPubKey, err := hex.DecodeString(u.scriptPubKey)
	if err != nil {
		return nil, err
	}
	return txscript.CalcSignatureHash(scriptPubKey, hashType, tx, idx)
}

func (StandardBtcUTXO) AddData(*txscript.ScriptBuilder) {
}

type ScriptBtcUTXO struct {
	StandardBtcUTXO

	Script          []byte
	UpdateSigScript func(builder *txscript.ScriptBuilder)
}

func (u ScriptBtcUTXO) Amount() btctypes.Amount {
	return u.amount
}

func (u ScriptBtcUTXO) TxHash() types.TxHash {
	return u.txHash
}

func (u ScriptBtcUTXO) ScriptPubKey() string {
	return u.scriptPubKey
}

func (u ScriptBtcUTXO) Vout() uint32 {
	return u.vout
}

func (u ScriptBtcUTXO) SigHash(hashType txscript.SigHashType, txBytes []byte, idx int) ([]byte, error) {
	tx := new(wire.MsgTx)
	if err := tx.Deserialize(bytes.NewBuffer(txBytes)); err != nil {
		return nil, err
	}
	return txscript.CalcSignatureHash(u.Script, hashType, tx, idx)
}

func (u ScriptBtcUTXO) AddData(builder *txscript.ScriptBuilder) {
	u.UpdateSigScript(builder)
}
