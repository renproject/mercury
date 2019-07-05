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
	Dust = btctypes.Amount(600)

	MinUTXOLimit     = 1
	MaxUTXOLimit     = 99
	MinConfirmations = 0
	MaxConfirmations = 99

	MainnetMercuryURL = "http://139.59.221.34/btc"
	TestnetMercuryURL = "http://206.189.83.88:5000/btc/testnet"
)

type Client interface {
	Network() btctypes.Network
	UTXOs(address btctypes.Address) (btctypes.UTXOs, error)
	Confirmations(hash btctypes.TxHash) (btctypes.Confirmations, error)
	BuildUnsignedTx(utxos btctypes.UTXOs, recipients btctypes.Recipients, refundTo btctypes.Address, gas btctypes.Amount) (btctypes.Tx, error)
	SubmitSignedTx(stx btctypes.Tx) error
}

// Client is a client which is used to talking with certain bitcoin network. It can interacting with the blockchain
// through Mercury server.
type client struct {
	network btctypes.Network
	client  *rpcclient.Client

	config chaincfg.Params
	url    string
}

// NewBtcClient returns a new Client of given bitcoin network.
func NewBtcClient(network btctypes.Network) (Client, error) {
	config := &rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	switch network {
	case btctypes.Mainnet:
		config.Host = MainnetMercuryURL
		c, err := rpcclient.New(config, nil)
		if err != nil {
			return &client{}, err
		}
		return &client{
			network: network,
			client:  c,
			config:  chaincfg.MainNetParams,
			url:     MainnetMercuryURL,
		}, nil
	case btctypes.Testnet:
		config.Host = TestnetMercuryURL
		c, err := rpcclient.New(config, nil)
		if err != nil {
			return &client{}, err
		}
		return &client{
			network: network,
			client:  c,
			config:  chaincfg.TestNet3Params,
			url:     TestnetMercuryURL,
		}, nil
	default:
		panic("unknown bitcoin network")
	}
}

func (c *client) Network() btctypes.Network {
	return c.network
}

// UTXOs returns the utxos of the given bitcoin address. It filters the utxos which have less confirmations than
// required. It times out if the context exceeded.
func (c *client) UTXOs(address btctypes.Address) (btctypes.UTXOs, error) {
	outputs, err := c.client.ListUnspent()
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
			TxHash:       output.TxID,
			Amount:       btctypes.Amount(amount),
			ScriptPubKey: output.ScriptPubKey,
			Vout:         output.Vout,
		}
	}

	return utxos, nil
}

// Confirmations returns the number of confirmation blocks of the given txHash.
func (c *client) Confirmations(hash btctypes.TxHash) (btctypes.Confirmations, error) {
	hashBytes, err := chainhash.NewHashFromStr(string(hash))
	if err != nil {
		return 0, fmt.Errorf("cannot parse hash: %v", err)
	}
	tx, err := c.client.GetTransaction(hashBytes)
	if err != nil {
		return 0, fmt.Errorf("cannot get tx from hash %s: %v", hash, err)
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
			Txid: utxo.TxHash,
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
	amounts[refundTo] = btcutil.Amount(amountToRefund)

	var lockTime int64
	tx, err := c.client.CreateRawTransaction(inputs, amounts, &lockTime)
	if err != nil {
		return btctypes.Tx{}, fmt.Errorf("cannot construct raw transaction: %v", err)
	}

	// Get the signature hashes we need to sign
	unsignedTx := btctypes.NewUnsignedTx(c.network, tx)
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

type PostTransactionRequest struct {
	SignedTransaction string `json:"stx"`
}

// SubmitSignedTx submits the signed transactions
func (c *client) SubmitSignedTx(stx btctypes.Tx) error {
	if !stx.IsSigned() {
		panic("pre-condition violation: cannot submit unsigned transaction")
	}

	_, err := c.client.SendRawTransaction(stx.Tx, false)
	return err
}
