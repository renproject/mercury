package btcrpc

import (
	"bytes"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/types/btctypes"
)

// Client is a RPC client which can send and retrieve information from bitcoin blockchain through JSON-RPC.
type Client interface {

	// Blockinfo returns a boolean indicates the health status of the Client.
	Blockinfo() btctypes.Network

	// GetUTXOs returns the utxos of the given address filtered by the given limit and confirmations .
	GetUTXOs(address btctypes.Addr, limit, confirmations int) ([]btctypes.UTXO, error)

	// Confirmations returns the number of block confirmations of the given txHash.
	Confirmations(txHash string) (int64, error)

	// PublishTransaction publish the raw tx to the blockchain.
	PublishTransaction(stx []byte) error
}

//
type nodeClient struct {
	client  *rpcclient.Client
	network btctypes.Network
}

func NewNodeClient(network btctypes.Network, host, username, password string) (Client, error) {
	config := &rpcclient.ConnConfig{
		Host:         host,
		User:         username,
		Pass:         password,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := rpcclient.New(config, nil)
	if err != nil {
		return nil, err
	}

	return &nodeClient{
		client:  client,
		network: network,
	}, nil
}

func (node *nodeClient) Blockinfo() btctypes.Network {
	return node.network
}

func (node *nodeClient) GetUTXOs(address btctypes.Addr, limit, confirmation int) ([]btctypes.UTXO, error) {
	addresses := []btcutil.Address{address}
	utxoResult, err := node.client.ListUnspentMinMaxAddresses(confirmation, 999999, addresses)
	if err != nil {
		return nil, err
	}
	if len(utxoResult) > limit {
		utxoResult = utxoResult[:limit]
	}

	utxos := make([]btctypes.UTXO, len(utxoResult))
	for i, unspent := range utxoResult {
		amount, err := btcutil.NewAmount(unspent.Amount)
		if err != nil {
			return utxos, err
		}
		utxos[i] = btctypes.UTXO{
			TxHash:       unspent.TxID,
			Amount:       btctypes.Value(amount),
			ScriptPubKey: unspent.ScriptPubKey,
			Vout:         unspent.Vout,
		}
	}

	return utxos, nil
}

func (node *nodeClient) Confirmations(hash string) (int64, error) {
	txHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return 0, err
	}

	tx, err := node.client.GetTransaction(txHash)
	if err != nil {
		return 0, err
	}
	return tx.Confirmations, nil
}

func (node *nodeClient) PublishTransaction(stx []byte) error {
	tx := wire.NewMsgTx(2)
	if err := tx.Deserialize(bytes.NewBuffer(stx)); err != nil {
		return err
	}

	_, err := node.client.SendRawTransaction(tx, false)
	return err
}
