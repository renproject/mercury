package zectypes

import (
	"github.com/renproject/mercury/types/btctypes"
)

type Amount = btctypes.Amount

const (
	ZAT = Amount(1)
	ZEC = Amount(1e8 * ZAT)
)

type Network = btctypes.Network

const (
	Localnet = btctypes.Localnet
	Mainnet  = btctypes.Mainnet
	Testnet  = btctypes.Testnet
)

var NewNetwork = btctypes.NewNetwork

type Address = btctypes.Address

var AddressFromBase58 = btctypes.AddressFromBase58

var AddressFromPubKey = btctypes.AddressFromPubKey

var SerializePublicKey = btctypes.SerializePublicKey

var RandomAddress = btctypes.RandomAddress

type UTXO = btctypes.UTXO

type UTXOs = btctypes.UTXOs

type Recipient = btctypes.Recipient

type Recipients = btctypes.Recipients

type Confirmations = btctypes.Confirmations
