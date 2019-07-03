package api

import (
	"encoding/hex"
	"io/ioutil"
	"net/http"

	"github.com/renproject/mercury/proxy"
	"golang.org/x/crypto/sha3"
)

func hashRequest(r *http.Request) (string, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	h := sha3.New256()
	h.Write(body)
	hash := hex.EncodeToString(h.Sum(nil))

	return hash, nil
}

func proxyRequest(proxy *proxy.Proxy, r *http.Request) func() ([]byte, error) {
	return func() ([]byte, error) {
		// Fetch the response from the API.
		resp, err := proxy.ProxyRequest(r)
		if err != nil {
			return nil, err
		}

		// Read the response and insert it into the store.
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return data, nil
	}
}