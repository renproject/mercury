package ethtypes

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/renproject/mercury/types"
)

const (
	EthMainnet  network = 0
	EthKovan    network = 1
	EthLocalnet network = 2

	MaticMainnet  network = 3
	MaticTestnet  network = 4
	MaticLocalnet network = 5
)

func (network network) String() string {
	switch network {
	case EthMainnet, MaticMainnet:
		return "mainnet"
	case EthKovan:
		return "kovan"
	case MaticTestnet:
		return "testnet"
	case EthLocalnet, MaticLocalnet:
		return "localnet"
	default:
		panic(types.ErrUnknownNetwork)
	}
}

func (network network) Chain() types.Chain {
	switch network {
	case EthMainnet, EthKovan, EthLocalnet:
		return types.Ethereum
	case MaticMainnet, MaticTestnet, MaticLocalnet:
		return types.Matic
	default:
		panic(types.ErrUnknownChain)
	}
}

type Network interface {
	types.Network
}

type network uint8

type Tx interface {
	types.Tx
}

type Address common.Address

func AddressFromPublicKey(publicKeyECDSA *ecdsa.PublicKey) Address {
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return Address(address)
}

func AddressFromHex(addr string) Address {
	return Address(common.HexToAddress(addr))
}

func (addr Address) Hex() string {
	return common.Address(addr).Hex()
}

type Hash common.Hash

type Event struct {
	Name        string
	Args        map[string]interface{}
	IndexedArgs []Hash
}
