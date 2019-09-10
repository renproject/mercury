package btcrpcclient

import (
	"context"
	"encoding/hex"
	"io"
	"time"

	"github.com/renproject/mercury/rpcclient"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

type ListUnspentResponse []ListUnspentObj

type ListUnspentObj struct {
	Address       string  `json:"address"`
	Category      string  `json:"category"`
	Amount        float64 `json:"amount"`
	TxID          string  `json:"txid"`
	Vout          uint32  `json:"vout"`
	Generated     bool    `json:"generated"`
	ScriptPubKey  string  `json:"scriptPubKey"`
	Confirmations int64   `json:"confirmations"`
	RedeemScript  string  `json:"redeemScript"`
	Spendable     bool    `json:"spendable"`
}

type RawTransactionVerbose struct {
	TxID          string `json:"txid"`
	Confirmations uint32 `json:"confirmations"`
}

type GetTxOutResponse struct {
	Confirmations int64        `json:"confirmations"`
	Value         float64      `json:"value"`
	ScriptPubKey  ScriptPubKey `json:"scriptPubKey"`
}

type ScriptPubKey struct {
	Hex string `json:"hex"`
}

type Client interface {
	ListUnspent(ctx context.Context, minConf, maxConf int64, addresses []btctypes.Address) (ListUnspentResponse, error)
	SendRawTransaction(ctx context.Context, stx btctypes.BtcTx) (string, error)
	GetTxOut(ctx context.Context, txid types.TxHash, i uint32) (GetTxOutResponse, error)
	GetRawTransactionVerbose(ctx context.Context, txid types.TxHash) (RawTransactionVerbose, error)
}

type rpcClient struct {
	client rpcclient.Client
}

func NewRPCClient(host, user, password string, retryDelay time.Duration) Client {
	return &rpcClient{
		rpcclient.NewClient(host, user, password, retryDelay),
	}
}

func (client *rpcClient) ListUnspent(ctx context.Context, minConf, maxConf int64, addresses []btctypes.Address) (ListUnspentResponse, error) {
	addrs := make([]string, len(addresses))
	for i := range addresses {
		addrs[i] = addresses[i].EncodeAddress()
	}

	resp := ListUnspentResponse{}
	if err := client.client.SendRequest(ctx, "listunspent", &resp, minConf, maxConf, addrs); err != nil && err != io.EOF {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) SendRawTransaction(ctx context.Context, stx btctypes.BtcTx) (string, error) {
	stxBytes, err := stx.Serialize()
	if err != nil {
		return "", err
	}
	resp := ""
	if err := client.client.SendRequest(ctx, "sendrawtransaction", &resp, hex.EncodeToString(stxBytes), false); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) GetTxOut(ctx context.Context, tx types.TxHash, i uint32) (GetTxOutResponse, error) {
	resp := GetTxOutResponse{}
	if err := client.client.SendRequest(ctx, "gettxout", &resp, tx, i); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) GetRawTransactionVerbose(ctx context.Context, txid types.TxHash) (RawTransactionVerbose, error) {
	resp := RawTransactionVerbose{}
	if err := client.client.SendRequest(ctx, "getrawtransaction", &resp, txid, 1); err != nil {
		return resp, err
	}
	return resp, nil
}
