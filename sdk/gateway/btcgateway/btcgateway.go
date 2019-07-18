package btcgateway

import (
	"crypto/ecdsa"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/types/btctypes"
)

// Gateway is an interface for interacting with Gateways
type Gateway interface {
	UTXO(op btctypes.OutPoint) (btctypes.UTXO, error)
	Address() btctypes.Address
	EstimateTxSize(numSpenderUTXOs, numGatewayUTXOs, numRecipients int) int
	Script() []byte
}

type gateway struct {
	client      btcclient.Client
	script      []byte
	gwAddr      btctypes.Address
	spenderAddr btctypes.Address
}

// New returns a new Gateway
func New(client btcclient.Client, spenderPubKey ecdsa.PublicKey, ghash []byte) Gateway {
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
	gwAddr, err := btctypes.AddressFromScript(script, client.Network())
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

func (gw *gateway) UTXO(op btctypes.OutPoint) (btctypes.UTXO, error) {
	utxo, err := gw.client.UTXO(op)
	if err != nil {
		return nil, err
	}

	return btctypes.NewUTXO(
		op,
		utxo.Amount(),
		utxo.ScriptPubKey(),
		utxo.Confirmations(),
		gw.Script(),
		func(builder *txscript.ScriptBuilder) {
			builder.AddData(gw.Script())
		},
	), nil
}

func (gw *gateway) Address() btctypes.Address {
	return gw.gwAddr
}

func (gw *gateway) Script() []byte {
	script := make([]byte, len(gw.script))
	copy(script, gw.script)
	return script
}

func (gw *gateway) EstimateTxSize(numSpenderUTXOs, numGatewayUTXOs, numRecipients int) int {
	scriptLen := len(gw.Script())
	return (113+scriptLen)*numGatewayUTXOs + gw.client.EstimateTxSize(numSpenderUTXOs, numRecipients)
}
