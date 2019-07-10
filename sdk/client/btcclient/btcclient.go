package btcclient

import (
	"encoding/hex"
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

type Client interface {
	Network() btctypes.Network
	UTXOs(txHash btctypes.TxHash) (btctypes.UTXOs, error)
	UTXOsFromAddress(address btctypes.Address) (btctypes.UTXOs, error)
	Confirmations(txHash btctypes.TxHash) (btctypes.Confirmations, error)
	BuildUnsignedTx(utxos btctypes.UTXOs, recipients btctypes.Recipients, refundTo btctypes.Address, gas btctypes.Amount) (btctypes.Tx, error)
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

// UTXOs returns the UTXOs for the given transaction hash.
func (c *client) UTXOs(txHash btctypes.TxHash) (btctypes.UTXOs, error) {
	txHashBytes, err := chainhash.NewHashFromStr(string(txHash))
	if err != nil {
		return nil, fmt.Errorf("cannot parse hash: %v", err)
	}
	tx, err := c.client.GetTransaction(txHashBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot get tx from hash %s: %v", txHash, err)
	}

	outputs := tx.Details
	var utxos btctypes.UTXOs
	for _, output := range outputs {
		txOut, err := c.client.GetTxOut(txHashBytes, output.Vout, true)
		if err != nil {
			return nil, fmt.Errorf("cannot get tx output from btc client: %v", err)
		}

		// If the transaction output has been spent, continue.
		if txOut == nil {
			continue
		}

		amount, err := btcutil.NewAmount(txOut.Value)
		if err != nil {
			return nil, fmt.Errorf("cannot parse amount received from btc client: %v", err)
		}
		utxo := btctypes.UTXO{
			TxHash:       btctypes.TxHash(tx.TxID),
			Amount:       btctypes.Amount(amount),
			ScriptPubKey: txOut.ScriptPubKey.Hex,
			Vout:         output.Vout,
		}
		utxos = append(utxos, utxo)
	}

	return utxos, nil
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

func (c *client) BuildUnsignedTx(utxos btctypes.UTXOs, recipients btctypes.Recipients, refundTo btctypes.Address, gas btctypes.Amount) (btctypes.Tx, error) {
	// Pre-condition checks
	if gas < Dust {
		panic(fmt.Errorf("pre-condition violation: gas=%v is too low", gas))
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
		// FIXME: Return an error.
		panic("newLessThanDustError()")
	}

	// Add an output for each recipient and sum the total amount that is being
	// transferred to recipients
	amountToRecipients := btctypes.Amount(0)
	amounts := make(map[btcutil.Address]btcutil.Amount, len(recipients))
	for _, recipient := range recipients {
		amountToRecipients += recipient.Amount
		amounts[recipient.Address] = btcutil.Amount(recipient.Amount)
	}

	// Check that we are not transferring more to recipients than available in
	// the UTXOs (accounting for gas)
	amountToRefund := amountFromUTXOs - amountToRecipients - gas
	if amountToRefund < 0 {
		// FIXME: Return an error.
		panic("newInsufficientAmountError")
	}

	// Add an output to refund the difference between what we are transferring
	// to recipients and what we are spending from the UTXOs (accounting for
	// gas)
	amounts[refundTo] += btcutil.Amount(amountToRefund)

	var lockTime int64
	wireTx, err := c.client.CreateRawTransaction(inputs, amounts, &lockTime)
	if err != nil {
		return btctypes.Tx{}, fmt.Errorf("cannot construct raw transaction: %v", err)
	}

	// Get the signature hashes we need to sign
	unsignedTx := btctypes.NewUnsignedTx(c.network, utxos, wireTx)
	fmt.Printf("before sig hashes: %v", unsignedTx.SignatureHashes())
	for _, utxo := range utxos {
		scriptPubKey, err := hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			return btctypes.Tx{}, err
		}
		if err := unsignedTx.AppendSignatureHash(scriptPubKey, txscript.SigHashAll); err != nil {
			return btctypes.Tx{}, err
		}
	}
	return unsignedTx, nil
}

// SubmitSignedTx submits the signed transactions
// returns the transaction hash as in hex
func (c *client) SubmitSignedTx(stx btctypes.Tx) (btctypes.TxHash, error) {
	// Pre-condition checks
	if !stx.IsSigned() {
		panic("pre-condition violation: cannot submit unsigned transaction")
	}
	if err := stx.Verify(); err != nil {
		panic(fmt.Errorf("pre-condition violation: transaction failed verification: %v", err))
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
