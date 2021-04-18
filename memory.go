package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/vmihailenco/msgpack/v5"
)

type MemoryStore struct {
	client            *cache.Cache
	DefaultExpiration time.Duration
}

type MemoryStoreOptions struct {
	DefaultExpiration time.Duration
	DefaultCacheItems map[string]cache.Item
	CleanupInterval   time.Duration
}

var MemoryStoreOptionsDefault = &MemoryStoreOptions{
	DefaultExpiration: time.Hour * 24,
	CleanupInterval:   time.Hour * 26,
}

func NewMemoryStore(options MemoryStoreOptions) *MemoryStore {
	var items = make(map[string]cache.Item)
	if options.DefaultCacheItems != nil {
		items = options.DefaultCacheItems
	}

	var client = cache.NewFrom(options.DefaultExpiration, options.CleanupInterval, items)
	return &MemoryStore{
		client:            client,
		DefaultExpiration: options.DefaultExpiration,
	}
}

func (c *MemoryStore) Get(key string, value interface{}) error {
	if !isPtr(value) {
		return ErrMustBePointer
	}

	val, found := c.client.Get(key)
	if !found {
		return ErrKeyNotFound
	}

	var err error
	switch v := val.(type) {
	case []byte:
		err = msgpack.Unmarshal(v, value)
	case string:
		err = msgpack.Unmarshal([]byte(v), value)

	}

	return err
}

func (c *MemoryStore) Set(key string, value interface{}, expiration ...time.Duration) error {
	if !isPtr(value) {
		return ErrMustBePointer
	}

	var exp = c.DefaultExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	bytes, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}
	c.client.Set(key, bytes, exp)
	return nil
}

func (c *MemoryStore) Delete(key string) error {
	c.client.Delete(key)
	return nil
}

func (c *MemoryStore) Type() string {
	return "memory"
}
