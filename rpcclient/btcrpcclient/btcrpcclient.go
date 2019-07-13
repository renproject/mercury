package btcrpcclient

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/renproject/mercury/rpcclient"
)

type ListReceivedByAddressResponse []ListReceivedByAddressObj

type ListReceivedByAddressObj struct {
	Address string   `json:"address"`
	TxIDs   []string `json:"txids"`
	Amount  float64  `json:"amount"`
}

type ListTransansactionsResponse []ListTransansactionsObj

type ListTransansactionsObj struct {
	Address  string  `json:"address"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	TxID     string  `json:"txid"`
}

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

type ScriptSig struct {
	Hex string `json:"hex"`
}
type VIN struct {
	ScriptSig ScriptSig `json:"scriptSig"`
}
type DecodeRawTransactionResponse struct {
	Vins []VIN `json:"vin"`
}

type OmniGetBalanceResponse struct {
	Token   int64 `json:"token"`
	Balance int64 `json:"balance"`
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
	ListTransansactions(accName string) (ListTransansactionsResponse, error)
	ListReceivedByAddress(address string) (ListReceivedByAddressObj, error)
	AddressTxIDs(address string) ([]string, error)
	SendRawTransaction(stx []byte) (string, error)
	GetTxOut(txid string, i uint32) (GetTxOutResponse, error)
	GetRawTransaction(txid string) (string, error)
	GetRawTransactionVerbose(txid string) (RawTransactionVerbose, error)
	ExtractScriptSig(txs []string) (string, error)
	OmniGetBalance(token int64, address string) (OmniGetBalanceResponse, error)
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
	if err := client.client.SendRequest("listunspent", &resp, minConf, maxConf, addresses); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) ListTransansactions(accName string) (ListTransansactionsResponse, error) {
	resp := ListTransansactionsResponse{}
	if err := client.client.SendRequest("listtransactions", &resp, accName, 999999, 0, true); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) ListReceivedByAddress(address string) (ListReceivedByAddressObj, error) {
	resp := ListReceivedByAddressResponse{}
	if err := client.client.SendRequest("listreceivedbyaddress", &resp, 0, true, true); err != nil {
		return ListReceivedByAddressObj{}, err
	}
	for _, addrObj := range resp {
		if addrObj.Address == address {
			return addrObj, nil
		}
	}
	return ListReceivedByAddressObj{}, fmt.Errorf("address not found")
}

func (client *rpcClient) AddressTxIDs(address string) ([]string, error) {
	resp, err := client.ListReceivedByAddress(address)
	if err != nil {
		return nil, err
	}
	return resp.TxIDs, nil
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

func (client *rpcClient) GetRawTransaction(txid string) (string, error) {
	resp := ""
	if err := client.client.SendRequest("getrawtransaction", &resp, txid); err != nil {
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

func (client *rpcClient) ExtractScriptSig(txs []string) (string, error) {
	resp := DecodeRawTransactionResponse{}
	if err := client.client.SendRequest("decoderawtransaction", &resp, txs); err != nil {
		return resp.Vins[0].ScriptSig.Hex, err
	}
	return resp.Vins[0].ScriptSig.Hex, nil
}

func (client *rpcClient) OmniGetBalance(token int64, address string) (OmniGetBalanceResponse, error) {
	resp := struct {
		Balance string `json:"balance"`
	}{}

	if err := client.client.SendRequest("omni_getbalance", &resp, address, token); err != nil {
		return OmniGetBalanceResponse{}, err
	}

	bal, ok := new(big.Float).SetString(resp.Balance)
	if !ok {
		return OmniGetBalanceResponse{}, fmt.Errorf("invalid amount: %s", resp.Balance)
	}

	balance, _ := new(big.Float).Mul(bal, big.NewFloat(100000000.0)).Int64()
	return OmniGetBalanceResponse{
		Token:   token,
		Balance: balance,
	}, nil
}
