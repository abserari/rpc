package rpc

import (
	"context"
	"net"
	"reflect"
	"sync"

	"github.com/yhyddr/rpc/codec"
)

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

// Precompute the reflect type for context.
var typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()

// Service represent RPC service, and method for call
type Service struct {
	name string
	desc string
}

// Server  is a RPC server to serve RPC requests
type Server struct {
	opts serverOptions
	lis  map[net.Listener]bool

	codec      codec.BaseCodec
	serviceMap sync.Map //map[string]*Service sync.Map's Load and LoadOrStore method for Register and Read Method
}

//ServerOptions to configure Server
type serverOptions struct {
}

// A ServerOption sets options such as credentials, codec and keepalive parameters, etc.
type ServerOption interface {
	apply(*serverOptions)
}

// EmptyServerOption does not alter the server configuration. It can be embedded
// in another structure to build custom server options.
//
// This API is EXPERIMENTAL.
type EmptyServerOption struct{}

func (EmptyServerOption) apply(*serverOptions) {}

// funcServerOption wraps a function that modifies serverOptions into an
// implementation of the ServerOption interface.
type funcServerOption struct {
	f func(*serverOptions)
}

func (fdo *funcServerOption) apply(do *serverOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(*serverOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

// NewServerWithCodec return custom codec server
func NewServerWithCodec(codec codec.BaseCodec) *Server {
	return &Server{
		codec: codec,
	}
}

// Register registe a Call
func (s *Server) Register(service *Service) {
	s.serviceMap.LoadOrStore(service.name, service)
}

// Invoke call service's method
func (s *Server) Invoke() error { return nil }

// Serve accepts incoming connections on the listener lis, creating a new
// ServerTransport and service goroutine for each. The service goroutines
// read  requests and then call the registered
// Serve returns when lis.Accept fails with fatal errors.
// lis will auto be closed when this method returns.
func (s *Server) Serve(lis net.Listener) error {
	defer lis.Close()
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		s.serveConn(conn)
	}
}

func (s *Server) serveConn(conn net.Conn) {}

// Endpoint is the fundamental building block of servers and clients.
// It represents a single RPC method.
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

// Nop is an endpoint that does nothing and returns a nil error.
// Useful for tests.
func Nop(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }

// Middleware is a chainable behavior modifier for endpoints.
type Middleware func(Endpoint) Endpoint

// Chain is a helper function for composing middlewares. Requests will
// traverse them in the order they're declared. That is, the first middleware
// is treated as the outermost middleware.
func Chain(outer Middleware, others ...Middleware) Middleware {
	return func(next Endpoint) Endpoint {
		for i := len(others) - 1; i >= 0; i-- { // reverse
			next = others[i](next)
		}
		return outer(next)
	}
}

// Failer may be implemented by Go kit response types that contain business
// logic error details. If Failed returns a non-nil error, the Go kit transport
// layer may interpret this as a business logic error, and may encode it
// differently than a regular, successful response.
//
// It's not necessary for your response types to implement Failer, but it may
// help for more sophisticated use cases. The addsvc example shows how Failer
// should be used by a complete application.
type Failer interface {
	Failed() error
}

func (s *Server) handleRequest(ctx context.Context, request interface{}) (response interface{}, err error) {
	return response, nil
}
