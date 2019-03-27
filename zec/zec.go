package zec

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury"
)

type zcash struct {
	network, host, user, password string
	client                        *rpcclient.Client
	client2                       RPCCLient
	params                        *chaincfg.Params
	initiated                     bool
}

func New(network, host, user, password string) mercury.BlockchainPlugin {
	return &zcash{
		host:     host,
		user:     user,
		password: password,
		network:  network,
	}
}

func (zec *zcash) Init() error {
	client, err := rpcclient.New(
		&rpcclient.ConnConfig{
			Host:         zec.host,
			User:         zec.user,
			Pass:         zec.password,
			HTTPPostMode: true,
			DisableTLS:   true,
		},
		nil,
	)
	if err != nil {
		return err
	}

	bcInfo, err := client.GetBlockChainInfo()
	if err != nil {
		return err
	}

	var params *chaincfg.Params
	switch bcInfo.Chain {
	case "main":
		params = &chaincfg.MainNetParams
	case "test":
		params = &chaincfg.TestNet3Params
	case "regtest":
		params = &chaincfg.RegressionNetParams
	default:
		return fmt.Errorf("unsupported bitcoin network: %s", bcInfo.Chain)
	}

	zec.client = client
	zec.client2 = NewRPCClient(zec.host, zec.user, zec.password)
	zec.params = params
	zec.initiated = true
	return nil
}

type UTXO struct {
	TxHash       string `json:"txHash"`
	Amount       int64  `json:"amount"`
	ScriptPubKey string `json:"scriptPubKey"`
	Vout         uint32 `json:"vout"`
}

func (zec *zcash) GetUTXOs(address string, limit, confitmations int64) ([]UTXO, error) {
	unspents, err := zec.client2.ListUnspent(0, 999999, address)
	if err != nil {
		return []UTXO{}, err
	}

	if len(unspents) == 0 {
		if err := zec.client.ImportAddressRescan(address, "", false); err != nil {
			return []UTXO{}, err
		}

		unspents, err = zec.client2.ListUnspent(0, 999999, address)
		if err != nil {
			return []UTXO{}, err
		}
	}

	utxos := []UTXO{}
	for _, unspent := range unspents {
		amount, err := btcutil.NewAmount(unspent.Amount)
		if err != nil {
			return utxos, err
		}
		utxos = append(utxos, UTXO{
			TxHash:       unspent.TxID,
			Amount:       int64(amount.ToUnit(btcutil.AmountSatoshi)),
			ScriptPubKey: unspent.ScriptPubKey,
			Vout:         unspent.Vout,
		})
	}
	return utxos, nil
}

func (zec *zcash) Confirmations(txHashStr string) (int64, error) {
	txHash, err := chainhash.NewHashFromStr(txHashStr)
	if err != nil {
		return 0, err
	}
	tx, err := zec.client.GetTransaction(txHash)
	if err != nil {
		return 0, err
	}
	return tx.Confirmations, nil
}

func (zec *zcash) ScriptFunded(address string, value int64) (bool, int64, error) {
	if err := zec.client.ImportAddressRescan(address, "", false); err != nil {
		return false, value, err
	}
	resp, err := zec.client2.ListReceivedByAddress(address)
	if err != nil {
		return false, value, err
	}
	amount := resp.Amount * math.Pow10(8)
	return int64(amount) >= value, int64(amount), nil
}

func (zec *zcash) ScriptRedeemed(address string, value int64) (bool, int64, error) {
	if err := zec.client.ImportAddressRescan(address, "", false); err != nil {
		return false, value, err
	}
	resp, err := zec.client2.ListReceivedByAddress(address)
	if err != nil {
		return false, value, err
	}
	amount := resp.Amount * math.Pow10(8)
	utxos, err := zec.GetUTXOs(address, 999999, 0)
	if err != nil {
		return false, value, err
	}
	var balance int64
	for _, utxo := range utxos {
		balance = balance + utxo.Amount
	}
	return int64(amount) >= value && balance == 0, balance, nil
}

func (zec *zcash) ScriptSpent(scriptAddress, spenderAddress string) (bool, string, error) {
	if err := zec.client.ImportAddressRescan(scriptAddress, "", false); err != nil {
		return false, "", err
	}

	if err := zec.client.ImportAddressRescan(spenderAddress, "", false); err != nil {
		return false, "", err
	}

	txs, err := zec.client2.ListTransansactions("")
	if err != nil {
		return false, "", err
	}

	var hash string
	for _, tx := range txs {
		if tx.Address == scriptAddress && tx.Category == "receive" {
			hash = reverse(tx.TxID)
		}
	}

	txList, err := zec.client2.AddressTxIDs(spenderAddress)
	if err != nil {
		return false, "", err
	}

	for _, txID := range txList {
		rawTx, err := zec.client2.GetRawTransaction(txID)
		if err != nil {
			return false, "", err
		}

		if strings.Contains(rawTx, hash) {
			scriptSig, err := zec.client2.ExtractScriptSig(rawTx)
			return err == nil, scriptSig, err
		}
	}

	return false, "", fmt.Errorf("could not find the transaction")
}

func (zec *zcash) PublishTransaction(stx []byte) error {
	_, err := zec.client2.SendRawTransaction(stx)
	return err
}

func reverse(hexStr string) string {
	hexBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	for left, right := 0, len(hexBytes)-1; left < right; left, right = left+1, right-1 {
		hexBytes[left], hexBytes[right] = hexBytes[right], hexBytes[left]
	}
	return hex.EncodeToString(hexBytes)
}
