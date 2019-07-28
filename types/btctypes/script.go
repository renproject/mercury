package btctypes

import (
	"github.com/btcsuite/btcd/txscript"
)

// Script is an interface for interacting with Scripts
type Script interface {
	Update(utxo UTXO) UTXO
	Bytes() []byte
	EstimateTxSize(numSpenderUTXOs, numGatewayUTXOs, numRecipients int) int
}

type script struct {
	data   []byte
	solver func(builder *txscript.ScriptBuilder)
}

// NewScript returns a new Script
func NewScript(data []byte, solver func(builder *txscript.ScriptBuilder)) Script {
	return &script{
		data:   data,
		solver: solver,
	}
}

func (s *script) Update(utxo UTXO) UTXO {
	return NewUTXO(
		NewOutPoint(utxo.TxHash(), utxo.Vout()),
		utxo.Amount(),
		utxo.ScriptPubKey(),
		utxo.Confirmations(),
		s.Bytes(),
		func(builder *txscript.ScriptBuilder) {
			if s.solver != nil {
				s.solver(builder)
			}
			builder.AddData(s.Bytes())
		},
	)
}

func (s *script) Bytes() []byte {
	script := make([]byte, len(s.data))
	copy(script, s.data)
	return script
}

func (s *script) EstimateTxSize(numSpenderUTXOs, numGatewayUTXOs, numRecipients int) int {
	scriptLen := len(s.Bytes())
	return (113+scriptLen)*numGatewayUTXOs + EstimateTxSize(numSpenderUTXOs, numRecipients)
}
