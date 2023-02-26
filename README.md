# seed
A RPC over HTTP server. 

## Example

cd into [examples](./examples) directory. Run `go build` . Run `./examples` .

```go
	// New RPCServer
	s := server.NewServer(mid.CommonMiddleware)

	// Register GreeterServicer
	gs := NewGreeterServicer()
	gs.Register(s)

	// Listen
	fmt.Println(`Listening on port 8080`)
	fmt.Println(`test cmd: curl -X POST  --data '{"name": "seed client"}' http://localhost:8080/v1/GreeterService.Greet`)
	http.Handle("/v1/", s)
	log.Fatal(http.ListenAndServe(":8080", nil))

```

### Routes
Routes are defined by the `RpcEndpoint`

```go
type RPCEndpoint struct {
	Roles   []string
	Handler func(GenericRequest, []byte) (any, error)
}
```

### Registering a Service
Service must implement the `RPCService` interface

```go
type RPCService interface {
	Register(s *Server)
}

// Register implements GreeterRpcService
func (gs GreeterServicer) Register(s *server.Server) {
	s.Register("GreeterService", "Greet", server.RPCEndpoint{Roles: []string{}, Handler: gs.GreetHandler})
}
```

```go
// GreetHandler validates input data prior to calling Greet
func (gs GreeterServicer) GreetHandler(r server.GenericRequest, b []byte) (any, error) {
	var gr GreetRequest
	if err := json.Unmarshal(b, &gr); err != nil {
		return nil, fmt.Errorf("Unmarshalling data: %w", err)
	}

	if err := validate.Check(gr); err != nil {
		return nil, fmt.Errorf("validating data: %w", err)
	}

	return gs.Greet(gr, r), nil
}
```

## Inspiration

##### Ardan Labs Service 
Pulled the [auth](./auth), [keystore](./keystore/), [validate](./validate/), and [values](./values/) code from [ardanlabs service repo](https://github.com/ardanlabs/service.git).

##### Oto
The rpc over http server code was pulled from [oto](https://github.com/pacedotdev/oto).

Refactor Highlights:
- The server has one http handler. 
    - Adds needed `http.Request` data from middleware into the `GenericRequest` struct.
- All Routes implement the RPCEndpoint func.
- Adds ability to add middleware.
- Adds ability to set roles per route.
