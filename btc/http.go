package btc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Handlers of the bitcoin blockchain
func (btc *bitcoin) AddRoutes(r *mux.Router) {
	r.HandleFunc(btc.AddRoutePrefix("/utxo/{address}"), btc.getUTXOhandler()).Queries("limit", "{limit}").Queries("confirmations", "{confirmations}").Methods("GET")
	r.HandleFunc(btc.AddRoutePrefix("/script/{state}/{address}"), btc.getScriptHandler()).Queries("value", "{value}").Methods("GET")
	r.HandleFunc(btc.AddRoutePrefix("/script/{state}/{address}"), btc.getScriptHandler()).Queries("spender", "{spender}").Methods("GET")
	r.HandleFunc(btc.AddRoutePrefix("/confirmations/{txHash}"), btc.getConfirmationsHandler()).Methods("GET")
	r.HandleFunc(btc.AddRoutePrefix("/tx"), btc.postTransaction()).Methods("POST")

	r.HandleFunc(btc.AddRoutePrefix("/omni/balance/{token}/{address}"), btc.getOmniBalanceHandler()).Methods("GET")
}

// Handlers of the bitcoin blockchain
func (btc *bitcoin) AddRoutePrefix(route string) string {
	return fmt.Sprintf("/%s%s", btc.prefix, route)
}

func (btc *bitcoin) Initiated() bool {
	return btc.initiated
}

func (btc *bitcoin) getUTXOhandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts := mux.Vars(r)
		addr := opts["address"]
		limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
		if err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		confirmations, err := strconv.ParseInt(r.URL.Query().Get("confirmations"), 10, 64)
		if err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		utxos, err := btc.client.GetUTXOs(ctx, addr, int(limit), int(confirmations))
		if err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		if err := json.NewEncoder(w).Encode(utxos); err != nil {
			btc.writeError(w, r, http.StatusInternalServerError, err)
			return
		}
	}
}

func (btc *bitcoin) getScriptHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts := mux.Vars(r)
		addr := opts["address"]
		state := opts["state"]

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		var resp GetScriptResponse
		var err error
		switch state {
		case "spent":
			status, script, err2 := btc.client.ScriptSpent(ctx, addr, r.URL.Query().Get("spender"))
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
			status, val, err2 := btc.client.ScriptFunded(ctx, addr, value)
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
			status, val, err2 := btc.client.ScriptRedeemed(ctx, addr, value)
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
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			btc.writeError(w, r, http.StatusInternalServerError, err)
			return
		}
	}
}

func (btc *bitcoin) getConfirmationsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts := mux.Vars(r)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		conf, err := btc.client.Confirmations(ctx, opts["txHash"])
		if err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		if err := json.NewEncoder(w).Encode(GetConfirmationsResponse(conf)); err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
	}
}

func (btc *bitcoin) postTransaction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := PostTransactionRequest{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		stx, err := hex.DecodeString(req.SignedTransaction)
		if err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		if err := btc.client.PublishTransaction(ctx, stx); err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func (btc *bitcoin) getOmniBalanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts := mux.Vars(r)
		token, err := strconv.ParseInt(opts["token"], 10, 64)
		if err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		bal, err := btc.client.OmniGetBalance(token, opts["address"])
		if err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		if err := json.NewEncoder(w).Encode(bal); err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}
	}
}

func (btc *bitcoin) writeError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	if statusCode >= 500 {
		btc.logger.Errorf("failed to call %s with error %v", r.URL.String(), err)
	}
	if statusCode >= 400 {
		btc.logger.Warningf("failed to call %s with error %v", r.URL.String(), err)
	}
	http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), statusCode)
}
