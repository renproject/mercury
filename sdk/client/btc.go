package client

import (
	"context"
	"net/http"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/renproject/mercury/types"
)

type BtcClient struct {
	rpcClient  *http.Client
	config     chaincfg.Params
	url        string
}

func NewBtcClient(network types.BtcNetwork)*BtcClient{
	switch network {
	case types.BtcMainnet:
		return &BtcClient{
			rpcClient: &http.Client{},
			config:    chaincfg.MainNetParams,
			url:       "https://ren-mercury.herokuapp.com/btc",
		}
	case types.BtcTestnet:
		return &BtcClient{
			rpcClient: &http.Client{},
			config:    chaincfg.TestNet3Params,
			url:       "https://ren-mercury.herokuapp.com/btc-testnet3",
		}
	default:
		panic("unknown bitcoin network")
	}
}

func (client *BtcClient) Balance(ctx context.Context, address types.BtcAddr, confirmed bool) (types.Satoshi, error) {
	panic("unimplemented")
}

func (client *BtcClient) UTXOs(ctx context.Context, address types.BtcAddr, limit, confitmations int) ([]types.UTXO, error){
	panic("unimplemented")
}

func (client *BtcClient) Confirmations(ctx context.Context, hash types.TxHash) (int64, error){
	panic("unimplemented")
}

func (client *BtcClient) PublishSTX(ctx context.Context, stx []byte) error {
	panic("unimplemented")
}
