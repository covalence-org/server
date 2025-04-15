package middleware

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client // Global Redis client instance

// InitRedis initializes the Redis client connection using environment variables.
func InitRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatalf("fatal: REDIS_ADDR environment variable not set!")
	}
	redisPassword := os.Getenv("REDIS_PASSWORD") // Get password, may be empty

	opts := &redis.Options{
		Addr:         redisAddr,
		Password:     redisPassword, // <-- REMOVE setting it unconditionally here
		DB:           0,             // Default DB
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
		PoolSize:     10,
	}

	RDB = redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ping should work regardless of AUTH now, as AUTH hasn't happened yet if no pass
	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("fatal: failed to connect to redis (ping failed) at %s: %v", redisAddr, err)
	}
	log.Println("successfully connected to redis (ping successful).")
}
