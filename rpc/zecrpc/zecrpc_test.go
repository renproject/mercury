package zecrpc_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/rpc/zecrpc"
)

var _ = Describe("ZEC RPC client", func() {
	Context("when interacting with the zec client", func() {
		It("should return the correct response", func() {
			testnetURL := os.Getenv("ZCASH_TESTNET_RPC_URL")
			testnetUser := os.Getenv("ZCASH_TESTNET_RPC_USERNAME")
			testnetPassword := os.Getenv("ZCASH_TESTNET_RPC_PASSWORD")

			client, err := NewNodeClient(testnetURL, testnetUser, testnetPassword)

			r, err := http.NewRequest("POST", "http://0.0.0.0:5000/zec/testnet", nil)
			Expect(err).ToNot(HaveOccurred())
			data := []byte(`{"jsonrpc":"2.0","id":1,"method":"getunconfirmedbalance","params":[]}`)

			// Handle request using ZCash client.
			resp, err := client.HandleRequest(r, data)
			Expect(err).ToNot(HaveOccurred())

			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			// Send request to ZCash node directly.
			httpClient := &http.Client{}
			req, err := http.NewRequest("POST", testnetURL, bytes.NewBuffer(data))
			req.SetBasicAuth(testnetUser, testnetPassword)
			nodeResp, err := httpClient.Do(req)
			Expect(err).ToNot(HaveOccurred())

			nodeRespBytes, err := ioutil.ReadAll(nodeResp.Body)
			Expect(err).ToNot(HaveOccurred())

			// Expect the result to be the same.
			Expect(bytes.Compare(respBytes, nodeRespBytes)).To(Equal(0))
		})
	})
})
