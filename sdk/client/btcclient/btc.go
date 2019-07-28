package btcclient

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/rpcclient"
	"github.com/renproject/mercury/rpcclient/btcrpcclient"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

const (
	Dust = btctypes.Amount(600)
)

var (
	ErrInvalidTxHash  = errors.New("invalid tx hash")
	ErrTxHashNotFound = errors.New("tx hash not found")
	ErrUTXOSpent      = errors.New("utxo spent or invalid index")
)

// Client is a client which is used to talking with certain Bitcoin network. It can interacting with the blockchain
// through Mercury server.
type client struct {
	client     btcrpcclient.Client
	network    btctypes.Network
	config     chaincfg.Params
	url        string
	gasStation BtcGasStation
	logger     logrus.FieldLogger
}

func MercuryURL(network btctypes.Network) string {
	var chainStr string
	switch network.Chain() {
	case types.Bitcoin:
		chainStr = "btc"
	case types.ZCash:
		chainStr = "zec"
	default:
		panic(types.ErrUnknownChain)
	}

	switch network {
	case btctypes.BtcMainnet, btctypes.ZecMainnet:
		return fmt.Sprintf("http://206.189.83.88:5000/%s/mainnet", chainStr)
	case btctypes.BtcTestnet, btctypes.ZecTestnet:
		return fmt.Sprintf("http://206.189.83.88:5000/%s/testnet", chainStr)
	case btctypes.BtcLocalnet, btctypes.ZecLocalnet:
		return fmt.Sprintf("http://0.0.0.0:5000/%s/testnet", chainStr)
	default:
		panic(types.ErrUnknownNetwork)
	}
}

// New returns a new Client of given Bitcoin network.
func New(logger logrus.FieldLogger, network btctypes.Network) (Client, error) {
	host := MercuryURL(network)
	gasStation := NewBtcGasStation(logger, 30*time.Minute)
	return &client{
		client:     btcrpcclient.NewRPCClient(host, "", ""),
		network:    network,
		config:     *network.Params(),
		url:        host,
		gasStation: gasStation,
		logger:     logger,
	}, nil
}

func (c *client) Network() btctypes.Network {
	return c.network
}

// UTXO returns the UTXO for the given transaction hash and index.
func (c *client) UTXO(op btctypes.OutPoint) (btctypes.UTXO, error) {
	if len(op.TxHash()) != 64 {
		return nil, ErrInvalidTxHash
	}

	tx, err := c.client.GetRawTransactionVerbose(string(op.TxHash()))
	if err != nil {
		return nil, ErrTxHashNotFound
	}

	txOut, err := c.client.GetTxOut(string(op.TxHash()), op.Vout())
	if err != nil {
		if err == rpcclient.ErrNullResult {
			return nil, ErrUTXOSpent
		}
		return nil, fmt.Errorf("cannot get tx output from btc client: %v", err)
	}

	amount, err := btcutil.NewAmount(txOut.Value)
	if err != nil {
		return nil, fmt.Errorf("cannot parse amount received from btc client: %v", err)
	}

	scriptPubKey, err := hex.DecodeString(txOut.ScriptPubKey.Hex)
	if err != nil {
		return nil, fmt.Errorf("cannot decode script pubkey")
	}

	return btctypes.NewUTXO(
		btctypes.NewOutPoint(types.TxHash(tx.TxID), op.Vout()),
		btctypes.Amount(amount),
		scriptPubKey,
		uint64(txOut.Confirmations),
		nil, nil,
	), nil
}

// UTXOsFromAddress returns the UTXOs for a given address. Important: this function will not return any UTXOs for
// addresses that have not been imported into the Bitcoin node.
func (c *client) UTXOsFromAddress(address btctypes.Address) (btctypes.UTXOs, error) {
	outputs, err := c.client.ListUnspent(0, 999999, []string{address.EncodeAddress()})
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve utxos from btc client: %v", err)
	}

	utxos := make(btctypes.UTXOs, len(outputs))
	for i, output := range outputs {
		amount, err := btcutil.NewAmount(output.Amount)
		if err != nil {
			return nil, fmt.Errorf("cannot parse amount received from btc client: %v", err)
		}

		scriptPubKey, err := hex.DecodeString(output.ScriptPubKey)
		if err != nil {
			return nil, fmt.Errorf("cannot decode script pubkey")
		}

		utxos[i] = btctypes.NewUTXO(
			btctypes.NewOutPoint(types.TxHash(output.TxID), output.Vout),
			btctypes.Amount(amount),
			scriptPubKey,
			uint64(output.Confirmations),
			nil, nil,
		)
	}

	return utxos, nil
}

// Confirmations returns the number of confirmation blocks of the given txHash.
func (c *client) Confirmations(txHash types.TxHash) (uint64, error) {
	tx, err := c.client.GetRawTransactionVerbose(string(txHash))
	if err != nil {
		return 0, fmt.Errorf("cannot get tx from hash %s: %v", txHash, err)
	}
	return uint64(tx.Confirmations), nil
}

func (c *client) BuildUnsignedTx(utxos btctypes.UTXOs, recipients btctypes.Recipients, refundTo btctypes.Address, gas btctypes.Amount) (btctypes.BtcTx, error) {
	// Pre-condition checks.
	if gas < Dust {
		return nil, fmt.Errorf("pre-condition violation: gas = %v is too low", gas)
	}

	amountFromUTXOs := utxos.Sum()
	if amountFromUTXOs < Dust {
		return nil, fmt.Errorf("pre-condition violation: amount=%v from utxos is less than dust=%v", amountFromUTXOs, Dust)
	}

	// Add an output for each recipient and sum the total amount that is being transferred to recipients.
	amountToRecipients := btctypes.Amount(0)
	for _, recipient := range recipients {
		amountToRecipients += recipient.Amount
	}

	// Check that we are not transferring more to recipients than available in the UTXOs (accounting for gas).
	amountToRefund := amountFromUTXOs - amountToRecipients - gas
	if amountToRefund < 0 {
		return nil, fmt.Errorf("insufficient balance: expected %v, got %v", amountToRecipients+gas, amountFromUTXOs)
	}

	// Add an output to refund the difference between the amount being transferred to recipients and the total amount
	// from the UTXOs.
	if amountToRefund > 0 {
		recipients = append(recipients, btctypes.NewRecipient(refundTo, amountToRefund))
	}

	// Get the signature hashes we need to sign.
	return c.createUnsignedTx(utxos, recipients)
}

// SubmitSignedTx submits the signed transaction and returns the transaction hash in hex.
func (c *client) SubmitSignedTx(stx btctypes.BtcTx) (types.TxHash, error) {
	// Pre-condition checks
	if !stx.IsSigned() {
		return "", errors.New("pre-condition violation: cannot submit unsigned transaction")
	}
	if err := c.VerifyTx(stx); err != nil {
		return "", fmt.Errorf("pre-condition violation: transaction failed verification: %v", err)
	}

	data, err := stx.Serialize()
	if err != nil {
		return "", fmt.Errorf("pre-condition violation: serialization failed: %v", err)
	}

	txHash, err := c.client.SendRawTransaction(data)
	if err != nil {
		return "", fmt.Errorf("cannot send raw transaction using btc client: %v", err)
	}
	return types.TxHash(txHash), nil
}

func (c *client) EstimateTxSize(numUTXOs, numRecipients int) int {
	return 146*numUTXOs + 33*numRecipients + 10
}

func (c *client) VerifyTx(tx btctypes.BtcTx) error {
	if c.network.Chain() != types.Bitcoin {
		return nil
	}

	data, err := tx.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize transaction")
	}
	msgTx := new(wire.MsgTx)
	msgTx.Deserialize(bytes.NewBuffer(data))

	for i, utxo := range tx.UTXOs() {
		engine, err := txscript.NewEngine(utxo.ScriptPubKey(), msgTx, i,
			txscript.StandardVerifyFlags, txscript.NewSigCache(10),
			txscript.NewTxSigHashes(msgTx), int64(utxo.Amount()))
		if err != nil {
			return err
		}
		if err := engine.Execute(); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) SuggestGasPrice(ctx context.Context, speed types.TxSpeed, txSizeInBytes int) btctypes.Amount {
	gasStationPrice, err := c.gasStation.GasRequired(ctx, speed, txSizeInBytes)
	if err == nil {
		return gasStationPrice
	}
	c.logger.Errorf("error getting btc gas information: %v", err)
	c.logger.Infof("using 10k sats as gas price")
	return 10000 * btctypes.SAT
}

func (c *client) createUnsignedTx(utxos btctypes.UTXOs, recipients btctypes.Recipients) (btctypes.BtcTx, error) {
	outputUTXOs := map[string]btctypes.UTXO{}
	msgTx := btctypes.NewMsgTx(c.network)
	for _, utxo := range utxos {
		hash, err := chainhash.NewHashFromStr(string(utxo.TxHash()))
		if err != nil {
			return nil, err
		}
		msgTx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(hash, utxo.Vout()), nil, nil))
	}
	for i, recipient := range recipients {
		script, err := btctypes.PayToAddrScript(recipient.Address, c.network)
		if err != nil {
			return nil, err
		}
		msgTx.AddTxOut(wire.NewTxOut(int64(recipient.Amount), script))
		outputUTXOs[recipient.Address.EncodeAddress()] = btctypes.NewUTXO(btctypes.NewOutPoint("", uint32(i)), recipient.Amount, script, 0, nil, nil)
	}
	return btctypes.NewUnsignedTx(c.network, utxos, msgTx, outputUTXOs)
}
