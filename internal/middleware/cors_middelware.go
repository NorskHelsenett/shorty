package middleware

import (
	"net/http"
	"slices"

	"github.com/NorskHelsenett/ror/pkg/rlog"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) headers
// It validates the Origin header against a list of allowed origins and
// sets appropriate CORS headers for the response
func CORSMiddleware(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if the origin is in the allowed list
		if origin != "" && slices.Contains(allowedOrigins, origin) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Expose-Headers", "X-Is-Admin")

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			rlog.Debug("CORS headers set for origin", rlog.String("origin", origin))
		} else if origin != "" {
			rlog.Debug("CORS request from unauthorized origin", rlog.String("origin", origin))
		}

		// Process the request
		next.ServeHTTP(w, r)
	})
}
