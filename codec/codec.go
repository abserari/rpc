package codec

import (
	"fmt"
	"reflect"
)

// BaseCodec contains the functionality of both Codec and encoding.Codec, but
// omits the name/string, which vary between the two and are not needed for
// anything besides the registry in the encoding package.
type BaseCodec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// ByteCodec uses raw slice pf bytes and don't encode/decode.
type ByteCodec struct{}

// Marshal returns raw slice of bytes.
func (c ByteCodec) Marshal(v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	if data, ok := v.(*[]byte); ok {
		return *data, nil
	}

	return nil, fmt.Errorf("%T is not a []byte", v)
}

// Unmarshal returns raw slice of bytes.
func (c ByteCodec) Unmarshal(data []byte, v interface{}) error {
	reflect.Indirect(reflect.ValueOf(v)).SetBytes(data)
	return nil
}
