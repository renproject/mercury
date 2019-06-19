package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/renproject/mercury/types"
)

var (
	// DefaultLimit is the default limit when querying utxos and balances.
	DefaultLimit = 999999

	// DefaultConfirmations is the default confirmations when querying utxos and balances.
	DefaultConfirmations = 0
)

// UnexpectedStatusCode returns an meaning error to be returned when getting unexpected status code.
func UnexpectedStatusCode (expected , got int)error{
	return fmt.Errorf("unexpected status code, expect %v, got %v", expected, got)
}

// BtcClient is a
type BtcClient struct {
	config     chaincfg.Params
	url        string
}

func NewBtcClient(network types.BtcNetwork)*BtcClient{
	switch network {
	case types.BtcMainnet:
		return &BtcClient{
			config:    chaincfg.MainNetParams,
			url:       "https://ren-mercury.herokuapp.com/btc",
		}
	case types.BtcTestnet:
		return &BtcClient{
			config:    chaincfg.TestNet3Params,
			url:       "https://ren-mercury.herokuapp.com/btc-testnet3",
		}
	default:
		panic("unknown bitcoin network")
	}
}

// Balance returns the balance of the given bitcoin address. It filters the utxos which have less confirmations than
// required. It times out if the context exceeded.
func (client *BtcClient) Balance(ctx context.Context, address types.BtcAddr, confirmations int) (types.BtcValue, error) {
	utxos, err  := client.UTXOs(ctx, address, DefaultLimit, confirmations)
	if err != nil {
		return types.BtcValue(0), err
	}

	// Add the amounts of each utxo to get the address balance.
	balance := types.BtcValue(0)
	for _,utxo := range utxos{
		balance +=  types.BtcValue(utxo.Amount)
	}

	return balance, nil
}

// UTXOs returns the utxos of the given bitcoin address. It filters the utxos which have less confirmations than
// required. It times out if the context exceeded.
func (client *BtcClient) UTXOs(ctx context.Context, address types.BtcAddr, limit, confirmations int) ([]types.UTXO, error){
	// Construct the http request.
	url := fmt.Sprintf("%v/utxo/%v?limit=%v&confirmations=%v", client.url, address.String(), limit, confirmations)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.WithContext(ctx)

	// Use a new http.Client to send the request.
	httpClient := &http.Client{}
	response ,err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	// Check the response code and decode the response.
	if response.StatusCode != http.StatusOK{
		return nil, UnexpectedStatusCode(http.StatusOK, response.StatusCode)
	}
	var utxos []types.UTXO
	err = json.NewDecoder(response.Body).Decode(&utxos)
	return utxos, err
}

// Confirmations returns the number of confirmation blocks of the given txHash.
func (client *BtcClient) Confirmations(ctx context.Context, hash types.TxHash) (int64, error){
	url := fmt.Sprintf("%v/confirmations/%v", client.url, hash)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	request.WithContext(ctx)

	// Use a new http.Client to send the request.
	httpClient := &http.Client{}
	response ,err := httpClient.Do(request)
	if err != nil {
		return 0, err
	}

	// Check the response code and decode the response.
	if response.StatusCode != http.StatusOK{
		return 0, UnexpectedStatusCode(http.StatusOK, response.StatusCode)
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(strings.TrimSpace(string(data)),10, 64)
}

// SubmitSTX submits the signed transactions
func (client *BtcClient) SubmitSTX(ctx context.Context, stx []byte) error {
	buf := bytes.NewBuffer(stx)
	url := fmt.Sprintf("%v/tx", client.url)
	request, err := http.NewRequest("GET", url, buf)
	if err != nil {
		return err
	}
	request.WithContext(ctx)

	// Use a new http.Client to send the request.
	httpClient := &http.Client{}
	response ,err := httpClient.Do(request)
	if err != nil {
		return err
	}

	// Check the response code and decode the response.
	if response.StatusCode != http.StatusOK{
		return UnexpectedStatusCode(http.StatusOK, response.StatusCode)
	}

	// todo : Check the response
	return nil
}
