package factory

import (
	errors "cache/internal/domain/errors"
	"cache/internal/store"
	LFUCache "cache/internal/store/cache/LFUCache"
	LRUCache "cache/internal/store/cache/LRUCache"
	TTLCache "cache/internal/store/cache/TTLCache"
	"time"
)

// CreateCache creates and returns a cache instance based on the specified cache type and capacity.
// Parameters:
//   cacheType: Type of cache ("LRU", "LFU", "TTL").
//   capacity: Maximum capacity of the cache.
// Returns:
//   Cache: A cache instance.
//   error: Nil if successful, otherwise an error indicating the cause.

func CreateCache(cacheType string, capacity int) (store.StoreInf, error) {
	var err error = nil
	switch cacheType {
	case "LRU":
		return LRUCache.NewCache(capacity), err
	case "LFU":
		return LFUCache.NewCache(capacity), err
	case "TTL":
		return TTLCache.NewCache(capacity, 10*time.Second, 7*time.Second), err
	default:
		return nil, errors.ErrInvalidCacheType
	}
}
