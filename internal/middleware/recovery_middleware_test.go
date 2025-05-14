package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRecoveryMiddleware_NoPanic tests that the middleware passes the request through
// normally when there is no panic in the next handler.
func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	// Create a dummy handler that does not panic.
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("All good"))
	})

	// Wrap the dummyHandler with the RecoveryMiddleware.
	handler := RecoveryMiddleware(dummyHandler)

	// Create a test request.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	// Serve the request.
	handler.ServeHTTP(rr, req)
	res := rr.Result()

	// Verify that the response is as expected.
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	body := rr.Body.String()
	if body != "All good" {
		t.Errorf("expected body %q, got %q", "All good", body)
	}
}

// TestRecoveryMiddleware_WithPanic tests that when the next handler panics,
// the middleware recovers and returns a 500 Internal Server Error.
func TestRecoveryMiddleware_WithPanic(t *testing.T) {
	// Create a dummy handler that panics.
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	// Wrap the dummyHandler with the RecoveryMiddleware.
	handler := RecoveryMiddleware(dummyHandler)

	// Create a test request.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	// Serve the request. Panic is expected to be recovered.
	handler.ServeHTTP(rr, req)
	res := rr.Result()

	// Verify that the response status is 500.
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, res.StatusCode)
	}

	// Verify that the response body contains "Internal Server Error".
	body := rr.Body.String()
	if !strings.Contains(body, "Internal Server Error") {
		t.Errorf("expected body to contain %q, got %q", "Internal Server Error", body)
	}
}
