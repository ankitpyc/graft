package cache

import (
	"fmt"
	"time"
)

type TTLCleaner struct {
	ttl      time.Duration
	interval time.Duration
	stop     chan bool
}

func NewTTLCleaner(duration time.Duration, interval time.Duration) *TTLCleaner {
	return &TTLCleaner{
		ttl:      duration,
		interval: interval,
	}
}

func (cleaner *TTLCleaner) Clean(cache *TTLCache) {
	for key, value := range cache.CacheMap {
		if time.Now().After(value.EntryTime.Add(cache.TTLExpiry)) {
			fmt.Println("Removing expired entry", key)
			cache.Evict(key)
		}
	}
}

func (cleaner *TTLCleaner) Run(cache *TTLCache) {
	timer := time.NewTimer(cleaner.interval)
	for {
		select {
		case <-timer.C:
			fmt.Println("Running TTLCleaner at ", time.Now())
			cleaner.Clean(cache)
		case <-cleaner.stop:
			timer.Stop()
		}
	}
}
