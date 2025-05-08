package redis

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/NorskHelsenett/shorty/internal/models"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startRedisContainer(t *testing.T) (*redis.Client, func()) {
	ctx := context.Background()

	// create Redis-container
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "Failed to start Redis container")

	// GEt endpoint (host og port)
	endpoint, err := redisC.Endpoint(ctx, "")
	require.NoError(t, err, "Failed to get Redis endpoint")

	// Redis-klienten with dynamic adress from container
	rdb := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	// Cleanup container funksjon
	cleanup := func() {
		err = redisC.Terminate(ctx)
		require.NoError(t, err, "Failed to terminate Redis container")
		rdb.Close()
	}

	return rdb, cleanup
}

func TestWithRedis(t *testing.T) {
	rdb, cleanup := startRedisContainer(t)
	defer cleanup()

	// Test conection to Redis
	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	require.NoError(t, err, "Failed to ping Redis")
	require.Equal(t, "PONG", pong, "Unexpected Redis ping result")

	// runs simple READ/WRITE-test with Redis
	t.Run("Basic Redis Set/Get Operation", func(t *testing.T) {
		err := rdb.Set(ctx, "key", "value", 0).Err()
		require.NoError(t, err, "Failed to set key in Redis")

		val, err := rdb.Get(ctx, "key").Result()
		require.NoError(t, err, "Failed to get key from Redis")
		require.Equal(t, "value", val, "Unexpected Redis get value")
	})
}

func TestUpdateOrCreate(t *testing.T) {
	rdb, cleanup := startRedisContainer(t)
	defer cleanup()

	type args struct {
		rdb      *redis.Client
		key      string
		newValue string
		user     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Create new key",
			args: args{
				rdb:      rdb,
				key:      "newKey",
				newValue: "http://new-url.com",
				user:     "testUser",
			},
			want:    "Path created successfully",
			wantErr: false,
		},
		{
			name: "Oppdatering av eksisterende nøkkel",
			args: args{
				rdb:      rdb,
				key:      "existingKey",
				newValue: "http://updated-url.com",
				user:     "editorUser",
			},
			want:    "Path updated successfully",
			wantErr: false,
		},
		{
			name: "Opprettelse av nøkkel med Tom verdi",
			args: args{
				rdb:      rdb,
				key:      "keyWithEmptyValue",
				newValue: "",
				user:     "testUser",
			},
			want:    "Path created successfully",
			wantErr: false,
		},
	}

	ctx := context.Background()

	// testdata
	err := rdb.HSet(ctx, "path:existingKey", "url", "http://existing-url.com").Err()
	require.NoError(t, err, "Failed to set up initial Redis data")

	// run testdata
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateOrCreatePath(tt.args.rdb, tt.args.key, tt.args.newValue, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateOrCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UpdateOrCreate() = %v, want %v", got, tt.want)
			}

			// Validate data in Redis
			pathKey := "path:" + tt.args.key
			result, err := rdb.HGetAll(ctx, pathKey).Result()
			if err != nil {
				t.Fatalf("Failed to validate Redis data for key %s: %v", pathKey, err)
			}

			if tt.args.newValue != "" && result["url"] != tt.args.newValue {
				t.Errorf("Expected 'url' to be %q but got %q", tt.args.newValue, result["url"])
			}
			if tt.args.user != "" && result["createdBy"] != tt.args.user && result["lastEditBy"] != tt.args.user {
				t.Errorf("Expected 'createdBy' or 'lastEditBy' to be %q but got %q and %q", tt.args.user, result["createdBy"], result["lastEditBy"])
			}
		})
	}
}

func TestGetURL(t *testing.T) {
	rdb, cleanup := startRedisContainer(t)
	defer cleanup()

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
			name: "Get path by existing key",
			args: args{
				rdb:   rdb,
				keyID: "existingKey",
			},
			want:    "http://existing-url.com",
			wantErr: false,
		},
		{
			name: "Get path by no existing key",
			args: args{
				rdb:   rdb,
				keyID: "missingKey",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Get path by no key",
			args: args{
				rdb:   rdb,
				keyID: "",
			},
			want:    "",
			wantErr: true,
		},
	}
	ctx := context.Background()
	err := rdb.HSet(ctx, "path:existingKey", "url", "http://existing-url.com").Err()
	require.NoError(t, err, "Failed to set up initial Redis data")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetURL(tt.args.rdb, tt.args.keyID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErr {
				require.EqualError(t, err, "URL not found", "Unexpected error message")
				return
			}

			if got != tt.want {
				t.Errorf("GetURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUrlExist(t *testing.T) {
	rdb, cleanup := startRedisContainer(t)
	defer cleanup()

	type args struct {
		rdb *redis.Client
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "existing key",
			args: args{
				rdb: rdb,
				key: "existingKey",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no existing key",
			args: args{
				rdb: rdb,
				key: "missingKey",
			},
			want:    false,
			wantErr: false,
		},
	}

	ctx := context.Background()
	err := rdb.HSet(ctx, "path:existingKey", "url", "http://existing-url.com").Err()
	require.NoError(t, err, "Failed to set up initial Redis data")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := URLExists(tt.args.rdb, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("URLExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("URLExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	rdb, cleanup := startRedisContainer(t)
	defer cleanup()

	type args struct {
		rdb *redis.Client
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "existing key",
			args: args{
				rdb: rdb,
				key: "existingKey",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no existing key",
			args: args{
				rdb: rdb,
				key: "missingKey",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "no key",
			args: args{
				rdb: rdb,
				key: "",
			},
			want:    false,
			wantErr: false,
		},
	}

	ctx := context.Background()
	err := rdb.HSet(ctx, "path:existingKey", "url", "http://existing-url.com").Err()
	require.NoError(t, err, "Failed to set up initial Redis data")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Delete(tt.args.rdb, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	rdb, cleanup := startRedisContainer(t)
	defer cleanup()

	tests := []struct {
		name    string
		prefix  string
		want    []models.RedirectPath
		wantErr bool
	}{
		{
			name:   "Gyldige paths",
			prefix: "path:",
			want: []models.RedirectPath{
				{Path: "abc", URL: "https://www.google.com", Owner: "user1"},
				{Path: "ghi", URL: "https://golang.org", Owner: ""},
				{Path: "def", URL: "https://example.com", Owner: "user2"},
				// "jkl" skal ikke returneres siden "url" mangler
			},
			wantErr: false,
		},
		{
			name:    "Ingen paths funnet",
			prefix:  "nopath:",
			want:    nil,
			wantErr: true, // forventer feilen "no paths found"
		},
	}

	ctx := context.Background()
	// Setting up test data – these keys have the prefix "path:"
	// 1. Valid key with both fields
	err := rdb.HSet(ctx, "path:abc", "url", "https://www.google.com", "createdBy", "user1").Err()
	if err != nil {
		t.Fatalf("feil ved oppretting av nøkkel: %v", err)
	}
	err = rdb.HSet(ctx, "path:def", "url", "https://example.com", "createdBy", "user2").Err()
	if err != nil {
		t.Fatalf("feil ved oppretting av nøkkel: %v", err)
	}
	// 2. Key missing "createdBy"
	err = rdb.HSet(ctx, "path:ghi", "url", "https://golang.org").Err() // createdBy mangler
	if err != nil {
		t.Fatalf("feil ved oppretting av nøkkel: %v", err)
	}
	// 3. Key missing "url" – should be skipped by GetAll
	err = rdb.HSet(ctx, "path:jkl", "createdBy", "userX").Err()
	if err != nil {
		t.Fatalf("feil ved oppretting av nøkkel: %v", err)
	}
	// 4. Key outside the prefix – this will be ignored when searching with prefix "path:".
	err = rdb.HSet(ctx, "other:foo", "url", "https://foo.com", "createdBy", "userFoo").Err()
	if err != nil {
		t.Fatalf("feil ved oppretting av nøkkel: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAll(rdb, tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAll() = %v, want %v", got, tt.want)
			}
			if err == nil {
				sortRedirectPaths(got)
				sortRedirectPaths(tt.want)
				if !equalRedirectPaths(got, tt.want) {
					t.Errorf("GetAll() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// sortRedirectPaths sorts a slice of RedirectPath by the Path field.
func sortRedirectPaths(paths []models.RedirectPath) {
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Path < paths[j].Path
	})
}

// equalRedirectPaths compares two slices of models.RedirectPath
func equalRedirectPaths(a, b []models.RedirectPath) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGetPathOwner(t *testing.T) {
	rdb, cleanup := startRedisContainer(t)
	defer cleanup()

	type args struct {
		rdb *redis.Client
		key string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "existing key",
			args: args{
				rdb: rdb,
				key: "abc",
			},
			want:    "user1",
			wantErr: false,
		},
		{
			name: "no key",
			args: args{
				rdb: rdb,
				key: "missingKey",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "no existing key",
			args: args{
				rdb: rdb,
				key: "tps",
			},
			want:    "",
			wantErr: true,
		},
	}
	ctx := context.Background()
	err := rdb.HSet(ctx, "path:abc", "url", "https://www.google.com", "createdBy", "user1").Err()
	if err != nil {
		t.Fatalf("feil ved oppretting av nøkkel: %v", err)
	}
	err = rdb.HSet(ctx, "path:tue", "url", "https://www.google.com", "createdBy", "user1").Err()
	if err != nil {
		t.Fatalf("feil ved oppretting av nøkkel: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPathOwner(tt.args.rdb, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPathOwner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetPathOwner() = %v, want %v", got, tt.want)
			}
		})
	}
}
