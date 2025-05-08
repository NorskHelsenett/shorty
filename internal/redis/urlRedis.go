package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/NorskHelsenett/shorty/internal/models"
	"github.com/go-redis/redis/v8"
)

var (
	// ErrURLNotFound is returned when a URL is not found in the database
	ErrURLNotFound = errors.New("URL not found")
	// ErrNoPathsFound is returned when no paths are found in the database
	ErrNoPathsFound = errors.New("no paths found")
)

// GetURL retrieves a URL by its key ID
func GetURL(rdb *redis.Client, keyID string) (string, error) {
	pathKey := "path:" + keyID
	url, err := rdb.HGet(context.Background(), pathKey, "url").Result()

	if err == redis.Nil {
		return "", ErrURLNotFound
	} else if err != nil {
		return "", err
	}

	return url, nil
}

// UpdateOrCreate creates or updates a URL in Redis
// Returns a descriptive message and any error that occurred
func UpdateOrCreatePath(rdb *redis.Client, key string, newValue string, user string) (string, error) {
	ctx := context.Background()
	pathKey := "path:" + key

	// Check if the key exists
	exists, err := rdb.Exists(ctx, pathKey).Result()
	if err != nil {
		rlog.Error("Failed to check if key exists", err, rlog.Any("key", key))
		return "", err
	}

	editTime := time.Now().Format(time.RFC3339)

	// Create or update
	if exists == 1 {
		// Update existing record
		err = rdb.HSet(ctx, pathKey,
			"url", newValue,
			"lastEditBy", user,
			"lastEditTime", editTime,
		).Err()
		if err != nil {
			rlog.Error("Failed to update path", err, rlog.Any("key", key), rlog.Any("value", newValue), rlog.String("user", user), rlog.Any("edit time", editTime))
			return "", err
		}
		rlog.Info("Path updated", rlog.Any("key", key), rlog.Any("value", newValue), rlog.String("user", user), rlog.Any("edit time", editTime))
		return "Path updated successfully", nil
	}

	// Create new record
	err = rdb.HSet(ctx, pathKey,
		"url", newValue,
		"createdBy", user,
		"lastEditBy", "",
		"createdTime", editTime,
	).Err()
	if err != nil {
		rlog.Error("Failed to create path", err, rlog.Any("key", key), rlog.Any("value", newValue), rlog.String("user", user), rlog.Any("edit time", editTime))
		return "", err
	}

	rlog.Info("Path created", rlog.Any("key", key), rlog.Any("value", newValue), rlog.Any("user", user), rlog.Any("edit time", editTime))
	return "Path created successfully", nil
}

// URLExists checks if a URL with the given key exists in the database
func URLExists(rdb *redis.Client, key string) (bool, error) {
	exist, err := rdb.Exists(context.Background(), "path:"+key).Result()
	if err != nil {
		return false, err
	}
	return exist > 0, nil
}

// Delete removes a redirect by key
// Returns true if the key was deleted, false if it didn't exist
func Delete(rdb *redis.Client, key string) (bool, error) {
	path := "path:" + key
	nDeleted, err := rdb.Del(context.Background(), path).Result()
	if err != nil {
		return false, err
	}
	return nDeleted > 0, nil
}

// GetAll retrieves all redirects with the given prefix
// Returns a slice of RedirectPath objects
func GetAll(rdb *redis.Client, prefix string) ([]models.RedirectPath, error) {
	ctx := context.Background()

	keys, err := rdb.Keys(ctx, prefix+"*").Result()
	if err != nil {
		return nil, err
	}

	redirectPaths := make([]models.RedirectPath, 0, len(keys))

	for _, key := range keys {
		url, err := rdb.HGet(ctx, key, "url").Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		createdBy, err := rdb.HGet(ctx, key, "createdBy").Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}

		// Extract path from key
		path := strings.TrimPrefix(key, prefix)
		path = strings.TrimLeft(path, ":")

		redirect := models.RedirectPath{
			Path:  path,
			URL:   url,
			Owner: createdBy,
		}
		redirectPaths = append(redirectPaths, redirect)
	}
	return redirectPaths, nil
}

// GetPathOwner retrieves the owner of a path
// Returns the owner's email or an error if not found
func GetPathOwner(rdb *redis.Client, key string) (string, error) {
	const errOwnerNotFound = "owner not found"
	pathKey := "path:" + key

	createdBy, err := rdb.HGet(context.Background(), pathKey, "createdBy").Result()
	if err == redis.Nil {
		return "", errors.New(errOwnerNotFound)
	} else if err != nil {
		return "", err
	}

	return createdBy, nil
}
