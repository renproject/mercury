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

	}

	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s", client.host), bytes.NewBuffer(data))
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
		ID:     uint64(rand.Int63()),
		Method: method,
		Params: ps,
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
		return &err{Data: c.Error}
	}
	if c.Result == nil {
		return fmt.Errorf("unexpected null result")
	}
	return json.Unmarshal(*c.Result, reply)
}
