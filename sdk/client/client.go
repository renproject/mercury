package client

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/renproject/mercury/sdk/client/btcclient"
)

type Client struct {
	BtcClient btcclient.Client
	EthClient ethclient.Client
}
