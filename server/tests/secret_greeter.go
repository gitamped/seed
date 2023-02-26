package tests

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gitamped/seed/auth"
	"github.com/gitamped/seed/server"
	"github.com/gitamped/seed/validate"
)

// SecretGreeterService is a polite API for greeting people.
type SecretGreeterService interface {
	// SecretGreet requires user to authenticated and authorized before sending lovely greeting.
	SecretGreet(SecretGreetRequest, server.GenericRequest) SecretGreetResponse
}

// Required to register endpoints with the Server
type SecretGreeterRpcService interface {
	SecretGreeterService
	// Registers RPCService with Server
	Register(s *server.Server)
}

// Implements interface
type SecretGreeterServicer struct{}

// SecretGreetHandler validates input data prior to calling SecretGreet
func (gs SecretGreeterServicer) SecretGreetHandler(g server.GenericRequest, b []byte) (any, error) {
	var gr SecretGreetRequest
	if err := json.Unmarshal(b, &gr); err != nil {
		return SecretGreetResponse{Error: "Invalid GreetRequest data."}, nil
	}

	if err := validate.Check(gr); err != nil {
		return SecretGreetResponse{Error: fmt.Errorf("validating data: %w", err).Error()}, nil
	}

	return gs.SecretGreet(gr, g), nil
}

// SecretGreet implements SecretGreeterRpcService
func (SecretGreeterServicer) SecretGreet(req SecretGreetRequest, gr server.GenericRequest) SecretGreetResponse {
	return SecretGreetResponse{
		SecretGreeting: fmt.Sprintf("Hello %s, meet at %s", req.Alias, gr.Values.Now.Add(time.Hour*2)),
	}
}

// Register implements SecretGreeterRpcService
func (gs SecretGreeterServicer) Register(s *server.Server) {
	s.Register("GreeterService", "SecretGreet", server.RPCEndpoint{Roles: []string{auth.RoleUser}, Handler: gs.SecretGreetHandler})
}

// Create new SecretGreeterServicer
func NewSecretGreeterServicer() SecretGreeterRpcService {
	return SecretGreeterServicer{}
}

// SecretGreetRequest is the request object for SecretGreeterService.SecretGreet.
type SecretGreetRequest struct {
	// Alias is the person to greet.
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
