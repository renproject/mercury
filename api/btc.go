package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/types/btctypes"
)

type BtcBackend struct {
	proxy *proxy.BtcProxy
}

func NewBtcBackend (proxy *proxy.BtcProxy) *BtcBackend{
	return &BtcBackend{
		proxy: proxy,
	}
}

func (btc *BtcBackend) AddHandler(r *mux.Router){
	network := btc.proxy.Network
	r.HandleFunc(btc.networkPrefix(network, "/utxo/{address}"), btc.utxoHandler()).Methods("GET")
	r.HandleFunc(btc.networkPrefix(network, "/balance/{address}"), btc.balanceHandler()).Methods("GET")
	r.HandleFunc(btc.networkPrefix(network, "/confirmations/{tx}"), btc.confirmationHandler()).Methods("GET")
}

func (btc *BtcBackend) networkPrefix (network btctypes.Network, path string) string {
	return fmt.Sprintf("/btc/%s%s", network, path)
}

func (btc *BtcBackend) utxoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		address, err:= btctypes.AddressFromBase58String(vars["address"], btc.proxy.Network)
		if err != nil {
			http.Error(w, "invalid btc address", http.StatusBadRequest)
			return
		}
		limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		confirmations, err := strconv.ParseInt(r.URL.Query().Get("confirmations"), 10, 64)
		if err != nil {
			http.Error(w, "invalid confirmations", http.StatusBadRequest)
			return
		}

		utxos, err := btc.proxy.GetUTXOs(address, int(limit), int(confirmations))
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get utxo of given address, err = %v", err), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(utxos); err != nil {
			http.Error(w, fmt.Sprintf("unable to decode utxos, err = %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func (btc *BtcBackend) balanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		address, err:= btctypes.AddressFromBase58String(vars["address"], btc.proxy.Network)
		if err != nil {
			http.Error(w, "invalid btc address", http.StatusBadRequest)
			return
		}
		confirmations, err := strconv.ParseInt(r.URL.Query().Get("confirmations"), 10, 64)
		if err != nil {
			http.Error(w, "invalid confirmations", http.StatusBadRequest)
			return
		}

		utxos, err := btc.proxy.GetUTXOs(address, 999999, int(confirmations))
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get utxo of given address, err = %v", err), http.StatusInternalServerError)
			return
		}
		amount := 0 * btctypes.Satoshi
		for _, utxo := range utxos{
			amount += utxo.Amount
		}

		if err := json.NewEncoder(w).Encode(amount); err != nil {
			http.Error(w, fmt.Sprintf("unable to decode balance, err = %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func (btc *BtcBackend) confirmationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		txHash := vars["tx"]

		confirmations, err := btc.proxy.Confirmations(txHash)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get confirmations with of tx [%v],  err = %v", txHash, err), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(confirmations); err != nil {
			http.Error(w, fmt.Sprintf("unable to decode balance, err = %v", err), http.StatusInternalServerError)
			return
		}
	}
}

