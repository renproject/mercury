package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Result json.RawMessage `json:"result"`
}

type Client interface {
	SendRequest(data []byte, response interface{}) error
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

func (client *client) SendRequest(data []byte, response interface{}) error {
	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s", client.host), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	request.SetBasicAuth(client.user, client.password)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(string(msg))
	}

	result := Response{}
	if err := json.Unmarshal(msg, &result); err != nil {
		return err
	}

	return json.Unmarshal(result.Result, response)
}
