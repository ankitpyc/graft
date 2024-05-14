package factory

import (
	Cache "cache/internal/domain"
	LFUCache "cache/internal/domain/LFUCache"
	LRUCache "cache/internal/domain/LRUCache"
	"errors"
)

var errInvalidSize = errors.New("Invalid size")
var errInvalidCacheType = errors.New("Invalid cache Type")

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
		return nil, errInvalidCacheType
	default:
		return nil, errInvalidCacheType
	}
}
