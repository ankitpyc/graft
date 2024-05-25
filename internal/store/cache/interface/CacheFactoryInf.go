package _interface

// CacheFactory represents a factory for creating caches.
type CacheFactory struct {
	CacheType string // CacheType represents the type of cache to create.
	Capacity  int    // Capacity represents the maximum capacity of the cache.
}
