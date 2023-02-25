package server

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gitamped/seed/auth"
	"github.com/gitamped/seed/mid"
	"github.com/gitamped/seed/validate"
	"github.com/gitamped/seed/values"
	"github.com/pkg/errors"
)

type RPCEndpoint struct {
	Roles   []string
	Handler func(GenericRequest, []byte) (any, error)
}

type RPCService interface {
	Register(s *Server)
}

type GenericRequest struct {
	// request context
	Ctx context.Context
	// claims for the request
	Claims auth.Claims
	// values of request
	Values *values.Values
}

// Register adds a handler for the specified service method.
func (s *Server) Register(service, path string, r RPCEndpoint) {
	s.Routes[fmt.Sprintf("%s%s.%s", s.Basepath, service, path)] = r
}

func NewServer(mw []mid.Middleware) *Server {
	s := &Server{
		Basepath: "/v1/",
		Routes:   make(map[string]RPCEndpoint),
		OnErr: func(w http.ResponseWriter, r *http.Request, err error) {
			errObj := struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			}
			if err := Encode(w, r, http.StatusInternalServerError, errObj); err != nil {
				log.Printf("failed to encode error: %s\n", err)
			}
		},

		NotFound: http.NotFoundHandler(),
	}

	s.Handler = mid.MultipleMiddleware(s.DefaultHandler, mw...)
	return s
}

type Server struct {
	// Basepath is the path prefix to match.
	// Default: /v1/
	Basepath string
	// the path and the rpc procedure to call
	Routes map[string]RPCEndpoint
	// The web handler. Default is
	Handler http.HandlerFunc
	// NotFound is the http.Handler to use when a resource is
	// not found.
	NotFound http.Handler
	// OnErr is called when there is an error.
	OnErr func(w http.ResponseWriter, r *http.Request, err error)
}

// ServeHTTP serves the request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Handler(w, r)
}

// Decode unmarshals the object in the request into v.
func Decode(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errors.Wrap(err, "decode json")
	}
	return nil
}

// Encode writes the response.
func Encode(w http.ResponseWriter, r *http.Request, status int, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "encode json")
	}
	var out io.Writer = w
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gzw := gzip.NewWriter(w)
		out = gzw
		defer gzw.Close()
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := out.Write(b); err != nil {
		return err
	}
	return nil
}

func Authorized(rpcRoles, userRoles []string) bool {
	// public endpoint if rpcRoles are not set
	if len(rpcRoles) == 0 {
		return true
	}
	for _, want := range rpcRoles {
		for _, has := range userRoles {
			if want == has {
				return true
			}
		}
	}
	return false
}

func Unauthorized(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
}

func StatusNotAcceptable(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "406 Not Acceptable", http.StatusNotAcceptable)
}

func (s *Server) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	g := GenericRequest{}

	ctx := r.Context()
	rpc, ok := s.Routes[r.URL.Path]

	if !ok {
		s.NotFound.ServeHTTP(w, r)
		return
	}

	if len(rpc.Roles) > 0 {
		claims, err := auth.GetClaims(ctx)
		if err != nil {
			Unauthorized(w, r)
			return
		}
		g.Claims = claims
	} else {
		g.Claims = auth.Claims{}
	}

	if ok := Authorized(rpc.Roles, g.Claims.Roles); !ok {
		Unauthorized(w, r)
		return
	}

	v, err := values.GetValues(ctx)
	if err != nil {
		StatusNotAcceptable(w, r)
		return
	}
	g.Values = v

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		StatusNotAcceptable(w, r)
		return
	}

	response, err := rpc.Handler(g, b)

	if err != nil {
		s.OnErr(w, r, err)
		return
	}

	if err := validate.Check(response); err != nil {
		s.OnErr(w, r, err)
		return
	}

	if err := Encode(w, r, http.StatusOK, response); err != nil {
		s.OnErr(w, r, err)
		return
	}
}
