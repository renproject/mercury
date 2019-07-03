package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/renproject/mercury/proxy"
	"github.com/sirupsen/logrus"
)

type BtcApi struct {
	proxy  *proxy.BtcProxy
	logger logrus.FieldLogger
}

func NewBtcApi(proxy *proxy.BtcProxy, logger logrus.FieldLogger) *BtcApi {
	return &BtcApi{
		proxy:  proxy,
		logger: logger,
	}
}

func (btc *BtcApi) AddHandler(r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/btc/%s", btc.proxy.Network), btc.jsonRPCHandler()).Methods("POST")
}

func (btc *BtcApi) jsonRPCHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := btc.proxy.ProxyRequest(r)
		if err != nil {
			btc.writeError(w, r, resp.StatusCode, err)
			return
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			btc.writeError(w, r, resp.StatusCode, err)
			return
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(data)
	}
}

func (btc *BtcApi) writeError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	if statusCode >= 500 {
		btc.logger.Errorf("failed to call %s with error: %v", r.URL.String(), err)
	}
	if statusCode >= 400 {
		btc.logger.Warningf("failed to call %s with error: %v", r.URL.String(), err)
	}
	http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), statusCode)
}
