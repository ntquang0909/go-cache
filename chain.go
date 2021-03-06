package cache

import (
	"sync"
	"time"
)

type Chain struct {
	caches []Cache
}

func NewChain(caches ...Cache) *Chain {
	var chain = &Chain{
		caches: caches,
	}

	return chain
}

func (c *Chain) Get(key string, value interface{}) error {
	for _, cache := range c.caches {
		var err = cache.Get(key, value)
		if err == nil {
			return nil
		}
	}

	return ErrKeyNotFound
}

// Set value by give key
func (c *Chain) Set(key string, value interface{}, expiration ...time.Duration) error {
	var wg sync.WaitGroup

	for _, cache := range c.caches {
		wg.Add(1)
		go func(wg *sync.WaitGroup, cache Cache) {
			defer wg.Done()

			cache.Set(key, value, expiration...)
		}(&wg, cache)
	}
	wg.Wait()

	return nil
}

// Delete by give key
func (c *Chain) Delete(key string) error {
	var wg sync.WaitGroup

	for _, cache := range c.caches {
		wg.Add(1)
		go func(wg *sync.WaitGroup, cache Cache) {
			defer wg.Done()

			cache.Delete(key)
		}(&wg, cache)
	}
	wg.Wait()

	return nil
}

func (c *Chain) Type() string {
	return "chain"
}
