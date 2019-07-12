package btcgateway

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/types/btctypes"
)

// Gateway is an interface for interacting with Gateways
type Gateway interface {
	UTXO(hash btctypes.TxHash, i uint32) (btctypes.UTXO, error)
	UTXOs() (btctypes.UTXOs, error)
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

func (gw *gateway) UTXOs() (btctypes.UTXOs, error) {
	utxos, err := gw.client.UTXOsFromAddress(gw.Address())
	if err != nil {
		return nil, err
	}
	scriptUTXOs := make(btctypes.UTXOs, len(utxos))
	for i := range scriptUTXOs {
		utxo, ok := utxos[i].(btctypes.StandardUTXO)
		if !ok {
			return nil, fmt.Errorf("unexpected utxo of type: %T", utxo)
		}
		scriptUTXOs[i] = btctypes.ScriptUTXO{
			StandardUTXO: utxo,
			Script:       gw.Script(),
			UpdateSigScript: func(builder *txscript.ScriptBuilder) {
				builder.AddData(gw.Script())
			},
		}
	}
	return scriptUTXOs, nil
}

func (gw *gateway) UTXO(hash btctypes.TxHash, i uint32) (btctypes.UTXO, error) {
	utxo, err := gw.client.UTXO(hash, i)
	if err != nil {
		return nil, err
	}

	stdUTXO, ok := utxo.(btctypes.StandardUTXO)
	if !ok {
		return nil, fmt.Errorf("unexpected utxo of type: %T", utxo)
	}

	return btctypes.ScriptUTXO{
		StandardUTXO: stdUTXO,
		Script:       gw.Script(),
		UpdateSigScript: func(builder *txscript.ScriptBuilder) {
			builder.AddData(gw.Script())
		},
	}, nil
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
