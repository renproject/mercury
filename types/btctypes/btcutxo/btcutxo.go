package btcutxo

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

type StandardBtcUTXO struct {
	outPoint
	amount        btctypes.Amount
	scriptPubKey  string
	confirmations types.Confirmations
}

func NewStandardBtcUTXO(txHash types.TxHash, amount btctypes.Amount, scriptPubKey string, vout uint32, confirmations types.Confirmations) StandardBtcUTXO {
	return StandardBtcUTXO{
		outPoint: outPoint{
			txHash: txHash,
			vout:   vout,
		},
		amount:        amount,
		scriptPubKey:  scriptPubKey,
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

func (u StandardBtcUTXO) SigHash(hashType txscript.SigHashType, tx MsgTx, idx int) ([]byte, error) {
	scriptPubKey, err := hex.DecodeString(u.scriptPubKey)
	if err != nil {
		return nil, err
	}
	return txscript.CalcSignatureHash(scriptPubKey, hashType, tx.(BtcMsgTx).MsgTx, idx)
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

func (u ScriptBtcUTXO) SigHash(hashType txscript.SigHashType, tx MsgTx, idx int) ([]byte, error) {
	return txscript.CalcSignatureHash(u.Script, hashType, tx.(BtcMsgTx).MsgTx, idx)
}

func (u ScriptBtcUTXO) AddData(builder *txscript.ScriptBuilder) {
	u.UpdateSigScript(builder)
}

type BtcMsgTx struct {
	*wire.MsgTx
}

func NewBtcMsgTx(msgTx *wire.MsgTx) BtcMsgTx {
	return BtcMsgTx{msgTx}
}

func (msgTx BtcMsgTx) InCount() int {
	return len(msgTx.TxIn)
}

func (msgTx BtcMsgTx) AddSigScript(i int, sigScript []byte) {
	msgTx.TxIn[i].SignatureScript = sigScript
}
