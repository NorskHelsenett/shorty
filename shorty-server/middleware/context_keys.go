// Package middleware provides HTTP middleware functions for the application
package middleware

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Context keys used in the application
const (
	// IsAdminKey represents whether the current user has admin privileges
	IsAdminKey contextKey = "isAdmin"

	// IsOwnerKey represents whether the current user owns the resource
	IsOwnerKey contextKey = "isOwner"

	// UserKey stores the authenticated user's email
	UserKey contextKey = "authenticatedUserEmail"
)
