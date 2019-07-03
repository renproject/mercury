package api

import (
	"fmt"
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

func NewEthApi(network ethtypes.Network, proxy *proxy.Proxy, cache *cache.Cache, logger logrus.FieldLogger) *EthApi {
	return &EthApi{
		network: network,
		proxy:   proxy,
		cache:   cache,
		logger:  logger,
	}
}

func (eth *EthApi) AddHandler(r *mux.Router) {
	var network string
	switch eth.network {
	case ethtypes.Kovan:
		network = "testnet"
	default:
		network = eth.network.String()
	}
	r.HandleFunc(fmt.Sprintf("/eth/%s", network), eth.jsonRPCHandler()).Methods("POST")
}

func (eth *EthApi) jsonRPCHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash, err := hashRequest(r)
		if err != nil {
			eth.writeError(w, r, http.StatusInternalServerError, err)
			return
		}

		// Check if the result has been cached and if not retrieve it (or wait if it is already being retrieved).
		resp, err := eth.cache.Get(hash, proxyRequest(eth.proxy, r))
		if err != nil {
			eth.writeError(w, r, http.StatusInternalServerError, err)
			return
		}

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
