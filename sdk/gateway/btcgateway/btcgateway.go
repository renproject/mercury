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
	UTXOs() (btctypes.UTXOs, error)
	BuildUnsignedTx(gwUTXOs btctypes.UTXOs, spenderUTXOs btctypes.UTXOs, gas btctypes.Amount) (GatewayTx, error)
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
	return gw.client.UTXOsFromAddress(gw.Address())
}

func (gw *gateway) BuildUnsignedTx(gwUTXOs btctypes.UTXOs, spenderUTXOs btctypes.UTXOs, gas btctypes.Amount) (GatewayTx, error) {
	amount := gwUTXOs.Sum()
	recipients := btctypes.Recipients{{Address: gw.spenderAddr, Amount: amount}}
	standardTx, err := gw.client.BuildUnsignedTx(
		append(spenderUTXOs, gwUTXOs...),
		recipients,
		gw.spenderAddr,
		gas,
	)
	if err != nil {
		return nil, fmt.Errorf("error generating new gateway: %v", err)
	}
	gtx := &gatewayTx{
		tx:           standardTx,
		gatewayUTXOs: gwUTXOs,
		spenderUTXOs: spenderUTXOs,
		script:       gw.Script(),
	}
	for i := len(spenderUTXOs); i < len(spenderUTXOs)+len(gwUTXOs); i++ {
		if err = gtx.tx.ReplaceSignatureHash(gw.Script(), txscript.SigHashAll, i); err != nil {
			return nil, fmt.Errorf("error setting gateway signature hashes: %v", err)
		}
	}
	return gtx, nil
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