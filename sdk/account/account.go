package account

import (
	"crypto/ecdsa"

	"github.com/renproject/mercury/sdk/account/btcaccount"
	"github.com/renproject/mercury/sdk/client"
	"github.com/sirupsen/logrus"
)

type Account struct {
	BtcAccount btcaccount.Account
	// EthAccount
	// ZecAccount
}

func NewAccount(logger logrus.FieldLogger, key *ecdsa.PrivateKey, client client.Client) (Account, error) {
	BtcAccount, err := btcaccount.New(logger, client.BtcClient, key)
	if err != nil {
		return Account{}, err
	}
	return Account{
		BtcAccount,
	}, nil
}
