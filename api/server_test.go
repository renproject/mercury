package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/renproject/kv"
	"github.com/renproject/mercury/api"
	. "github.com/renproject/mercury/api"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc"
	"github.com/renproject/mercury/rpcclient/btcrpcclient"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/phi"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Server", func() {
	Context("when sending concurrent requests", func() {
		It("should not fail on concurrent requests", func() {
			// Initialise Bitcoin API.
			btcTestnetURL := os.Getenv("BITCOIN_TESTNET_RPC_URL")
			btcTestnetUser := os.Getenv("BITCOIN_TESTNET_RPC_USERNAME")
			btcTestnetPassword := os.Getenv("BITCOIN_TESTNET_RPC_PASSWORD")
			logger := logrus.StandardLogger()
			btcTestnetNodeClient := rpc.NewClient(btcTestnetURL, btcTestnetUser, btcTestnetPassword)
			btcTestnetProxy := proxy.NewProxy(btcTestnetNodeClient)
			btcCache := cache.New(kv.NewJSON(kv.NewMemDB()), logger)
			btcTestnetAPI := api.NewApi(btctypes.BtcTestnet, btcTestnetProxy, btcCache, logger)
			server := NewServer(logrus.StandardLogger(), "5000", btcTestnetAPI)
			go server.Run()
			time.Sleep(2 * time.Second)

			phi.ParForAll(5, func(i int) {
				buf := bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "method": "listunspent", "id": 1, "params": [0, 999999, "mwdXtp8ow61jcG1EXYVy5aZqksxvtrnNsL"]}`))
				resp, err := http.Post("http://127.0.0.1:5000/btc/testnet/", "application/json", buf)
				if err != nil {
					logger.Println(err)
					return
				}
				luResp := btcrpcclient.ListUnspentResponse{}
				respBytes, err := ioutil.ReadAll(resp.Body)
				if err := json.NewDecoder(resp.Body).Decode(&luResp); err != nil {
					logger.Println(err)
					return
				}
				fmt.Println(string(respBytes))
			})
		})
	})
})
