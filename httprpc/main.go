package main

import (
	"log"
	"net/http"
)

func sum(a, b int) (int, string) {
	return a + b, "sum"
}
func main() {
	server:= NewServer()

	server.AddFunction("sum", sum)

	http.Handle("/rpc", server)

	if err := http.ListenAndServe("127.0.0.1:8181", nil); err != nil {
		log.Fatal(err)
	}
}
