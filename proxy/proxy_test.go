package proxy_test

import (
	"context"
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
			mockClient := NewMockClient()
			proxy := NewProxy(mockClient, mockClient)

			req, err := http.NewRequest("POST", "", nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := proxy.ProxyRequest(context.Background(), req, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should receive a response if the first client is working", func() {
			mockClient := NewMockClient()
			errClient := NewMockErrorClient()
			proxy := NewProxy(mockClient, errClient)

			req, err := http.NewRequest("POST", "", nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := proxy.ProxyRequest(context.Background(), req, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should receive a response if the first client is faulty and second is working", func() {
			mockClient := NewMockClient()
			errClient := NewMockErrorClient()
			proxy := NewProxy(errClient, mockClient)

			req, err := http.NewRequest("POST", "", nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := proxy.ProxyRequest(context.Background(), req, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not receive a response if all clients are faulty", func() {
			errClient := NewMockErrorClient()
			proxy := NewProxy(errClient, errClient)

			req, err := http.NewRequest("POST", "", nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := proxy.ProxyRequest(context.Background(), req, nil)
			Expect(resp).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})
})

type mockClient struct {
}

func NewMockClient() rpc.Client {
	return mockClient{}
}

func (mockClient) HandleRequest(r *http.Request, data []byte) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
	}, nil
}

type mockErrorClient struct {
}

func NewMockErrorClient() rpc.Client {
	return mockErrorClient{}
}

func (mockErrorClient) HandleRequest(r *http.Request, data []byte) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
	}, errors.New("error")
}
