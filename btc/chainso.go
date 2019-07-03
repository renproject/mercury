package btc

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
)

type chainSoClient struct {
	URL     string
	network string
	token   string
}

func NewCS(network string) (BitcoinClient, error) {
	client := &chainSoClient{
		network: network,
		URL:     "https://chain.so/api/v2",
	}
	network = strings.ToLower(network)
	switch network {
	case "mainnet":
		client.token = "BTC"
	case "testnet", "testnet3", "":
		client.token = "BTCTEST"
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
	Confirmations int    `json:"confirmations"`
}

type RawAddress struct {
	Balance  string `json:"balance"`
	Received string `json:"received_value"`
	Txs      []Tx   `json:"txs"`
}

func (client chainSoClient) Init() error {
	return nil
}

func (btc chainSoClient) GetUTXO(_ context.Context, txHash string, vout int64) (UTXO, error) {
	panic("unimplemented")
}

func (client chainSoClient) GetUTXOs(ctx context.Context, address string, limit, confitmations int) ([]UTXO, error) {
	unspent, err := client.GetUnspentOutputs(ctx, address)
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
	return utxos, nil
}

func (client chainSoClient) balance(ctx context.Context, address string, confirmations int) (int64, error) {
	utxos, err := client.GetUTXOs(ctx, address, 999999, confirmations)
	if err != nil {
		return 0, nil
	}
	var balance int64
	for _, utxo := range utxos {
		balance = balance + utxo.Amount
	}
	return balance, err
}

func (client chainSoClient) GetUnspentOutputs(ctx context.Context, address string) (UnspentTxResponse, error) {
	utxos := UnspentTxResponse{}
	err := backoff(ctx, func() error {
		csoResp := ChainSoResponse{}
		resp, err := http.Get(fmt.Sprintf("%s/get_tx_unspent/%s/%s", client.URL, client.token, address))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get unspent txs: %s", respBytes)
		}

		if err := json.Unmarshal(respBytes, &csoResp); err != nil {
			return err
		}

		return json.Unmarshal(csoResp.Data, &utxos)
	})
	return utxos, err
}

func (client chainSoClient) PublishTransaction(ctx context.Context, stx []byte) error {
	return backoff(ctx, func() error {
		resp, err := http.Get(fmt.Sprintf("%s/send_tx/%s", client.URL, client.token))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get unspent txs: %s", respBytes)
		}
		return nil
	})
}

func (client chainSoClient) Health() bool {
	return true
}

func (client chainSoClient) OmniGetBalance(token int64, address string) (OmniGetBalanceResponse, error) {
	return OmniGetBalanceResponse{}, fmt.Errorf("chain.so api doesnot support omnilayer")
}

func (client chainSoClient) Confirmations(ctx context.Context, txHashStr string) (int64, error) {
	return 0, fmt.Errorf("TODO: chain.so api doesnot support omnilayer")
}

func (client chainSoClient) GetRawAddressInformation(ctx context.Context, addr string) (RawAddress, error) {
	addressInfo := RawAddress{}
	err := backoff(ctx, func() error {
		csoResp := ChainSoResponse{}
		resp, err := http.Get(fmt.Sprintf("%s/address/%s/%s", client.URL, client.token, addr))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get unspent txs: %s", respBytes)
		}

		if err := json.Unmarshal(respBytes, &csoResp); err != nil {
			return err
		}

		return json.Unmarshal(csoResp.Data, &addressInfo)
	})
	return addressInfo, err
}

func (client chainSoClient) ScriptSpent(ctx context.Context, script, spender string) (bool, string, error) {
	return false, "", fmt.Errorf("TODO: chain.so api doesnot support omnilayer")
}

func (client chainSoClient) ScriptFunded(ctx context.Context, address string, value int64) (bool, int64, error) {
	rawAddress, err := client.GetRawAddressInformation(ctx, address)
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

func (client chainSoClient) ScriptRedeemed(ctx context.Context, address string, value int64) (bool, int64, error) {
	rawAddress, err := client.GetRawAddressInformation(ctx, address)
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
	amt = big.NewFloat(0).Mul(amt, big.NewFloat(10e8))
	value, _ := amt.Int64()
	return value, nil
}

func floatToInt(val float64) int64 {
	value, _ := new(big.Float).Mul(new(big.Float).SetFloat64(val), big.NewFloat(10e8)).Int64()
	return value
}
