package account

import (
	"crypto/ecdsa"

	"github.com/renproject/mercury/sdk/account/btcaccount"
	"github.com/renproject/mercury/sdk/client"
	"github.com/sirupsen/logrus"
)

type Account struct {
	BtcAccount *btcaccount.Account
	// EthAccount
	// ZecAccount
}

func NewAccount(logger logrus.FieldLogger, key *ecdsa.PrivateKey, client client.Client) Account {
	return Account{
		BtcAccount: btcaccount.NewBtcAccount(logger, client.BtcClient, key),
	}
}
