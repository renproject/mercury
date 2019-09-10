package bnctypes

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/tendermint/tendermint/libs/bech32"
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

func AddressFromBech32(address string) (Address, error) {
	network := networkFromPrefix(address[:3])
	pkh, err := types.GetFromBech32(address, network.ChainNetwork().Bech32Prefixes())
	if err != nil {
		return Address{}, err
	}
	return Address{PubKeyHash: pkh, Network: network}, nil
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

func PubKeyHashToAddress(pkHash []byte, network Network) (string, error) {
	return bech32.ConvertAndEncode(network.ChainNetwork().Bech32Prefixes(), pkHash)
}

func (addr Address) AccAddress() types.AccAddress {
	return types.AccAddress(addr.PubKeyHash)
}

func (addr Address) String() string {
	addrString, err := PubKeyHashToAddress(addr.PubKeyHash, addr.Network)
	if err != nil {
		panic(err)
	}
	return addrString
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

func networkFromPrefix(prefix string) Network {
	switch prefix {
	case "bnb":
		return Mainnet
	case "tbn":
		return Testnet
	default:
		panic(fmt.Sprintf("invalid bnc prefix: %s", prefix))
	}
}
