package btcutxo

import (
	"fmt"
	"io"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/iqoption/zecutil"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

const (
	BtcVersion = 2

	ZecExpiryHeight = uint32(10000000)
	ZecVersion      = 4
)

type UTXO interface {
	Amount() btctypes.Amount
	TxHash() types.TxHash
	ScriptPubKey() string
	Vout() uint32
	Confirmations() types.Confirmations
	SigHash(hashType txscript.SigHashType, tx MsgTx, idx int) ([]byte, error)
	AddData(builder *txscript.ScriptBuilder)
}

type UTXOs []UTXO

func (utxos UTXOs) Sum() btctypes.Amount {
	total := btctypes.Amount(0)
	for _, utxo := range utxos {
		total += utxo.Amount()
	}
	return total
}

func (utxos *UTXOs) Filter(confs types.Confirmations) UTXOs {
	newList := UTXOs{}
	for _, utxo := range *utxos {
		if utxo.Confirmations() >= confs {
			newList = append(newList, utxo)
		}
	}
	return newList
}

func NewStandardUTXO(chain types.Chain, txhash types.TxHash, amount btctypes.Amount, scriptPubKey string, vout uint32, confirmations types.Confirmations) UTXO {
	switch chain {
	case types.Bitcoin:
		return StandardBtcUTXO{
			outPoint{
				txhash,
				vout,
			},
			amount,
			scriptPubKey,
			confirmations,
		}
	case types.ZCash:
		return StandardZecUTXO{
			outPoint{
				txhash,
				vout,
			},
			amount,
			scriptPubKey,
			confirmations,
		}
	default:
		panic(fmt.Sprintf("unknown blockchain: %d", chain))
	}
}

func NewScriptUTXO(utxo UTXO, script []byte, updateSigScript func(builder *txscript.ScriptBuilder)) UTXO {
	switch utxo := utxo.(type) {
	case StandardBtcUTXO:
		return ScriptBtcUTXO{
			StandardBtcUTXO: utxo,
			Script:          script,
			UpdateSigScript: updateSigScript,
		}
	case StandardZecUTXO:
		return ScriptZecUTXO{
			StandardZecUTXO: utxo,
			Script:          script,
			UpdateSigScript: updateSigScript,
		}
	default:
		panic(fmt.Sprintf("unknown standard utxo: %T", utxo))
	}
}

type MsgTx interface {
	Serialize(buffer io.Writer) error
	TxHash() chainhash.Hash
	InCount() int
	AddTxIn(txIn *wire.TxIn)
	AddTxOut(txOut *wire.TxOut)
	AddSigScript(i int, sigScript []byte)
}

type OutPoint interface {
	TxHash() types.TxHash
	Vout() uint32
}

type outPoint struct {
	txHash types.TxHash
	vout   uint32
}

func NewOutPoint(txHash types.TxHash, vout uint32) OutPoint {
	return outPoint{
		txHash: txHash,
		vout:   vout,
	}
}

func (op outPoint) TxHash() types.TxHash {
	return op.txHash
}

func (op outPoint) Vout() uint32 {
	return op.vout
}

func NewMsgTx(network btctypes.Network) MsgTx {
	switch network.Chain() {
	case types.Bitcoin:
		return NewBtcMsgTx(wire.NewMsgTx(BtcVersion))
	case types.ZCash:
		return NewZecMsgTx(&zecutil.MsgTx{
			MsgTx:        wire.NewMsgTx(ZecVersion),
			ExpiryHeight: ZecExpiryHeight,
		})
	default:
		panic(types.ErrUnknownChain)
	}
}
