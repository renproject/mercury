package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/types/zectypes"
	"github.com/sirupsen/logrus"
)

type ZecApi struct {
	network zectypes.Network
	proxy   *proxy.Proxy
	cache   *cache.Cache
	logger  logrus.FieldLogger
}

func NewZecApi(network zectypes.Network, proxy *proxy.Proxy, cache *cache.Cache, logger logrus.FieldLogger) *ZecApi {
	return &ZecApi{
		network: network,
		proxy:   proxy,
		cache:   cache,
		logger:  logger,
	}
}

func (zec *ZecApi) AddHandler(r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/zec/%s", zec.network), zec.jsonRPCHandler()).Methods("POST")
}

func (zec *ZecApi) jsonRPCHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash, err := hashRequest(r)
		if err != nil {
			zec.writeError(w, r, http.StatusInternalServerError, err)
			return
		}

		// Check if the result has been cached and if not retrieve it (or wait if it is already being retrieved).
		resp, err := zec.cache.Get(hash, proxyRequest(zec.proxy, r))
		if err != nil {
			zec.writeError(w, r, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resp)
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
