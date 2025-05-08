package middleware

import (
	"fmt"
	"net/http"

	"github.com/NorskHelsenett/ror/pkg/rlog"
)

// RecoveryMiddleware catches any runtime panic in HTTP handlers and
// return a 500 Internal Error to the client instead of crashing the application
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				rlog.Error("Panic recovered", fmt.Errorf("%v", err))
				// Sets HTTP status code to 500 (Internal Server Error)
				w.WriteHeader(http.StatusInternalServerError)
				// Writes error message to response body
				_, _ = w.Write([]byte("Internal Server Error"))
			}
		}()
		next.ServeHTTP(w, r) // Calls the next handler in the chain
	})
}
