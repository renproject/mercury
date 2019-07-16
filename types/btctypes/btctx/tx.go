package btctx

import (
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes/btcaddress"
	"github.com/renproject/mercury/types/btctypes/btcutxo"
)

type BtcTx interface {
	types.Tx
	UTXOs() btcutxo.UTXOs
	OutputUTXO(address btcaddress.Address) btcutxo.UTXO
}
