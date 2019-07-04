package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

type BtcApi struct {
	network btctypes.Network
	proxy   *proxy.Proxy
	cache   *cache.Cache
	logger  logrus.FieldLogger
}

// NewBtcApi returns a new BtcApi.
func NewBtcApi(network btctypes.Network, proxy *proxy.Proxy, cache *cache.Cache, logger logrus.FieldLogger) *BtcApi {
	return &BtcApi{
		network: network,
		proxy:   proxy,
		cache:   cache,
		logger:  logger,
	}
}

// AddHandler implements the `BlockchainApi` interface.
func (btc *BtcApi) AddHandler(r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/btc/%s", btc.network), btc.jsonRPCHandler()).Methods("POST")
}

func (btc *BtcApi) jsonRPCHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			btc.writeError(w, r, http.StatusBadRequest, err)
			return
		}

		hash, err := hashData(data)
		if err != nil {
			btc.writeError(w, r, http.StatusInternalServerError, err)
			return
		}

		// Check if the result has been cached and if not retrieve it (or wait if it is already being retrieved).
		resp, err := btc.cache.Get(hash, proxyRequest(btc.proxy, r, data))
		if err != nil {
			btc.writeError(w, r, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func (btc *BtcApi) writeError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	if statusCode >= 500 {
		btc.logger.Errorf("failed to call %s: %v", r.URL.String(), err)
	} else if statusCode >= 400 {
		btc.logger.Warningf("failed to call %s: %v", r.URL.String(), err)
	}
	http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), statusCode)
}
