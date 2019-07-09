package btcclient

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	UTXOs(ctx context.Context, address btctypes.Address, limit, confirmations int) (btctypes.UTXOs, error)
	Confirmations(ctx context.Context, hash btctypes.TxHash) (btctypes.Confirmations, error)
	BuildUnsignedTx(refundTo btctypes.Address, recipients btctypes.Recipients, utxos btctypes.UTXOs, gas btctypes.Amount) (btctypes.Tx, error)
	SubmitSignedTx(ctx context.Context, stx btctypes.Tx) (btctypes.TxHash, error)
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

func (c *client) Network() btctypes.Network {
	return c.network
}

// UTXOs returns the utxos of the given bitcoin address. It filters the utxos which have less confirmations than
// required. It times out if the context exceeded.
func (c *client) UTXOs(ctx context.Context, address btctypes.Address, limit, confirmations int) (btctypes.UTXOs, error) {
	// Pre-condition checks
	checkConfirmationPreCondition(confirmations)
	checkUTXOLimitPreCondition(limit)

	// Construct the http request.
	url := fmt.Sprintf("%v/utxo/%v?limit=%v&confirmations=%v", c.url, address.EncodeAddress(), limit, confirmations)
	// log.Printf("url = %v", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating fetch utxo request for address=%v, %v", address.EncodeAddress(), err)
	}
	request.WithContext(ctx)

	var utxos btctypes.UTXOs
	err = c.sendRequest(request, http.StatusOK, &utxos)
	if err != nil {
		return nil, fmt.Errorf("error submitting fetch utxo request for address=%v, %v", address.EncodeAddress(), err)
	}
	return utxos, nil
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

func (c *client) BuildUnsignedTx(refundTo btctypes.Address, recipients btctypes.Recipients, utxos btctypes.UTXOs, gas btctypes.Amount) (btctypes.Tx, error) {
	// Pre-condition checks
	if gas < Dust {
		panic(fmt.Errorf("pre-condition violation: gas=%v is too low", gas))
	}

	wireTx := wire.NewMsgTx(wire.TxVersion)

	// Add the UTXOs to the wire transactions
	for _, utxo := range utxos {
		hash, err := chainhash.NewHashFromStr(utxo.TxHash)
		if err != nil {
			return btctypes.Tx{}, err
		}
		sourceUTXO := wire.NewOutPoint(hash, utxo.Vout)
		sourceTxIn := wire.NewTxIn(sourceUTXO, nil, nil)
		wireTx.AddTxIn(sourceTxIn)
	}
	amountFromUTXOs := utxos.Sum()
	if amountFromUTXOs < Dust {
		// FIXME: Return an error.
		panic("newLessThanDustError()")
	}

	// Add an output for each recipient and sum the total amount that is being
	// transferred to recipients
	amountToRecipients := btctypes.Amount(0)
	for _, recipient := range recipients {
		payToAddrScript, err := txscript.PayToAddrScript(recipient.Address)
		if err != nil {
			return btctypes.Tx{}, err
		}
		amountToRecipients += recipient.Amount
		sourceTxOut := wire.NewTxOut(int64(recipient.Amount), payToAddrScript)
		wireTx.AddTxOut(sourceTxOut)
	}
	if amountToRecipients < Dust {
		// FIXME: Return an error.
		panic("newLessThanDustError()")
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
	payToAddrScript, err := txscript.PayToAddrScript(refundTo)
	if err != nil {
		return btctypes.Tx{}, err
	}
	sourceTxOut := wire.NewTxOut(int64(amountToRefund), payToAddrScript)
	wireTx.AddTxOut(sourceTxOut)

	// Get the signature hashes we need to sign
	unsignedTx := btctypes.NewUnsignedTx(c.network, wireTx)
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
// returns the transaction hash as in hex
func (c *client) SubmitSignedTx(ctx context.Context, stx btctypes.Tx) (btctypes.TxHash, error) {
	// Pre-condition checks
	if !stx.IsSigned() {
		panic("pre-condition violation: cannot submit unsigned transaction")
	}
	if err := stx.Verify(); err != nil {
		panic(fmt.Errorf("pre-condition violation: transaction failed verification: %v", err))
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
		return "", fmt.Errorf("error building 'submit signed transaction' http request: %v", err)
	}
	request.WithContext(ctx)

	// Use a new http.Client to send the request.
	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("error executing 'submit signed transaction' http request: %v", err)
	}
	defer response.Body.Close()

	expectedStatusCode := http.StatusCreated
	// Check the response code and decode the response.
	if response.StatusCode != expectedStatusCode {
		body, err := ioutil.ReadAll(response.Body)
		fmt.Printf("body: %v", string(body))
		if err != nil {
			return "", fmt.Errorf("error reading http response body: %v", err)
		}
		return "", types.NewErrHTTPResponse(expectedStatusCode, response.StatusCode, body)
	}

	return stx.Hash(), nil
}

func (c *client) sendRequest(request *http.Request, statusCode int, result interface{}) error {
	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("error sending http request: %v", err)
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
