package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

const (
	ErrorCodeInvalidJSON    = -32700
	ErrorCodeInvalidRequest = -32600
)

type Api struct {
	network types.Network
	proxy   *proxy.Proxy
	cache   *cache.Cache
	logger  logrus.FieldLogger
}

// NewApi returns a new Api.
func NewApi(network types.Network, proxy *proxy.Proxy, cache *cache.Cache, logger logrus.FieldLogger) *Api {
	return &Api{
		network: network,
		proxy:   proxy,
		cache:   cache,
		logger:  logger,
	}
}

// AddHandler implements the `BlockchainApi` interface.
func (api *Api) AddHandler(r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/%s", api.network.Chain(), api.network), api.jsonRPCHandler()).Methods("POST")
}

func (api *Api) jsonRPCHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeError(w, r, api.logger, http.StatusBadRequest, ErrorCodeInvalidJSON, err)
			return
		}

		method, id, err := GetMethodAndID(data)
		if err != nil {
			writeError(w, r, api.logger, http.StatusBadRequest, ErrorCodeInvalidRequest, fmt.Errorf("cannot get the method: %v", err))
			return
		}

		level := WhitelistLevel(api.network, method)
		if level == 0 {
			writeError(w, r, api.logger, http.StatusMethodNotAllowed, id, fmt.Errorf("method unavailable: %s", method))
			return
		}

		hash, err := HashData(data)
		if err != nil {
			writeError(w, r, api.logger, http.StatusInternalServerError, id, err)
			return
		}

		// Check if the result has been cached and if not retrieve it (or wait if it is already being retrieved).
		resp, err := api.cache.Get(level, hash, FetchResponse(api.proxy, r, data))
		if err != nil {
			writeError(w, r, api.logger, http.StatusInternalServerError, id, err)
			return
		}

		var result Result
		if err := json.Unmarshal(resp, &result); err != nil {
			writeError(w, r, api.logger, http.StatusInternalServerError, id, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(result.StatusCode)
		w.Write(result.Data)
	}
}

type Result struct {
	Data       []byte
	StatusCode int
}

func HashData(data []byte) (string, error) {
	h := sha3.New256()
	h.Write(data)
	hash := hex.EncodeToString(h.Sum(nil))
	return hash, nil
}

func GetMethodAndID(data []byte) (string, int, error) {
	req := struct {
		Method string `json:"method"`
		ID     int    `json:"id"`
	}{}
	if err := json.Unmarshal(data, &req); err != nil {
		return "", -1, err
	}
	return req.Method, req.ID, nil
}

func FetchResponse(proxy *proxy.Proxy, r *http.Request, data []byte) func() ([]byte, error) {
	return func() ([]byte, error) {
		// TODO: Update the timeout as per requirements.
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		// Fetch the response from the API.
		resp, err := proxy.ProxyRequest(ctx, r, data)
		if err != nil {
			return nil, err
		}

		// Read the response and return it.
		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		result := Result{
			Data:       respData,
			StatusCode: resp.StatusCode,
		}

		return json.Marshal(result)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, logger logrus.FieldLogger, statusCode, id int, err error) {
	resp := struct {
		Error string `json:"error"`
		ID    int    `json:"id"`
	}{
		Error: err.Error(),
		ID:    id,
	}

	if statusCode >= 500 {
		logger.Errorf("failed to call %s: %v", r.URL.String(), err)
	} else if statusCode >= 400 {
		logger.Warningf("failed to call %s: %v", r.URL.String(), err)
	}

	errMsg, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal the error message: %v", err), statusCode)
		return
	}
	http.Error(w, string(errMsg), statusCode)
}
