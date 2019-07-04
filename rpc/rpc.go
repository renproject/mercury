package rpc

import "net/http"

// Client is a RPC client which can send and retrieve information from a blockchain through JSON-RPC. `data` is the
// request data we want to send to the ZCash node, and `r` is the original request in case we need to access any query
// parameters or other fields.
type Client interface {
	HandleRequest(r *http.Request, data []byte) (*http.Response, error)
}
