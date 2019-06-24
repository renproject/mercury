package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/sirupsen/logrus"
)

type EthBackend struct {
	proxy  *proxy.EthProxy
	tags   map[string]string
	logger logrus.FieldLogger
}

func NewEthBackend(proxy *proxy.EthProxy, tags map[string]string, logger logrus.FieldLogger) *EthBackend {
	return &EthBackend{
		proxy:  proxy,
		tags:   tags,
		logger: logger,
	}
}

func (eth *EthBackend) AddHandler(r *mux.Router) {
	r.HandleFunc(eth.addNetworkPrefix(""), eth.jsonRPCHandler()).Queries("tag", "{tag}").Methods("POST")
	r.HandleFunc(eth.addNetworkPrefix(""), eth.jsonRPCHandler()).Methods("POST")
}

func (eth *EthBackend) addNetworkPrefix(route string) string {
	var prefix string
	switch eth.proxy.Network {
	case ethtypes.EthKovan:
		prefix = "eth-kovan"
	default:
		prefix = "eth"
	}
	return fmt.Sprintf("/%s%s", prefix, route)
}

func (eth *EthBackend) jsonRPCHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := eth.tags[r.URL.Query().Get("tag")]
		if apiKey == "" {
			apiKey = eth.tags[""]
		}

		resp, err := eth.proxy.ForwardRequest(r, apiKey)
		if err != nil {
			eth.writeError(w, r, resp.StatusCode, err)
			return
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			eth.writeError(w, r, resp.StatusCode, err)
			return
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(data)
	}
}

func (eth *EthBackend) writeError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	if statusCode >= 500 {
		eth.logger.Errorf("failed to call %s with error %v", r.URL.String(), err)
	}
	if statusCode >= 400 {
		eth.logger.Warningf("failed to call %s with error %v", r.URL.String(), err)
	}
	http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), statusCode)
}
