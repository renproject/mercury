package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

type ZecApi struct {
	network btctypes.Network
	proxy   *proxy.Proxy
	cache   *cache.Cache
	logger  logrus.FieldLogger
}

// NewZecApi returns a new ZecApi.
func NewZecApi(network btctypes.Network, proxy *proxy.Proxy, cache *cache.Cache, logger logrus.FieldLogger) *ZecApi {
	return &ZecApi{
		network: network,
		proxy:   proxy,
		cache:   cache,
		logger:  logger,
	}
}

// AddHandler implements the `BlockchainApi` interface.
func (zec *ZecApi) AddHandler(r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/zec/%s", zec.network), zec.jsonRPCHandler()).Methods("POST")
}

func (zec *ZecApi) jsonRPCHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			zec.writeError(w, r, http.StatusBadRequest, err)
			return
		}

		hash, err := HashData(data)
		if err != nil {
			zec.writeError(w, r, http.StatusInternalServerError, err)
			return
		}

		// Check if the result has been cached and if not retrieve it (or wait if it is already being retrieved).
		resp, err := zec.cache.Get(hash, FetchResponse(zec.proxy, r, data))
		if err != nil {
			zec.writeError(w, r, http.StatusInternalServerError, err)
			return
		}

		var result Result
		if err := json.Unmarshal(resp, &result); err != nil {
			zec.writeError(w, r, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(result.StatusCode)
		w.Write(result.Data)
	}
}

func (zec *ZecApi) writeError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	if statusCode >= 500 {
		zec.logger.Errorf("failed to call %s: %v", r.URL.String(), err)
	} else if statusCode >= 400 {
		zec.logger.Warningf("failed to call %s: %v", r.URL.String(), err)
	}
	http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), statusCode)
}
