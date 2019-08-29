package bnctypes

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/tendermint/tendermint/libs/bech32"
)

type Address types.AccAddress

func AddressFromBech32(address string, network Network) (Address, error) {
	addr, err := types.GetFromBech32(address, network.ChainNetwork().Bech32Prefixes())
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

func AddressFromPubKey(pubKey ecdsa.PublicKey, network Network) (Address, error) {
	pkHash := btcutil.Hash160((*btcec.PublicKey)(&pubKey).SerializeCompressed())
	bech32Addr, err := bech32.ConvertAndEncode(network.ChainNetwork().Bech32Prefixes(), pkHash)
	if err != nil {
		panic(err)
	}
	return AddressFromBech32(bech32Addr, network)
}

func (addr Address) AccAddress() types.AccAddress {
	return types.AccAddress(addr)
}

func (addr Address) String() string {
	return types.AccAddress(addr).String()
}

func PrivKeyFromKeyStore(keystore string, password string) (*ecdsa.PrivateKey, error) {
	keyManager, err := keys.NewKeyStoreKeyManager(keystore, password)
	if err != nil {
		panic(err)
	}

	privKey, err := keyManager.ExportAsPrivateKey()
	if err != nil {
		panic(err)
	}

	fmt.Println(privKey)
	return nil, nil
}
