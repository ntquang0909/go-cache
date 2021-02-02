package cache

import (
	"bytes"
	"encoding/gob"
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
