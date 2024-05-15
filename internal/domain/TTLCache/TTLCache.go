package cache

import (
	"cache/internal/domain"
	"sync"
	"time"
)

type TTLCache struct {
	CacheMap map[domain.Key]*TTLEntry
	MinLevel int
	Capacity int
}

type TTLEntry struct {
	ttl time.Duration
	key domain.Key
	val domain.Key
	mu  sync.RWMutex
}
