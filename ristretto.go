package cache

import (
	"reflect"
	"time"

	"github.com/dgraph-io/ristretto"
)

type RistrettoStore struct {
	client            *ristretto.Cache
	cost              int64
	DefaultExpiration time.Duration
}

type RistrettoStoreOptions struct {
	NumCounters int64
	MaxCost     int64
	BufferItems int64
	DefaultCost int64
}

var RistrettoStoreOptionsDefault = &RistrettoStoreOptions{
	NumCounters: 1e7,     // number of keys to track frequency of (10M).
	MaxCost:     1 << 30, // maximum cost of cache (1GB).
	BufferItems: 64,      // number of keys per Get buffer.
	DefaultCost: 8,
}

func NewRistrettoStore(options *RistrettoStoreOptions) *RistrettoStore {
	client, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: options.NumCounters,
		MaxCost:     options.MaxCost,
		BufferItems: options.BufferItems,
	})
	if err != nil {
		panic(err)
	}

	return &RistrettoStore{
		client: client,
		cost:   options.DefaultCost,
	}
}

func (c *RistrettoStore) Get(key string, value interface{}) error {
	if !isPtr(value) {
		return ErrMustBePointer
	}

	val, found := c.client.Get(key)
	if !found {
		return ErrKeyNotFound
	}

	var i = reflect.ValueOf(val)
	var o = reflect.ValueOf(value)

	if i.Kind() != reflect.Ptr {
		i = toPtr(i)
	}

	if o.Kind() != reflect.Ptr {
		o = toPtr(o)
	}

	o.Elem().Set(i.Elem())

	return nil
}

func (c *RistrettoStore) Set(key string, value interface{}, expiration ...time.Duration) error {
	if !isPtr(value) {
		return ErrMustBePointer
	}

	var success = c.client.Set(key, value, c.getCost())
	if !success {
		return ErrRistrettoWrite
	}
	return nil
}

func (c *RistrettoStore) Delete(key string) error {
	c.client.Del(key)

	return nil
}

func (c *RistrettoStore) Type() string {
	return "ristretto"
}

func (c *RistrettoStore) getCost() int64 {
	if c.cost > 0 {
		return c.cost
	}

	return 8
}
