package btcutxo

import (
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
	SigHash(hashType txscript.SigHashType, txBytes []byte, idx int) ([]byte, error)
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
