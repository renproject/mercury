package btcclient

import (
	"context"

	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

type Client interface {
	Network() btctypes.Network
	UTXO(op btctypes.OutPoint) (btctypes.UTXO, error)
	UTXOsFromAddress(address btctypes.Address) (btctypes.UTXOs, error)
	Confirmations(txHash types.TxHash) (uint64, error)
	BuildUnsignedTx(utxos btctypes.UTXOs, recipients btctypes.Recipients, refundTo btctypes.Address, gas btctypes.Amount) (btctypes.BtcTx, error)
	SubmitSignedTx(stx btctypes.BtcTx) (types.TxHash, error)
	EstimateTxSize(numUTXOs, numRecipients int) int
	SuggestGasPrice(ctx context.Context, speed types.TxSpeed, txSizeInBytes int) btctypes.Amount
}
