package proxy_test

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/proxy"

	"github.com/renproject/mercury/rpc"
)

var _ = Describe("Proxies", func() {
	Context("when creating a proxy", func() {
		It("should receive a response if all clients are working", func() {
			mockClient := NewMockBtcClient()
			proxy := NewProxy(mockClient, mockClient)

			req, err := http.NewRequest("POST", "", nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := proxy.ProxyRequest(req)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should receive a response if the first client is working", func() {
			mockClient := NewMockBtcClient()
			errClient := NewMockBtcErrorClient()
			proxy := NewProxy(mockClient, errClient)

			req, err := http.NewRequest("POST", "", nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := proxy.ProxyRequest(req)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should receive a response if the first client is faulty and second is working", func() {
			mockClient := NewMockBtcClient()
			errClient := NewMockBtcErrorClient()
			proxy := NewProxy(errClient, mockClient)

			req, err := http.NewRequest("POST", "", nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := proxy.ProxyRequest(req)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not receive a response if all clients are faulty", func() {
			errClient := NewMockBtcErrorClient()
			proxy := NewProxy(errClient, errClient)

			req, err := http.NewRequest("POST", "", nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := proxy.ProxyRequest(req)
			Expect(resp).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})
})

type mockBtcClient struct {
}

func NewMockBtcClient() rpc.Client {
	return mockBtcClient{}
}

func (mockBtcClient) HandleRequest(r *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
	}, nil
}

type mockBtcErrorClient struct {
}

func NewMockBtcErrorClient() rpc.Client {
	return mockBtcErrorClient{}
}

func (mockBtcErrorClient) HandleRequest(r *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
	}, errors.New("error")
}
