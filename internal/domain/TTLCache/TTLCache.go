package cache

import (
	"cache/internal/domain"
	_interface "cache/internal/domain/interface"
	"fmt"
	"log"
	"sync"
	"time"
)

type TTLCache struct {
	CacheMap  map[domain.Key]*TTLEntry
	Capacity  int
	TTLExpiry time.Duration
	Interval  time.Duration
	Expora    *TTLCleaner
	evicted   chan domain.Key
}

func (cache *TTLCache) EvictKey() {
	panic("implement me")
}

func NewTTLCache(capacity int, interval time.Duration, ttl time.Duration) _interface.Cache {
	ttlCache := &TTLCache{
		CacheMap: make(map[domain.Key]*TTLEntry),
		Capacity: capacity,
		Expora:   NewTTLCleaner(ttl, interval),
	}
	go ttlCache.Expora.Run(ttlCache)
	return ttlCache
}

func (cache *TTLCache) Get(key domain.Key) domain.Key {
	if entry, ok := cache.CacheMap[key]; ok {
		cache.updateEntry(entry)
		return entry.val
	}
	log.Print("Entry not found in cache for key: ", key)
	return -1
}

func (cache *TTLCache) Evict(key domain.Key) {
	delete(cache.CacheMap, key)
}

func (cache *TTLCache) GetAllCacheData() {
	for key, val := range cache.CacheMap {
		fmt.Println(key, " --> ", val)
	}
}

func (cache *TTLCache) Put(key domain.Key, val domain.Key) {
	entry := &TTLEntry{
		val:       val,
		key:       key,
		EntryTime: time.Now(),
		mu:        sync.RWMutex{},
	}
	log.Print("added new entry", entry.key)
	cache.CacheMap[key] = entry
}

func (cache *TTLCache) updateEntry(entry *TTLEntry) {
	entry.EntryTime = time.Now()
	cache.CacheMap[entry.key] = entry
}

type TTLEntry struct {
	key       domain.Key
	val       domain.Key
	EntryTime time.Time
	mu        sync.RWMutex
}
