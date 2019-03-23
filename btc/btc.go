package btc

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury"
)

type bitcoin struct {
	client  *rpcclient.Client
	client2 RPCCLient
	params  *chaincfg.Params
	network string
}

func New(host, user, password string) (mercury.BlockchainPlugin, error) {
	client, err := rpcclient.New(
		&rpcclient.ConnConfig{
			Host:         host,
			User:         user,
			Pass:         password,
			HTTPPostMode: true,
			DisableTLS:   true,
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	bcInfo, err := client.GetBlockChainInfo()
	if err != nil {
		return nil, err
	}

	var params *chaincfg.Params
	var network string
	switch bcInfo.Chain {
	case "main":
		params = &chaincfg.MainNetParams
	case "test":
		params = &chaincfg.TestNet3Params
	case "regtest":
		params = &chaincfg.RegressionNetParams
	default:
		return nil, fmt.Errorf("unsupported bitcoin network: %s", bcInfo.Chain)
	}

	return &bitcoin{
		client:  client,
		client2: NewRPCClient(host, user, password),
		params:  params,
		network: network,
	}, nil
}

type UTXO struct {
	TxHash       string `json:"txHash"`
	Amount       int64  `json:"amount"`
	ScriptPubKey string `json:"scriptPubKey"`
	Vout         uint32 `json:"vout"`
}

func (btc *bitcoin) GetUTXOs(address string, limit, confitmations int64) ([]UTXO, error) {
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
		utxos = append(utxos, UTXO{
			TxHash:       unspent.TxID,
			Amount:       int64(unspent.Amount * math.Pow(10, 8)),
			ScriptPubKey: unspent.ScriptPubKey,
			Vout:         unspent.Vout,
		})
	}
	return utxos, nil
}

func (btc *bitcoin) Confirmations(txHashStr string) (int64, error) {
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

func (btc *bitcoin) ScriptFunded(address string, value int64) (bool, int64, error) {
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

func (btc *bitcoin) ScriptRedeemed(address string, value int64) (bool, int64, error) {
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
	utxos, err := btc.GetUTXOs(address, 999999, 0)
	if err != nil {
		return false, value, err
	}
	var balance int64
	for _, utxo := range utxos {
		balance = balance + utxo.Amount
	}
	return int64(amount.ToUnit(btcutil.AmountSatoshi)) >= value && balance == 0, balance, nil
}

func (btc *bitcoin) ScriptSpent(scriptAddress, spenderAddress string) (bool, string, error) {
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

func (btc *bitcoin) PublishTransaction(stx []byte) error {
	tx := wire.NewMsgTx(2)
	if err := tx.Deserialize(bytes.NewBuffer(stx)); err != nil {
		return err
	}
	_, err := btc.client.SendRawTransaction(tx, false)
	return err
}

func (btc *bitcoin) Network() string {
	return btc.params.Name
}
