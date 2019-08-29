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
	btctypes.Script
	UTXO(ctx context.Context, op btctypes.OutPoint) (btctypes.UTXO, error)
	Spender() btctypes.Address
	BaseScript() btctypes.Script
}

type gateway struct {
	spender btctypes.Address
	client  btcclient.Client
	btctypes.Script
}

// New returns a new Gateway
func New(client btcclient.Client, spenderPubKey ecdsa.PublicKey, ghash []byte) Gateway {
	pubKeyBytes := btctypes.SerializePublicKey(spenderPubKey)
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
		panic("invariant violation: invalid gateway script")
	}
	spenderAddr, err := client.AddressFromPubKey(spenderPubKey)
	if err != nil {
		panic("invariant violation: invalid gateway spender address")
	}
	return &gateway{spenderAddr, client, btctypes.NewScript(script, client.Network())}
}

func (gw *gateway) UTXO(ctx context.Context, op btctypes.OutPoint) (btctypes.UTXO, error) {
	utxo, err := gw.client.UTXO(ctx, op)
	if err != nil {
		return nil, err
	}
	utxo.SetScript(gw.BaseScript().Bytes())
	return utxo, nil
}

func (gw *gateway) Spender() btctypes.Address {
	return gw.spender
}

func (gw *gateway) BaseScript() btctypes.Script {
	return gw.Script
}
