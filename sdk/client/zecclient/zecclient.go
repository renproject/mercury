package zecclient

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/types/zectypes"
)

const (
	Dust               = zectypes.Amount(600)
	MainnetMercuryURL  = "206.189.83.88:5000/zec/mainnet"
	TestnetMercuryURL  = "206.189.83.88:5000/zec/testnet"
	LocalnetMercuryURL = "0.0.0.0:5000/zec/testnet"
)

var (
	ErrInvalidTxHash  = errors.New("invalid tx hash")
	ErrTxHashNotFound = errors.New("tx hash not found")
	ErrUTXOSpent      = errors.New("utxo spent or invalid index")
)

// TODO: This should use a common interface to Bitcoin (types should be generic)
type Client interface {
	Network() zectypes.Network
	UTXO(txHash zectypes.TxHash, index uint32) (zectypes.UTXO, error)
	UTXOsFromAddress(address zectypes.Address) (zectypes.UTXOs, error)
	Confirmations(txHash zectypes.TxHash) (zectypes.Confirmations, error)
	BuildUnsignedTx(utxos zectypes.UTXOs, recipients zectypes.Recipients, refundTo zectypes.Address, gas zectypes.Amount) (zectypes.Tx, error)
	SubmitSignedTx(stx zectypes.Tx) (zectypes.TxHash, error)
}

// Client is a client which is used to talking with certain ZCash network. It can interacting with the blockchain
// through Mercury server.
type client struct {
	network zectypes.Network
	// FIXME: We do not want to rely on the Bitcoin RPC client in this package as there may be
	// subtle differences.
	client *rpcclient.Client

	config chaincfg.Params
	url    string
}

// New returns a new Client of given ZCash network.
func New(network zectypes.Network) (Client, error) {
	config := &rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	switch network {
	case zectypes.Mainnet:
		config.Host = MainnetMercuryURL
	case zectypes.Testnet:
		config.Host = TestnetMercuryURL
	case zectypes.Localnet:
		config.Host = LocalnetMercuryURL
	default:
		panic("unknown zcash network")
	}

	c, err := rpcclient.New(config, nil)
	if err != nil {
		return &client{}, err
	}

	return &client{
		network: network,
		client:  c,
		config:  *network.Params(),
		url:     config.Host,
	}, nil
}

func (c *client) Network() zectypes.Network {
	return c.network
}

// UTXO returns the UTXO for the given transaction hash and index.
func (c *client) UTXO(txHash zectypes.TxHash, index uint32) (zectypes.UTXO, error) {
	txHashBytes, err := chainhash.NewHashFromStr(string(txHash))
	if err != nil {
		return zectypes.UTXO{}, ErrInvalidTxHash
	}
	tx, err := c.client.GetRawTransactionVerbose(txHashBytes)
	if err != nil {
		return zectypes.UTXO{}, ErrTxHashNotFound
	}

	txOut, err := c.client.GetTxOut(txHashBytes, index, true)
	if err != nil {
		return zectypes.UTXO{}, fmt.Errorf("cannot get tx output from zec client: %v", err)
	}

	// Check if UTXO has been spent.
	if txOut == nil {
		return zectypes.UTXO{}, ErrUTXOSpent
	}

	amount, err := btcutil.NewAmount(txOut.Value)
	if err != nil {
		return zectypes.UTXO{}, fmt.Errorf("cannot parse amount received from zec client: %v", err)
	}
	return zectypes.UTXO{
		TxHash:       zectypes.TxHash(tx.Txid),
		Amount:       zectypes.Amount(amount),
		ScriptPubKey: txOut.ScriptPubKey.Hex,
		Vout:         index,
	}, nil
}

// UTXOsFromAddress returns the UTXOs for a given address. Important: this function will not return any UTXOs for
// addresses that have not been imported into the Bitcoin node.
func (c *client) UTXOsFromAddress(address zectypes.Address) (zectypes.UTXOs, error) {
	outputs, err := c.client.ListUnspentMinMaxAddresses(0, 999999, []btcutil.Address{address})
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve utxos from zec client: %v", err)
	}

	utxos := make(zectypes.UTXOs, len(outputs))
	for i, output := range outputs {
		amount, err := btcutil.NewAmount(output.Amount)
		if err != nil {
			return nil, fmt.Errorf("cannot parse amount received from zec client: %v", err)
		}
		utxos[i] = zectypes.UTXO{
			TxHash:       zectypes.TxHash(output.TxID),
			Amount:       zectypes.Amount(amount),
			ScriptPubKey: output.ScriptPubKey,
			Vout:         output.Vout,
		}
	}

	return utxos, nil
}

// Confirmations returns the number of confirmation blocks of the given txHash.
func (c *client) Confirmations(txHash zectypes.TxHash) (zectypes.Confirmations, error) {
	txHashBytes, err := chainhash.NewHashFromStr(string(txHash))
	if err != nil {
		return 0, fmt.Errorf("cannot parse hash: %v", err)
	}
	tx, err := c.client.GetTransaction(txHashBytes)
	if err != nil {
		return 0, fmt.Errorf("cannot get tx from hash %s: %v", txHash, err)
	}

	return zectypes.Confirmations(tx.Confirmations), nil
}

func (c *client) BuildUnsignedTx(utxos zectypes.UTXOs, recipients zectypes.Recipients, refundTo zectypes.Address, gas zectypes.Amount) (zectypes.Tx, error) {
	panic("unimplemented")
}

// SubmitSignedTx submits the signed transaction and returns the transaction hash in hex.
func (c *client) SubmitSignedTx(stx zectypes.Tx) (zectypes.TxHash, error) {
	panic("unimplemented")
}
