package types

import (
	"crypto/ecdsa"
	"errors"
)

// ErrUnknownBtcNetwork is returned when the given bitcoin network is unknwon to us.
var ErrUnknownBtcNetwork = errors.New("unknown bitcoin network")

type BtcValue int64

const (
	Satoshi BtcValue = 1
	BTC              = 100000000 * Satoshi
)

type BtcNetwork uint8

const (
	BtcMainnet BtcNetwork = 1
	BtcTestnet BtcNetwork = 2
)

func (network BtcNetwork) String() string {
	switch network {
	case BtcMainnet:
		return "mainnet"
	case BtcTestnet:
		return "testnet"
	default:
		panic(ErrUnknownBtcNetwork)
	}
}

type BtcAddr struct {
}

func DecodeBase58Address(addr string, network BtcNetwork) (BtcAddr, error) {
	panic("unimplemented")

}

func AddressFromPubKey(pubKey *ecdsa.PublicKey) BtcAddr {
	panic("unimplemented")
}

type TxHash string

type UTXO struct {
	TxHash       string `json:"txHash"`
	Amount       int64  `json:"amount"`
	ScriptPubKey string `json:"scriptPubKey"`
	Vout         uint32 `json:"vout"`
}
