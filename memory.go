package cache

import (
	"reflect"
	"time"

	"github.com/patrickmn/go-cache"
)

type MemoryStore struct {
	client            *cache.Cache
	DefaultExpiration time.Duration
	logger            Logger
}

type MemoryStoreOptions struct {
	DefaultExpiration time.Duration
	DefaultCacheItems map[string]cache.Item
	CleanupInterval   time.Duration
	Logger            Logger
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
		logger:            options.Logger,
	}
}

func (c *MemoryStore) Get(key string, value interface{}) error {
	var o = reflect.ValueOf(value)
	if o.Kind() != reflect.Ptr {
		c.Logger().Printf("%s: Get key = %s value = %v [ERROR] %v\n", c.Type(), key, value, ErrMustBePointer)
		return ErrMustBePointer
	}

	val, found := c.client.Get(key)
	if !found {
		c.Logger().Printf("%s: Get key = %s [ERROR] %v\n", c.Type(), key, ErrKeyNotFound)
		return ErrKeyNotFound
	}

	var i = reflect.ValueOf(val)

	if i.Kind() != reflect.Ptr {
		i = toPtr(i)
	}

	o.Elem().Set(i.Elem())

	return nil
}

func (c *MemoryStore) Set(key string, value interface{}, expiration ...time.Duration) error {
	var v = reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr {
		c.Logger().Printf("%s: Set key = %s value = %v [ERROR] %v\n", c.Type(), key, value, ErrMustBePointer)
		return ErrMustBePointer
	}

	var exp = c.DefaultExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	c.client.Set(key, value, exp)
	return nil
}

func (c *MemoryStore) Delete(key string) error {
	c.client.Delete(key)
	return nil
}

func (c *MemoryStore) Type() string {
	return "memory"
}

func (c *MemoryStore) Logger() Logger {
	if c.logger != nil {
		return c.logger
	}
	return DefaultLogger
}
