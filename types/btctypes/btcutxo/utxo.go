package btcutxo

import (
	"fmt"

	"github.com/btcsuite/btcd/txscript"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
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
			txhash,
			amount,
			scriptPubKey,
			vout,
			confirmations,
		}
	case types.ZCash:
		return StandardZecUTXO{
			txhash,
			amount,
			scriptPubKey,
			vout,
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
	IsMsgTx()
}
