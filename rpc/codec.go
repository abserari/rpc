package rpc

// BaseCodec contains the functionality of both Codec and encoding.Codec, but
// omits the name/string, which vary between the two and are not needed for
// anything besides the registry in the encoding package.
type BaseCodec interface {
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
}
