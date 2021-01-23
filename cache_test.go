package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var instance Cache

type CacheItem struct {
	Name string
}

func testStore(t *testing.T) {
	// Test string
	var key = "test_key"
	var strIn = "Hello world"
	var err = instance.Set(key, &strIn)
	assert.NoError(t, err)

	var strOut string
	err = instance.Get(key, &strOut)
	assert.NoError(t, err)

	assert.Equal(t, strOut, strIn)

	// Test struct
	var itemIn = CacheItem{
		Name: "Test",
	}
	err = instance.Set(key, &itemIn)
	assert.NoError(t, err)

	var itemOut CacheItem
	err = instance.Get(key, &itemOut)
	assert.NoError(t, err)

	assert.Equal(t, itemOut, itemIn)

	// Test bool
	var boolIn = true
	err = instance.Set(key, &boolIn)
	assert.NoError(t, err)

	var boolOut bool
	err = instance.Get(key, &boolOut)
	assert.NoError(t, err)

	assert.Equal(t, boolIn, boolOut)
}
func TestRedisCache(t *testing.T) {
	instance = NewRedisStore(&RedisStoreOptions{
		Address: "localhost:6379",
	})
	testStore(t)

}

func TestMemoryCache(t *testing.T) {
	instance = NewMemoryStore(MemoryStoreOptions{})
	testStore(t)

}

func TestMemcacheCache(t *testing.T) {
	/*
		Install memcached macOS:
		brew install memcached
		ps -few | grep memcached
	*/
	instance = NewMemcacheStore(&MemcacheStoreOptions{
		Servers: []string{"localhost:11211"},
	})

	testStore(t)

}
