package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/NorskHelsenett/ror/pkg/rlog"
)

var allowedOrigins []string

func LoadOrigins(allowedOriginsString string) {
	if allowedOriginsString != "" {
		allowedOrigins = strings.Split(allowedOriginsString, ";")
		rlog.Info("CORS allowed origins loaded", rlog.Any("origins", allowedOrigins))
	} else {
		rlog.Warn("No CORS origins configured - ALLOW_ORIGINS environment variable is empty")
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) headers
// It validates the Origin header against a list of allowed origins and
// sets appropriate CORS headers for the response
func CORSMiddleware(next http.Handler) http.Handler {
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
			rlog.Warn("CORS request from unauthorized origin",
				rlog.String("origin", origin),
				rlog.Any("allowedOrigins", allowedOrigins))
		}

		// Process the request
		next.ServeHTTP(w, r)
	})
}
