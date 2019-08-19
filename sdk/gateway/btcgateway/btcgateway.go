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
	Address() btctypes.Address
	Spender() btctypes.Address
}

type gateway struct {
	addr    btctypes.Address
	spender btctypes.Address
	client  btcclient.Client
	script  btctypes.Script
}

// New returns a new Gateway
func New(client btcclient.Client, spenderPubKey ecdsa.PublicKey, ghash []byte, segwit bool) Gateway {
	pubKeyBytes := btctypes.SerializePublicKey(spenderPubKey, client.Network(), segwit)
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
	scriptAddr, err := btctypes.AddressFromScript(script, client.Network(), segwit)
	if err != nil {
		panic("invariant violation: invalid gateway script address")
	}
	spenderAddr, err := btctypes.AddressFromPubKey(spenderPubKey, client.Network(), segwit)
	if err != nil {
		panic("invariant violation: invalid gateway spender address")
	}
	return &gateway{scriptAddr, spenderAddr, client, btctypes.NewScript(script, nil)}
}

func (gw *gateway) UTXO(ctx context.Context, op btctypes.OutPoint) (btctypes.UTXO, error) {
	utxo, err := gw.client.UTXO(ctx, op)
	if err != nil {
		return nil, err
	}
	return gw.script.Update(utxo), nil
}

func (gw *gateway) Update(utxo btctypes.UTXO) btctypes.UTXO {
	return gw.script.Update(utxo)
}

func (gw *gateway) Address() btctypes.Address {
	return gw.addr
}

func (gw *gateway) Spender() btctypes.Address {
	return gw.spender
}

func (gw *gateway) Bytes() []byte {
	return gw.script.Bytes()
}

func (gw *gateway) EstimateTxSize(numSpenderUTXOs, numGatewayUTXOs, numRecipients int) int {
	scriptLen := len(gw.Bytes())
	return (113+scriptLen)*numGatewayUTXOs + gw.client.EstimateTxSize(numSpenderUTXOs, numRecipients)
}
