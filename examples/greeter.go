package main

import (
	"encoding/json"
	"fmt"

	"github.com/gitamped/seed/server"
	"github.com/gitamped/seed/validate"
)

// GreeterService is a polite API for greeting people.
type GreeterService interface {
	// Greet prepares a lovely greeting.
	Greet(GreetRequest, server.GenericRequest) GreetResponse
}

// Required to register endpoints with the Server
type GreeterRpcService interface {
	GreeterService
	// Registers RPCService with Server
	Register(s *server.Server)
}

// Implements interface
type GreeterServicer struct{}

// GreetHandler validates input data prior to calling Greet
func (gs GreeterServicer) GreetHandler(g server.GenericRequest, b []byte) (any, error) {
	var gr GreetRequest
	if err := json.Unmarshal(b, &gr); err != nil {
		return nil, fmt.Errorf("Unmarshalling data: %w", err)
	}

	if err := validate.Check(gr); err != nil {
		return nil, fmt.Errorf("validating data: %w", err)
	}

	return gs.Greet(gr, g), nil
}

// Greet implements GreeterRpcService
func (GreeterServicer) Greet(req GreetRequest, gr server.GenericRequest) GreetResponse {
	return GreetResponse{
		Greeting: fmt.Sprintf("Hello %s, the current time is %s", req.Name, gr.Values.Now),
	}
}

// Register implements GreeterRpcService
func (gs GreeterServicer) Register(s *server.Server) {
	s.Register("GreeterService", "Greet", server.RPCEndpoint{Roles: []string{}, Handler: gs.GreetHandler})
}

// Create new GreeterServicer
func NewGreeterServicer() GreeterRpcService {
	return GreeterServicer{}
}

// GreetRequest is the request object for GreeterService.Greet.
type GreetRequest struct {
	// Name is the person to greet.
	// It is required.
	Name string
}

// GreetResponse is the response object containing a
// person's greeting.
type GreetResponse struct {
	// Greeting is a nice message welcoming somebody.
	Greeting string
}
