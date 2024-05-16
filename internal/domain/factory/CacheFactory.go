package factory

import (
	LFUCache "cache/internal/domain/LFUCache"
	LRUCache "cache/internal/domain/LRUCache"
	TTLCache "cache/internal/domain/TTLCache"
	err "cache/internal/domain/errors"
	Cache "cache/internal/domain/interface"
	"time"
)

type CacheFactory struct {
	CacheType string
	capacity  int
}

func CreateCache(cacheType string, capacity int) (Cache.Cache, error) {
	switch cacheType {
	case "LRU":
		return LRUCache.NewCache(capacity), nil
	case "LFU":
		return LFUCache.NewCache(capacity), nil
	case "TTL":
		return TTLCache.NewTTLCache(capacity, 10*time.Second, 7*time.Second), nil
	default:
		return nil, err.ErrInvalidCacheType
	}
}
