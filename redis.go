package cache

import (
	"context"
	"reflect"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStore client
type RedisStore struct {
	client            *redis.Client
	DefaultExpiration time.Duration
	logger            Logger
}

// RedisStoreOptions options
type RedisStoreOptions struct {
	Address           string
	DB                int
	Password          string
	DefaultExpiration time.Duration

	MaxRetries int
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff time.Duration

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration

	// Maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize int
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	MaxConnAge time.Duration
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout time.Duration
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout time.Duration
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency time.Duration

	Logger Logger
}

func NewRedisStore(options *RedisStoreOptions) *RedisStore {
	var opt = &redis.Options{
		Addr:     options.Address,
		DB:       options.DB,
		Password: options.Password,
	}
	if options.DialTimeout > 0 {
		opt.DialTimeout = options.DialTimeout
	}

	if options.MaxRetries > 0 {
		opt.MaxRetries = options.MaxRetries
	}

	if options.MinRetryBackoff > 0 {
		opt.MinRetryBackoff = options.MinRetryBackoff
	}

	if options.MaxRetryBackoff > 0 {
		opt.MaxRetryBackoff = options.MaxRetryBackoff
	}

	if options.ReadTimeout > 0 {
		opt.ReadTimeout = options.ReadTimeout
	}

	if options.WriteTimeout > 0 {
		opt.WriteTimeout = options.WriteTimeout
	}

	if options.PoolSize > 0 {
		opt.PoolSize = options.PoolSize
	}
	if options.MinIdleConns > 0 {
		opt.MinIdleConns = options.MinIdleConns
	}
	if options.MaxConnAge > 0 {
		opt.MaxConnAge = options.MaxConnAge
	}
	if options.PoolTimeout > 0 {
		opt.PoolTimeout = options.PoolTimeout
	}

	if options.IdleTimeout > 0 {
		opt.IdleTimeout = options.IdleTimeout
	}
	if options.IdleCheckFrequency > 0 {
		opt.IdleCheckFrequency = options.IdleCheckFrequency
	}

	var client = redis.NewClient(opt)

	return &RedisStore{
		client:            client,
		DefaultExpiration: options.DefaultExpiration,
		logger:            options.Logger,
	}
}

func (c *RedisStore) Get(key string, value interface{}) error {
	if !isPtr(value) {
		return ErrMustBePointer
	}

	val, err := c.client.Get(context.TODO(), key).Result()

	if err != nil {
		c.Logger().Printf("%s: Get key = %s error %v\n", c.Type(), key, err)
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return err
	}

	err = decode([]byte(val), value)
	if err != nil {
		c.Logger().Printf("%s: Decode key = %s error %v\n", c.Type(), key, err)
		return ErrUnmarshal
	}
	return nil
}

func (c *RedisStore) Set(key string, value interface{}, expiration ...time.Duration) error {
	var v = reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr {
		c.Logger().Printf("%s: Set key = %s value = %v error %v\n", c.Type(), key, value, ErrMustBePointer)
		return ErrMustBePointer
	}

	cacheEntry, err := encode(value)
	if err != nil {
		c.Logger().Printf("%s: Encode key = %s value = %v error %v\n", c.Type(), key, v.Interface(), err)
		return ErrMarshal
	}
	var exp = c.DefaultExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	err = c.client.Set(context.TODO(), key, cacheEntry, exp).Err()
	if err != nil {
		c.Logger().Printf("%s: Set key = %s value = %v error %v\n", c.Type(), key, v.Interface(), err)
		return err
	}
	return nil
}

func (c *RedisStore) Delete(key string) error {
	var err = c.client.Del(context.TODO(), key).Err()
	if err != nil {
		c.Logger().Printf("%s: Delete key = %s value = %v error %v\n", c.Type(), key, err)
		return err
	}
	return nil
}

func (c *RedisStore) Type() string {
	return "redis"
}

func (c *RedisStore) Logger() Logger {
	if c.logger != nil {
		return c.logger
	}
	return DefaultLogger
}
