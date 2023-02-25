package mid

import (
	"net/http"

	"github.com/gitamped/seed/values"
)

func ValuesMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := values.SetValues(r.Context())
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}
