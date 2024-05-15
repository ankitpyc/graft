package factory

import (
	Cache "cache/internal/domain"
	LFUCache "cache/internal/domain/LFUCache"
	LRUCache "cache/internal/domain/LRUCache"
	err "cache/internal/domain/errors"
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
		return nil, err.ErrInvalidCacheType
	default:
		return nil, err.ErrInvalidCacheType
	}
}
