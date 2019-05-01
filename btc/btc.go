package btc

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury"
	"github.com/sirupsen/logrus"
)

type fullnodeClient struct {
	host, user, password, network string
	client                        *rpcclient.Client
	client2                       RPCCLient
	params                        *chaincfg.Params
	initiated                     bool
}

type bitcoin struct {
	initiated bool
	prefix    string
	logger    logrus.FieldLogger
	client    BitcoinClient
}

func New(prefix string, client BitcoinClient, logger logrus.FieldLogger) mercury.BlockchainPlugin {
	return &bitcoin{prefix: prefix, client: client, logger: logger}
}

func NewFN(network, host, user, password string) BitcoinClient {
	return &fullnodeClient{
		host:     host,
		user:     user,
		password: password,
		network:  network,
	}
}

type BitcoinClient interface {
	Init() error

	Health() bool
	GetUTXOs(ctx context.Context, address string, limit, confitmations int64) ([]UTXO, error)
	Confirmations(ctx context.Context, txHashStr string) (int64, error)
	ScriptFunded(ctx context.Context, address string, value int64) (bool, int64, error)
	ScriptRedeemed(ctx context.Context, address string, value int64) (bool, int64, error)
	ScriptSpent(ctx context.Context, scriptAddress, spenderAddress string) (bool, string, error)
	PublishTransaction(ctx context.Context, stx []byte) error
	OmniGetBalance(token int64, address string) (OmniGetBalanceResponse, error)
}

func (btc *bitcoin) Init() error {
	if err := btc.client.Init(); err != nil {
		return err
	}
	btc.initiated = true
	return nil
}

func (btc *bitcoin) Prefix() string {
	return btc.prefix
}

func (btc *bitcoin) Health() bool {
	return btc.client.Health()
}

func (btc *fullnodeClient) Init() error {
	client, err := rpcclient.New(
		&rpcclient.ConnConfig{
			Host:         btc.host,
			User:         btc.user,
			Pass:         btc.password,
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

	btc.client = client
	btc.client2 = NewRPCClient(btc.host, btc.user, btc.password)
	btc.params = params
	btc.initiated = true
	return nil
}

type UTXO struct {
	TxHash       string `json:"txHash"`
	Amount       int64  `json:"amount"`
	ScriptPubKey string `json:"scriptPubKey"`
	Vout         uint32 `json:"vout"`
}

func (btc *fullnodeClient) GetUTXOs(_ context.Context, address string, limit, confitmations int64) ([]UTXO, error) {
	net := btc.params
	addr, err := btcutil.DecodeAddress(address, net)
	if err != nil {
		return []UTXO{}, err
	}

	unspents, err := btc.client.ListUnspentMinMaxAddresses(0, 999999, []btcutil.Address{addr})
	if err != nil {
		return []UTXO{}, err
	}

	if len(unspents) == 0 {
		if err := btc.client.ImportAddressRescan(address, "", false); err != nil {
			return []UTXO{}, err
		}

		unspents, err = btc.client.ListUnspentMinMaxAddresses(0, 999999, []btcutil.Address{addr})
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

func (btc *fullnodeClient) Confirmations(_ context.Context, txHashStr string) (int64, error) {
	txHash, err := chainhash.NewHashFromStr(txHashStr)
	if err != nil {
		return 0, err
	}
	tx, err := btc.client.GetTransaction(txHash)
	if err != nil {
		return 0, err
	}
	return tx.Confirmations, nil
}

func (btc *fullnodeClient) ScriptFunded(_ context.Context, address string, value int64) (bool, int64, error) {
	if err := btc.client.ImportAddressRescan(address, "scripts", false); err != nil {
		return false, value, err
	}
	net := btc.params
	addr, err := btcutil.DecodeAddress(address, net)
	if err != nil {
		return false, value, err
	}
	amount, err := btc.client.GetReceivedByAddressMinConf(addr, 0)
	if err != nil {
		return false, value, err
	}
	return int64(amount.ToUnit(btcutil.AmountSatoshi)) >= value, int64(amount.ToUnit(btcutil.AmountSatoshi)), nil
}

func (btc *fullnodeClient) ScriptRedeemed(ctx context.Context, address string, value int64) (bool, int64, error) {
	if err := btc.client.ImportAddressRescan(address, "scripts", false); err != nil {
		return false, value, err
	}
	net := btc.params
	addr, err := btcutil.DecodeAddress(address, net)
	if err != nil {
		return false, value, err
	}
	amount, err := btc.client.GetReceivedByAddressMinConf(addr, 0)
	if err != nil {
		return false, value, err
	}
	utxos, err := btc.GetUTXOs(ctx, address, 999999, 0)
	if err != nil {
		return false, value, err
	}
	var balance int64
	for _, utxo := range utxos {
		balance = balance + utxo.Amount
	}
	return int64(amount.ToUnit(btcutil.AmountSatoshi)) >= value && balance == 0, balance, nil
}

func (btc *fullnodeClient) ScriptSpent(_ context.Context, scriptAddress, spenderAddress string) (bool, string, error) {
	randAcc := [32]byte{}
	rand.Read(randAcc[:])
	randAccName := base64.StdEncoding.EncodeToString(randAcc[:])

	if err := btc.client.ImportAddressRescan(scriptAddress, randAccName, false); err != nil {
		return false, "", err
	}

	if err := btc.client.ImportAddressRescan(spenderAddress, "", false); err != nil {
		return false, "", err
	}

	txs, err := btc.client2.ListTransansactions(randAccName)
	if err != nil {
		return false, "", err
	}

	var hash string
	for _, tx := range txs {
		if tx.Address == scriptAddress && tx.Category == "receive" {
			hash = tx.TxID
		}
	}

	txList, err := btc.client2.ListReceivedByAddress(spenderAddress)
	if err != nil {
		return false, "", err
	}

	for _, txID := range txList[0].TxIDs {
		txHash, err := chainhash.NewHashFromStr(txID)
		if err != nil {
			return false, "", err
		}

		tx, err := btc.client.GetRawTransaction(txHash)
		if err != nil {
			return false, "", err
		}

		for _, txIn := range tx.MsgTx().TxIn {
			if txIn.PreviousOutPoint.Hash.String() == hash {
				return true, hex.EncodeToString(txIn.SignatureScript), nil
			}
		}
	}

	return false, "", fmt.Errorf("could not find the transaction")
}

func (btc *fullnodeClient) PublishTransaction(_ context.Context, stx []byte) error {
	tx := wire.NewMsgTx(2)
	if err := tx.Deserialize(bytes.NewBuffer(stx)); err != nil {
		return err
	}
	_, err := btc.client.SendRawTransaction(tx, false)
	return err
}

func (btc *fullnodeClient) OmniGetBalance(token int64, address string) (OmniGetBalanceResponse, error) {
	return btc.client2.OmniGetBalance(token, address)
}

func (btc *fullnodeClient) Health() bool {
	_, err := btc.client.GetBlockChainInfo()
	return err != nil
}
