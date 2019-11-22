package btctypes

import (
	"bytes"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/renproject/mercury/types"
)

type ZecNetwork struct {
	p2shPrefix        []byte
	p2pkhPrefix       []byte
	spendingKeyPrefix []byte
	viewingKeyPrefix  []byte
	zAddressPrefix    []byte
	upgradeParams     []upgradeParam
	params            *chaincfg.Params
	netString         string
}

var ZecMainnet = ZecNetwork{
	p2pkhPrefix:       []byte{0x1C, 0xB8},
	p2shPrefix:        []byte{0x1C, 0xBD},
	spendingKeyPrefix: []byte{0xAB, 0x36},
	viewingKeyPrefix:  []byte{0xA8, 0xAB, 0xD3},
	zAddressPrefix:    []byte{0x16, 0x9A},
	upgradeParams: []upgradeParam{
		{0, []byte{0x00, 0x00, 0x00, 0x00}},
		{207500, []byte{0x19, 0x1B, 0xA8, 0x5B}},
		{280000, []byte{0xBB, 0x09, 0xB8, 0x76}},
	},
	netString: "mainnet",
	params:    &chaincfg.MainNetParams,
}

var ZecTestnet = ZecNetwork{
	p2pkhPrefix:       []byte{0x1D, 0x25},
	p2shPrefix:        []byte{0x1C, 0xBA},
	spendingKeyPrefix: []byte{0xAC, 0x08},
	viewingKeyPrefix:  []byte{0x16, 0xB6},
	zAddressPrefix:    []byte{0xA8, 0xAC, 0x0C},
	upgradeParams: []upgradeParam{
		{0, []byte{0x00, 0x00, 0x00, 0x00}},
		{207500, []byte{0x19, 0x1B, 0xA8, 0x5B}},
		{280000, []byte{0xBB, 0x09, 0xB8, 0x76}},
		{584000, []byte{0x60, 0x0E, 0xB4, 0x2B}},
	},
	netString: "testnet",
	params:    &chaincfg.TestNet3Params,
}

var ZecLocalnet = ZecNetwork{
	p2pkhPrefix:       []byte{0x1D, 0x25},
	p2shPrefix:        []byte{0x1C, 0xBA},
	spendingKeyPrefix: []byte{0xAC, 0x08},
	viewingKeyPrefix:  []byte{0x16, 0xB6},
	zAddressPrefix:    []byte{0xA8, 0xAC, 0x0C},
	upgradeParams: []upgradeParam{
		{0, []byte{0x00, 0x00, 0x00, 0x00}},
		{40, []byte{0x19, 0x1B, 0xA8, 0x5B}},
		{60, []byte{0xBB, 0x09, 0xB8, 0x76}},
		{80, []byte{0x60, 0x0E, 0xB4, 0x2B}},
	},
	netString: "localnet",
	params:    &chaincfg.RegressionNetParams,
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

func (net ZecNetwork) UpgradeParams() []upgradeParam {
	return net.upgradeParams
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
