package rpc

import "net/http"

// Client is a RPC client which can send and retrieve information from a blockchain through JSON-RPC.
type Client interface {
	HandleRequest(r *http.Request, data []byte) (*http.Response, error)
}
