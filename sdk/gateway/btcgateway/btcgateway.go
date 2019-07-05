package btcgateway

import (
	"context"
	"crypto/ecdsa"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/types/btctypes"
)

// Gateway is an interface for interacting with Gateways
type Gateway interface {
	UTXOs(ctx context.Context, limit, confirmations int) (btctypes.UTXOs, error)
	BuildUnsignedTx(gwUTXOs btctypes.UTXOs, spenderUTXOs btctypes.UTXOs, gas btctypes.Amount) (btctypes.Tx, error)
	Address() btctypes.Address
}

type gateway struct {
	client      btcclient.Client
	script      []byte
	gwAddr      btctypes.Address
	spenderAddr btctypes.Address
}

// New returns a new Gateway
func New(client btcclient.Client, spenderPubKey *ecdsa.PublicKey, ghash []byte) Gateway {
	pubKeyBytes := btctypes.SerializePublicKey(spenderPubKey, client.Network())
	pubKeyHash160 := btcutil.Hash160(pubKeyBytes)
	b := txscript.NewScriptBuilder()
	b.AddData(ghash)
	b.AddOp(txscript.OP_DROP)
	b.AddOp(txscript.OP_DUP)
	b.AddOp(txscript.OP_HASH160)
	b.AddData(pubKeyHash160)
	b.AddOp(txscript.OP_EQUALVERIFY)
	b.AddOp(txscript.OP_CHECKSIG)
	script, err := b.Script()
	if err != nil {
		panic("invariant violation: invalid bitcoin gateway script")
	}
	gwAddr, err := btcutil.NewAddressScriptHash(script, client.Network().Params())
	if err != nil {
		panic("invariant violation: invalid bitcoin gateway script address")
	}
	spenderAddr, err := btctypes.AddressFromPubKey(spenderPubKey, client.Network())
	if err != nil {
		panic("invariant violation: invalid bitcoin gateway spender address")
	}
	return &gateway{
		client:      client,
		script:      script,
		gwAddr:      gwAddr,
		spenderAddr: spenderAddr,
	}
}

func (gw *gateway) UTXOs(ctx context.Context, limit, confirmations int) (btctypes.UTXOs, error) {
	return gw.client.UTXOs(ctx, gw.Address(), limit, confirmations)
}

func (gw *gateway) BuildUnsignedTx(gwUTXOs btctypes.UTXOs, spenderUTXOs btctypes.UTXOs, gas btctypes.Amount) (btctypes.Tx, error) {
	amount := gwUTXOs.Sum()
	tx, err := gw.client.BuildUnsignedTx(
		gw.spenderAddr,
		btctypes.Recipients{
			btctypes.Recipient{
				Address: gw.spenderAddr,
				Amount:  amount,
			},
		},
		append(spenderUTXOs, gwUTXOs...),
		gas,
	)
	if err != nil {
		// FIXME: Return an error.
		panic("newGatewayTxError()")
	}
	for i := len(spenderUTXOs); i < len(spenderUTXOs)+len(gwUTXOs); i++ {
		if err := tx.ReplaceSignatureHash(gw.script, txscript.SigHashAll, i); err != nil {
			return btctypes.Tx{}, err
		}
	}
	return tx, nil
}

func (gw *gateway) Address() btctypes.Address {
	return gw.gwAddr
}

func (gw *gateway) Script() []byte {
	script := make([]byte, len(gw.script))
	copy(script, gw.script)
	return script
}

func (gw *gateway) ScriptLen() int {
	return len(gw.script)
}
