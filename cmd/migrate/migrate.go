package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/testutil/btcaccount"
	"github.com/renproject/mercury/testutil/hdutil"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

var (
	paramMnemonic   string
	paramPassphrase string
	paramChain      string
)

func main() {
	flag.StringVar(&paramMnemonic, "mnemonic", os.Getenv("MNEMONIC"), "Mercury ECDSA distributed key mnemonic")
	flag.StringVar(&paramPassphrase, "passphrase", os.Getenv("PASSPHRASE"), "Mercury ECDSA distributed key passphrase")
	flag.StringVar(&paramChain, "chain", "", "Mercury chain to use")
	flag.Parse()

	chain := types.NewChain(paramChain)
	btcNet := btctypes.NewNetwork(chain, "testnet")
	extPrivKey, err := hdutil.DeriveExtendedPrivKey(paramMnemonic, paramPassphrase, btcNet)
	if err != nil {
		panic(err)
	}
	oldPrivKey, err := hdutil.DerivePrivKey(extPrivKey, 44, 1, 0, 0, 0)
	if err != nil {
		panic(err)
	}

	newPrivKey, err := hdutil.DerivePrivKey(extPrivKey, 44, 1, 0, 0, 0)
	if err != nil {
		panic(err)
	}

	client, err := btcclient.New(logrus.StandardLogger(), btcNet)
	if err != nil {
		panic(err)
	}

	acc, err := btcaccount.NewAccount(client, oldPrivKey)
	if err != nil {
		panic(err)
	}

	btcPubKey := (btcec.PublicKey)(newPrivKey.PublicKey)
	newAddr, err := btctypes.AddressFromPubKeyHash(btcutil.Hash160(btcPubKey.SerializeCompressed()), btcNet)
	if err != nil {
		panic(err)
	}

	txHash, err := acc.Transfer(context.Background(), newAddr, 0, types.Fast, true)
	if err != nil {
		panic(err)
	}

	fmt.Printf("transaction successful: %s\n", txHash)
}
