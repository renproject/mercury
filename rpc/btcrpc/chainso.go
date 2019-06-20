package btcrpc

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

// chainso endpoint
const GET_TX_UNSPENT = "get_tx_unspent"

func ErrUnsuccessfulChainsoResponse(status string) error {
	return fmt.Errorf("unsuccessful response from chain.so, status = %v", status)
}

type ChainsoResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

type Tx struct {
	Txid          string `json:"txid"`
	OutputNo      uint32 `json:"output_no"`
	ScriptAsm     string `json:"scriptAsm"`
	ScriptHex     string `json:"scriptHex"`
	Value         string `json:"value"`
	Confirmations int    `json:"confirmations"`
	Time          int64  `json:"time"`
}

type GetTxUnspentResponse struct {
	Network string `json:"network"`
	Address string `json:"address"`
	Txs     []Tx   `json:"txs"`
}

type chainsoClient struct {
	network btctypes.Network
}

func NewChainsoClient(network btctypes.Network) Client {
	return &chainsoClient{
		network: network,
	}
}

func (chainso *chainsoClient) Blockinfo() btctypes.Network {
	return chainso.network
}

func (chainso *chainsoClient) GetUTXOs(address btctypes.Addr, limit, confirmations int) ([]btctypes.UTXO, error) {
	client := &http.Client{}
	url := chainsoUrl(chainso.network, GET_TX_UNSPENT, address.String())
	log.Print("url", url)
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
			TxHash:       utxo.Txid,
			Amount:       value,
			ScriptPubKey: utxo.ScriptHex,
			Vout:         utxo.OutputNo,
		}
	}
	return utxos, nil
}

func (*chainsoClient) Confirmations(txHashStr string) (int64, error) {
	panic("implement me")
}

func (*chainsoClient) PublishTransaction(stx []byte) error {
	panic("implement me")
}

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

func parseChainsoValue(value string) (btctypes.Value, error) {
	valueFloat, err := strconv.ParseFloat(value, 10)
	if err != nil {
		return 0, err
	}
	valueInSat := valueFloat * math.Pow10(8)
	return btctypes.Value(valueInSat), nil
}
