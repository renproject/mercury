package btcclient

import (
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/mercury/types/btctypes/btcaddress"
	"github.com/renproject/mercury/types/btctypes/btcutxo"
)

type Client interface {
	Chain() btctypes.Chain
	Network() btctypes.Network
	UTXO(txHash types.TxHash, index uint32) (btcutxo.UTXO, error)
	UTXOsFromAddress(address btcaddress.Address) (btcutxo.UTXOs, error)
	Confirmations(txHash types.TxHash) (types.Confirmations, error)
	BuildUnsignedTx(utxos btcutxo.UTXOs, recipients btcaddress.Recipients, refundTo btcaddress.Address, gas btctypes.Amount) (types.Tx, error)
	SubmitSignedTx(stx types.Tx) (types.TxHash, error)
	EstimateTxSize(numUTXOs, numRecipients int) int
	GasStation() GasStation
}