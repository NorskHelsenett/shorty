package config

import (
	"context"
	"fmt"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// create redis-client
func NewClient() (*redis.Client, error) {
	rlog.Info("Connecting to Redis server")
	password := viper.GetString("REDIS_PASSWORD")
	host := viper.GetString("REDIS_HOST")
	port := viper.GetString("REDIS_PORT")

	if host == "" {
		host = "localhost"

	}

	if port == "" {
		port = "6379"
	}

	rlog.Debug("Redis config", rlog.String("host", host), rlog.String("port", port))

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})

	// checks if the server is up and running
	_, err := client.Ping(context.Background()).Result()

	if err != nil {
		rlog.Error("Redis connection failed", err)
		return nil, err
	}
	rlog.Info("Redis connected")

	return client, nil
}
