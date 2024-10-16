package cache

import (
	"ad_service/internal/config"
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Cache struct holds the Redis client instance
type Cache struct {
	Client *redis.Client
}

// NewCache initializes and returns a new Cache instance connected to Redis
func NewCache() *Cache {

	// Load configuration from Viper
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}
	redisHost := cfg.Redis.Host
	redisPort := cfg.Redis.Port
	redisPassword := cfg.Redis.Password
	redisDB := cfg.Redis.DB

	// Create a Redis client using configuration
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword, // Password from config (can be empty)
		DB:       redisDB,       // DB number from config
	})

	// Test the connection to Redis
	errRetry := retry(func() error {
		_, err := rdb.Ping(context.Background()).Result()
		return err
	}, 3, 2*time.Second) // Retry 3 times with 2s delay if connection fails

	if errRetry != nil {
		log.Fatalf("Could not connect to Redis after multiple attempts: %v", err)
	}

	return &Cache{Client: rdb}
}

// Get retrieves a value from Redis by key, with tracing
func (c *Cache) Get(key string, ctx context.Context) (string, error) {
	// Start a new span for the Get operation
	tracer := otel.Tracer("cache")
	ctx, span := tracer.Start(ctx, "Redis Get")
	defer span.End()

	// Add key as attribute for tracing
	span.SetAttributes(attribute.String("redis.key", key))
	// Get the value associated with the key
	result, err := c.Client.Get(ctx, key).Result()

	if err == redis.Nil {
		span.SetAttributes(attribute.String("Cache", "miss"))
		return "", nil // Cache miss
	} else if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Error in Redis GET operation")
		return "", err // Other Redis errors
	}

	// Successfully retrieved from cache
	span.SetAttributes(attribute.String("Cache", "retrieved"))
	return result, nil
}

// Set stores a value in Redis with an expiration time, with tracing
func (c *Cache) Set(key string, value string, expiration time.Duration, ctx context.Context) error {
	// Start a new span for the Set operation
	tracer := otel.Tracer("cache")
	ctx, span := tracer.Start(ctx, "Redis Set")
	defer span.End()

	// Add key and expiration as attributes for tracing
	span.SetAttributes(
		attribute.String("redis.key", key),
		attribute.String("redis.value", value),
		attribute.Int64("redis.expiration", int64(expiration.Seconds())),
	)
	// Set the key-value pair with the specified expiration time
	err := c.Client.Set(ctx, key, value, expiration).Err()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Error in Redis SET operation")
		return err
	}

	// Successfully stored in cache
	span.SetAttributes(attribute.String("Cache", "set"))
	return nil
}

// Delete: removes a specific key from the Redis cache
func (c *Cache) Delete(key string, ctx context.Context) error {
	// Start a new span for the Delete operation
	tracer := otel.Tracer("cache")
	ctx, span := tracer.Start(ctx, "Redis Delete")
	defer span.End()

	// Delete the key from the Redis cache.
	err := c.Client.Del(ctx, key).Err()
	span.SetAttributes(attribute.String("redis.key", key))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Error in Redis DELETE operation")
		return err
	}
	span.SetAttributes(attribute.String("Cache", "deleted"))
	return nil
}

// retry is a helper function to retry Redis connection
func retry(operation func() error, attempts int, delay time.Duration) error {
	for i := 0; i < attempts; i++ {
		if err := operation(); err != nil {
			if i == attempts-1 {
				return err // Return the final error if all attempts fail
			}
			time.Sleep(delay)
			continue
		}
		break
	}
	return nil
}
