package rpc

import (
	"net"
	"sync"
)

// Server  is a RPC server to serve RPC requests
type Server struct {
	net.Listener
	codec   baseCodec
	callMap sync.Map //map[string]*Service sync.Map's Load and LoadOrStore method for Register and Read Method
}

// Service represent RPC service, and method for call
type Service struct {
	name string
	desc string
}

// Request - Call's request info
type Request struct {
}

// Reply - Call's response
type Reply struct {
}

// NewServerWithCodec return custom codec server
func NewServerWithCodec(codec baseCodec) *Server {
	return &Server{
		codec: codec,
	}
}

// Serve accepts incoming connections on the listener lis, creating a new
// ServerTransport and service goroutine for each. The service goroutines
// read gRPC requests and then call the registered handlers to reply to them.
// Serve returns when lis.Accept fails with fatal errors.  lis will be closed when
// this method returns.
// Serve will return a non-nil error unless Stop or GracefulStop is called.
func (s *Server) Serve(lis net.Listener) error {
	return nil
}

// Register registe a Call
func (s *Server) Register(service *Service) {
	s.callMap.LoadOrStore(service.name, service)
}

// Invoke call service's method
func (s *Server) Invoke() error { return nil }

//Close to close server
func (s *Server) Close() error {
	return s.Listener.Close()
}
