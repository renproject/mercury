// TODO: Note assumptions (same request will always return the same response)
package cache

import (
	"fmt"
	"sync"

	"github.com/renproject/kv"
	"github.com/sirupsen/logrus"
)

type Cache struct {
	locks  sync.Map
	store  kv.Store
	logger logrus.FieldLogger
}

func New(store kv.Store, logger logrus.FieldLogger) *Cache {
	return &Cache{
		locks:  sync.Map{},
		store:  store,
		logger: logger,
	}
}

// Get checks if the data for a given hash exists in the store, and if not, uses f() to retrieve the result. Any
// requests that are sent while the result is being retrieved, wait until the first function call returns. This prevents
// the function f() from being called multiple times for the same request.
func (cache *Cache) Get(hash string, f func() ([]byte, error)) ([]byte, error) {
	// Check if the result already exists in the store.
	var data []byte
	if err := cache.store.Get(hash, &data); err == nil {
		return data, nil
	}

	// If not, check to see if a mutex exists.
	v, ok := cache.locks.Load(hash)
	if !ok {
		mu := new(sync.RWMutex)
		mu.Lock()
		cache.locks.Store(hash, mu)

		data, err := f()
		if err != nil {
			mu.RUnlock()
			return nil, err
		}

		if err := cache.store.Insert(hash, data); err != nil {
			cache.logger.Errorf("cannot store response data: %v", err)
		}

		cache.locks.Delete(hash)
		mu.Unlock()
		return data, nil
	}

	// Wait for the response to be written to the store.
	mu := v.(*sync.RWMutex)
	mu.RLock()
	if err := cache.store.Get(hash, &data); err != nil {
		mu.RUnlock()
		return nil, fmt.Errorf("cannot get response: %v", err)
	}
	mu.RUnlock()

	return data, nil
}