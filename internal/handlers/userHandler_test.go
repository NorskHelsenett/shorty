package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NorskHelsenett/shorty/internal/middleware"
	"github.com/NorskHelsenett/shorty/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// --- Fake implementations for redisdb functions ---

var (
	// FakeAddAdminUser simulates adding an admin user.
	FakeAddAdminUser = func(rdb *redis.Client, userID, email string) (string, error) {
		// if email is "exists@example.com", simulate user already exists.
		if email == "exists@example.com" {
			return "exists", nil
		}
		// otherwise simulate a success status string (e.g., "created").
		return "created", nil
	}

	// FakeGetAllAdminEmails simulates retrieval of admin emails.
	FakeGetAllAdminEmails = func(rdb *redis.Client) ([]string, error) {
		// return two dummy email strings
		return []string{
			"a@example.com",
			"b@example.com",
		}, nil
	}

	// FakeDeleteUser simulates the deletion of an admin user.
	FakeDeleteUser = func(rdb *redis.Client, email string) error {
		if email == "notfound@example.com" {
			return errors.New("email not found")
		}
		return nil
	}

	// FakeAdminUserExists simulates the existence check for an admin user.
	FakeAdminUserExists = func(rdb *redis.Client, email string) bool {
		// for testing, if the email equals "exists@example.com", return true.
		return email == "exists@example.com"
	}
)

func init() {
	AddAdminUser = FakeAddAdminUser
	GetAllAdminEmails = FakeGetAllAdminEmails
	DeleteUser = FakeDeleteUser
	AdminUserExists = FakeAdminUserExists
}

// --- Helper: dummy context with admin flag ---
func contextWithAdmin(isAdmin bool) context.Context {
	return context.WithValue(context.Background(), middleware.IsAdminKey, isAdmin)
}

// --- Test for AddUserRedirect ---
func TestAddUserRedirect(t *testing.T) {
	// Create a dummy redis.Client.
	var fakeRdb *redis.Client // not used by fake functions

	// Prepare a valid RedirectUser object.
	validUser := models.RedirectUser{
		Email: "valid@example.com",
	}
	// Prepare an object with an already existing email.
	existingUser := models.RedirectUser{
		Email: "exists@example.com",
	}
	// Prepare a struct with empty email.
	emptyEmail := models.RedirectUser{
		Email: "",
	}

	tests := []struct {
		name           string
		isAdmin        bool
		body           interface{}
		expectedStatus int
		expectedBody   string // substring expected in response
	}{
		{
			name:           "Non-admin gets forbidden",
			isAdmin:        false,
			body:           validUser,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden",
		},
		{
			name:           "Invalid JSON in body",
			isAdmin:        true,
			body:           "this is not JSON",
			expectedStatus: http.StatusInternalServerError, // due to ReadAll not failing but unmarshal fails
			expectedBody:   "Impossible to unmarshal body of request",
		},
		{
			name:           "Empty email returns bad request",
			isAdmin:        true,
			body:           emptyEmail,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Email is required",
		},
		{
			name:           "Existing user returns conflict",
			isAdmin:        true,
			body:           existingUser,
			expectedStatus: http.StatusConflict,
			expectedBody:   "already exists",
		},
		{
			name:           "Valid user returns OK",
			isAdmin:        true,
			body:           validUser,
			expectedStatus: http.StatusOK,
			expectedBody:   "created",
		},
	}

	handler := AddUserRedirect(fakeRdb)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			// If body is a string, use as is, otherwise marshal.
			switch v := tc.body.(type) {
			case string:
				reqBody = []byte(v)
			default:
				reqBody, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/admin/user", bytes.NewReader(reqBody))
			req = req.WithContext(contextWithAdmin(tc.isAdmin))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()
			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyStr := string(bodyBytes)

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d; got %d", tc.expectedStatus, resp.StatusCode)
			}
			if !bytes.Contains(bodyBytes, []byte(tc.expectedBody)) {
				t.Errorf("expected response body to contain %q; got %q", tc.expectedBody, bodyStr)
			}
		})
	}
}

// --- Test for GetAllUsersRedirect ---
func TestGetAllUsersRedirect(t *testing.T) {
	var fakeRdb *redis.Client

	handler := GetAllUsersRedirect(fakeRdb)

	tests := []struct {
		name           string
		isAdmin        bool
		expectedStatus int
		expectedEmails int // number of emails expected in JSON array
	}{
		{
			name:           "Non-admin gets forbidden",
			isAdmin:        false,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Admin gets list of email redirects successfully",
			isAdmin:        true,
			expectedStatus: http.StatusOK,
			expectedEmails: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin/user", nil)
			req = req.WithContext(contextWithAdmin(tc.isAdmin))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)
			resp := rr.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d; got %d", tc.expectedStatus, resp.StatusCode)
			}

			if tc.isAdmin {
				var response []models.RedirectUser
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode JSON response: %v", err)
				}
				if len(response) != tc.expectedEmails {
					t.Errorf("expected %d emails; got %d", tc.expectedEmails, len(response))
				}
			}
		})
	}
}

// --- Test for DeleteUserRedirect ---
func TestDeleteUserRedirect(t *testing.T) {
	var fakeRdb *redis.Client

	// We need a router to set URL variables (mux.Vars).
	router := mux.NewRouter()
	router.HandleFunc("/admin/user/{id}", DeleteUserRedirect(fakeRdb))

	tests := []struct {
		name           string
		isAdmin        bool
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Missing admin status returns unauthorized",
			isAdmin:        false,
			url:            "/admin/user/test@example.com",
			expectedStatus: http.StatusForbidden, // in our code, non-admin gets forbidden
			expectedBody:   "Forbidden",
		},
		{
			name:           "Missing email parameter returns bad request",
			isAdmin:        true,
			url:            "/admin/user/", // no id provided
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Email is required in URL",
		},
		{
			name:           "Email not found returns 404",
			isAdmin:        true,
			url:            "/admin/user/notfound@example.com",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Email not found",
		},
		{
			name:           "Valid delete returns OK",
			isAdmin:        true,
			url:            "/admin/user/valid@example.com",
			expectedStatus: http.StatusOK,
			expectedBody:   "User deleted successfully",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, tc.url, nil)
			req = req.WithContext(contextWithAdmin(tc.isAdmin))
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)
			resp := rr.Result()
			defer resp.Body.Close()
			bodyBytes, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d; got %d", tc.expectedStatus, resp.StatusCode)
			}
			if !bytes.Contains(bodyBytes, []byte(tc.expectedBody)) {
				t.Errorf("expected response body to contain %q; got %q", tc.expectedBody, string(bodyBytes))
			}
		})
	}
}

// --- Test for CheckUserEmailRedirect ---
func TestCheckUserEmailRedirect(t *testing.T) {
	var fakeRdb *redis.Client

	handler := CheckUserEmailRedirect(fakeRdb)

	// Create a request body with valid email and one with empty email.
	validCheck := models.RedirectUser{
		Email: "exists@example.com",
	}
	emptyCheck := models.RedirectUser{
		Email: "",
	}

	tests := []struct {
		name           string
		isAdmin        bool
		body           interface{}
		expectedStatus int
		expectedBody   string // substring expected in response
	}{
		{
			name:           "Non-admin gets forbidden",
			isAdmin:        false,
			body:           validCheck,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden",
		},
		{
			name:           "Empty email returns bad request",
			isAdmin:        true,
			body:           emptyCheck,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Email is required",
		},
		{
			name:           "Valid check returns email exists true",
			isAdmin:        true,
			body:           validCheck,
			expectedStatus: http.StatusOK,
			expectedBody:   `"Exists":true`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.body)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/admin/user/check", bytes.NewReader(reqBody))
			req = req.WithContext(contextWithAdmin(tc.isAdmin))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)
			resp := rr.Result()
			defer resp.Body.Close()
			bodyBytes, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d; got %d", tc.expectedStatus, resp.StatusCode)
			}

			if !bytes.Contains(bodyBytes, []byte(tc.expectedBody)) {
				t.Errorf("expected response body to contain %q; got %q", tc.expectedBody, string(bodyBytes))
			}
		})
	}
}
