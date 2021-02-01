package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	rediscache "github.com/go-redis/cache/v7"
	"github.com/go-redis/redis/v7"
	"github.com/vmihailenco/msgpack/v4"
)

const (
	// DefaultRedisPort is the port used by redis by default.
	DefaultRedisPort = "6379"
	// Redis will be shared across multiple services. Each should have their own
	// key prefix to avoid collision and overwriting keys.
	keyPrefix = "poc"
	// localCacheSize size for use when Redis is not being used (usually for unit tests).
	localCacheSize = 5000
	//defaultExpiration is the default expiration time for a redis cache key.
	defaultExpiration = time.Hour * 2
	// defaultRetries is the default number of times a cache request is retried
	defaultRetries = 10
)

// ErrCacheMiss is returned when a key is not contained in cache.
var ErrCacheMiss = errors.New("cache: key is missing")

// Cache defines an interface for getting and setting values to a cache.
type Cache interface {
	// Get retrieves value belonging to key from the cache.
	// If no value exists an error of ErrCacheMiss should be returned.
	// A pointer type should be passed into value:
	//     var val string
	//     Get("test", &val)
	Get(ctx context.Context, key string, value interface{}, retry bool) error

	// Set sets value in a cache under key, with an expiration.
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration, retry bool) error
}

// NewCache returns a Cache object. If a host is provided then the underlying
// implementation will connect to redis at that host. Otherwise, an in-memory cache will be used.
func NewCache(host string) Cache {
	// All values will be marshaled upon Set and unmarshaled upon Get.
	codec := &rediscache.Codec{
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}

	// Use in-memory cache if a redis server hostname is not provided.
	if host == "" {
		codec.UseLocalCache(localCacheSize, time.Minute*5)
	} else {
		client := redis.NewClient(&redis.Options{
			Addr: host + ":" + DefaultRedisPort,
		})
		codec.Redis = client
	}

	return &Redis{
		codec: codec,
	}
}

// Redis is an implementation of Cache with an underlying redis server.
type Redis struct {
	codec *rediscache.Codec
}

// Call the redis get with retries
func (r *Redis) Get(ctx context.Context, key string, value interface{}, retry bool) error {
	var err error
	retries := 1
	if retry {
		retries = defaultRetries
	}

	for i := 0; i < retries; i++ {
		// do not retry if no value is present
		if err = r.get(key, value); err == nil {
			return nil
		} else if err == ErrCacheMiss {
			return err
		}

		log.Debug().Msgf("failed to get cache item. retry #%v for key: %s, with error: %v", i, key, err)
	}

	return err
}

// Call the redis set with retries
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration, retry bool) error {
	var err error
	retries := 1
	if retry {
		retries = defaultRetries
	}

	for i := 0; i < retries; i++ {
		if err = r.set(key, value, expiration); err == nil {
			return nil
		}

		log.Debug().Msgf("failed to set cache item. retry #%v for key: %s, with error: %v", i, key, err)
	}

	return err
}

// Get retrieves key from the cache and reads it into value.
// A pointer type should be passed into value.
func (r *Redis) get(key string, value interface{}) error {
	err := r.codec.Get(keyWithPrefix(key), &value)
	if err != nil {
		// Use this package's error value to hide the redis implementation.
		if err == rediscache.ErrCacheMiss {
			return ErrCacheMiss
		}

		return err
	}

	return nil
}

// Set adds value to the cache. If no expiration is supplied, default to 2 hours.
func (r *Redis) set(key string, value interface{}, expiration time.Duration) error {
	if expiration == 0 {
		expiration = defaultExpiration
	}

	item := &rediscache.Item{
		Key:        keyWithPrefix(key),
		Object:     value,
		Expiration: expiration,
	}

	return r.codec.Set(item)
}

func keyWithPrefix(key string) string {
	return fmt.Sprintf("%s/%s", keyPrefix, key)
}
