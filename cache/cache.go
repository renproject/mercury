// Package cache allows clients to fetch result from a store without having to execute intensive code numerous times. An
// incoming request first checks to see if the result already exists in the store, if not it executes a function that
// returns the result. Any additional incoming requests wait until this function has finished executing before
// attempting to retrieve data from the store.
package cache

import (
	"errors"
	"sync"

	"github.com/renproject/kv"
	"github.com/sirupsen/logrus"
)

var (
	// ErrNoResponse is returned when one task is waiting for another to retrieve a response, but the first one fails.
	ErrNoResponse = errors.New("cannot get response, please try again later")
)

type Cache struct {
	locks  sync.Map
	store  kv.Store
	logger logrus.FieldLogger
}

// New returns a new Cache.
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
func (cache *Cache) Get(level int64, hash string, f func() ([]byte, error)) ([]byte, error) {
	if level == 2 {
		return f()
	}

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
		defer mu.Unlock()
		defer cache.locks.Delete(hash)

		// Store mutex.
		cache.locks.Store(hash, mu)

		data, err := f()
		if err != nil {
			return nil, err
		}

		if err := cache.store.Insert(hash, data); err != nil {
			cache.logger.Errorf("cannot store response data: %v", err)
		}

		return data, nil
	}

	// Wait for the response to be written to the store.
	mu := v.(*sync.RWMutex)
	mu.RLock()
	if err := cache.store.Get(hash, &data); err != nil {
		mu.RUnlock()
		return nil, ErrNoResponse
	}
	mu.RUnlock()

	return data, nil
}
