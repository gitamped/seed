package tests

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gitamped/seed/auth"
	"github.com/gitamped/seed/server"
	"github.com/gitamped/seed/validate"
)

// GreeterService is a polite API for greeting people.
type GreeterService interface {
	// SecretGreet requires user to authenticated and authorized before sending lovely greeting.
	SecretGreet(SecretGreetRequest, server.GenericRequest) SecretGreetResponse
}

// Required to register endpoints with the Server
type GreeterRpcService interface {
	GreeterService
	// Registers RPCService with Server
	Register(s *server.Server)
}

// Implements interface
type GreeterServicer struct{}

// SecretGreetHandler validates input data prior to calling SecretGreet
func (gs GreeterServicer) SecretGreetHandler(g server.GenericRequest, b []byte) (any, error) {
	var gr SecretGreetRequest
	if err := json.Unmarshal(b, &gr); err != nil {
		return SecretGreetResponse{Error: "Invalid GreetRequest data."}, nil
	}

	if err := validate.Check(gr); err != nil {
		return SecretGreetResponse{Error: fmt.Errorf("validating data: %w", err).Error()}, nil
	}

	return gs.SecretGreet(gr, g), nil
}

// SecretGreet implements GreeterRpcService
func (GreeterServicer) SecretGreet(req SecretGreetRequest, gr server.GenericRequest) SecretGreetResponse {
	return SecretGreetResponse{
		SecretGreeting: fmt.Sprintf("Hello %s, meet at %s", req.Alias, gr.Values.Now.Add(time.Hour*2)),
	}
}

// Register implements GreeterRpcService
func (gs GreeterServicer) Register(s *server.Server) {
	s.Register("GreeterService", "SecretGreet", server.RPCEndpoint{Roles: []string{auth.RoleUser}, Handler: gs.SecretGreetHandler})
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
	// Error message if request was not successful
	Error string
}

// SecretGreetRequest is the request object for GreeterService.SecretGreet.
type SecretGreetRequest struct {
	// Alias is the person to greet.
	// It is required.
	Alias string `json:"alias" validate:"gte=1"`
}

// SecretGreetResponse is the response object containing a
// person's greeting.
type SecretGreetResponse struct {
	// SecretGreeting is a secret message.
	SecretGreeting string
	// Error message if request was not successful
	Error string
}
