package mid

import (
	"net/http"
	"strings"

	"github.com/gitamped/seed/auth"
)

func AuthMiddleware(a *auth.Auth) Middleware {
	m := func(h http.HandlerFunc) http.HandlerFunc {
		handler := func(w http.ResponseWriter, r *http.Request) {
			// Expecting: bearer <token>
			authStr := r.Header.Get("authorization")

			// Parse the authorization header.
			parts := strings.Split(authStr, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				h.ServeHTTP(w, r)
			} else {
				// Validate the token is signed by us.
				claims, err := a.ValidateToken(parts[1])
				if err == nil {
					// Add claims to the context, so they can be retrieved later.
					ctx := auth.SetClaims(r.Context(), claims)
					r = r.WithContext(ctx)
				}
				h.ServeHTTP(w, r)
			}
		}
		return handler
	}
	return m
}
