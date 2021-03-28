package cache

import (
	"fmt"
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
		fmt.Printf("%s: Get cache error %v\n", cache.Type(), err)
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

			var err = cache.Set(key, value, expiration...)
			if err != nil {
				fmt.Printf("%s: Set cache key = %s error %v\n", cache.Type(), key, err)
			}
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

			var err = cache.Delete(key)
			if err != nil {
				fmt.Printf("%s: Delete cache key = %s error %v\n", cache.Type(), key, err)
			}
		}(&wg, cache)
	}
	wg.Wait()

	return nil
}
