package btcrpcclient

import (
	"context"
	"encoding/hex"
	"io"
	"time"

	"github.com/renproject/mercury/rpcclient"
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
	ListUnspent(ctx context.Context, minConf, maxConf int64, addresses []string) (ListUnspentResponse, error)
	SendRawTransaction(ctx context.Context, stx []byte) (string, error)
	GetTxOut(ctx context.Context, txid string, i uint32) (GetTxOutResponse, error)
	GetRawTransactionVerbose(ctx context.Context, txid string) (RawTransactionVerbose, error)
}

type rpcClient struct {
	client rpcclient.Client
}

func NewRPCClient(host, user, password string, retryDelay time.Duration) Client {
	return &rpcClient{
		rpcclient.NewClient(host, user, password, retryDelay),
	}
}

func (client *rpcClient) ListUnspent(ctx context.Context, minConf, maxConf int64, addresses []string) (ListUnspentResponse, error) {
	resp := ListUnspentResponse{}
	if err := client.client.SendRequest(ctx, "listunspent", &resp, minConf, maxConf, addresses); err != nil && err != io.EOF {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) SendRawTransaction(ctx context.Context, stx []byte) (string, error) {
	resp := ""
	if err := client.client.SendRequest(ctx, "sendrawtransaction", &resp, hex.EncodeToString(stx), false); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) GetTxOut(ctx context.Context, tx string, i uint32) (GetTxOutResponse, error) {
	resp := GetTxOutResponse{}
	if err := client.client.SendRequest(ctx, "gettxout", &resp, tx, i); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) GetRawTransactionVerbose(ctx context.Context, txid string) (RawTransactionVerbose, error) {
	resp := RawTransactionVerbose{}
	if err := client.client.SendRequest(ctx, "getrawtransaction", &resp, txid, 1); err != nil {
		return resp, err
	}
	return resp, nil
}
