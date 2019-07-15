package rpc



type Server struct {
	Service sync.Map //use Load and LoadOrStore method for Register and Read Method
}
//represent RPC service
type Call struct {
	ServiceMethod string
	Id            uint64
}