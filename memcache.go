package cache

import (
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/vmihailenco/msgpack/v5"
)

type MemcacheStore struct {
	client            *memcache.Client
	DefaultExpiration time.Duration
}

type MemcacheStoreOptions struct {
	Servers           []string
	DefaultExpiration time.Duration
	MaxIdleConns      int
	Timeout           time.Duration
}

func NewMemcacheStore(options *MemcacheStoreOptions) *MemcacheStore {
	if len(options.Servers) == 0 {
		panic(ErrMemcacheServerRequired)
	}

	var client = memcache.New(options.Servers...)
	if options.MaxIdleConns > 0 {
		client.MaxIdleConns = options.MaxIdleConns
	}
	if options.Timeout > 0 {
		client.Timeout = options.Timeout
	}
	return &MemcacheStore{
		client:            client,
		DefaultExpiration: options.DefaultExpiration,
	}
}

func (c *MemcacheStore) Get(key string, value interface{}) error {
	if !isPtr(value) {
		return ErrMustBePointer
	}

	val, err := c.client.Get(key)
	if err != nil {
		return err
	}

	err = msgpack.Unmarshal(val.Value, value)
	if err != nil {
		return ErrUnmarshal
	}
	return nil
}

func (c *MemcacheStore) Set(key string, value interface{}, expiration ...time.Duration) error {
	if !isPtr(value) {
		return ErrMustBePointer
	}

	cacheEntry, err := msgpack.Marshal(value)
	if err != nil {
		return ErrMarshal
	}
	var exp = c.DefaultExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	var item = memcache.Item{
		Key:        key,
		Expiration: int32(exp.Seconds()),
		Value:      cacheEntry,
	}
	err = c.client.Set(&item)
	if err != nil {
		return err
	}
	return nil
}

func (c *MemcacheStore) Delete(key string) error {
	var err = c.client.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (c *MemcacheStore) Type() string {
	return "memcache"
}
