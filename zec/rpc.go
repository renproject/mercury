package zec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Result json.RawMessage `json:"result"`
}

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
	host     string
	user     string
	password string
}

func NewRPCClient(host, user, password string) RPCCLient {
	return &rpcClient{
		host, user, password,
	}
}

func (client *rpcClient) ListUnspent(minConf, maxConf int64, address string) (ListUnspentResponse, error) {
	resp := ListUnspentResponse{}
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"listunspent\", \"params\": [%d, %d, [\"%s\"]]  }", minConf, maxConf, address))
	if err := client.sendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) ListTransansactions(accName string) (ListTransansactionsResponse, error) {
	resp := ListTransansactionsResponse{}
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"listtransactions\", \"params\": [\"%s\", 999999, 0, true] }", accName))
	if err := client.sendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) ListReceivedByAddress(address string) (ListReceivedByAddressObj, error) {
	resp := ListReceivedByAddressResponse{}
	req := []byte("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"listreceivedbyaddress\", \"params\": [0, true, true] }")
	if err := client.sendRequest(req, &resp); err != nil {
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
	if err := client.sendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) GetRawTransaction(txid string) (string, error) {
	resp := ""
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"getrawtransaction\", \"params\": [\"%s\"] }", txid))
	if err := client.sendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) ExtractScriptSig(tx string) (string, error) {
	resp := DecodeRawTransactionResponse{}
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"decoderawtransaction\", \"params\": [\"%s\"] }", tx))
	if err := client.sendRequest(req, &resp); err != nil {
		return resp.Vins[0].ScriptSig.Hex, err
	}
	return resp.Vins[0].ScriptSig.Hex, nil
}

func (client *rpcClient) sendRequest(data []byte, response interface{}) error {
	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s", client.host), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	request.SetBasicAuth(client.user, client.password)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(msg))
	}
	fmt.Println(msg)

	result := Response{}
	if err := json.Unmarshal(msg, &result); err != nil {
		return err
	}

	return json.Unmarshal(result.Result, response)
}
