package btcclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

var (
	// DefaultLimit is the default limit when querying utxos and balances.
	DefaultLimit = 999999

	// DefaultConfirmations is the default confirmations when querying utxos and balances.
	DefaultConfirmations = 0
)

// Client is a client which is used to talking with certain bitcoin network. It can interacting with the blockchain
// through Mercury server.
type Client struct {
	Network btctypes.Network

	config chaincfg.Params
	url    string
}

// NewBtcClient returns a new Client of given bitcoin network.
func NewBtcClient(network btctypes.Network) *Client {
	switch network {
	case btctypes.Mainnet:
		return &Client{
			Network: network,
			config:  chaincfg.MainNetParams,
			url:     "https://ren-mercury.herokuapp.com/btc",
		}
	case btctypes.Testnet:
		return &Client{
			Network: network,
			config:  chaincfg.TestNet3Params,
			url:     "https://ren-mercury.herokuapp.com/btc-testnet3",
		}
	default:
		panic("unknown bitcoin network")
	}
}

// Balance returns the balance of the given bitcoin address. It filters the utxos which have less confirmations than
// required. It times out if the context exceeded.
func (client *Client) Balance(ctx context.Context, address btctypes.Addr, confirmations int) (btctypes.Amount, error) {
	utxos, err := client.UTXOs(ctx, address, DefaultLimit, confirmations)
	if err != nil {
		return btctypes.Amount(0), err
	}

	// Add the amounts of each utxo to get the address balance.
	balance := btctypes.Amount(0)
	for _, utxo := range utxos {
		balance += btctypes.Amount(utxo.Amount)
	}
	return balance, nil
}

// UTXOs returns the utxos of the given bitcoin address. It filters the utxos which have less confirmations than
// required. It times out if the context exceeded.
func (client *Client) UTXOs(ctx context.Context, address btctypes.Addr, limit, confirmations int) ([]btctypes.UTXO, error) {
	// Construct the http request.
	url := fmt.Sprintf("%v/utxo/%v?limit=%v&confirmations=%v", client.url, address.EncodeAddress(), limit, confirmations)
	log.Printf("url = %v", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.WithContext(ctx)

	var utxos []btctypes.UTXO
	err = client.sendRequest(request, http.StatusOK, &utxos)
	return utxos, err
}

// Confirmations returns the number of confirmation blocks of the given txHash.
func (client *Client) Confirmations(ctx context.Context, hash string) (int64, error) {
	url := fmt.Sprintf("%v/confirmations/%v", client.url, hash)
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

	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}

// SubmitSTX submits the signed transactions
func (client *Client) SubmitSTX(ctx context.Context, stx []byte) error {
	buf := bytes.NewBuffer(stx)
	url := fmt.Sprintf("%v/tx", client.url)
	request, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return err
	}
	request.WithContext(ctx)

	// Use a new http.Client to send the request.
	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	// Check the response code and decode the response.
	if response.StatusCode != http.StatusOK {
		return types.UnexpectedStatusCode(http.StatusOK, response.StatusCode)
	}

	// todo : Check the response
	return nil
}

func (client *Client) sendRequest(request *http.Request, statusCode int, result interface{}) error {
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
