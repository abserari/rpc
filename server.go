package rpc

import (
	"context"
	"sync"
)

//Server  mean instance to call service and listen&serve Request
type Server struct {
	codec   baseCodec
	callMap sync.Map //map[string]*Call sync.Map's Load and LoadOrStore method for Register and Read Method
}

//Service represent RPC service, and method for call
type Service struct {
	name string
}

//Request - Call's request info
type Request struct {
}

//Reply - Call's response
type Reply struct {
}

//Endpoint represent RPC Call Chain
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

// NewServerWithCodec return custom codec server
func NewServerWithCodec(codec baseCodec) *Server {
	return &Server{
		codec: codec,
	}
}

// Register registe a Call
func (s *Server) Register(service *Service) {
	s.callMap.LoadOrStore(service.name, service)
}

func (s *Server) Invoke() error {}
