package rpc

import (
	"bytes"
	"encoding/gob"
)

type Codec interface {
	Marshal(v interface{})([]byte,error)
	Unmarshal(data []byte,v interface{})error
	Name() string
}

type GobCodec int

func (c GobCodec) Marshal(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (c GobCodec) Unmarshal(b []byte, v interface{}) error {
	r := bytes.NewReader(b)
	dec := gob.NewDecoder(r)
	return dec.Decode(v)
}

func (c GobCodec) Name() string {
	return "gob"
}