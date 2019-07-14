package btcclient

import (
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/mercury/types/btctypes/btcaddress"
	"github.com/renproject/mercury/types/btctypes/btctx"
	"github.com/renproject/mercury/types/btctypes/btcutxo"
)

type Client interface {
	Chain() btctypes.Chain
	Network() btctypes.Network
	UTXO(txHash types.TxHash, index uint32) (btcutxo.UTXO, error)
	UTXOsFromAddress(address btcaddress.Address) (btcutxo.UTXOs, error)
	Confirmations(txHash types.TxHash) (types.Confirmations, error)
	BuildUnsignedTx(utxos btcutxo.UTXOs, recipients btcaddress.Recipients, refundTo btcaddress.Address, gas btctypes.Amount) (btctx.BtcTx, error)
	SubmitSignedTx(stx btctx.BtcTx) (types.TxHash, error)
	EstimateTxSize(numUTXOs, numRecipients int) int
	SuggestGasPrice(ctx context.Context, speed types.TxSpeed, txSizeInBytes int) btctypes.Amount
}
