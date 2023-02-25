package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gitamped/seed/mid"
	"github.com/gitamped/seed/server"
)

func main() {
	// New RPCServer
	s := server.NewServer(mid.CommonMiddleware)

	// Register GreeterServicer
	gs := NewGreeterServicer()
	gs.Register(s)

	//
	fmt.Println(`Listening on port 8080`)
	fmt.Println(`test cmd: curl -X POST  --data '{"name": "seed server"}' http://localhost:8080/v1/GreeterService.Greet`)
	http.Handle("/v1/", s)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
