package values

/*
	https://github.com/ardanlabs/service
	Apache License Version 2.0
	Copyright (c) Ardan Labs
*/
import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// key is how request values are stored/retrieved.
const key ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

// GetValues returns the values from the context.
func GetValues(ctx context.Context) (*Values, error) {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return nil, errors.New("web value missing from context")
	}
	return v, nil
}

// SetValues sets the values from the context.
func SetValues(ctx context.Context) context.Context {
	// Set the context with the required values to
	// process the request.
	traceId := "empty"
	u, err := uuid.NewRandom()
	if err != nil {
		traceId = u.String()
	}
	v := Values{
		TraceID: traceId,
		Now:     time.Now().UTC(),
	}
	ctx = context.WithValue(ctx, key, &v)
	return ctx
}

// GetTraceID returns the trace id from the context.
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}
	return v.TraceID
}

// SetStatusCode sets the status code back into the context.
func SetStatusCode(ctx context.Context, statusCode int) error {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return errors.New("web value missing from context")
	}
	v.StatusCode = statusCode
	return nil
}
