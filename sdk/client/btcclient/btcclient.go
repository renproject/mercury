package btcclient

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

const (
	Dust = btctypes.Amount(600)

	MinUTXOLimit     = 1
	MaxUTXOLimit     = 99
	MinConfirmations = 0
	MaxConfirmations = 99

	MainnetMercuryURL = "http://139.59.221.34/btc"
	TestnetMercuryURL = "http://139.59.221.34/btc-testnet3"
)

type Client interface {
	Network() btctypes.Network
	Balance(ctx context.Context, address btctypes.Address, limit, confirmations int) (btctypes.Amount, error)
	UTXOs(ctx context.Context, address btctypes.Address, limit, confirmations int) ([]btctypes.UTXO, error)
	Confirmations(ctx context.Context, hash btctypes.TxHash) (btctypes.Confirmations, error)
	BuildUnsignedTx(utxos []btctypes.UTXO, recipients ...btctypes.Recipient) (btctypes.Tx, error)
	SubmitSignedTx(ctx context.Context, stx btctypes.Tx) error
	Config() *chaincfg.Params
}

// Client is a client which is used to talking with certain bitcoin network. It can interacting with the blockchain
// through Mercury server.
type client struct {
	network btctypes.Network

	config chaincfg.Params
	url    string
}

// NewBtcClient returns a new Client of given bitcoin network.
func NewBtcClient(network btctypes.Network) Client {
	switch network {
	case btctypes.Mainnet:
		return &client{
			network: network,
			config:  chaincfg.MainNetParams,
			url:     MainnetMercuryURL,
		}
	case btctypes.Testnet:
		return &client{
			network: network,
			config:  chaincfg.TestNet3Params,
			url:     TestnetMercuryURL,
		}
	default:
		panic("unknown bitcoin network")
	}
}

func (c *client) Config() *chaincfg.Params {
	return &c.config
}

func (c *client) Network() btctypes.Network {
	return c.network
}

// Balance returns the balance of the given bitcoin address. It filters the utxos which have less confirmations than
// required. It times out if the context exceeded. Limit must be greater than MinUTXOLimit and less than MaxUTXOLimit.
// Confirmations must be greater than MinConfirmationsLimit and less than MaxConfirmationsLimit.
func (c *client) Balance(ctx context.Context, address btctypes.Address, limit, confirmations int) (btctypes.Amount, error) {
	// Pre-condition checks
	checkConfirmationPreCondition(confirmations)
	checkUTXOLimitPreCondition(limit)

	utxos, err := c.UTXOs(ctx, address, limit, confirmations)
	if err != nil {
		return btctypes.Amount(0), err
	}
	balance := btctypes.Amount(0)
	for _, utxo := range utxos {
		balance += btctypes.Amount(utxo.Amount)
	}

	// Post-condition checks
	checkBalancePostCondition(balance)
	return balance, nil
}

// UTXOs returns the utxos of the given bitcoin address. It filters the utxos which have less confirmations than
// required. It times out if the context exceeded.
func (c *client) UTXOs(ctx context.Context, address btctypes.Address, limit, confirmations int) ([]btctypes.UTXO, error) {
	// Pre-condition checks
	checkConfirmationPreCondition(confirmations)
	checkUTXOLimitPreCondition(limit)

	// Construct the http request.
	url := fmt.Sprintf("%v/utxo/%v?limit=%v&confirmations=%v", c.url, address.EncodeAddress(), limit, confirmations)
	log.Printf("url = %v", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.WithContext(ctx)

	var utxos []btctypes.UTXO
	err = c.sendRequest(request, http.StatusOK, &utxos)
	return utxos, err
}

// Confirmations returns the number of confirmation blocks of the given txHash.
func (c *client) Confirmations(ctx context.Context, hash btctypes.TxHash) (btctypes.Confirmations, error) {
	url := fmt.Sprintf("%v/confirmations/%v", c.url, hash)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	request.WithContext(ctx)

	// TODO: CURRENT MERCURY DOESN'T RETURN A JSON RESPONSE FOR CONFIRMATIONS.
	// Use a new http.Client to send the request.
	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return 0, err
	}

	// Check the response code and decode the response.
	if response.StatusCode != http.StatusOK {
		return 0, types.UnexpectedStatusCode(http.StatusOK, response.StatusCode)
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	confirmations, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, err
	}

	return btctypes.Confirmations(confirmations), nil
}

func (c *client) BuildUnsignedTx(utxos []btctypes.UTXO, recipients ...btctypes.Recipient) (btctypes.Tx, error) {
	newTx := wire.NewMsgTx(wire.TxVersion)

	// Fill the utxos we want to use as newTx inputs.
	for _, utxo := range utxos {
		hash, err := chainhash.NewHashFromStr(utxo.TxHash)
		if err != nil {
			return btctypes.Tx{}, err
		}

		sourceUtxo := wire.NewOutPoint(hash, utxo.Vout)
		sourceTxIn := wire.NewTxIn(sourceUtxo, nil, nil)
		newTx.AddTxIn(sourceTxIn)
	}

	// Fill newTx output with the target address we want to receive the funds.
	for _, recipient := range recipients {
		outputPkScript, err := txscript.PayToAddrScript(recipient.Address)
		if err != nil {
			return btctypes.Tx{}, err
		}
		sourceTxOut := wire.NewTxOut(int64(recipient.Amount), outputPkScript)
		newTx.AddTxOut(sourceTxOut)
	}

	// Get the signature hashes we need to sign
	sigHashes := make([]btctypes.SignatureHash, len(utxos))
	for i, utxo := range utxos {
		script, err := hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			return btctypes.Tx{}, err
		}
		sigHashes[i], err = txscript.CalcSignatureHash(script, txscript.SigHashAll, newTx, i)
		if err != nil {
			return btctypes.Tx{}, err
		}
	}

	return btctypes.NewUnsignedTx(c.network, newTx, sigHashes), nil
}

type PostTransactionRequest struct {
	SignedTransaction string `json:"stx"`
}

// SubmitSignedTx submits the signed transactions
func (c *client) SubmitSignedTx(ctx context.Context, stx btctypes.Tx) error {
	// Pre-condition checks
	if !stx.IsSigned() {
		panic("pre-condition violation: cannot submit unsigned transaction")
	}

	req := PostTransactionRequest{
		SignedTransaction: hex.EncodeToString(stx.Serialize()),
	}

	// reqBytes = stx.Serialize()
	reqBytes, err := json.Marshal(req)

	fmt.Println(string(reqBytes))
	buf := bytes.NewBuffer(reqBytes)
	url := fmt.Sprintf("%v/tx", c.url)
	fmt.Printf("sending post request to: %v\n", url)
	request, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return fmt.Errorf("error building 'submit signed transaction' http request: %v", err)
	}
	request.WithContext(ctx)

	// Use a new http.Client to send the request.
	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("error executing 'submit signed transaction' http request: %v", err)
	}
	defer response.Body.Close()

	expectedStatusCode := http.StatusCreated
	// Check the response code and decode the response.
	if response.StatusCode != expectedStatusCode {
		body, err := ioutil.ReadAll(response.Body)
		fmt.Printf("body: %v", string(body))
		if err != nil {
			return fmt.Errorf("error reading http response body: %v", err)
		}
		return types.NewErrHTTPResponse(expectedStatusCode, response.StatusCode, body)
	}

	return nil
}

func (c *client) sendRequest(request *http.Request, statusCode int, result interface{}) error {
	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	// Check the response code and decode the response.
	if response.StatusCode != statusCode {
		return types.UnexpectedStatusCode(statusCode, response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(&result)
}

func checkUTXOLimitPreCondition(limit int) {
	// Pre-condition checks
	if limit < MinUTXOLimit {
		panic(fmt.Errorf("pre-condition violation: limit=%v is too low", limit))
	}
	if limit > MaxUTXOLimit {
		panic(fmt.Errorf("pre-condition violation: limit=%v is too high", limit))
	}
}

func checkConfirmationPreCondition(confirmations int) {
	if confirmations < MinConfirmations {
		panic(fmt.Errorf("pre-condition violation: confirmations=%v is to low", confirmations))
	}
	if confirmations > MaxConfirmations {
		panic(fmt.Errorf("pre-condition violation: confirmations=%v is too high", confirmations))
	}
}

func checkBalancePostCondition(balance btctypes.Amount) {
	// Post-condition checks
	if balance < btctypes.Amount(0) {
		panic(fmt.Errorf("post-condition violation: balance=%v is too low", balance))
	}
}
