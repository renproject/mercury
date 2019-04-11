package eth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	"github.com/renproject/libeth-go"
	"github.com/renproject/mercury"
)

type ethereum struct {
	account   libeth.Account
	network   string
	privKey   string
	tags      map[string]string
	initiated bool
}

func New(network, privKey string, tags map[string]string) mercury.BlockchainPlugin {
	return &ethereum{
		network: network,
		privKey: privKey,
		tags:    tags,
	}
}

func (eth *ethereum) Init() error {
	privKey, err := crypto.HexToECDSA(eth.privKey)
	if err != nil {
		return err
	}

	var network string
	switch eth.network {
	case "eth":
		network = "mainnet"
	case "eth-ropsten":
		network = "ropsten"
	case "eth-kovan":
		network = "kovan"
	default:
		return fmt.Errorf("unsupported network: %s", network)
	}

	client, err := libeth.NewInfuraClient(network, eth.tags[""])
	if err != nil {
		return err
	}

	account, err := libeth.NewAccount(client, privKey)
	if err != nil {
		return err
	}

	eth.account = account
	eth.initiated = true

	return nil
}

func (eth *ethereum) Initiated() bool {
	return eth.initiated
}

// Handlers of the bitcoin blockchain
func (eth *ethereum) AddRoutes(r *mux.Router) {
	r.HandleFunc(eth.AddRoutePrefix(""), eth.jsonRPCHandler()).Queries("tag", "{tag}").Methods("POST")
	r.HandleFunc(eth.AddRoutePrefix(""), eth.jsonRPCHandler()).Methods("POST")
	r.HandleFunc(eth.AddRoutePrefix("/relay"), eth.relayHandler()).Methods("POST")
}

func (eth *ethereum) AddRoutePrefix(route string) string {
	return fmt.Sprintf("/%s%s", eth.network, route)
}

func (eth *ethereum) jsonRPCHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := eth.tags[r.URL.Query().Get("tag")]
		if apiKey == "" {
			apiKey = eth.tags[""]
		}

		var network string
		switch eth.network {
		case "eth":
			network = "mainnet"
		case "eth-ropsten":
			network = "ropsten"
		case "eth-kovan":
			network = "kovan"
		default:
			http.Error(w, fmt.Sprintf("{ \"error\": \"unsupported network: %s\" }", network), http.StatusBadRequest)
			return
		}

		resp, err := http.Post(fmt.Sprintf("https://%s.infura.io/v3/%s", network, apiKey), "application/json", r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), resp.StatusCode)
			return
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), resp.StatusCode)
			return
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(data)
	}
}

func (eth *ethereum) relayHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := RelayRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), http.StatusBadRequest)
			return
		}
		resp, err := eth.Relay(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), http.StatusBadRequest)
			return
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), http.StatusInternalServerError)
			return
		}
	}
}
