package btcrpcclient

import (
	"encoding/hex"
	"io"

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
	ListUnspent(minConf, maxConf int64, addresses []string) (ListUnspentResponse, error)
	SendRawTransaction(stx []byte) (string, error)
	GetTxOut(txid string, i uint32) (GetTxOutResponse, error)
	GetRawTransactionVerbose(txid string) (RawTransactionVerbose, error)
}

type rpcClient struct {
	client rpcclient.Client
}

func NewRPCClient(host, user, password string) Client {
	return &rpcClient{
		rpcclient.NewClient(host, user, password),
	}
}

func (client *rpcClient) ListUnspent(minConf, maxConf int64, addresses []string) (ListUnspentResponse, error) {
	resp := ListUnspentResponse{}
	if err := client.client.SendRequest("listunspent", &resp, minConf, maxConf, addresses); err != nil && err != io.EOF {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) SendRawTransaction(stx []byte) (string, error) {
	resp := ""
	if err := client.client.SendRequest("sendrawtransaction", &resp, hex.EncodeToString(stx), false); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) GetTxOut(tx string, i uint32) (GetTxOutResponse, error) {
	resp := GetTxOutResponse{}
	if err := client.client.SendRequest("gettxout", &resp, tx, i); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) GetRawTransactionVerbose(txid string) (RawTransactionVerbose, error) {
	resp := RawTransactionVerbose{}
	if err := client.client.SendRequest("getrawtransaction", &resp, txid, 1); err != nil {
		return resp, err
	}
	return resp, nil
}
