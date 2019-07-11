package btcclient

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/types/btctypes"
)

const (
	Dust               = btctypes.Amount(600)
	MainnetMercuryURL  = "206.189.83.88:5000/btc/mainnet"
	TestnetMercuryURL  = "206.189.83.88:5000/btc/testnet"
	LocalnetMercuryURL = "0.0.0.0:5000/btc/testnet"
)

var (
	ErrInvalidTxHash  = errors.New("invalid tx hash")
	ErrTxHashNotFound = errors.New("tx hash not found")
	ErrUTXOSpent      = errors.New("utxo spent or invalid index")
)

type Client interface {
	Network() btctypes.Network
	UTXO(txHash btctypes.TxHash, index uint32) (btctypes.UTXO, error)
	UTXOsFromAddress(address btctypes.Address) (btctypes.UTXOs, error)
	Confirmations(txHash btctypes.TxHash) (btctypes.Confirmations, error)
	BuildUnsignedTx(utxos btctypes.UTXOs, recipients btctypes.Recipients, refundTo btctypes.Address, gas btctypes.Amount) (btctypes.StandardTx, error)
	SubmitSignedTx(stx btctypes.Tx) (btctypes.TxHash, error)
	EstimateTxSize(numUTXOs, numRecipients int) int
}

// Client is a client which is used to talking with certain Bitcoin network. It can interacting with the blockchain
// through Mercury server.
type client struct {
	network btctypes.Network
	client  *rpcclient.Client

	config chaincfg.Params
	url    string
}

// New returns a new Client of given Bitcoin network.
func New(network btctypes.Network) (Client, error) {
	config := &rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	switch network {
	case btctypes.Mainnet:
		config.Host = MainnetMercuryURL
	case btctypes.Testnet:
		config.Host = TestnetMercuryURL
	case btctypes.Localnet:
		config.Host = LocalnetMercuryURL
	default:
		panic("unknown bitcoin network")
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

func (c *client) Network() btctypes.Network {
	return c.network
}

// UTXO returns the UTXO for the given transaction hash and index.
func (c *client) UTXO(txHash btctypes.TxHash, index uint32) (btctypes.UTXO, error) {
	txHashBytes, err := chainhash.NewHashFromStr(string(txHash))
	if err != nil {
		return btctypes.UTXO{}, ErrInvalidTxHash
	}
	tx, err := c.client.GetRawTransactionVerbose(txHashBytes)
	if err != nil {
		return btctypes.UTXO{}, ErrTxHashNotFound
	}

	txOut, err := c.client.GetTxOut(txHashBytes, index, true)
	if err != nil {
		return btctypes.UTXO{}, fmt.Errorf("cannot get tx output from btc client: %v", err)
	}

	// Check if UTXO has been spent.
	if txOut == nil {
		return btctypes.UTXO{}, ErrUTXOSpent
	}

	amount, err := btcutil.NewAmount(txOut.Value)
	if err != nil {
		return btctypes.UTXO{}, fmt.Errorf("cannot parse amount received from btc client: %v", err)
	}
	return btctypes.UTXO{
		TxHash:       btctypes.TxHash(tx.Txid),
		Amount:       btctypes.Amount(amount),
		ScriptPubKey: txOut.ScriptPubKey.Hex,
		Vout:         index,
	}, nil
}

// UTXOsFromAddress returns the UTXOs for a given address. Important: this function will not return any UTXOs for
// addresses that have not been imported into the Bitcoin node.
func (c *client) UTXOsFromAddress(address btctypes.Address) (btctypes.UTXOs, error) {
	outputs, err := c.client.ListUnspentMinMaxAddresses(0, 999999, []btcutil.Address{address})
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve utxos from btc client: %v", err)
	}

	utxos := make(btctypes.UTXOs, len(outputs))
	for i, output := range outputs {
		amount, err := btcutil.NewAmount(output.Amount)
		if err != nil {
			return nil, fmt.Errorf("cannot parse amount received from btc client: %v", err)
		}
		utxos[i] = btctypes.UTXO{
			TxHash:       btctypes.TxHash(output.TxID),
			Amount:       btctypes.Amount(amount),
			ScriptPubKey: output.ScriptPubKey,
			Vout:         output.Vout,
		}
	}

	return utxos, nil
}

// Confirmations returns the number of confirmation blocks of the given txHash.
func (c *client) Confirmations(txHash btctypes.TxHash) (btctypes.Confirmations, error) {
	txHashBytes, err := chainhash.NewHashFromStr(string(txHash))
	if err != nil {
		return 0, fmt.Errorf("cannot parse hash: %v", err)
	}
	tx, err := c.client.GetTransaction(txHashBytes)
	if err != nil {
		return 0, fmt.Errorf("cannot get tx from hash %s: %v", txHash, err)
	}

	return btctypes.Confirmations(tx.Confirmations), nil
}

func (c *client) BuildUnsignedTx(utxos btctypes.UTXOs, recipients btctypes.Recipients, refundTo btctypes.Address, gas btctypes.Amount) (btctypes.StandardTx, error) {
	// Pre-condition checks
	if gas < Dust {
		return nil, fmt.Errorf("pre-condition violation: gas = %v is too low", gas)
	}

	inputs := make([]btcjson.TransactionInput, len(utxos))
	for i, utxo := range utxos {
		inputs[i] = btcjson.TransactionInput{
			Txid: string(utxo.TxHash),
			Vout: utxo.Vout,
		}
	}

	amountFromUTXOs := utxos.Sum()
	if amountFromUTXOs < Dust {
		return nil, fmt.Errorf("pre-condition violation: amount=%v from utxos is less than dust=%v", amountFromUTXOs, Dust)
	}

	// Add an output for each recipient and sum the total amount that is being
	// transferred to recipients
	amountToRecipients := btctypes.Amount(0)
	outputs := make(map[btcutil.Address]btcutil.Amount, len(recipients))
	for _, recipient := range recipients {
		amountToRecipients += recipient.Amount
		outputs[recipient.Address] = btcutil.Amount(recipient.Amount)
	}

	// Check that we are not transferring more to recipients than available in
	// the UTXOs (accounting for gas)
	amountToRefund := amountFromUTXOs - amountToRecipients - gas
	if amountToRefund < 0 {
		return nil, fmt.Errorf("insufficient balance: expected %v, got %v", amountToRecipients+gas, amountFromUTXOs)
	}

	// Add an output to refund the difference between what we are transferring
	// to recipients and what we are spending from the UTXOs (accounting for
	// gas)
	outputs[refundTo] += btcutil.Amount(amountToRefund)

	var lockTime int64
	wireTx, err := c.client.CreateRawTransaction(inputs, outputs, &lockTime)
	if err != nil {
		return nil, fmt.Errorf("cannot construct raw transaction: %v", err)
	}

	// Get the signature hashes we need to sign
	return btctypes.NewUnsignedTx(c.network, utxos, wireTx)
}

// SubmitSignedTx submits the signed transaction and returns the transaction hash in hex.
func (c *client) SubmitSignedTx(stx btctypes.Tx) (btctypes.TxHash, error) {
	// Pre-condition checks
	if !stx.IsSigned() {
		return "", errors.New("pre-condition violation: cannot submit unsigned transaction")
	}
	if err := c.VerifyTx(stx); err != nil {
		return "", fmt.Errorf("pre-condition violation: transaction failed verification: %v", err)
	}

	txHash, err := c.client.SendRawTransaction(stx.Tx(), false)
	if err != nil {
		return "", fmt.Errorf("cannot send raw transaction using btc client: %v", err)
	}
	return btctypes.TxHash(txHash.String()), nil
}

func (c *client) EstimateTxSize(numUTXOs, numRecipients int) int {
	return 146*numUTXOs + 33*numRecipients + 10
}

func (c *client) VerifyTx(tx btctypes.Tx) error {
	for i, utxo := range tx.UTXOs() {
		scriptPubKey, err := hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			return err
		}
		engine, err := txscript.NewEngine(scriptPubKey, tx.Tx(), i,
			txscript.StandardVerifyFlags, txscript.NewSigCache(10),
			txscript.NewTxSigHashes(tx.Tx()), int64(utxo.Amount))
		if err != nil {
			return err
		}
		if err := engine.Execute(); err != nil {
			return err
		}
	}
	return nil
}
