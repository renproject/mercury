package btctypes

import "github.com/renproject/mercury/types"

// BaseScript is an interface for interacting with Scripts
type Script interface {
	Update(utxo UTXO) UTXO
	Bytes() []byte
	EstimateTxSize(numSpenderUTXOs, numGatewayUTXOs, numRecipients int) int
	Address() Address
}

type BtcScript struct {
	segWitAddress Address

	Script
}

type ZecScript struct {
	Script
}

type script struct {
	address Address
	data    []byte
}

// NewScript returns a new Script
func NewScript(data []byte, network Network) Script {
	address, err := AddressFromScript(data, network)
	if err != nil {
		panic("invariant violation: failed to calcucalte address of a btc script")
	}
	baseScript := &script{address, data}
	switch network.Chain() {
	case types.Bitcoin:
		segWitAddress, err := SegWitAddressFromScript(data, network)
		if err != nil {
			panic("invariant violation: failed to calcucalte SegWit address of a btc script")
		}
		return &BtcScript{
			Script:        baseScript,
			segWitAddress: segWitAddress,
		}
	case types.ZCash:
		return &ZecScript{baseScript}
	default:
		panic(types.ErrUnknownChain)
	}
}

func (s *script) Update(utxo UTXO) UTXO {
	return NewUTXO(
		NewOutPoint(utxo.TxHash(), utxo.Vout()),
		utxo.Amount(),
		utxo.ScriptPubKey(),
		utxo.Confirmations(),
		s.Bytes(),
	)
}

func (s *script) Bytes() []byte {
	script := make([]byte, len(s.data))
	copy(script, s.data)
	return script
}

func (s *script) Address() Address {
	return s.address
}

func (s *script) EstimateTxSize(numSpenderUTXOs, numGatewayUTXOs, numRecipients int) int {
	scriptLen := len(s.Bytes())
	return (113+scriptLen)*numGatewayUTXOs + EstimateTxSize(numSpenderUTXOs, numRecipients)
}

func (btcScript *BtcScript) SegWitaddress() Address {
	return btcScript.segWitAddress
}
