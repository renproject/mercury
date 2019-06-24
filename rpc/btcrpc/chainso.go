package btcrpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

// Chain.so function endpoints
const (
	GET_TX_UNSPENT  = "get_tx_unspent"
	IS_TX_CONFIRMED = "is_tx_confirmed"
	SEND_TX         = "send_tx"
)

// ErrUnsuccessfulChainsoResponse is returned when the response from chain.so fails.
func ErrUnsuccessfulChainsoResponse(status string) error {
	return fmt.Errorf("unsuccessful response from chain.so, status = %v", status)
}

// ChainsoResponse defines the response interface we receive from chain.so.
type ChainsoResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

// Tx has the information regarding a bitcoin transaction.
type Tx struct {
	TxId          string `json:"txid"`
	OutputNo      uint32 `json:"output_no"`
	ScriptAsm     string `json:"scriptAsm"`
	ScriptHex     string `json:"scriptHex"`
	Value         string `json:"value"`
	Confirmations int    `json:"confirmations"`
	Time          int64  `json:"time"`
}

// GetTxUnspentResponse is an successful response when calling the `GET_TX_UNSPENT` endpoint.
type GetTxUnspentResponse struct {
	Network string `json:"network"`
	Address string `json:"address"`
	Txs     []Tx   `json:"txs"`
}

// IsTxConfirmedResponse is an successful response when calling the `IS_TX_CONFIRMED` endpoint.
type IsTxConfirmedResponse struct {
	TxId          string `json:"txid"`
	Network       string `json:"network"`
	Confirmations int64  `json:"confirmations"`
	IsConfirmed   bool   `json:"is_confirmed"`
}

// SendTxResponse is an successful response when calling the `SEND_TX` endpoint.
type SendTxResponse struct {
	TxId    string `json:"txid"`
	Network string `json:"network"`
}

// chainsoClient implements the `Client` interface by interacting with chain.so API.
type chainsoClient struct {
	network btctypes.Network
}

// NewChainsoClient returns a chainsoClient which can talk to given network.
func NewChainsoClient(network btctypes.Network) Client {
	return &chainsoClient{
		network: network,
	}
}

// BlockInfo implements the `Client` interface.
func (chainso *chainsoClient) BlockInfo() btctypes.Network {
	return chainso.network
}

// GetUTXOs implements the `Client` interface.
func (chainso *chainsoClient) GetUTXOs(address btctypes.Addr, limit, confirmations int) ([]btctypes.UTXO, error) {
	url := chainsoUrl(chainso.network, GET_TX_UNSPENT, address.String())
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	var response GetTxUnspentResponse
	if err := validateChainsoResponse(resp, &response); err != nil {
		return nil, err
	}

	utxos := make([]btctypes.UTXO, len(response.Txs))
	for i, utxo := range response.Txs {
		value, err := parseChainsoValue(utxo.Value)
		if err != nil {
			return nil, err
		}
		utxos[i] = btctypes.UTXO{
			TxHash:       utxo.TxId,
			Amount:       value,
			ScriptPubKey: utxo.ScriptHex,
			Vout:         utxo.OutputNo,
		}
	}
	return utxos, nil
}

// Confirmations implements the `Client` interface.
func (chainso *chainsoClient) Confirmations(txHash string) (int64, error) {
	url := chainsoUrl(chainso.network, IS_TX_CONFIRMED, txHash)
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}

	var response IsTxConfirmedResponse
	if err := validateChainsoResponse(resp, &response); err != nil {
		return 0, err
	}

	return response.Confirmations, nil
}

// PublishTransaction implements the `Client` interface.
func (chainso *chainsoClient) PublishTransaction(stx []byte) error {
	url := chainsoUrl(chainso.network, SEND_TX)
	stxJson := struct {
		TxHex string `json:"tx_hex"`
	}{hex.EncodeToString(stx)}
	data, err := json.Marshal(stxJson)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)

	client := &http.Client{}
	resp, err := client.Post(url, "application/json", buf)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return types.UnexpectedStatusCode(http.StatusOK, resp.StatusCode)
	}
	var response SendTxResponse
	return json.NewDecoder(resp.Body).Decode(&response)
}

// chainsoUrl constructs the endpoint of chain.so API with given query name and parameters.
func chainsoUrl(network btctypes.Network, query string, params ...string) string {
	var net string
	switch network {
	case btctypes.Mainnet:
		net = "BTC"
	case btctypes.Testnet:
		net = "BTCTEST"
	default:
		panic(types.ErrUnknownNetwork)
	}
	endPoint := fmt.Sprintf("https://chain.so/api/v2/%v/%v", query, net)
	for _, param := range params {
		endPoint = fmt.Sprintf("%v/%v", endPoint, param)
	}

	return endPoint
}

// validateChainsoResponse checks the status code and body of the response. It tries to unmarshal the response into the
// given object if the response has a `success` status.
func validateChainsoResponse(resp *http.Response, data interface{}) error {
	if resp.StatusCode != http.StatusOK {
		return types.UnexpectedStatusCode(http.StatusOK, resp.StatusCode)
	}

	var chainsoResp ChainsoResponse
	if err := json.NewDecoder(resp.Body).Decode(&chainsoResp); err != nil {
		return err
	}
	if chainsoResp.Status != "success" {
		return ErrUnsuccessfulChainsoResponse(chainsoResp.Status)
	}

	return json.Unmarshal(chainsoResp.Data, &data)
}

// parseChainsoValue parses the bitcoin amount returned by chain.so API to `btctypes.Value` (BTC -> Satoshi)
func parseChainsoValue(value string) (btctypes.Value, error) {
	valueFloat, err := strconv.ParseFloat(value, 10)
	if err != nil {
		return 0, err
	}
	valueInSat := valueFloat * math.Pow10(8)
	return btctypes.Value(valueInSat), nil
}
