package btctx

import (
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes/btcutxo"
)

type BtcTx interface {
	types.Tx
	UTXOs() btcutxo.UTXOs
}
