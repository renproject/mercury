package cache_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/mercury/cache"

	"github.com/renproject/kv"
	"github.com/renproject/phi"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Cache", func() {
	rand.Seed(0)

	Context("when sending multiple identical requests", func() {
		It("should only forward a single request", func() {
			store := kv.NewTable(kv.NewMemDB(kv.JSONCodec), "test")
			logger := logrus.StandardLogger()
			cache := New(store, logger)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte{})
			}))
			defer server.Close()

			numRequests := 0
			phi.ParBegin(func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				_, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				_, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				_, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				_, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				_, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				_, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				_, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				_, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
			})

			Expect(numRequests).To(Equal(1))
		})

		It("should return the same result for each request", func() {
			store := kv.NewTable(kv.NewMemDB(kv.JSONCodec), "test")
			logger := logrus.StandardLogger()
			cache := New(store, logger)

			response := []byte("response")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(response)
			}))
			defer server.Close()

			numRequests := 0
			phi.ParBegin(func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				resp, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(response))
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				resp, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(response))
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				resp, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(response))
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				resp, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(response))
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				resp, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(response))
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				resp, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(response))
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				resp, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(response))
			}, func() {
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				resp, err := cache.Get(1, "hash", getResponse(server.URL, &numRequests))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(response))
			})
		})
	})
})

func getResponse(url string, numRequests *int) func() ([]byte, error) {
	return func() ([]byte, error) {
		*numRequests++

		req := map[string]string{"": ""}
		reqBytes, err := json.Marshal(req)
		Expect(err).ToNot(HaveOccurred())

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
		Expect(err).ToNot(HaveOccurred())

		data, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		// Add 3 second timeout to simulate latency.
		time.Sleep(3 * time.Second)

		return data, nil
	}
}
