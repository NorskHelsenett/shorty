package middleware

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

// dummyHandler is a simple next handler used to verify that the middleware passes the request along.
func dummyHandler(w http.ResponseWriter, r *http.Request) {
	// Write a message so we know the next handler was executed.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("next handler called"))
}

// TestCORSMiddleware_AllowedOrigin_GET tests that a GET request from an allowed origin receives the correct CORS headers.
func TestCORSMiddleware_AllowedOrigin_GET(t *testing.T) {
	// Define allowed origins and load them globally.
	LoadOrigins("http://allowed.com")
	// Wrap the dummyHandler with the CORSMiddleware.
	handler := CORSMiddleware(http.HandlerFunc(dummyHandler))

	// Create a GET request with the Origin header set to an allowed origin.
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Header.Set("Origin", "http://allowed.com")
	rr := httptest.NewRecorder()

	// Serve the request.
	handler.ServeHTTP(rr, req)
	res := rr.Result()

	// Check that the CORS headers are set correctly.
	if got, want := res.Header.Get("Access-Control-Allow-Origin"), "http://allowed.com"; got != want {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, want)
	}
	if got, want := res.Header.Get("Access-Control-Allow-Methods"), "GET, POST, PATCH, PUT, DELETE, OPTIONS"; got != want {
		t.Errorf("Access-Control-Allow-Methods = %q, want %q", got, want)
	}
	if got, want := res.Header.Get("Access-Control-Allow-Headers"), "Content-Type, Authorization"; got != want {
		t.Errorf("Access-Control-Allow-Headers = %q, want %q", got, want)
	}
	if got, want := res.Header.Get("Access-Control-Allow-Credentials"), "true"; got != want {
		t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, want)
	}
	// Check that the Vary header contains "Origin".
	if got := res.Header.Values("Vary"); !slices.Contains(got, "Origin") {
		t.Errorf("header Vary does not contain \"Origin\"; got %v", got)
	}
	if got, want := res.Header.Get("Access-Control-Expose-Headers"), "X-Is-Admin"; got != want {
		t.Errorf("Access-Control-Expose-Headers = %q, want %q", got, want)
	}

	// Ensure that the next handler was called by checking the response body.
	body := rr.Body.String()
	if !strings.Contains(body, "next handler called") {
		t.Errorf("expected body to contain %q, got %q", "next handler called", body)
	}
}

// TestCORSMiddleware_AllowedOrigin_OPTIONS tests that an OPTIONS (preflight) request from an allowed origin is handled correctly.
func TestCORSMiddleware_AllowedOrigin_OPTIONS(t *testing.T) {
	LoadOrigins("http://allowed.com")
	handler := CORSMiddleware(http.HandlerFunc(dummyHandler))

	// Create an OPTIONS request simulating a preflight request.
	req := httptest.NewRequest(http.MethodOptions, "http://example.com/test", nil)
	req.Header.Set("Origin", "http://allowed.com")
	rr := httptest.NewRecorder()

	// Serve the request.
	handler.ServeHTTP(rr, req)
	res := rr.Result()

	// For preflight requests, the middleware should respond with StatusOK.
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	// Check that the Access-Control-Allow-Origin header is set properly.
	if got, want := res.Header.Get("Access-Control-Allow-Origin"), "http://allowed.com"; got != want {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, want)
	}

	// The body should be empty because the middleware handles the OPTIONS request and does not call the next handler.
	if body := rr.Body.String(); body != "" {
		t.Errorf("expected empty body for OPTIONS request, got %q", body)
	}
}

// TestCORSMiddleware_UnauthorizedOrigin tests that requests from a disallowed origin do not receive CORS headers.
func TestCORSMiddleware_UnauthorizedOrigin(t *testing.T) {
	LoadOrigins("http://allowed.com")
	handler := CORSMiddleware(http.HandlerFunc(dummyHandler))

	// Create a GET request with an Origin header that is not allowed.
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Header.Set("Origin", "http://disallowed.com")
	rr := httptest.NewRecorder()

	// Serve the request.
	handler.ServeHTTP(rr, req)
	res := rr.Result()

	// Expect that no Access-Control-Allow-Origin header is set.
	if origin := res.Header.Get("Access-Control-Allow-Origin"); origin != "" {
		t.Errorf("expected no Access-Control-Allow-Origin header, got %q", origin)
	}
	// Verify that the next handler was still called.
	body := rr.Body.String()
	if !strings.Contains(body, "next handler called") {
		t.Errorf("expected body to contain %q, got %q", "next handler called", body)
	}
}

// TestCORSMiddleware_NoOriginHeader tests that requests without an Origin header do not have CORS headers added.
func TestCORSMiddleware_NoOriginHeader(t *testing.T) {
	LoadOrigins("http://allowed.com")
	handler := CORSMiddleware(http.HandlerFunc(dummyHandler))

	// Create a GET request without an Origin header.
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	rr := httptest.NewRecorder()

	// Serve the request.
	handler.ServeHTTP(rr, req)
	res := rr.Result()

	// Expect that no CORS headers are set if the Origin header is missing.
	if origin := res.Header.Get("Access-Control-Allow-Origin"); origin != "" {
		t.Errorf("expected no Access-Control-Allow-Origin header, got %q", origin)
	}
	// Verify that the next handler was called.
	body := rr.Body.String()
	if !strings.Contains(body, "next handler called") {
		t.Errorf("expected body to contain %q, got %q", "next handler called", body)
	}
}

// TestCORSMiddleware_NoAllowedOrigins tests that when no origins are configured, no CORS headers are set.
func TestCORSMiddleware_NoAllowedOrigins(t *testing.T) {
	LoadOrigins("") // Empty string - no origins configured
	handler := CORSMiddleware(http.HandlerFunc(dummyHandler))

	// Create a GET request with an Origin header.
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	// Serve the request.
	handler.ServeHTTP(rr, req)
	res := rr.Result()

	// Expect that no CORS headers are set when no origins are configured.
	if origin := res.Header.Get("Access-Control-Allow-Origin"); origin != "" {
		t.Errorf("expected no Access-Control-Allow-Origin header, got %q", origin)
	}
	// Verify that the next handler was called.
	body := rr.Body.String()
	if !strings.Contains(body, "next handler called") {
		t.Errorf("expected body to contain %q, got %q", "next handler called", body)
	}
}
