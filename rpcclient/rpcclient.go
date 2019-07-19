package rpcclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

// request represents a JSON-RPC request sent by a client.
type request struct {
	Version string            `json:"jsonrpc"`
	ID      int64             `json:"id"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

// response represents a JSON-RPC response returned to a client.
type response struct {
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
	ID     int64            `json:"id"`
}

// errObj is a wrapper for a JSON interface value.
type errObj struct {
	Data interface{}
}

func (e *errObj) Error() string {
	return fmt.Sprintf("%v", e.Data)
}

type Client interface {
	SendRequest(method string, response interface{}, params ...interface{}) error
}

type client struct {
	host     string
	user     string
	password string
}

func NewClient(host, user, password string) Client {
	return &client{
		host, user, password,
	}
}

func (client *client) SendRequest(method string, response interface{}, params ...interface{}) error {
	data, err := encodeRequest(method, params)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", client.host, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	request.SetBasicAuth(client.user, client.password)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	return decodeResponse(resp.Body, response)
}

// encodeRequest encodes parameters for a JSON-RPC client request.
func encodeRequest(method string, params []interface{}) ([]byte, error) {
	ps := make([]json.RawMessage, len(params))

	var err error
	for i := range ps {
		ps[i], err = json.Marshal(params[i])
		if err != nil {
			return nil, err
		}
	}

	req := &request{
		Version: "2.0",
		ID:      rand.Int63(),
		Method:  method,
		Params:  ps,
	}
	return json.Marshal(req)
}

// decodeResponse decodes the response body of a client request into the interface reply.
func decodeResponse(r io.Reader, reply interface{}) error {
	var c response
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return err
	}

	if c.Error != nil {
		return &errObj{Data: c.Error}
	}
	if c.Result == nil {
		return ErrNullResult
	}
	return json.Unmarshal(*c.Result, reply)
}

var ErrNullResult = fmt.Errorf("unexpected null result")
