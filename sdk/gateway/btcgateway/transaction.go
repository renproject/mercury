package btcgateway

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/renproject/mercury/types/btctypes"
)

// GatewayTx is a transaction which supports injecting signatures
type GatewayTx interface {
	btctypes.Tx

	InjectSignatures(sigs []*btcec.Signature, serializedPubKey btctypes.SerializedPubKey) error
}

type gatewayTx struct {
	tx           btctypes.StandardTx
	gatewayUTXOs btctypes.UTXOs
	spenderUTXOs btctypes.UTXOs
	script       []byte
}

func (gtx *gatewayTx) SignatureHashes() []btctypes.SignatureHash {
	return gtx.tx.SignatureHashes()
}

func (gtx *gatewayTx) IsSigned() bool {
	return gtx.tx.IsSigned()
}

func (gtx *gatewayTx) UTXOs() btctypes.UTXOs {
	return gtx.tx.UTXOs()
}

func (gtx *gatewayTx) Tx() *wire.MsgTx {
	return gtx.tx.Tx()
}

func (gtx *gatewayTx) InjectSignatures(sigs []*btcec.Signature, serializedPubKey btctypes.SerializedPubKey) error {
	customScript := func(builder *txscript.ScriptBuilder, index int) {
		if index >= len(gtx.spenderUTXOs) && index < len(gtx.spenderUTXOs)+len(gtx.gatewayUTXOs) {
			builder.AddData(gtx.script)
		}
	}
	return gtx.tx.InjectSignatures(sigs, serializedPubKey, customScript)
}
