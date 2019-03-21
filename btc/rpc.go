package btc

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
	TxIDs []string `json:"txids"`
}

type ListTransansactionsResponse []ListTransansactionsObj

type ListTransansactionsObj struct {
	Address  string  `json:"address"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	TxID     string  `json:"txid"`
}

type RPCCLient interface {
	ListTransansactions(accName string) (ListTransansactionsResponse, error)
	ListReceivedByAddress(address string) (ListReceivedByAddressResponse, error)
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

func (client *rpcClient) ListTransansactions(accName string) (ListTransansactionsResponse, error) {
	resp := ListTransansactionsResponse{}
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"listtransactions\", \"params\": [\"%s\", 999999, 0, true] }", accName))
	if err := client.sendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) ListReceivedByAddress(address string) (ListReceivedByAddressResponse, error) {
	resp := ListReceivedByAddressResponse{}
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"listreceivedbyaddress\", \"params\": [0, false, true, \"%s\"] }", address))
	if err := client.sendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
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

	result := Response{}
	if err := json.Unmarshal(msg, &result); err != nil {
		return err
	}

	return json.Unmarshal(result.Result, response)
}
