package types

import "encoding/json"

// JSONError defines a JSON error object that is compatible with the JSON-RPC 2.0 specification. See
// https://www.jsonrpc.org/specification for more information.
type JSONError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// JSONRequest defines a JSON request object that is compatible with the JSON-RPC 2.0 specification. See
// https://www.jsonrpc.org/specification for more information.
type JSONRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// JSONResponse defines a JSON response object that is compatible with the JSON-RPC 2.0 specification. See
// https://www.jsonrpc.org/specification for more information.
type JSONResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Error   *JSONError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	ID      interface{}     `json:"id"`
}
