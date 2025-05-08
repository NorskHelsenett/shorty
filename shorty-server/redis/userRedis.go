package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/go-redis/redis/v8"
)

var (
	// ErrUserNotFound is returned when a user is not found in the database
	ErrUserNotFound = errors.New("user not found")
	// ErrNoUsersFound is returned when no users are found in the database
	ErrNoUsersFound = errors.New("no users found")
	// ErrUserExists is returned when trying to create a user that already exists
	ErrUserExists = errors.New("user already exists")
	// ErrEmailNotFound is returned when an email is not found in the database
	ErrEmailNotFound = errors.New("email not found")
)

// AddAdminUser creates a new admin user in Redis
// Returns "created" if successful, "exists" if the user already exists, or an error
func AddAdminUser(rdb *redis.Client, userID string, email string) (string, error) {
	ctx := context.Background()

	// Generate Keys (user and email mapping)
	key := "user:" + userID
	emailKey := "email:" + email

	// Check if the email already exists in db
	existingUserEmail, err := rdb.Get(ctx, emailKey).Result()
	if err != nil && err != redis.Nil {
		return "", fmt.Errorf("failed to check if email exists: %w", err)
	}
	if existingUserEmail != "" {
		rlog.Info("Admin user already exists", rlog.String("email", email))
		return "exists", ErrUserExists
	}

	// Create user as Redis-hash
	if err := rdb.HSet(ctx, key, map[string]interface{}{
		"id":    userID,
		"email": email,
		"admin": true,
	}).Err(); err != nil {
		return "", fmt.Errorf("failed to create user hash: %w", err)
	}

	// Save email-to-userID mapping
	if err := rdb.Set(ctx, emailKey, userID, 0).Err(); err != nil {
		// Cleanup the user hash if email mapping fails
		rdb.Del(ctx, key)
		return "", fmt.Errorf("failed to create email mapping: %w", err)
	}

	// Add userID to "users" set
	if err := rdb.SAdd(ctx, "users", userID).Err(); err != nil {
		// Attempt to clean up if adding to set fails
		rdb.Del(ctx, key, emailKey)
		return "", fmt.Errorf("failed to add user to users set: %w", err)
	}

	rlog.Info("User created successfully", rlog.String("userID", userID), rlog.String("email", email))
	return "created", nil
}

// GetUserByEmail retrieves user data by email
// Returns the user's data as a map or an error if the user doesn't exist
func GetUserByEmail(rdb *redis.Client, email string) (map[string]string, error) {
	ctx := context.Background()
	emailKey := "email:" + email

	// Retrieve the userID associated with the email
	userID, err := rdb.Get(ctx, emailKey).Result()
	if err == redis.Nil {
		rlog.Info("Email not found in database", rlog.String("email", email))
		return nil, ErrEmailNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to retrieve userID for email: %w", err)
	}

	// Retrieve the user data using the associated userID
	userData, err := GetUser(rdb, userID)
	if err != nil {
		rlog.Error("User not found for email", err, rlog.String("email", email), rlog.String("userID", userID))
		return nil, fmt.Errorf("failed to retrieve user data: %w", err)
	}

	return userData, nil
}

// AdminUserExists checks if a user with the given email exists and has admin privileges
// Returns true if the user exists and is an admin, false otherwise
func AdminUserExists(rdb *redis.Client, email string) bool {
	ctx := context.Background()
	emailKey := "email:" + email

	// Retrieve the userID associated with the email
	userID, err := rdb.Get(ctx, emailKey).Result()
	if err == redis.Nil {
		rlog.Debug("Email not found in database", rlog.String("email", email))
		return false
	} else if err != nil {
		rlog.Error("Error checking admin status", err, rlog.String("email", email))
		return false
	}

	rlog.Debug("Admin status check",
		rlog.String("email", email),
		rlog.String("userID", userID))

	return true
}

// GetUser retrieves a user by ID from Redis
// Returns the user data as a map or an error if the user doesn't exist
func GetUser(rdb *redis.Client, userID string) (map[string]string, error) {
	ctx := context.Background()
	key := "user:" + userID

	// Fetch all hash fields associated with the user
	userData, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user data: %w", err)
	}

	// Check if the user data is empty
	if len(userData) == 0 {
		return nil, ErrUserNotFound
	}

	return userData, nil
}

// GetAllAdminEmails retrieves emails of all admin users
// Returns a slice of email strings or an error if no users are found
func GetAllAdminEmails(rdb *redis.Client) ([]string, error) {
	ctx := context.Background()

	// Get all userIDs from the "users" set
	userIDs, err := rdb.SMembers(ctx, "users").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user IDs: %w", err)
	}

	// Check if there are no userIDs
	if len(userIDs) == 0 {
		return nil, ErrNoUsersFound
	}

	// Pre-allocate email slice with capacity based on userIDs
	emails := make([]string, 0, len(userIDs))

	// Iterate through all userIDs and fetch their data
	for _, userID := range userIDs {
		userData, err := GetUser(rdb, userID)
		if err != nil {
			rlog.Warn("Skipping user due to error",
				rlog.String("userID", userID),
				rlog.String("error", err.Error()))
			continue
		}

		email, exists := userData["email"]
		if exists {
			emails = append(emails, email)
		}
	}

	if len(emails) == 0 {
		return nil, ErrNoUsersFound
	}

	return emails, nil
}

// DeleteUser removes a user by email from Redis
// Returns nil on success or an error if the user doesn't exist or deletion fails
func DeleteUser(rdb *redis.Client, email string) error {
	ctx := context.Background()
	rlog.Info("Deleting user", rlog.String("email", email))

	emailKey := "email:" + email

	// Retrieve the userID associated with the email
	userID, err := rdb.Get(ctx, emailKey).Result()
	if err == redis.Nil {
		rlog.Info("User not found for deletion", rlog.String("email", email))
		return ErrEmailNotFound
	} else if err != nil {
		return fmt.Errorf("failed to retrieve userID for email: %w", err)
	}

	userKey := "user:" + userID

	// Start a Redis transaction with MULTI/EXEC
	pipe := rdb.Pipeline()

	// Delete user hash
	pipe.Del(ctx, userKey)

	// Delete email-to-userID mapping
	pipe.Del(ctx, emailKey)

	// Remove userID from "users" set
	pipe.SRem(ctx, "users", userID)

	// Execute all commands in the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rlog.Info("User successfully deleted", rlog.String("email", email), rlog.String("userID", userID))
	return nil
}
