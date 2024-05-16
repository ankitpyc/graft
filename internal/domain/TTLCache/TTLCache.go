package cache

import (
	"cache/internal/domain"
	_interface "cache/internal/domain/interface"
	"fmt"
	"log"
	"sync"
	"time"
)

// TTLCache implements a Time-To-Live (TTL) cache algorithm.
type TTLCache struct {
	CacheMap  map[domain.Key]*TTLEntry // CacheMap stores key-value pairs along with their associated TTL entries.
	Capacity  int                      // Capacity represents the maximum number of items the cache can hold.
	TTLExpiry time.Duration            // TTLExpiry represents the time duration after which entries expire.
	Interval  time.Duration            // Interval represents the interval at which the cache cleaner runs to check for expired entries.
	Expora    *TTLCleaner              // Expora is the TTL cleaner responsible for evicting expired entries.
	evicted   chan domain.Key          // evicted is a channel to communicate evicted keys.
}

// EvictKey is a placeholder function for TTLCache to satisfy the Cache interface.
// TTLCache uses a TTL cleaner to handle eviction of expired entries.
func (cache *TTLCache) EvictKey() {
	panic("implement me")
}

// NewCache creates a new instance of TTLCache with the specified capacity, TTL expiry duration, and cleaning interval.
func NewCache(capacity int, interval time.Duration, ttl time.Duration) _interface.Cache {
	ttlCache := &TTLCache{
		CacheMap: make(map[domain.Key]*TTLEntry),
		Capacity: capacity,
		Expora:   NewTTLCleaner(ttl, interval),
	}

	// Start the TTL cleaner routine to periodically check and evict expired entries.
	go ttlCache.Expora.Run(ttlCache)

	return ttlCache
}

// Get retrieves the value associated with the given key from the cache.
// If the key doesn't exist in the cache, it returns -1.
func (cache *TTLCache) Get(key domain.Key) domain.Key {
	if entry, ok := cache.CacheMap[key]; ok {
		cache.updateEntry(entry)
		return entry.val
	}
	log.Print("Entry not found in cache for key: ", key)
	return -1
}

// Evict removes the entry associated with the given key from the cache.
func (cache *TTLCache) Evict(key domain.Key) {
	delete(cache.CacheMap, key)
}

// GetAllCacheData prints all keys along with their corresponding values in the cache.
func (cache *TTLCache) GetAllCacheData() {
	for key, val := range cache.CacheMap {
		fmt.Println(key, " --> ", val)
	}
}

// Put adds a new key-value pair to the cache.
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

// updateEntry updates the entry's last accessed time in the cache.
func (cache *TTLCache) updateEntry(entry *TTLEntry) {
	entry.EntryTime = time.Now()
	cache.CacheMap[entry.key] = entry
}

// TTLEntry represents an entry in the TTL cache with key, value, and entry time.
type TTLEntry struct {
	key       domain.Key   // key is the unique identifier for the cache entry.
	val       domain.Key   // val is the value associated with the cache entry.
	EntryTime time.Time    // EntryTime represents the time when the entry was added to the cache.
	mu        sync.RWMutex // mu provides synchronization for concurrent access to the entry.
}
