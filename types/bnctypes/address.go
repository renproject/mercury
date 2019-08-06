package bnctypes

import (
	"crypto/ecdsa"

	"github.com/binance-chain/go-sdk/common/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
)

type Address types.AccAddress

func AddressFromBech32(address string) (Address, error) {
	addr, err := types.AccAddressFromBech32(address)
	if err != nil {
		return Address{}, err
	}
	return Address(addr), nil
}

func AddressFromHex(address string) (Address, error) {
	addr, err := types.AccAddressFromHex(address)
	if err != nil {
		return Address{}, err
	}
	return Address(addr), nil
}

func AddressFromPubKey(pubKey ecdsa.PublicKey) Address {
	addr := Address{}
	copy(addr[:], btcutil.Hash160((*btcec.PublicKey)(&pubKey).SerializeCompressed()))
	return addr
}

func (addr Address) AccAddress() types.AccAddress {
	return types.AccAddress(addr)
}

func (addr Address) String() string {
	return types.AccAddress(addr).String()
}
