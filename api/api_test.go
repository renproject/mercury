package api_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/api"

	"github.com/renproject/kv"
	"github.com/renproject/mercury/cache"
	"github.com/renproject/mercury/proxy"
	"github.com/renproject/mercury/rpc/ethrpc"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/sirupsen/logrus"
)

var _ = Describe("APIs", func() {
	Context("when interacting with the ETH pass-through API", func() {
		It("should return the same response as infura", func() {
			infuraAPIKey := os.Getenv("INFURA_KEY_DEFAULT")

			store := kv.NewJSON(kv.NewMemDB())
			logger := logrus.StandardLogger()
			cache := cache.New(store, logger)

			client, err := ethrpc.NewInfuraClient(ethtypes.Kovan, infuraAPIKey, nil)
			Expect(err).ToNot(HaveOccurred())
			proxy := proxy.NewProxy(client)
			ethTestnetAPI := NewEthApi(ethtypes.Kovan, proxy, cache, logger)
			server := NewServer(logger, "8378", ethTestnetAPI)

			// Run server and allow it to start before proceeding.
			go server.Run()
			time.Sleep(time.Second)

			data := []byte(`{"jsonrpc":"2.0","id":1,"method":"eth_gasPrice","params":[]}`)

			// Send request to pass-through server.
			resp, err := http.Post("http://0.0.0.0:8378/eth/testnet", "application/json", bytes.NewBuffer(data))
			Expect(err).ToNot(HaveOccurred())

			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			// Send request to Infura.
			infuraResp, err := http.Post(fmt.Sprintf("https://kovan.infura.io/v3/%s", infuraAPIKey), "application/json", bytes.NewBuffer(data))
			Expect(err).ToNot(HaveOccurred())

			infuraRespBytes, err := ioutil.ReadAll(infuraResp.Body)
			Expect(err).ToNot(HaveOccurred())

			// Expect the result to be the same.
			Expect(bytes.Compare(respBytes, infuraRespBytes)).To(Equal(0))
		})
	})
})
