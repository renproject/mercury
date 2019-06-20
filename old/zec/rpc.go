package zec

import (
	"fmt"

	"github.com/renproject/mercury/old/rpc"
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

type RPCCLient interface {
	ListUnspent(minConf, maxConf int64, address string) (ListUnspentResponse, error)
	ListTransansactions(accName string) (ListTransansactionsResponse, error)
	ListReceivedByAddress(address string) (ListReceivedByAddressObj, error)
	AddressTxIDs(address string) ([]string, error)
	SendRawTransaction(stx []byte) (string, error)
	GetRawTransaction(txid string) (string, error)
	ExtractScriptSig(tx string) (string, error)
}

type rpcClient struct {
	client rpc.Client
}

func NewRPCClient(host, user, password string) RPCCLient {
	return &rpcClient{
		rpc.NewClient(host, user, password),
	}
}

func (client *rpcClient) ListUnspent(minConf, maxConf int64, address string) (ListUnspentResponse, error) {
	resp := ListUnspentResponse{}
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"listunspent\", \"params\": [%d, %d, [\"%s\"]]  }", minConf, maxConf, address))
	if err := client.client.SendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) ListTransansactions(accName string) (ListTransansactionsResponse, error) {
	resp := ListTransansactionsResponse{}
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"listtransactions\", \"params\": [\"%s\", 999999, 0, true] }", accName))
	if err := client.client.SendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) ListReceivedByAddress(address string) (ListReceivedByAddressObj, error) {
	resp := ListReceivedByAddressResponse{}
	req := []byte("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"listreceivedbyaddress\", \"params\": [0, true, true] }")
	if err := client.client.SendRequest(req, &resp); err != nil {
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
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"sendrawtransaction\", \"params\": [\"%x\", false] }", stx))
	if err := client.client.SendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) GetRawTransaction(txid string) (string, error) {
	resp := ""
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"getrawtransaction\", \"params\": [\"%s\"] }", txid))
	if err := client.client.SendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) ExtractScriptSig(tx string) (string, error) {
	resp := DecodeRawTransactionResponse{}
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"decoderawtransaction\", \"params\": [\"%s\"] }", tx))
	if err := client.client.SendRequest(req, &resp); err != nil {
		return resp.Vins[0].ScriptSig.Hex, err
	}
	return resp.Vins[0].ScriptSig.Hex, nil
}
