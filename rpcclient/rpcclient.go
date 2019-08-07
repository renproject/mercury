package rpcclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
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
	SendRequest(ctx context.Context, method string, response interface{}, params ...interface{}) error
}

type client struct {
	host     string
	user     string
	password string

	retryDelay time.Duration
}

func NewClient(host, user, password string, retryDelay time.Duration) Client {
	return &client{
		host, user, password, retryDelay,
	}
}

func (client *client) SendRequest(ctx context.Context, method string, response interface{}, params ...interface{}) error {
	data, err := encodeRequest(method, params)
	if err != nil {
		return err
	}

	return retry(ctx, client.retryDelay, func() error {
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
	})
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
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	if err := json.Unmarshal(buf.Bytes(), &c); err != nil {
		return fmt.Errorf("cannot decode response body = %s, err = %v", buf.String(), err)
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

func retry(ctx context.Context, delay time.Duration, fn func() error) error {
	ticker := time.NewTicker(delay)
	if err := fn(); err == nil {
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := fn(); err == nil {
				return nil
			}
		}
	}
}
