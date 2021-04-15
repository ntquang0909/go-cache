package cache

import (
	"reflect"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type MemcacheStore struct {
	client            *memcache.Client
	DefaultExpiration time.Duration
	logger            Logger
}

type MemcacheStoreOptions struct {
	Servers           []string
	DefaultExpiration time.Duration
	MaxIdleConns      int
	Timeout           time.Duration
	Logger            Logger
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
		logger:            options.Logger,
		DefaultExpiration: options.DefaultExpiration,
	}
}

func (c *MemcacheStore) Get(key string, value interface{}) error {
	if !isPtr(value) {
		c.Logger().Printf("%s: Get key = %s value = %v [ERROR] %v\n", c.Type(), key, value, ErrMustBePointer)
		return ErrMustBePointer
	}

	val, err := c.client.Get(key)
	if err != nil {
		return err
	}

	err = decode(val.Value, value)
	if err != nil {
		c.Logger().Printf("%s: Decode key = %s [ERROR] %v\n", c.Type(), key, err)
		return ErrUnmarshal
	}
	return nil
}

func (c *MemcacheStore) Set(key string, value interface{}, expiration ...time.Duration) error {
	var v = reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr {
		c.Logger().Printf("%s: Set key = %s value = %v [ERROR] %v\n", c.Type(), key, value, ErrMustBePointer)
		return ErrMustBePointer
	}

	cacheEntry, err := encode(value)
	if err != nil {
		c.Logger().Printf("%s: Encode key = %s value = %v [ERROR] %v\n", c.Type(), key, v.Interface(), err)
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
		c.Logger().Printf("%s: Set key = %s value = %v [ERROR] %v\n", c.Type(), key, v.Interface(), err)
		return err
	}
	return nil
}

func (c *MemcacheStore) Delete(key string) error {
	var err = c.client.Delete(key)
	if err != nil {
		c.Logger().Printf("%s: Delete key = %s [ERROR] %v\n", c.Type(), key, err)
		return err
	}
	return nil
}

func (c *MemcacheStore) Type() string {
	return "memcache"
}

func (c *MemcacheStore) Logger() Logger {
	if c.logger != nil {
		return c.logger
	}
	return DefaultLogger
}
