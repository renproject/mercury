package btc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
)

type PreviousOut struct {
	TransactionHash  string `json:"hash"`
	Value            uint64 `json:"value"`
	TransactionIndex uint64 `json:"tx_index"`
	VoutNumber       uint8  `json:"n"`
	Address          string `json:"addr"`
}

type Input struct {
	PrevOut PreviousOut `json:"prev_out"`
	Script  string      `json:"script"`
}

type Output struct {
	Value           uint64 `json:"value"`
	TransactionHash string `json:"hash"`
	Script          string `json:"script"`
}

type Transaction struct {
	TransactionHash  string   `json:"hash"`
	Version          uint8    `json:"ver"`
	VinSize          uint32   `json:"vin_sz"`
	VoutSize         uint32   `json:"vout_sz"`
	Size             int64    `json:"size"`
	RelayedBy        string   `json:"relayed_by"`
	BlockHeight      int64    `json:"block_height"`
	TransactionIndex uint64   `json:"tx_index"`
	Inputs           []Input  `json:"inputs"`
	Outputs          []Output `json:"out"`
}

type Block struct {
	BlockHash         string        `json:"hash"`
	Version           uint8         `json:"ver"`
	PreviousBlockHash string        `json:"prev_block"`
	MerkleRoot        string        `json:"mrkl_root"`
	Time              int64         `json:"time"`
	Bits              int64         `json:"bits"`
	Nonce             int64         `json:"nonce"`
	TransactionCount  int           `json:"n_tx"`
	Size              int64         `json:"size"`
	BlockIndex        uint64        `json:"block_index"`
	MainChain         bool          `json:"main_chain"`
	Height            int64         `json:"height"`
	ReceivedTime      int64         `json:"received_time"`
	RelayedBy         string        `json:"relayed_by"`
	Transactions      []Transaction `json:"tx"`
}

type Blocks struct {
	Blocks []Block `json:"block"`
}

type SingleAddress struct {
	Address      string        `json:"address"`
	Received     int64         `json:"total_received"`
	Sent         int64         `json:"total_sent"`
	Balance      int64         `json:"final_balance"`
	Transactions []Transaction `json:"txs"`
}

type Address struct {
	PublicKeyHash    string `json:"hash160"`
	Address          string `json:"address"`
	TransactionCount int64  `json:"n_tx"`
	Received         int64  `json:"total_received"`
	Sent             int64  `json:"total_sent"`
	Balance          int64  `json:"final_balance"`
}

type MultiAddress struct {
	Addresses    []Address     `json:"addresses"`
	Transactions []Transaction `json:"txs"`
}

type UnspentOutput struct {
	TransactionAge          string `json:"tx_age"`
	TransactionHash         string `json:"tx_hash_big_endian"`
	TransactionIndex        uint32 `json:"tx_index"`
	TransactionOutputNumber uint32 `json:"tx_output_n"`
	ScriptPubKey            string `json:"script"`
	Amount                  int64  `json:"value"`
}

type Unspent struct {
	Outputs []UnspentOutput `json:"unspent_outputs"`
}

type LatestBlock struct {
	Hash       string `json:"hash"`
	Time       int64  `json:"time"`
	BlockIndex int64  `json:"block_index"`
	Height     int64  `json:"height"`
}

type blockchainInfoClient struct {
	URL    string
	Params *chaincfg.Params
}

func NewBI(network string) (BitcoinClient, error) {
	network = strings.ToLower(network)
	switch network {
	case "mainnet":
		return &blockchainInfoClient{
			URL:    "https://blockchain.info",
			Params: &chaincfg.MainNetParams,
		}, nil
	case "testnet", "testnet3", "":
		return &blockchainInfoClient{
			URL:    "https://testnet.blockchain.info",
			Params: &chaincfg.TestNet3Params,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
}

func (client *blockchainInfoClient) Init() error {
	return nil
}

func (client *blockchainInfoClient) GetUTXOs(ctx context.Context, address string, limit, confitmations int) ([]UTXO, error) {
	unspent, err := client.GetUnspentOutputs(ctx, address, limit, confitmations)
	if err != nil {
		return nil, err
	}

	utxos := []UTXO{}
	for _, output := range unspent.Outputs {
		utxos = append(utxos, UTXO{
			TxHash:       output.TransactionHash,
			Amount:       output.Amount,
			ScriptPubKey: output.ScriptPubKey,
			Vout:         output.TransactionOutputNumber,
		})
	}
	return utxos, nil
}

func (client *blockchainInfoClient) balance(ctx context.Context, address string, confirmations int) (int64, error) {
	utxos, err := client.GetUTXOs(ctx, address, 999999, confirmations)
	if err != nil {
		return 0, nil
	}
	var balance int64
	for _, utxo := range utxos {
		balance = balance + Amount
	}
	return balance, err
}

func (client *blockchainInfoClient) GetUnspentOutputs(ctx context.Context, address string, limit, confitmations int) (Unspent, error) {
	if limit == 0 {
		limit = 250
	}
	utxos := Unspent{}
	err := backoff(ctx, func() error {
		resp, err := http.Get(fmt.Sprintf("%s/unspent?active=%s&confirmations=%d&limit=%d", client.URL, address, confitmations, limit))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if string(respBytes) == "No free outputs to spend" {
			return nil
		}
		return json.Unmarshal(respBytes, &utxos)
	})
	return utxos, err
}

func (client *blockchainInfoClient) GetRawTransaction(ctx context.Context, txhash string) (Transaction, error) {
	transaction := Transaction{}
	err := backoff(ctx, func() error {
		resp, err := http.Get(fmt.Sprintf("%s/rawtx/%s", client.URL, txhash))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		txBytes, err := ioutil.ReadAll(resp.Body)
		return json.Unmarshal(txBytes, &transaction)
	})
	return transaction, err
}

func (client *blockchainInfoClient) Confirmations(ctx context.Context, txhash string) (int64, error) {
	tx, err := client.GetRawTransaction(ctx, txhash)
	if err != nil {
		return 0, err
	}
	if tx.BlockHeight != 0 {
		latest, err := client.LatestBlock(ctx)
		if err != nil {
			return 0, err
		}
		return 1 + (latest.Height - tx.BlockHeight), nil
	}
	return 0, nil
}

func (client *blockchainInfoClient) GetRawAddressInformation(ctx context.Context, addr string) (SingleAddress, error) {
	addressInfo := SingleAddress{}
	err := backoff(ctx, func() error {
		resp, err := http.Get(fmt.Sprintf("%s/rawaddr/%s", client.URL, addr))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		addrBytes, err := ioutil.ReadAll(resp.Body)
		return json.Unmarshal(addrBytes, &addressInfo)
	})
	return addressInfo, err
}

func (client *blockchainInfoClient) LatestBlock(ctx context.Context) (LatestBlock, error) {
	latestBlock := LatestBlock{}
	err := backoff(ctx, func() error {
		resp, err := http.Get(fmt.Sprintf("%s/latestblock", client.URL))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		latestBlockBytes, err := ioutil.ReadAll(resp.Body)
		return json.Unmarshal(latestBlockBytes, &latestBlock)
	})
	return latestBlock, err
}

func (client *blockchainInfoClient) PublishTransaction(ctx context.Context, stx []byte) error {
	data := url.Values{}
	data.Set("tx", hex.EncodeToString(stx))
	err := backoff(ctx, func() error {
		httpClient := &http.Client{}
		r, err := http.NewRequest("POST", fmt.Sprintf("%s/pushtx", client.URL), strings.NewReader(data.Encode())) // URL-encoded payload
		if err != nil {
			return err
		}
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := httpClient.Do(r)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		stxResultBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		stxResult := string(stxResultBytes)
		if !strings.Contains(stxResult, "Transaction Submitted") {
			return fmt.Errorf("unable to submit tx: %s", stxResult)
		}
		return nil
	})
	return err
}

func (client *blockchainInfoClient) ScriptSpent(ctx context.Context, script, spender string) (bool, string, error) {
	addrInfo, err := client.GetRawAddressInformation(ctx, script)
	if err != nil || addrInfo.Sent == 0 {
		return false, "", err
	}
	for _, tx := range addrInfo.Transactions {
		for i := range tx.Inputs {
			if tx.Inputs[i].PrevOut.Address == addrInfo.Address {
				return true, tx.Inputs[i].Script, nil
			}
		}
	}
	return true, "", fmt.Errorf("could not find a spending transaction")
}

func (client *blockchainInfoClient) ScriptFunded(ctx context.Context, address string, value int64) (bool, int64, error) {
	rawAddress, err := client.GetRawAddressInformation(ctx, address)
	if err != nil {
		return false, 0, err
	}
	return rawAddress.Received >= value, rawAddress.Received, nil
}

func (client *blockchainInfoClient) ScriptRedeemed(ctx context.Context, address string, value int64) (bool, int64, error) {
	rawAddress, err := client.GetRawAddressInformation(ctx, address)
	if err != nil {
		return false, 0, err
	}
	return rawAddress.Received >= value && rawAddress.Balance == 0, rawAddress.Balance, nil
}

func (client *blockchainInfoClient) NetworkParams() *chaincfg.Params {
	return client.Params
}

func (client *blockchainInfoClient) OmniGetBalance(token int64, address string) (OmniGetBalanceResponse, error) {
	return OmniGetBalanceResponse{}, fmt.Errorf("blockchain info doesnot support omnilayer")
}

func (client *blockchainInfoClient) Health() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.LatestBlock(ctx)
	return err != nil
}

func backoff(ctx context.Context, f func() error) error {
	duration := time.Duration(1000)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := f()
			if err == nil {
				return nil
			}
			fmt.Printf("Error: %v, will try again in %d sec\n", err, duration)
			time.Sleep(duration * time.Millisecond)
			duration = time.Duration(float64(duration) * 1.6)
		}
	}
}
