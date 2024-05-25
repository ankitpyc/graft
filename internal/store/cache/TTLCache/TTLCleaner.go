package cache

import (
	"fmt"
	"time"
)

// TTLCleaner represents a cleaner responsible for evicting expired entries from the cache.
type TTLCleaner struct {
	ttl      time.Duration // ttl represents the time-to-live duration for cache entries.
	interval time.Duration // interval represents the interval at which the cleaner runs to check for expired entries.
	stop     chan bool     // stop is a channel to signal the cleaner to stop.
}

// NewTTLCleaner creates a new instance of TTLCleaner with the specified TTL duration and cleaning interval.
func NewTTLCleaner(duration time.Duration, interval time.Duration) *TTLCleaner {
	return &TTLCleaner{
		ttl:      duration,
		interval: interval,
		stop:     make(chan bool),
	}
}

// Clean iterates through the cache and evicts entries that have expired based on their TTL.
func (cleaner *TTLCleaner) Clean(cache *TTLCache) {
	for key, value := range cache.CacheMap {
		value.mu.Lock()
		if time.Now().After(value.EntryTime.Add(cache.TTLExpiry)) {
			fmt.Println("Removing expired entry", key)
			cache.Evict(key)
		}
		value.mu.Unlock()
	}
}

// Run starts the cleaner routine, which periodically checks for expired entries in the cache and evicts them.
func (cleaner *TTLCleaner) Run(cache *TTLCache) {
	timer := time.NewTicker(cleaner.interval)
	for {
		select {
		case <-timer.C:
			fmt.Println("Running TTLCleaner at ", time.Now())
			cleaner.Clean(cache)
		case <-cleaner.stop:
			timer.Stop()
			return
		}
	}
}
