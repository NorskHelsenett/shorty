package redis

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
)

func TestGetURL(t *testing.T) {
	// Create a new mock with a redis client
	db, mock := redismock.NewClientMock()

	type args struct {
		rdb   *redis.Client
		keyID string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Key not found",
			args:    args{rdb: db, keyID: "nonexistent"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Valid URL",
			args:    args{rdb: db, keyID: "existing"},
			want:    "https://example.com",
			wantErr: false,
		},
	}

	// Set up expectations

	// When key "nonexistent" is used, redis returns nil
	mock.ExpectHGet("path:nonexistent", "url").RedisNil()
	// When key "existing" is used, redis returns the valid URL
	mock.ExpectHGet("path:existing", "url").SetVal("https://example.com")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetURL(tt.args.rdb, tt.args.keyID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetURL() = %v, want %v", got, tt.want)
			}
		})
	}

	// Ensure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %s", err)
	}
}

func TestUpdateOrCreatePath(t *testing.T) {

	db, mock := redismock.NewClientMock()

	user := "testuser"
	key := "mykey"
	newValue := "https://example.com"

	t.Run("Update success", func(t *testing.T) {
		pathKey := "path:" + key

		// Expect Exists to return 1 (key exists)
		mock.ExpectExists(pathKey).SetVal(1)

		// Capture the expected timestamp.
		expectedTime := time.Now().Format(time.RFC3339)
		// Expect an HSet call with the update fields
		mock.ExpectHSet(pathKey,
			"url", newValue,
			"lastEditBy", user,
			"lastEditTime", expectedTime,
		).SetVal(1)

		// Call the function under test.
		msg, err := UpdateOrCreatePath(db, key, newValue, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg != "Path updated successfully" {
			t.Errorf("unexpected message: got %q, want %q", msg, "Path updated successfully")
		}
	})

	t.Run("Create success", func(t *testing.T) {
		pathKey := "path:" + key

		// Expect Exists to return 0 (key does not exist)
		mock.ExpectExists(pathKey).SetVal(0)

		// Capture expected timestamp for creation.
		expectedTime := time.Now().Format(time.RFC3339)
		// Expect HSet call with the create fields
		mock.ExpectHSet(pathKey,
			"url", newValue,
			"createdBy", user,
			"lastEditBy", "",
			"createdTime", expectedTime,
		).SetVal(1)

		msg, err := UpdateOrCreatePath(db, key, newValue, user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg != "Path created successfully" {
			t.Errorf("unexpected message: got %q, want %q", msg, "Path created successfully")
		}
	})

	t.Run("Exists error", func(t *testing.T) {
		pathKey := "path:" + key

		// Simulate an error on Exists
		mock.ExpectExists(pathKey).SetErr(errors.New("exists error"))

		_, err := UpdateOrCreatePath(db, key, newValue, user)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if err.Error() != "exists error" {
			t.Errorf("unexpected error: got %q, want %q", err.Error(), "exists error")
		}
	})

	t.Run("HSet update error", func(t *testing.T) {
		pathKey := "path:" + key

		// Expect Exists to return 1 (update branch)
		mock.ExpectExists(pathKey).SetVal(1)
		expectedTime := time.Now().Format(time.RFC3339)
		// Simulate an error during HSet call in update branch.
		mock.ExpectHSet(pathKey,
			"url", newValue,
			"lastEditBy", user,
			"lastEditTime", expectedTime,
		).SetErr(errors.New("hset update error"))

		_, err := UpdateOrCreatePath(db, key, newValue, user)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if err.Error() != "hset update error" {
			t.Errorf("unexpected error: got %q, want %q", err.Error(), "hset update error")
		}
	})

	t.Run("HSet create error", func(t *testing.T) {
		pathKey := "path:" + key

		// Expect Exists to return 0 (create branch)
		mock.ExpectExists(pathKey).SetVal(0)
		expectedTime := time.Now().Format(time.RFC3339)
		// Simulate an error during HSet call in create branch.
		mock.ExpectHSet(pathKey,
			"url", newValue,
			"createdBy", user,
			"lastEditBy", "",
			"createdTime", expectedTime,
		).SetErr(errors.New("hset create error"))

		_, err := UpdateOrCreatePath(db, key, newValue, user)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if err.Error() != "hset create error" {
			t.Errorf("unexpected error: got %q, want %q", err.Error(), "hset create error")
		}
	})

	// Ensure that all expectations were met.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %s", err)
	}
}

func TestURLExists(t *testing.T) {
	// Create a new mock using redis/v8 and redismock/v8.
	db, mock := redismock.NewClientMock()

	t.Run("Key exists", func(t *testing.T) {
		key := "existing"
		pathKey := "path:" + key

		// Expect Exists to return 1 (key exists)
		mock.ExpectExists(pathKey).SetVal(1)

		exists, err := URLExists(db, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Errorf("expected key to exist, got %v", exists)
		}
	})

	t.Run("Key does not exist", func(t *testing.T) {
		key := "nonexisting"
		pathKey := "path:" + key

		// Expect Exists to return 0 (key does not exist)
		mock.ExpectExists(pathKey).SetVal(0)

		exists, err := URLExists(db, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Errorf("expected key to not exist, got %v", exists)
		}
	})

	t.Run("Exists error", func(t *testing.T) {
		key := "error"
		pathKey := "path:" + key

		// Expect Exists to return an error
		mock.ExpectExists(pathKey).SetErr(errors.New("exists error"))

		exists, err := URLExists(db, key)
		if err == nil {
			t.Fatal("expected an error but got nil")
		}
		if exists {
			t.Errorf("expected exists to be false on error, got %v", exists)
		}
	})

	// Ensure that all expectations were met.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestDelete(t *testing.T) {
	// Create a new Redis mock client.
	db, mock := redismock.NewClientMock()

	t.Run("Delete successful", func(t *testing.T) {
		key := "existing"
		path := "path:" + key

		// Expect Del to return 1 as the number of deleted keys.
		mock.ExpectDel(path).SetVal(1)

		deleted, err := Delete(db, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !deleted {
			t.Errorf("expected deletion to be true; got %v", deleted)
		}
	})

	t.Run("No key deleted", func(t *testing.T) {
		key := "nonexistent"
		path := "path:" + key

		// Expect Del to return 0 as no keys were deleted.
		mock.ExpectDel(path).SetVal(0)

		deleted, err := Delete(db, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if deleted {
			t.Errorf("expected deletion to be false; got %v", deleted)
		}
	})

	t.Run("Del returns error", func(t *testing.T) {
		key := "error"
		path := "path:" + key

		// Expect Del to return an error.
		mock.ExpectDel(path).SetErr(errors.New("delete error"))

		deleted, err := Delete(db, key)
		if err == nil {
			t.Fatal("expected error, but got nil")
		}
		if deleted {
			t.Errorf("expected deletion to be false when error occurs; got %v", deleted)
		}
	})

	// Verify that all expectations were met.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestGetAll(t *testing.T) {
	const prefix = "myprefix"

	t.Run("error on keys", func(t *testing.T) {
		// Create a new redis mock instance.
		db, mock := redismock.NewClientMock()

		// Simulate error on Keys command.
		mock.ExpectKeys(prefix + "*").SetErr(errors.New("keys error"))

		_, err := GetAll(db, prefix)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "keys error" {
			t.Errorf("unexpected error: got %q, want %q", err.Error(), "keys error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("no keys returned", func(t *testing.T) {
		db, mock := redismock.NewClientMock()

		// Expect Keys call returning an empty slice.
		mock.ExpectKeys(prefix + "*").SetVal([]string{})

		result, err := GetAll(db, prefix)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 results, got %d", len(result))
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("valid key with url and createdBy", func(t *testing.T) {
		db, mock := redismock.NewClientMock()

		// Assume the key is stored as "myprefix:somepath"
		fullKey := prefix + ":somepath"
		mock.ExpectKeys(prefix + "*").SetVal([]string{fullKey})

		// Expect HGet for url returns a valid URL.
		mock.ExpectHGet(fullKey, "url").SetVal("https://example.com")
		// Expect HGet for createdBy returns the owner.
		mock.ExpectHGet(fullKey, "createdBy").SetVal("owner1")

		results, err := GetAll(db, prefix)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		got := results[0]
		if got.Path != "somepath" {
			t.Errorf("expected path %q, got %q", "somepath", got.Path)
		}
		if got.URL != "https://example.com" {
			t.Errorf("expected URL %q, got %q", "https://example.com", got.URL)
		}
		if got.Owner != "owner1" {
			t.Errorf("expected owner %q, got %q", "owner1", got.Owner)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("skip key when url is missing", func(t *testing.T) {
		db, mock := redismock.NewClientMock()

		// Return one key whose url field does not exist.
		fullKey := prefix + ":nopage"
		mock.ExpectKeys(prefix + "*").SetVal([]string{fullKey})
		mock.ExpectHGet(fullKey, "url").RedisNil()

		results, err := GetAll(db, prefix)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("error when retrieving url", func(t *testing.T) {
		db, mock := redismock.NewClientMock()

		fullKey := prefix + ":errorpage"
		mock.ExpectKeys(prefix + "*").SetVal([]string{fullKey})
		// Simulate a non-nil error (other than redis.Nil) on HGet for url.
		mock.ExpectHGet(fullKey, "url").SetErr(errors.New("hget url error"))

		_, err := GetAll(db, prefix)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "hget url error" {
			t.Errorf("unexpected error: got %q, want %q", err.Error(), "hget url error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}

func TestGetPathOwner(t *testing.T) {
	// Create a new Redis mock client.
	db, mock := redismock.NewClientMock()

	t.Run("Owner exists", func(t *testing.T) {
		key := "somekey"
		pathKey := "path:" + key

		// Expect HGet to return a valid owner.
		mock.ExpectHGet(pathKey, "createdBy").SetVal("owner1")

		owner, err := GetPathOwner(db, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if owner != "owner1" {
			t.Errorf("expected owner 'owner1', got %q", owner)
		}
	})

	t.Run("Owner not found", func(t *testing.T) {
		key := "missing"
		pathKey := "path:" + key

		// Simulate redis.Nil which means the field was not found.
		mock.ExpectHGet(pathKey, "createdBy").RedisNil()

		owner, err := GetPathOwner(db, key)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if err.Error() != "owner not found" {
			t.Errorf("expected error 'owner not found', got %q", err.Error())
		}
		if owner != "" {
			t.Errorf("expected empty owner, got %q", owner)
		}
	})

	t.Run("Error retrieving owner", func(t *testing.T) {
		key := "failure"
		pathKey := "path:" + key

		// Simulate a non-redis.Nil error.
		mock.ExpectHGet(pathKey, "createdBy").SetErr(errors.New("hget error"))

		owner, err := GetPathOwner(db, key)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if err.Error() != "hget error" {
			t.Errorf("expected error 'hget error', got %q", err.Error())
		}
		if owner != "" {
			t.Errorf("expected empty owner, got %q", owner)
		}
	})

	// Ensure all expectations are met.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestValidatePathInput(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Valid path and URL",
			key:     "test",
			value:   "https://example.com",
			wantErr: false,
		},
		{
			name:        "Reserved key - admin",
			key:         "admin",
			value:       "https://example.com",
			wantErr:     true,
			errContains: "reserved key",
		},
		{
			name:        "Reserved key - api (case insensitive)",
			key:         "API",
			value:       "https://example.com",
			wantErr:     true,
			errContains: "reserved key",
		},
		{
			name:        "Reserved key - health",
			key:         "health",
			value:       "https://example.com",
			wantErr:     true,
			errContains: "reserved key",
		},
		{
			name:        "Reserved key - metrics",
			key:         "metrics",
			value:       "https://example.com",
			wantErr:     true,
			errContains: "reserved key",
		},
		{
			name:        "Reserved key - swagger",
			key:         "swagger",
			value:       "https://example.com",
			wantErr:     true,
			errContains: "reserved key",
		},
		{
			name:        "Self-redirect HTTPS",
			key:         "test",
			value:       "https://k.nhn.no/test",
			wantErr:     true,
			errContains: "cannot redirect to",
		},
		{
			name:        "Self-redirect HTTP",
			key:         "test",
			value:       "http://k.nhn.no/test",
			wantErr:     true,
			errContains: "cannot redirect to",
		},
		{
			name:        "Forbidden target - admin user HTTPS",
			key:         "test",
			value:       "https://k.nhn.no/admin/user",
			wantErr:     true,
			errContains: "cannot redirect to",
		},
		{
			name:        "Forbidden target - admin HTTPS",
			key:         "test",
			value:       "https://k.nhn.no/admin",
			wantErr:     true,
			errContains: "cannot redirect to",
		},
		{
			name:        "Forbidden target - admin user HTTP",
			key:         "test",
			value:       "http://k.nhn.no/admin/user",
			wantErr:     true,
			errContains: "cannot redirect to",
		},
		{
			name:        "Forbidden target - admin HTTP",
			key:         "test",
			value:       "http://k.nhn.no/admin",
			wantErr:     true,
			errContains: "cannot redirect to",
		},
		{
			name:    "Valid URL with subdomain",
			key:     "test",
			value:   "https://sub.example.com",
			wantErr: false,
		},
		{
			name:    "Valid URL with path",
			key:     "test",
			value:   "https://example.com/path/to/resource",
			wantErr: false,
		},
		{
			name:    "Different key - not a loop",
			key:     "prod",
			value:   "https://k.nhn.no/test",
			wantErr: false,
		},
		// NOTE: URL format validation (empty, invalid protocol, etc.)
		// is handled by IsURL() in the handler layer
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePathInput(tt.key, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePathInput() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validatePathInput() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("validatePathInput() unexpected error = %v", err)
				}
			}
		})
	}
}
