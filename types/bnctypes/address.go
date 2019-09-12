package bnctypes

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
)

type Recipient struct {
	Address Address
	Coins   Coins
}

type Recipients []Recipient

type Address struct {
	Network    Network
	PubKeyHash []byte
}

func (recipients Recipients) SendTx(from Address) msg.SendMsg {
	types.Network = from.Network.ChainNetwork()
	var inputCoins types.Coins
	ops := make([]msg.Output, len(recipients))
	for i, recipient := range recipients {
		ops[i] = msg.Output{
			Address: recipient.Address.AccAddress(),
			Coins:   types.Coins(recipient.Coins),
		}
		ops[i].Coins = ops[i].Coins.Sort()
		inputCoins = inputCoins.Plus(ops[i].Coins)
	}
	ips := []msg.Input{{Address: from.AccAddress(), Coins: inputCoins}}
	return msg.SendMsg{Inputs: ips, Outputs: ops}
}

func AddressFromBech32(address string, network Network) (Address, error) {
	types.Network = network.ChainNetwork()
	accAddr, err := types.AccAddressFromBech32(address)
	if err != nil {
		return Address{}, err
	}
	return Address{PubKeyHash: accAddr, Network: network}, nil
}

func AddressFromHex(address string, network Network) (Address, error) {
	pkh, err := hex.DecodeString(address)
	if err != nil {
		return Address{}, err
	}
	return Address{PubKeyHash: pkh, Network: network}, nil
}

func AddressFromPubKey(pubKey ecdsa.PublicKey, network Network) Address {
	pkHash := btcutil.Hash160((*btcec.PublicKey)(&pubKey).SerializeCompressed())
	return Address{network, pkHash}
}

func (addr Address) AccAddress() types.AccAddress {
	types.Network = addr.Network.ChainNetwork()
	return types.AccAddress(addr.PubKeyHash)
}

func (addr Address) String() string {
	types.Network = addr.Network.ChainNetwork()
	return types.AccAddress(addr.PubKeyHash).String()
}
