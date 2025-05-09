package redis

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-redis/redismock/v8"
)

func TestAddAdminUser(t *testing.T) {
	// Use a background context for testing, and create a Redis mock client.

	db, mock := redismock.NewClientMock()

	userID := "user123"
	email := "user@example.com"
	userKey := "user:" + userID
	emailKey := "email:" + email

	// ----------------------------
	// Case 1: Email already exists
	t.Run("Email already exists", func(t *testing.T) {
		// Expect GET on the emailKey to return an existing user.
		mock.ExpectGet(emailKey).SetVal("existingUser")

		msg, err := AddAdminUser(db, userID, email)
		if msg != "exists" {
			t.Errorf("expected msg 'exists'; got %q", msg)
		}
		if !errors.Is(err, ErrUserExists) {
			t.Errorf("expected error %q; got %v", ErrUserExists, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	// Reset mock for next tests.
	mock.ClearExpect()
	// ----------------------------
	// Case 2: Successful creation
	t.Run("Successful creation", func(t *testing.T) {
		// Expect GET to return redis.Nil (email does not exist).
		mock.ExpectGet(emailKey).RedisNil()
		// Expect HSet to succeed.
		mock.ExpectHSet(userKey, map[string]interface{}{
			"id":    userID,
			"email": email,
			"admin": true,
		}).SetVal(1)
		// Expect Set (for email mapping) to succeed.
		mock.ExpectSet(emailKey, userID, 0).SetVal("OK")
		// Expect SAdd to succeed.
		mock.ExpectSAdd("users", userID).SetVal(1)

		msg, err := AddAdminUser(db, userID, email)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg != "created" {
			t.Errorf("expected message 'created'; got %q", msg)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	mock.ClearExpect()
	// ----------------------------
	// Case 3: Error during creation of the user hash (HSet failure)
	t.Run("Error on HSet", func(t *testing.T) {
		mock.ExpectGet(emailKey).RedisNil()
		mock.ExpectHSet(userKey, map[string]interface{}{
			"id":    userID,
			"email": email,
			"admin": true,
		}).SetErr(errors.New("hset failure"))

		_, err := AddAdminUser(db, userID, email)
		if err == nil || !strings.Contains(err.Error(), "failed to create user hash") {
			t.Errorf("expected error on HSet, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	mock.ClearExpect()
	// ----------------------------
	// Case 4: Error during email mapping (Set failure) triggers deletion of user hash
	t.Run("Error on Set (email mapping)", func(t *testing.T) {
		mock.ExpectGet(emailKey).RedisNil()
		mock.ExpectHSet(userKey, map[string]interface{}{
			"id":    userID,
			"email": email,
			"admin": true,
		}).SetVal(1)
		mock.ExpectSet(emailKey, userID, 0).SetErr(errors.New("set failure"))
		// Expect Del to be called for the user hash cleanup.
		mock.ExpectDel(userKey).SetVal(1)

		_, err := AddAdminUser(db, userID, email)
		if err == nil || !strings.Contains(err.Error(), "failed to create email mapping") {
			t.Errorf("expected email mapping error; got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	mock.ClearExpect()
	// ----------------------------
	// Case 5: Error while adding to the users set (SAdd failure) triggers deletion of user hash and email mapping
	t.Run("Error on SAdd", func(t *testing.T) {
		mock.ExpectGet(emailKey).RedisNil()
		mock.ExpectHSet(userKey, map[string]interface{}{
			"id":    userID,
			"email": email,
			"admin": true,
		}).SetVal(1)
		mock.ExpectSet(emailKey, userID, 0).SetVal("OK")
		mock.ExpectSAdd("users", userID).SetErr(errors.New("sadd failure"))
		// Expect Del to be called for cleanup of both user hash and email mapping.
		mock.ExpectDel(userKey, emailKey).SetVal(1)

		_, err := AddAdminUser(db, userID, email)
		if err == nil || !strings.Contains(err.Error(), "failed to add user to users set") {
			t.Errorf("expected SAdd error; got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}

func TestGetUserByEmail(t *testing.T) {

	// Create the mock client.
	db, mock := redismock.NewClientMock()

	email := "user@example.com"
	emailKey := "email:" + email
	userID := "12345"

	t.Run("Email not found", func(t *testing.T) {
		mock.ExpectGet(emailKey).RedisNil()

		_, err := GetUserByEmail(db, email)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrEmailNotFound {
			t.Errorf("expected error %q, got %q", ErrEmailNotFound.Error(), err.Error())
		}
	})

	t.Run("GetUser returns error", func(t *testing.T) {
		// Reset expectations for a new test.
		mock.ExpectGet(emailKey).SetVal(userID)
		// For GetUser, simulate error on HGetAll.
		userKey := "user:" + userID
		mock.ExpectHGetAll(userKey).SetErr(errors.New("hgetall error"))

		_, err := GetUserByEmail(db, email)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to retrieve user data") {
			t.Errorf("unexpected error: %q", err.Error())
		}
	})

	t.Run("Successful retrieval", func(t *testing.T) {
		// Reset expectations.
		mock.ExpectGet(emailKey).SetVal(userID)
		userKey := "user:" + userID
		userData := map[string]string{
			"email": "user@example.com",
			"name":  "Test User",
		}
		mock.ExpectHGetAll(userKey).SetVal(userData)

		data, err := GetUserByEmail(db, email)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if data["email"] != "user@example.com" {
			t.Errorf("expected email %q, got %q", "user@example.com", data["email"])
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %s", err)
	}
}

func TestAdminUserExists(t *testing.T) {
	db, mock := redismock.NewClientMock()
	email := "admin@example.com"
	emailKey := "email:" + email
	userID := "adminid"

	t.Run("User not found", func(t *testing.T) {
		mock.ExpectGet(emailKey).RedisNil()
		if AdminUserExists(db, email) {
			t.Errorf("expected false; got true")
		}
	})

	t.Run("Error retrieving user", func(t *testing.T) {
		mock.ExpectGet(emailKey).SetErr(errors.New("get error"))
		if AdminUserExists(db, email) {
			t.Errorf("expected false on error; got true")
		}
	})

	t.Run("User exists", func(t *testing.T) {
		mock.ExpectGet(emailKey).SetVal(userID)
		if !AdminUserExists(db, email) {
			t.Errorf("expected true; got false")
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %s", err)
	}
}

func TestGetUser(t *testing.T) {
	db, mock := redismock.NewClientMock()
	userID := "12345"
	userKey := "user:" + userID

	t.Run("User not found", func(t *testing.T) {
		// HGetAll returns an empty map.
		mock.ExpectHGetAll(userKey).SetVal(map[string]string{})
		_, err := GetUser(db, userID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrUserNotFound {
			t.Errorf("expected error %q, got %q", ErrUserNotFound.Error(), err.Error())
		}
	})

	t.Run("Error retrieving user", func(t *testing.T) {
		mock.ExpectHGetAll(userKey).SetErr(errors.New("hgetall error"))
		_, err := GetUser(db, userID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to retrieve user data") {
			t.Errorf("unexpected error: %q", err.Error())
		}
	})

	t.Run("Successful retrieval", func(t *testing.T) {
		data := map[string]string{
			"email": "user@example.com",
			"name":  "Regular User",
		}
		mock.ExpectHGetAll(userKey).SetVal(data)
		res, err := GetUser(db, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res["email"] != "user@example.com" {
			t.Errorf("expected email %q, got %q", "user@example.com", res["email"])
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %s", err)
	}
}

func TestGetAllAdminEmails(t *testing.T) {
	db, mock := redismock.NewClientMock()

	t.Run("Error retrieving user IDs", func(t *testing.T) {
		mock.ExpectSMembers("users").SetErr(errors.New("smembers error"))
		_, err := GetAllAdminEmails(db)
		if err == nil || !strings.Contains(err.Error(), "failed to retrieve user IDs") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("No user IDs found", func(t *testing.T) {
		mock.ExpectSMembers("users").SetVal([]string{})
		_, err := GetAllAdminEmails(db)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrNoUsersFound {
			t.Errorf("expected error %q, got %q", ErrNoUsersFound.Error(), err.Error())
		}
	})

	t.Run("Successful retrieval with skipping errors", func(t *testing.T) {
		// Suppose there are two userIDs; one returns valid user data and the other returns an error.
		userIDs := []string{"1", "2"}
		mock.ExpectSMembers("users").SetVal(userIDs)

		// For user "1", return valid data with an email.
		key1 := "user:" + "1"
		data1 := map[string]string{"email": "user1@example.com"}
		mock.ExpectHGetAll(key1).SetVal(data1)

		// For user "2", simulate error (or missing user data).
		key2 := "user:" + "2"
		mock.ExpectHGetAll(key2).SetErr(errors.New("hgetall error"))
		emails, err := GetAllAdminEmails(db)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(emails) != 1 || emails[0] != "user1@example.com" {
			t.Errorf("expected [\"user1@example.com\"], got %v", emails)
		}
	})

	t.Run("No emails from user data", func(t *testing.T) {
		// SMembers returns one userID, but HGetAll returns a map without an email.
		mock.ExpectSMembers("users").SetVal([]string{"3"})
		key3 := "user:" + "3"
		mock.ExpectHGetAll(key3).SetVal(map[string]string{"name": "No Email"})

		_, err := GetAllAdminEmails(db)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrNoUsersFound {
			t.Errorf("expected error %q, got %q", ErrNoUsersFound.Error(), err.Error())
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %s", err)
	}
}
