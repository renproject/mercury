package btcclient

import (
	"context"
	"crypto/ecdsa"

	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

type Client interface {
	Network() btctypes.Network
	UTXO(ctx context.Context, op btctypes.OutPoint) (btctypes.UTXO, error)
	UTXOsFromAddress(ctx context.Context, address btctypes.Address) (btctypes.UTXOs, error)
	Confirmations(ctx context.Context, txHash types.TxHash) (uint64, error)
	BuildUnsignedTx(utxos btctypes.UTXOs, recipients btctypes.Recipients, refundTo btctypes.Address, gas btctypes.Amount) (btctypes.BtcTx, error)
	SubmitSignedTx(ctx context.Context, stx btctypes.BtcTx) (types.TxHash, error)
	EstimateTxSize(numUTXOs, numRecipients int) int // Depricated
	SuggestGasPrice(ctx context.Context, speed types.TxSpeed, txSizeInBytes int) btctypes.Amount
	SerializePublicKey(pubkey ecdsa.PublicKey) []byte
	AddressFromBase58(addr string) (btctypes.Address, error)
	AddressFromPubKey(pubkey ecdsa.PublicKey) (btctypes.Address, error)
	AddressFromScript(script []byte) (btctypes.Address, error)
	PayToAddrScript(address btctypes.Address) ([]byte, error)
}
