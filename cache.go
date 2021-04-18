package cache

import (
	"errors"
	"log"
	"os"
	"time"
)

// Default
const (
	NoExpiration = time.Duration(0)
)

// Errors
var (
	ErrKeyNotFound            = errors.New("cache: Key not found")
	ErrUnmarshal              = errors.New("cache: Unmarshal error")
	ErrMarshal                = errors.New("cache: Marshal error")
	ErrMustBePointer          = errors.New("cache: Must be a pointer")
	ErrMemcacheServerRequired = errors.New("cache: Memcache must have a valid server")
	ErrRistrettoWrite         = errors.New("cache: Ristretto write error")
)

var DefaultLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

// Cache interface
type Cache interface {
	Get(key string, value interface{}) error

	Set(key string, value interface{}, expire ...time.Duration) error

	Delete(key string) error

	Type() string
}
