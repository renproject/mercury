package btc

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/renproject/mercury/rpc"
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
	OmniGetBalance(token int64, address string) (OmniGetBalanceResponse, error)
}

type OmniGetBalanceResponse struct {
	Token   int64 `json:"token"`
	Balance int64 `json:"balance"`
}

type rpcClient struct {
	rpc.Client
}

func NewRPCClient(host, user, password string) RPCCLient {
	return &rpcClient{rpc.NewClient(host, user, password)}
}

func (client *rpcClient) ListTransansactions(accName string) (ListTransansactionsResponse, error) {
	resp := ListTransansactionsResponse{}
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"listtransactions\", \"params\": [\"%s\", 999999, 0, true] }", accName))
	if err := client.SendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) ListReceivedByAddress(address string) (ListReceivedByAddressResponse, error) {
	resp := ListReceivedByAddressResponse{}
	req := []byte(fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"listreceivedbyaddress\", \"params\": [0, false, true, \"%s\"] }", address))
	if err := client.SendRequest(req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (client *rpcClient) OmniGetBalance(token int64, address string) (OmniGetBalanceResponse, error) {
	resp := struct {
		Balance string `json:"balance"`
	}{}
	req := []byte(fmt.Sprintf(
		"{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"omni_getbalance\", \"params\": [\"%s\", %d] }",
		address,
		token,
	))

	if err := client.SendRequest(req, &resp); err != nil {
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
