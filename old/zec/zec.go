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
	"github.com/sirupsen/logrus"
)

type zcash struct {
	prefix    string
	client    ZCashClient
	initiated bool
	logger    logrus.FieldLogger
}

type fnClient struct {
	host     string
	user     string
	password string
	network  string

	client    *rpcclient.Client
	client2   RPCCLient
	params    *chaincfg.Params
	initiated bool
}

type ZCashClient interface {
	Init() error

	Health() bool
	GetUTXOs(address string, limit, confitmations int64) ([]UTXO, error)
	Confirmations(txHashStr string) (int64, error)
	ScriptFunded(address string, value int64) (bool, int64, error)
	ScriptRedeemed(address string, value int64) (bool, int64, error)
	ScriptSpent(scriptAddress, spenderAddress string) (bool, string, error)
	PublishTransaction(stx []byte) error
}

func NewFN(network, host, user, password string) ZCashClient {
	return &fnClient{
		host:     host,
		user:     user,
		password: password,
	}
}

func New(prefix string, client ZCashClient, logger logrus.FieldLogger) mercury.BlockchainPlugin {
	return &zcash{
		prefix: prefix,
		logger: logger,
		client: client,
	}
}

func (zec *fnClient) Init() error {
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

func (zec *fnClient) Health() bool {
	_, err := zec.client.GetBlockChainInfo()
	return err != nil
}
func (zec *zcash) Health() bool {
	return zec.client.Health()
}

func (zec *zcash) Init() error {
	err := zec.client.Init()
	if err != nil {
		zec.logger.Error(err)
		return err
	}
	zec.initiated = true
	return nil
}

func (zec *zcash) Prefix() string {
	return zec.prefix
}

type UTXO struct {
	TxHash       string `json:"txHash"`
	Amount       int64  `json:"amount"`
	ScriptPubKey string `json:"scriptPubKey"`
	Vout         uint32 `json:"vout"`
}

func (zec *fnClient) GetUTXOs(address string, limit, confitmations int64) ([]UTXO, error) {
	unspents, err := ListUnspent(0, 999999, address)
	if err != nil {
		return []UTXO{}, err
	}

	if len(unspents) == 0 {
		if err := zec.client.ImportAddressRescan(address, "", false); err != nil {
			return []UTXO{}, err
		}

		unspents, err = ListUnspent(0, 999999, address)
		if err != nil {
			return []UTXO{}, err
		}
	}

	utxos := []UTXO{}
	for _, unspent := range unspents {
		amount, err := btcutil.NewAmount(Amount)
		if err != nil {
			return utxos, err
		}
		utxos = append(utxos, UTXO{
			TxHash:       TxID,
			Amount:       int64(amount.ToUnit(btcutil.AmountSatoshi)),
			ScriptPubKey: ScriptPubKey,
			Vout:         Vout,
		})
	}
	if len(utxos) > int(limit) {
		return utxos[:limit], nil
	}
	return utxos, nil
}

func (zec *fnClient) Confirmations(txHashStr string) (int64, error) {
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

func (zec *fnClient) ScriptFunded(address string, value int64) (bool, int64, error) {
	if err := zec.client.ImportAddressRescan(address, "", false); err != nil {
		return false, value, err
	}
	resp, err := ListReceivedByAddress(address)
	if err != nil {
		return false, value, err
	}
	amount := Amount * math.Pow10(8)
	return int64(amount) >= value, int64(amount), nil
}

func (zec *fnClient) ScriptRedeemed(address string, value int64) (bool, int64, error) {
	if err := zec.client.ImportAddressRescan(address, "", false); err != nil {
		return false, value, err
	}
	resp, err := ListReceivedByAddress(address)
	if err != nil {
		return false, value, err
	}
	amount := Amount * math.Pow10(8)
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

func (zec *fnClient) ScriptSpent(scriptAddress, spenderAddress string) (bool, string, error) {
	if err := zec.client.ImportAddressRescan(scriptAddress, "", false); err != nil {
		return false, "", err
	}

	if err := zec.client.ImportAddressRescan(spenderAddress, "", false); err != nil {
		return false, "", err
	}

	txs, err := ListTransansactions("")
	if err != nil {
		return false, "", err
	}

	var hash string
	for _, tx := range txs {
		if Address == scriptAddress && Category == "receive" {
			hash = reverse(TxID)
		}
	}

	txList, err := AddressTxIDs(spenderAddress)
	if err != nil {
		return false, "", err
	}

	for _, txID := range txList {
		rawTx, err := GetRawTransaction(txID)
		if err != nil {
			return false, "", err
		}

		if strings.Contains(rawTx, hash) {
			scriptSig, err := ExtractScriptSig(rawTx)
			return err == nil, scriptSig, err
		}
	}

	return false, "", fmt.Errorf("could not find the transaction")
}

func (zec *fnClient) PublishTransaction(stx []byte) error {
	_, err := SendRawTransaction(stx)
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
