package rpc

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
)

// Data presents the data transported between server and client.
type Data struct {
	Name string        // service name
	Args []interface{} // request's or response's body except error
	Err  string        // remote server error
}

func encode(data Data) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decode(b []byte) (Data, error) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	var data Data
	if err := decoder.Decode(&data); err != nil {
		return Data{}, err
	}

	return data, nil
}

// Transport struct
type Transport struct {
	conn net.Conn
}

// NewTransport creates a transport on connection
func NewTransport(conn net.Conn) *Transport {
	return &Transport{conn}
}

// Send data
func (t *Transport) Send(req Data) error {
	b, err := encode(req) // Encode req into bytes
	if err != nil {
		return err
	}

	buf := make([]byte, 4+len(b))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(b))) // Set Header field
	copy(buf[4:], b)                                    // Set Data field
	_, err = t.conn.Write(buf)

	return err
}

// Receive data
func (t *Transport) Receive() (Data, error) {
	header := make([]byte, 4)
	_, err := io.ReadFull(t.conn, header)
	if err != nil {
		return Data{}, err
	}

	dataLen := binary.BigEndian.Uint32(header) // Read Header filed
	data := make([]byte, dataLen)              // Read Data Field
	_, err = io.ReadFull(t.conn, data)
	if err != nil {
		return Data{}, err
	}

	rsp, err := decode(data) // Decode rsp from bytes
	return rsp, err
}

// Client struct
type Client struct {
	conn net.Conn
}

// NewClient creates a new client
func NewClient(conn net.Conn) *Client {
	return &Client{conn}
}

// Call transforms a function prototype into a function
func (c *Client) Call(name string, fptr interface{}) {
	container := reflect.ValueOf(fptr).Elem()

	f := func(req []reflect.Value) []reflect.Value {
		cliTransport := NewTransport(c.conn)

		errorHandler := func(err error) []reflect.Value {
			outArgs := make([]reflect.Value, container.Type().NumOut())
			for i := 0; i < len(outArgs)-1; i++ {
				outArgs[i] = reflect.Zero(container.Type().Out(i))
			}

			outArgs[len(outArgs)-1] = reflect.ValueOf(&err).Elem()
			return outArgs
		}

		// package request arguments
		inArgs := make([]interface{}, 0, len(req))
		for i := range req {
			inArgs = append(inArgs, req[i].Interface())
		}

		// send request to server
		err := cliTransport.Send(Data{Name: name, Args: inArgs})
		if err != nil { // local network error or encode error
			return errorHandler(err)
		}

		// receive response from server
		rsp, err := cliTransport.Receive()
		if err != nil { // local network error or decode error
			return errorHandler(err)
		}

		if rsp.Err != "" { // remote server error
			return errorHandler(errors.New(rsp.Err))
		}

		if len(rsp.Args) == 0 {
			rsp.Args = make([]interface{}, container.Type().NumOut())
		}

		// unpackage response arguments
		numOut := container.Type().NumOut()
		outArgs := make([]reflect.Value, numOut)

		for i := 0; i < numOut; i++ {
			if i != numOut-1 { // unpackage arguments (except error)
				if rsp.Args[i] == nil { // if argument is nil (gob will ignore "Zero" in transmission), set "Zero" value
					outArgs[i] = reflect.Zero(container.Type().Out(i))
				} else {
					outArgs[i] = reflect.ValueOf(rsp.Args[i])
				}
			} else { // unpackage error argument
				outArgs[i] = reflect.Zero(container.Type().Out(i))
			}
		}

		return outArgs
	}

	container.Set(reflect.MakeFunc(container.Type(), f))
}

// Server struct
type Server struct {
	addr  string
	funcs map[string]reflect.Value
}

// NewServer creates a new server
func NewServer(addr string) *Server {
	return &Server{addr: addr, funcs: make(map[string]reflect.Value)}
}

// Run server To do - go func realize asynchronism
func (s *Server) Run() {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Printf("listen on %s err: %v\n", s.addr, err)
		return
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("accept err: %v\n", err)
			continue
		}

		go func() {
			srvTransport := NewTransport(conn)

			for {
				// read request from client
				req, err := srvTransport.Receive()
				if err != nil {
					if err != io.EOF {
						log.Printf("read err: %v\n", err)
					}
					return
				}

				// get method by name
				f, ok := s.funcs[req.Name]
				if !ok { // if method requested does not exist
					e := fmt.Sprintf("func %s does not exist", req.Name)
					log.Println(e)

					if err = srvTransport.Send(Data{Name: req.Name, Err: e}); err != nil {
						log.Printf("transport write err: %v\n", err)
					}
					continue
				}

				log.Printf("func %s is called\n", req.Name)

				// unpackage request arguments
				inArgs := make([]reflect.Value, len(req.Args))
				for i := range req.Args {
					inArgs[i] = reflect.ValueOf(req.Args[i])
				}

				// invoke requested method - Call wraped f to func prototype
				out := f.Call(inArgs)

				// package response arguments (except error)
				outArgs := make([]interface{}, len(out)-1)
				for i := 0; i < len(out)-1; i++ {
					outArgs[i] = out[i].Interface()
				}

				// package error argument
				var e string
				if _, ok := out[len(out)-1].Interface().(error); !ok {
					e = ""
				} else {
					e = out[len(out)-1].Interface().(error).Error()
				}

				// send response to client
				err = srvTransport.Send(Data{Name: req.Name, Args: outArgs, Err: e})
				if err != nil {
					log.Printf("transport write err: %v\n", err)
				}
			}
		}()
	}
}

// Register a method via name
func (s *Server) Register(name string, f interface{}) {
	if _, ok := s.funcs[name]; ok {
		return
	}

	s.funcs[name] = reflect.ValueOf(f)
}
