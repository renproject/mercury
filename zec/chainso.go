package zec

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
)

type chainSoClient struct {
	token   string
	network string
	URL     string
}

func NewCS(network string) (ZCashClient, error) {
	client := &chainSoClient{
		network: network,
		URL:     "https://chain.so/api/v2",
	}
	network = strings.ToLower(network)
	switch network {
	case "mainnet":
		client.token = "ZEC"
	case "testnet", "testnet3", "":
		client.token = "ZECTEST"
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
	return client, nil
}

type ChainSoResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

type UnspentTxResponse struct {
	Network string `json:"network"`
	Address string `json:"address"`
	Txs     []Tx   `json:"txs"`
}

type Tx struct {
	ID            string `json:"txid"`
	Value         string `json:"value"`
	ScriptHex     string `json:"script_hex"`
	OutNo         uint32 `json:"output_no"`
	Confirmations int64  `json:"confirmations"`
}

type RawAddress struct {
	Balance  string `json:"balance"`
	Received string `json:"received_value"`
	Txs      []Tx   `json:"txs"`
}

func (client chainSoClient) Init() error {
	return nil
}

func (client chainSoClient) Network() string {
	return client.network
}

func (client chainSoClient) GetUTXOs(address string, limit, confitmations int64) ([]UTXO, error) {
	unspent, err := client.GetUnspentOutputs(address)
	if err != nil {
		return nil, err
	}

	utxos := []UTXO{}
	for _, output := range unspent.Txs {
		if output.Confirmations >= confitmations {
			amount, err := strToInt(output.Value)
			if err != nil {
				return nil, fmt.Errorf("unable to convert %s into sat: %v", output.Value, err)
			}

			utxos = append(utxos, UTXO{
				TxHash:       output.ID,
				Amount:       amount,
				ScriptPubKey: output.ScriptHex,
				Vout:         output.OutNo,
			})
		}
	}
	if len(utxos) > int(limit) {
		return utxos[:limit], nil
	}
	return utxos, nil
}

func (client chainSoClient) balance(address string, confirmations int64) (int64, error) {
	utxos, err := client.GetUTXOs(address, 999999, confirmations)
	if err != nil {
		return 0, nil
	}
	var balance int64
	for _, utxo := range utxos {
		balance = balance + utxo.Amount
	}
	return balance, err
}

func (client chainSoClient) GetUnspentOutputs(address string) (UnspentTxResponse, error) {
	utxos := UnspentTxResponse{}
	csoResp := ChainSoResponse{}
	resp, err := http.Get(fmt.Sprintf("%s/get_tx_unspent/%s/%s", client.URL, client.token, address))
	if err != nil {
		return utxos, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return utxos, err
	}
	if resp.StatusCode != http.StatusOK {
		return utxos, fmt.Errorf("failed to get unspent txs: %s", respBytes)
	}

	if err := json.Unmarshal(respBytes, &csoResp); err != nil {
		return utxos, err
	}

	return utxos, json.Unmarshal(csoResp.Data, &utxos)
}

func (client chainSoClient) PublishTransaction(stx []byte) error {
	txObj := struct {
		TxHex string `json:"tx_hex"`
	}{
		TxHex: hex.EncodeToString(stx),
	}

	fmt.Println(hex.EncodeToString(stx))

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(txObj); err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s/send_tx/%s", client.URL, client.token), "encoding/json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println(fmt.Errorf("failed to publish transaction txs: %s", respBytes))
		return fmt.Errorf("failed to publish transaction txs: %s", respBytes)
	}
	return nil
}

func (client chainSoClient) Health() bool {
	return true
}

func (client chainSoClient) Confirmations(txHashStr string) (int64, error) {
	return 0, fmt.Errorf("TODO: chain.so api doesnot support confirmations")
}

func (client chainSoClient) GetRawAddressInformation(addr string) (RawAddress, error) {
	addressInfo := RawAddress{}
	csoResp := ChainSoResponse{}
	resp, err := http.Get(fmt.Sprintf("%s/address/%s/%s", client.URL, client.token, addr))
	if err != nil {
		return addressInfo, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return addressInfo, err
	}
	if resp.StatusCode != http.StatusOK {
		return addressInfo, fmt.Errorf("failed to get unspent txs: %s", respBytes)
	}

	if err := json.Unmarshal(respBytes, &csoResp); err != nil {
		return addressInfo, err
	}

	return addressInfo, json.Unmarshal(csoResp.Data, &addressInfo)
}

func (client chainSoClient) ScriptSpent(script, spender string) (bool, string, error) {
	return false, "", fmt.Errorf("TODO: chain.so api doesnot support omnilayer")
}

func (client chainSoClient) ScriptFunded(address string, value int64) (bool, int64, error) {
	rawAddress, err := client.GetRawAddressInformation(address)
	if err != nil {
		return false, 0, err
	}

	received, err := strToInt(rawAddress.Received)
	if err != nil {
		return false, 0, err
	}

	balance, err := strToInt(rawAddress.Balance)
	if err != nil {
		return false, 0, err
	}

	return received >= value, balance, nil
}

func (client chainSoClient) ScriptRedeemed(address string, value int64) (bool, int64, error) {
	rawAddress, err := client.GetRawAddressInformation(address)
	if err != nil {
		return false, 0, err
	}
	received, err := strToInt(rawAddress.Received)
	if err != nil {
		return false, 0, err
	}

	balance, err := strToInt(rawAddress.Balance)
	if err != nil {
		return false, 0, err
	}

	return received >= value && balance == 0, balance, nil
}

func strToInt(val string) (int64, error) {
	amt, ok := new(big.Float).SetString(val)
	if !ok {
		return -1, fmt.Errorf("failed to convert %s to float", val)
	}
	value, _ := new(big.Float).Mul(amt, new(big.Float).SetInt(big.NewInt(100000000))).Int64()
	return value, nil
}
