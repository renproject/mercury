package cache_test

import (
	"bytes"
	"context"
	"math/rand"
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
		Context("when the request is of CachedAccess level", func() {
			It("should only forward a single request", func() {
				db := kv.NewMemDB(kv.JSONCodec)
				store := kv.NewTable(db, "test")
				ttl := kv.NewTTLCache(context.Background(), db, "ttl", time.Second)
				logger := logrus.StandardLogger()
				cache := New(store, ttl, logger)

				counter := 0
				f := func() ([]byte, error) {
					counter++
					return []byte(time.Now().String()), nil
				}

				phi.ParForAll(10, func(i int) {
					time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
					_, err := cache.Get(1, "hash", f)
					Expect(err).ToNot(HaveOccurred())
				})

				Expect(counter).To(Equal(1))
			})

			It("should return the same result for each request", func() {
				db := kv.NewMemDB(kv.JSONCodec)
				store := kv.NewTable(db, "test")
				ttl := kv.NewTTLCache(context.Background(), db, "ttl", time.Second)
				logger := logrus.StandardLogger()
				cache := New(store, ttl, logger)

				counter := 0
				f := func() ([]byte, error) {
					counter++
					return []byte(time.Now().String()), nil
				}

				numRequests := 10
				responses := make([][]byte, numRequests)
				phi.ParForAll(numRequests, func(i int) {
					time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
					resp, err := cache.Get(1, "hash", f)
					Expect(err).ToNot(HaveOccurred())
					responses[i] = resp
				})
				for _, resp := range responses {
					Expect(bytes.Equal(resp, responses[0])).Should(BeTrue())
				}
			})
		})

		Context("when the request is FullAccess level", func() {
			It("should return the same result for each request and only be forwards once", func() {
				db := kv.NewMemDB(kv.JSONCodec)
				store := kv.NewTable(db, "test")
				ttl := kv.NewTTLCache(context.Background(), db, "ttl", time.Second)
				logger := logrus.StandardLogger()
				cache := New(store, ttl, logger)

				expire := time.After(4 * time.Second)
				counter := 0
				f := func() ([]byte, error) {
					counter++
					return []byte(time.Now().String()), nil
				}

				// Expect result to be cached and only one request will be forwarded
				numRequests := 10
				responses := make([][]byte, numRequests)
				phi.ParForAll(numRequests, func(i int) {
					time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
					resp, err := cache.Get(2, "hash", f)
					Expect(err).ToNot(HaveOccurred())
					responses[i] = resp
				})
				Expect(counter).Should(Equal(1))
				for _, resp := range responses {
					Expect(bytes.Equal(resp, responses[0])).Should(BeTrue())
				}

				// Expect the result to be expired and the request be forwarded again.
				<-expire
				resp, err := cache.Get(2, "hash", f)
				Expect(err).ToNot(HaveOccurred())
				Expect(counter).Should(Equal(2))
				Expect(bytes.Equal(resp, responses[0])).ShouldNot(BeTrue())
			})
		})
	})
})
