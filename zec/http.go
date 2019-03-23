package zec

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Handlers of the zcash blockchain
func (zec *zcash) AddRoutes(r *mux.Router) {
	r.HandleFunc(zec.AddRoutePrefix("/utxo/{address}"), zec.getUTXOhandler()).Queries("limit", "{limit}").Queries("confirmations", "{confirmations}").Methods("GET")
	r.HandleFunc(zec.AddRoutePrefix("/script/{state}/{address}"), zec.getScriptHandler()).Queries("value", "{value}").Methods("GET")
	r.HandleFunc(zec.AddRoutePrefix("/script/{state}/{address}"), zec.getScriptHandler()).Queries("spender", "{spender}").Methods("GET")
	r.HandleFunc(zec.AddRoutePrefix("/confirmations/{txHash}"), zec.getConfirmationsHandler()).Methods("GET")
	r.HandleFunc(zec.AddRoutePrefix("/tx"), zec.postTransaction()).Methods("POST")
}

func (zec *zcash) AddRoutePrefix(route string) string {
	network := zec.Network()
	if network == "mainnet" {
		network = ""
	}
	return fmt.Sprintf("/%s-%s%s", "zec", network, route)
}

func (zec *zcash) getUTXOhandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts := mux.Vars(r)
		addr := opts["address"]
		limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, err)
			return
		}
		confirmations, err := strconv.ParseInt(r.URL.Query().Get("confirmations"), 10, 64)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, err)
			return
		}
		utxos, err := zec.GetUTXOs(addr, limit, confirmations)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, err)
			return
		}
		if err := json.NewEncoder(w).Encode(utxos); err != nil {
			writeError(w, r, http.StatusInternalServerError, err)
			return
		}
	}
}

func (zec *zcash) getScriptHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts := mux.Vars(r)
		addr := opts["address"]
		state := opts["state"]

		var resp GetScriptResponse
		var err error
		switch state {
		case "spent":
			status, script, err2 := zec.ScriptSpent(addr, r.URL.Query().Get("spender"))
			if err2 != nil {
				err = err2
				break
			}
			resp.Script = script
			resp.Status = status
		case "funded":
			value, err2 := strconv.ParseInt(r.URL.Query().Get("value"), 10, 64)
			if err2 != nil {
				err = err2
				break
			}
			status, val, err2 := zec.ScriptFunded(addr, value)
			if err2 != nil {
				err = err2
				break
			}
			resp.Value = val
			resp.Status = status
		case "redeemed":
			value, err2 := strconv.ParseInt(r.URL.Query().Get("value"), 10, 64)
			if err2 != nil {
				err = err2
				break
			}
			status, val, err2 := zec.ScriptRedeemed(addr, value)
			if err2 != nil {
				err = err2
				break
			}
			resp.Value = val
			resp.Status = status
		default:
			err = fmt.Errorf("unsupported script state: %s", state)
		}
		if err != nil {
			writeError(w, r, http.StatusBadRequest, err)
			return
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			writeError(w, r, http.StatusInternalServerError, err)
			return
		}
	}
}

func (zec *zcash) getConfirmationsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts := mux.Vars(r)
		conf, err := zec.Confirmations(opts["txHash"])
		if err != nil {
			writeError(w, r, http.StatusBadRequest, err)
			return
		}
		if err := json.NewEncoder(w).Encode(GetConfirmationsResponse(conf)); err != nil {
			writeError(w, r, http.StatusBadRequest, err)
			return
		}
	}
}

func (zec *zcash) postTransaction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := PostTransactionRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, r, http.StatusBadRequest, err)
			return
		}
		stx, err := hex.DecodeString(req.SignedTransaction)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, err)
			return
		}
		if err := zec.PublishTransaction(stx); err != nil {
			writeError(w, r, http.StatusBadRequest, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), statusCode)
}
