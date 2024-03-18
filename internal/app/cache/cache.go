package cache

import (
	"context"
	"crypto/md5"
	"fmt"

	redis "github.com/redis/go-redis/v9"
)

type Options struct {
	Host string
	Port int
}

type Cache interface {
	Add(ctx context.Context, input string, language string, translation string) error
	Has(ctx context.Context, input string, language string) (string, bool)
	Get(ctx context.Context, input string, language string) string
}

type cache struct {
	client redis.Client
}

func NewCache(opts Options) Cache {
	return &cache{
		client: *redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", opts.Host, opts.Port),
			Password: "",
			DB:       0,
		}),
	}
}

// Add adds an key/value pair to the cache.
func (c *cache) Add(ctx context.Context, input string, language string, translation string) error {
	hashedKey := hashKey(fmt.Sprintf("%s%s", input, language))
	return c.client.Set(ctx, hashedKey, translation, 0).Err()
}

// Has checks if an key exists in the cache.
func (c *cache) Has(ctx context.Context, input string, language string) (string, bool) {
	hashedKey := hashKey(fmt.Sprintf("%s%s", input, language))
	result, err := c.client.Exists(ctx, hashedKey).Result()
	if err != nil {
		return hashedKey, false
	}
	return hashedKey, result > 0
}

// Get returns an key/value pair from the cache by its key.
func (c *cache) Get(ctx context.Context, input string, language string) string {
	hashedKey := hashKey(fmt.Sprintf("%s%s", input, language))
	return c.client.Get(ctx, hashedKey).Val()
}

// hashKey returns the md5 hash of the key.
func hashKey(key string) string {
	keyBytes := []byte(key)
	return fmt.Sprintf("%x", md5.Sum(keyBytes))
}
