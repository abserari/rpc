package rpc

import "sync"

//Server  mean instance to call service and listen&serve Request
type Server struct {
	Codec   baseCodec
	CallMap sync.Map //map[string]*Call sync.Map's Load and LoadOrStore method for Register and Read Method
}

//Call represent RPC service
type Call struct {
	name string
}

//Request - Call's request info
type Request struct {
}

//Reply - Call's response
type Reply struct {
}

// NewServerWithCodec return custom codec server
func NewServerWithCodec(codec baseCodec) *Server {
	return &Server{
		Codec: codec,
	}
}

// Register registe a Call
func (server *Server) Register(call *Call) {
	server.CallMap.LoadOrStore(call.name, *Call)
}
