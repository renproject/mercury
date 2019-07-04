package api

import (
	"encoding/hex"
	"io/ioutil"
	"net/http"

	"github.com/renproject/mercury/proxy"
	"golang.org/x/crypto/sha3"
)

func hashData(data []byte) (string, error) {
	h := sha3.New256()
	h.Write(data)
	hash := hex.EncodeToString(h.Sum(nil))
	return hash, nil
}

func proxyRequest(proxy *proxy.Proxy, r *http.Request, data []byte) func() ([]byte, error) {
	return func() ([]byte, error) {
		// Fetch the response from the API.
		resp, err := proxy.ProxyRequest(r, data)
		if err != nil {
			return nil, err
		}

		// Read the response and insert it into the store.
		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return respData, nil
	}
}
