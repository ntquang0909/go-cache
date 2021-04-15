package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
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

	Logger() Logger
}

type Logger interface {
	Printf(format string, values ...interface{})
}

// ToPtr wraps the given value with pointer: V => *V, *V => **V, etc.
func toPtr(v reflect.Value) reflect.Value {
	pt := reflect.PtrTo(v.Type()) // create a *T type.
	pv := reflect.New(pt.Elem())  // create a reflect.Value of type *T.
	pv.Elem().Set(v)              // sets pv to point to underlying value of v.
	return pv
}

// isPtr check pointer
func isPtr(val interface{}) bool {
	v := reflect.ValueOf(val)
	return v.Kind() == reflect.Ptr
}

func printJSON(val interface{}) {
	data, _ := json.MarshalIndent(val, "", "   ")
	fmt.Println(string(data))
}
