// Package middleware provides HTTP middleware components for authentication and authorization
package middleware

import (
	"context"
	"fmt"
	"net/http"
	redisdb "shorty/redis"
	"strings"
	"time"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

// User represents an authenticated user
type User struct {
	Email string `json:"email"`
}

// AuthenticationMiddlewareWrapper creates a mux-compatible middleware for authentication
func AuthenticationMiddlewareWrapper(rdb *redis.Client) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return AuthenticationMiddleware(next, rdb)
	}
}

// AuthenticationMiddleware validates bearer tokens and adds authenticated user to context
func AuthenticationMiddleware(next http.Handler, rdb *redis.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the Authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			rlog.Info("Authentication failed: Missing or invalid Authorization header")
			http.Error(w, "Unauthorized: Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the token and extract user information
		user, err := validateAccessToken(token)
		if err != nil {
			rlog.Error("Authentication failed", err)
			http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		// Add user information to the request context
		ctx := context.WithValue(r.Context(), UserKey, user.Email)
		r = r.WithContext(ctx)

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}

// validateAccessToken verifies an OIDC token and extracts user information
func validateAccessToken(token string) (User, error) {
	ctx := context.Background()

	// Initialize an OIDC provider
	provider, err := oidc.NewProvider(ctx, viper.GetString("OIDC_PROVIDER_URL"))
	if err != nil {
		rlog.Error("OIDC provider initialization failed", err)
		return User{}, fmt.Errorf("could not initialize OIDC provider: %w", err)
	}

	// Create a token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID:                   viper.GetString("OIDC_CLIENT_ID"),
		SkipIssuerCheck:            viper.GetBool("SKIPISSUERCHECK"),
		InsecureSkipSignatureCheck: viper.GetBool("INSECURE_SKIP_SIGNATURE_CHECK"),
	})
	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		rlog.Error("Token verification failed", err)
		if strings.Contains(err.Error(), "expired") {
			return User{}, fmt.Errorf("token has expired: %w", err)
		}
		return User{}, fmt.Errorf("invalid token: %w", err)
	}

	// Check if the token has expired
	if idToken.Expiry.Before(time.Now()) {
		return User{}, fmt.Errorf("token has expired")
	}

	// Parse user information from the token claims
	var user User
	if err := idToken.Claims(&user); err != nil {
		rlog.Error("Failed to parse user claims", err)
		return User{}, fmt.Errorf("unable to parse user claims: %w", err)
	}

	if user.Email == "" {
		return User{}, fmt.Errorf("token does not contain an email claim")
	}

	rlog.Debug("User authenticated successfully", rlog.String("email", user.Email))
	return user, nil
}

// AddAdminStatusMiddlewareWrapper creates a mux-compatible middleware for adding admin status
func AddAdminStatusMiddlewareWrapper(rdb *redis.Client) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return AddAdminStatusMiddleware(next, rdb)
	}
}

// AddAdminStatusMiddleware checks if the current user is an admin and adds this information to the request context
func AddAdminStatusMiddleware(next http.Handler, rdb *redis.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the user from the request context
		email, ok := r.Context().Value(UserKey).(string)
		isAdminUser := false

		if !ok || email == "" {
			rlog.Debug("No user email found in context")
			// Set default (non-admin) status and continue
			ctx := context.WithValue(r.Context(), IsAdminKey, isAdminUser)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		// Check if user is an admin
		isAdminUser = redisdb.AdminUserExists(rdb, email)
		rlog.Debug("Setting admin status",
			rlog.Any("isAdmin", isAdminUser),
			rlog.String("email", email))

		// Add admin status to the response header
		w.Header().Set("X-Is-Admin", fmt.Sprintf("%t", isAdminUser))

		// Add admin status to request context
		ctx := context.WithValue(r.Context(), IsAdminKey, isAdminUser)
		r = r.WithContext(ctx)

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}

// IsOwnerMiddlewareWrapper creates a mux-compatible middleware for checking resource ownership
func IsOwnerMiddlewareWrapper(rdb *redis.Client) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return IsOwnerMiddleware(next, rdb)
	}
}

// IsOwnerMiddleware checks if the current user owns the requested resource
// Only the path owner (creator) and admins can modify or delete paths
func IsOwnerMiddleware(next http.Handler, rdb *redis.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip ownership check for GET and POST requests
		method := r.Method
		if method == http.MethodGet || method == http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Retrieve the user's email from the request context
		email, ok := r.Context().Value(UserKey).(string)
		if !ok || email == "" {
			rlog.Warn("No valid user found in context")
			http.Error(w, "Unauthorized: No valid user found in context", http.StatusUnauthorized)
			return
		}

		// Get the path ID from route parameters
		params := mux.Vars(r)
		pathID := params["id"]
		if pathID == "" {
			rlog.Warn("Path ID not found in request")
			http.Error(w, "Bad Request: Path ID not found", http.StatusBadRequest)
			return
		}

		// Get the path owner from database
		pathOwner, err := redisdb.GetPathOwner(rdb, pathID)
		if err != nil {
			rlog.Error("Failed to get path owner", err,
				rlog.String("pathID", pathID),
				rlog.String("requestedBy", email))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Check if the user is the owner of the resource
		isOwner := pathOwner == email
		rlog.Debug("Ownership check",
			rlog.Any("isOwner", isOwner),
			rlog.String("pathID", pathID),
			rlog.String("requestedBy", email),
			rlog.String("owner", pathOwner))

		// Add ownership status to the request context
		ctx := context.WithValue(r.Context(), IsOwnerKey, isOwner)
		r = r.WithContext(ctx)

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}
