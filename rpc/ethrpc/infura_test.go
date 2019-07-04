package ethrpc_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/rpc/ethrpc"

	"github.com/renproject/mercury/types/ethtypes"
)

var _ = Describe("Infura RPC client", func() {
	Context("when interacting with the infura client", func() {
		It("should return the correct response", func() {
			infuraAPIKey := os.Getenv("INFURA_KEY_DEFAULT")

			client, err := NewInfuraClient(ethtypes.Kovan, infuraAPIKey, nil)
			Expect(err).ToNot(HaveOccurred())

			r, err := http.NewRequest("POST", "http://0.0.0.0:5000/eth/testnet", nil)
			Expect(err).ToNot(HaveOccurred())
			data := []byte(`{"jsonrpc":"2.0","id":1,"method":"eth_gasPrice","params":[]}`)

			// Handle request using Infura client.
			resp, err := client.HandleRequest(r, data)
			Expect(err).ToNot(HaveOccurred())

			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			// Send request to Infura directly.
			infuraResp, err := http.Post(fmt.Sprintf("https://kovan.infura.io/v3/%s", infuraAPIKey), "application/json", bytes.NewBuffer(data))
			Expect(err).ToNot(HaveOccurred())

			infuraRespBytes, err := ioutil.ReadAll(infuraResp.Body)
			Expect(err).ToNot(HaveOccurred())

			// Expect the result to be the same.
			Expect(bytes.Compare(respBytes, infuraRespBytes)).To(Equal(0))
		})
	})
})
