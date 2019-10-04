package btctypes

import (
	"bytes"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/renproject/mercury/types"
)

type ZecNetwork struct {
	p2shPrefix  []byte
	p2pkhPrefix []byte
	params      *chaincfg.Params
	netString   string
}

var ZecMainnet = ZecNetwork{
	p2pkhPrefix: []byte{0x1C, 0xB8},
	p2shPrefix:  []byte{0x1C, 0xBD},
	netString:   "mainnet",
	params:      &chaincfg.MainNetParams,
}

var ZecTestnet = ZecNetwork{
	p2pkhPrefix: []byte{0x1D, 0x25},
	p2shPrefix:  []byte{0x1C, 0xBA},
	netString:   "testnet",
	params:      &chaincfg.TestNet3Params,
}

var ZecLocalnet = ZecNetwork{
	p2pkhPrefix: []byte{0x1D, 0x25},
	p2shPrefix:  []byte{0x1C, 0xBA},
	netString:   "localnet",
	params:      &chaincfg.RegressionNetParams,
}

func NewZecNetwork(network string) ZecNetwork {
	switch network {
	case "mainnet":
		return ZecMainnet
	case "testnet", "testnet3":
		return ZecTestnet
	case "localnet", "localhost":
		return ZecLocalnet
	default:
		panic(types.ErrUnknownNetwork)
	}
}

func (net ZecNetwork) Chain() types.Chain {
	return types.ZCash
}

func (net ZecNetwork) String() string {
	return net.netString
}

func (net ZecNetwork) SegWitEnabled() bool {
	return false
}

func (net ZecNetwork) Params() *chaincfg.Params {
	return net.params
}

func (net ZecNetwork) Prefix(addrType AddressType) []byte {
	switch addrType {
	case P2PKH:
		return net.p2pkhPrefix
	case P2SH:
		return net.p2shPrefix
	default:
		panic(ErrUnknownAddressType)
	}
}

func ParsePrefix(prefix []byte) (AddressType, Network) {
	if bytes.Equal(prefix, ZecTestnet.p2pkhPrefix) {
		return P2PKH, ZecTestnet
	}
	if bytes.Equal(prefix, ZecTestnet.p2shPrefix) {
		return P2SH, ZecTestnet
	}

	if bytes.Equal(prefix, ZecMainnet.p2shPrefix) {
		return P2SH, ZecMainnet
	}

	if bytes.Equal(prefix, ZecMainnet.p2pkhPrefix) {
		return P2PKH, ZecMainnet
	}
	panic("unknown prefix")
}
