package cache

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
)

func decode(data []byte, in interface{}) error {
	var buf = bytes.NewBuffer(data)
	var dec = gob.NewDecoder(buf)
	var err = dec.Decode(in)
	return err
}

func encode(in interface{}) ([]byte, error) {
	var buf bytes.Buffer
	var enc = gob.NewEncoder(&buf)
	var err = enc.Encode(in)
	return buf.Bytes(), err
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
