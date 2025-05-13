package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NorskHelsenett/shorty/internal/middleware"
	"github.com/NorskHelsenett/shorty/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func TestIsUrl(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Gyldige URL-er
		{"Valid URL with domain", args{"http://google.com"}, true},
		{"Valid URL with subdomain", args{"http://sub.google.com"}, true},
		{"Valid URL with IP", args{"http://192.168.0.1"}, true},
		{"Valid URL with IP and port", args{"http://192.168.0.1:8080"}, true},
		{"Valid URL with path", args{"http://google.com/search"}, true},
		{"Invalid IP URL", args{"http://192.168.1.300"}, true},

		// Ugyldige URL-er
		{"Invalid URL without schema", args{"google.com"}, false},
		{"Invalid URL with missing TLD", args{"http://google"}, false},
		{"Invalid URL with special characters", args{"http://!@#$%^&*()"}, false},

		{"Empty string", args{""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsURL(tt.args.str); got != tt.want {
				t.Errorf("IsUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ----- Fake implementations for redisdb functions -----

// FakeURLExists simulates redisdb.URLExists.
// Returns true when id=="exists", false when id=="notexists" and an error for id=="internal_error".
var fakeURLExists = func(rdb *redis.Client, id string) (bool, error) {
	switch id {
	case "internal_error":
		return false, fmt.Errorf("internal error")
	case "exists":
		return true, nil
	}
	return false, nil
}

// FakeGetURL simulates redisdb.GetURL.
// Returns a valid URL when id=="id1". Otherwise returns an error.
var fakeGetURL = func(rdb *redis.Client, id string) (string, error) {
	if id == "id1" {
		return "https://example.com", nil
	}
	return "", fmt.Errorf("not found")
}

// FakeDelete simulates redisdb.Delete.
// Returns success except when id=="fail".
var fakeDelete = func(rdb *redis.Client, id string) (bool, error) {
	if id == "fail" {
		return false, fmt.Errorf("delete failed")
	}
	return true, nil
}

// FakeUpdateOrCreatePath simulates redisdb.UpdateOrCreatePath.
// Returns "success" unless the provided URL is "fail".
var fakeUpdateOrCreatePath = func(rdb *redis.Client, id, urlStr, user string) (string, error) {
	if urlStr == "fail" {
		return "", fmt.Errorf("update failed")
	}
	return "success", nil
}

// types used by GetAllRedirects
type FakeRedirectPath struct {
	Path  string
	URL   string
	Owner string
}

// FakeGetAll simulates redisdb.GetAll.
// Ignores the key and returns a fixed list.
var fakeGetAll = func(rdb *redis.Client, key string) ([]models.RedirectPath, error) {
	return []models.RedirectPath{
		{Path: "p1", URL: "https://p1.com", Owner: "user1"},
		{Path: "p2", URL: "https://p2.com", Owner: "user2"},
	}, nil
}

// ----- Override the redisdb functions for testing -----
func init() {
	// Override via our package-level variables.
	URLExists = fakeURLExists
	// Although handlers call redisdb.GetURL, make sure our handlers use our variable.
	GetURL = fakeGetURL
	Delete = fakeDelete
	UpdateOrCreatePath = fakeUpdateOrCreatePath
	GetAll = fakeGetAll
}

// ----- Helpers for context values -----
func contextWithRoles(isAdmin, isOwner bool, user string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.IsAdminKey, isAdmin)
	ctx = context.WithValue(ctx, middleware.IsOwnerKey, isOwner)
	ctx = context.WithValue(ctx, middleware.UserKey, user)
	return ctx
}

// ----- Tests for CheckURL -----
func TestCheckURL(t *testing.T) {
	// No need for a real Redis client in these tests.
	var fakeRdb *redis.Client

	tests := []struct {
		name           string
		id             string
		expectedOK     bool
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "Missing key",
			id:             "",
			expectedOK:     false,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Missing key parameter",
		},
		{
			name:           "Internal error",
			id:             "internal_error",
			expectedOK:     false,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Internal server error",
		},
		{
			name:           "URL not exists",
			id:             "notexists",
			expectedOK:     false,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "URL does not exist",
		},
		{
			name:           "URL exists",
			id:             "exists",
			expectedOK:     true,
			expectedStatus: http.StatusOK,
			expectedMsg:    "",
		},
	}

	for _, tc := range tests {
		ok, status, msg := CheckURL(fakeRdb, tc.id)
		if ok != tc.expectedOK {
			t.Errorf("[%s] expected ok=%v but got %v", tc.name, tc.expectedOK, ok)
		}
		if status != tc.expectedStatus {
			t.Errorf("[%s] expected status=%d but got %d", tc.name, tc.expectedStatus, status)
		}
		if msg != tc.expectedMsg {
			t.Errorf("[%s] expected msg=%q but got %q", tc.name, tc.expectedMsg, msg)
		}
	}
}

// ----- Test for Redirect handler -----
func TestRedirect(t *testing.T) {
	var fakeRdb *redis.Client

	// Create a handler using our fake redis functions.
	handler := Redirect(fakeRdb)

	// Case 1: valid id returns URL from fakeGetURL (id "id1")
	req := httptest.NewRequest(http.MethodGet, "/id1", nil)
	// Set the mux variable "id" since our handler uses mux.Vars.
	req = mux.SetURLVars(req, map[string]string{"id": "id1"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Our fakeGetURL returns "https://example.com" for id "id1"
	if rr.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, rr.Code)
	}
	// Check Location header.
	loc := rr.Header().Get("Location")
	if loc != "https://example.com" {
		t.Errorf("Expected redirect to %s, got %s", "https://example.com", loc)
	}

	// Case 2: id not found should return default redirect ("https://nhn.no")
	req = httptest.NewRequest(http.MethodGet, "/something", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, rr.Code)
	}
	loc = rr.Header().Get("Location")
	if loc != "https://nhn.no" {
		t.Errorf("Expected default redirect to %s, got %s", "https://nhn.no", loc)
	}
}

// ----- Test for DeleteRedirect handler -----
func TestDeleteRedirect(t *testing.T) {
	var fakeRdb *redis.Client

	// We need to set URL variable "id" for mux.
	router := mux.NewRouter()
	router.HandleFunc("/admin/{id}", DeleteRedirect(fakeRdb))

	tests := []struct {
		name           string
		url            string
		ctxFunc        func(r *http.Request) *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Non-admin and non-owner forbidden",
			url:  "/admin/exists",
			ctxFunc: func(r *http.Request) *http.Request {
				return r.WithContext(contextWithRoles(false, false, "user1"))
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden",
		},
		{
			name: "Missing URL in redis -> URL not found",
			url:  "/admin/notexists",
			ctxFunc: func(r *http.Request) *http.Request {
				return r.WithContext(contextWithRoles(true, false, "user1"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "URL does not exist",
		},
		{
			name: "Delete fails",
			url:  "/admin/fail", // fakeDelete returns error when id=="fail"
			ctxFunc: func(r *http.Request) *http.Request {
				return r.WithContext(contextWithRoles(true, false, "user1"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to delete URL",
		},
		{
			name: "Successful deletion",
			url:  "/admin/exists",
			ctxFunc: func(r *http.Request) *http.Request {
				// With admin role so deletion is allowed.
				return r.WithContext(contextWithRoles(true, false, "user1"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Path deleted successfully",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, tc.url, nil)
			// Set URL variables via mux helper.
			vars := mux.Vars(req)
			// If the url is missing the email (id), then vars will be empty.
			// test cases supply a valid URL.
			if vars == nil {
				req = mux.SetURLVars(req, map[string]string{"id": strings.TrimPrefix(tc.url, "/admin/")})
			}
			req = tc.ctxFunc(req)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d but got %d", tc.expectedStatus, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tc.expectedBody) {
				t.Errorf("Expected body to contain %q but got %q", tc.expectedBody, rr.Body.String())
			}
		})
	}
}

// ----- Test for UpdateRedirect handler -----
func TestUpdateRedirect(t *testing.T) {
	var fakeRdb *redis.Client

	router := mux.NewRouter()
	router.HandleFunc("/admin/{id}", UpdateRedirect(fakeRdb))

	tests := []struct {
		name           string
		id             string
		body           interface{}
		ctxFunc        func(r *http.Request) *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Not admin or owner is forbidden",
			id:   "exists",
			body: map[string]string{"url": "https://updated.com"},
			ctxFunc: func(r *http.Request) *http.Request {
				return r.WithContext(contextWithRoles(false, false, "user1"))
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden",
		},
		{
			name: "Missing user from context returns unauthorized",
			id:   "exists",
			body: map[string]string{"url": "https://updated.com"},
			ctxFunc: func(r *http.Request) *http.Request {
				// Return a request with empty user.
				ctx := context.WithValue(r.Context(), middleware.IsAdminKey, true)
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user userEmail not found",
		},
		{
			name: "Successful update",
			id:   "exists",
			body: map[string]string{"url": "https://updated.com"},
			ctxFunc: func(r *http.Request) *http.Request {
				return r.WithContext(contextWithRoles(true, false, "user1"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Path updated successfully",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPatch, "/admin/"+tc.id, bytes.NewReader(reqBody))
			req = tc.ctxFunc(req)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tc.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tc.expectedBody, rr.Body.String())
			}
		})
	}
}

// ----- Test for AddRedirect handler -----
func TestAddRedirect(t *testing.T) {
	var fakeRdb *redis.Client

	// The AddRedirect handler is mapped to "/admin/"
	handler := AddRedirect(fakeRdb)

	tests := []struct {
		name           string
		body           interface{}
		ctxFunc        func(r *http.Request) *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Invalid JSON returns bad request",
			body: "invalid json",
			ctxFunc: func(r *http.Request) *http.Request {
				return r.WithContext(contextWithRoles(true, false, "user1"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request format",
		},
		{
			name: "Empty URL returns bad request",
			body: map[string]string{"url": "", "path": "testpath"},
			ctxFunc: func(r *http.Request) *http.Request {
				return r.WithContext(contextWithRoles(true, false, "user1"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "URL cannot be empty",
		},
		{
			name: "Invalid URL format",
			body: map[string]string{"url": "notaurl", "path": "testpath"},
			ctxFunc: func(r *http.Request) *http.Request {
				return r.WithContext(contextWithRoles(true, false, "user1"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid URL format",
		},
		{
			name: "Path already exists returns conflict",
			// For this fake, note that AddRedirect checks URLExists
			// For id equal to "exists", fakeURLExists returns true.
			body: map[string]string{"url": "https://valid.com", "path": "exists"},
			ctxFunc: func(r *http.Request) *http.Request {
				return r.WithContext(contextWithRoles(true, false, "user1"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   "Path already exists",
		},
		{
			name: "Missing user in context returns unauthorized",
			body: map[string]string{"url": "https://valid.com", "path": "newpath"},
			ctxFunc: func(r *http.Request) *http.Request {
				// Do not set user key.
				ctx := context.WithValue(r.Context(), middleware.IsAdminKey, true)
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authentication required",
		},
		{
			name: "Successful addition",
			body: map[string]string{"url": "https://valid.com", "path": "newpath"},
			ctxFunc: func(r *http.Request) *http.Request {
				return r.WithContext(contextWithRoles(true, false, "user1"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var reqBody []byte
			switch v := tc.body.(type) {
			case string:
				reqBody = []byte(v)
			default:
				var err error
				reqBody, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("Failed to marshal body: %v", err)
				}
			}
			req := httptest.NewRequest(http.MethodPost, "/admin/", bytes.NewReader(reqBody))
			req = tc.ctxFunc(req)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tc.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tc.expectedBody, rr.Body.String())
			}
		})
	}
}

// ----- Test for GetAllRedirects handler -----
func TestGetAllRedirects(t *testing.T) {
	var fakeRdb *redis.Client

	handler := GetAllRedirects(fakeRdb)

	// this test simulate a context with a user and admin status.
	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	req = req.WithContext(contextWithRoles(true, false, "user1"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d but got %d", http.StatusOK, rr.Code)
	}
	// Decode the JSON response.
	var responses []models.RedirectAllPaths
	if err := json.Unmarshal(rr.Body.Bytes(), &responses); err != nil {
		t.Errorf("Failed to unmarshal JSON response: %v", err)
	}
	// Expecting 2 elements from our fakeGetAll.
	if len(responses) != 2 {
		t.Errorf("Expected 2 redirect entries but got %d", len(responses))
	}
}

// ----- Test for IsURL helper function -----
func TestIsURL(t *testing.T) {
	validURLs := []string{
		"https://example.com",
		"http://www.example.com",
		"https://sub.domain.co",
	}
	invalidURLs := []string{
		"notaurl",
		"htp://example.com",
		"",
	}

	for _, urlStr := range validURLs {
		if !IsURL(urlStr) {
			t.Errorf("Expected %q to be a valid URL", urlStr)
		}
	}

	for _, urlStr := range invalidURLs {
		if IsURL(urlStr) {
			t.Errorf("Expected %q to be an invalid URL", urlStr)
		}
	}

	// Test an IP address host.
	ipURL := "http://127.0.0.1"
	if !IsURL(ipURL) {
		t.Errorf("Expected %q to be valid", ipURL)
	}
}
