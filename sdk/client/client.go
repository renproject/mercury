package client

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
)

// request represents a JSON-RPC request sent by a client.
type request struct {
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
	ID     uint64            `json:"id"`
}

// response represents a JSON-RPC response returned to a client.
type response struct {
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
	ID     uint64           `json:"id"`
}

// err is a wrapper for a JSON interface value.
type err struct {
	Data interface{}
}

func (e *err) Error() string {
	return fmt.Sprintf("%v", e.Data)
}

// EncodeRequest encodes parameters for a JSON-RPC client request.
func EncodeRequest(method string, params []json.RawMessage) ([]byte, error) {
	req := &request{
		ID:     uint64(rand.Int63()),
		Method: method,
		Params: params,
	}
	return json.Marshal(req)
}

// DecodeResponse decodes the response body of a client request into the interface reply.
func DecodeResponse(r io.Reader, reply interface{}) error {
	var c response
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return err
	}
	if c.Error != nil {
		return &err{Data: c.Error}
	}
	if c.Result == nil {
		return fmt.Errorf("unexpected null result")
	}
	return json.Unmarshal(*c.Result, reply)
}
