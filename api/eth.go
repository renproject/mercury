package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/sirupsen/logrus"
)

type EthApi struct {
	network ethtypes.Network
	proxy   *proxy.Proxy
	cache   *cache.Cache
	logger  logrus.FieldLogger
}

// NewEthApi returns a new EthApi.
func NewEthApi(network ethtypes.Network, proxy *proxy.Proxy, cache *cache.Cache, logger logrus.FieldLogger) *EthApi {
	return &EthApi{
		network: network,
		proxy:   proxy,
		cache:   cache,
		logger:  logger,
	}
}

// AddHandler implements the `BlockchainApi` interface.
func (eth *EthApi) AddHandler(r *mux.Router) {
	var network string
	switch eth.network {
	case ethtypes.Kovan:
		network = "testnet"
	default:
		network = eth.network.String()
	}
	r.HandleFunc(fmt.Sprintf("/eth/%s", network), eth.jsonRPCHandler()).Queries("tag", "{tag}").Methods("POST")
	r.HandleFunc(fmt.Sprintf("/eth/%s", network), eth.jsonRPCHandler()).Methods("POST")
}

func (eth *EthApi) jsonRPCHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			eth.writeError(w, r, http.StatusBadRequest, err)
			return
		}

		hash, err := HashData(data)
		if err != nil {
			eth.writeError(w, r, http.StatusInternalServerError, err)
			return
		}

		// Check if the result has been cached and if not retrieve it (or wait if it is already being retrieved).
		resp, err := eth.cache.Get(hash, FetchResponse(eth.proxy, r, data))
		if err != nil {
			eth.writeError(w, r, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func (eth *EthApi) writeError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	if statusCode >= 500 {
		eth.logger.Errorf("failed to call %s: %v", r.URL.String(), err)
	} else if statusCode >= 400 {
		eth.logger.Warningf("failed to call %s: %v", r.URL.String(), err)
	}
	http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), statusCode)
}
