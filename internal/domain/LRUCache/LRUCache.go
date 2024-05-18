package cache

import (
	"cache/internal/domain"
	"fmt"
	"sync"
)

// LRUCache implements a Least Recently Used (LRU) cache algorithm.
type LRUCache struct {
	Store    map[domain.Key]*domain.ListNode // Store stores key-value pairs along with their associated list nodes.
	List     *domain.DoubleLinkedList        // List is a doubly linked list used to maintain the LRU order of cache entries.
	capacity int                             // capacity represents the maximum number of items the cache can hold.
	mu       sync.RWMutex                    // mutex to provide thread safe operations
}

// NewCache creates a new instance of LRUCache with the specified capacity.
func NewCache(capacity int) *LRUCache {
	return &LRUCache{
		Store:    make(map[domain.Key]*domain.ListNode),
		capacity: capacity,
		List: &domain.DoubleLinkedList{
			Head: nil,
			Tail: nil,
		},
	}
}

// Put adds a new key-value pair to the cache.
// If the cache is at full capacity, it evicts the least recently used item before adding the new one.
func (cache *LRUCache) Put(k domain.Key, val domain.Key) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	fmt.Printf("Cache size %d cache Capacity %d", len(cache.Store), cache.capacity)
	// Evict the least recently used item if the cache is at full capacity.
	if len(cache.Store) == cache.capacity {
		cache.EvictKey()
	}

	// Create a new list node for the key-value pair.
	node := &domain.ListNode{
		Val:  val,
		Key:  k,
		Prev: nil,
		Next: nil,
	}

	// Update the Store with the new node.
	cache.Store[k] = node

	// Insert the new node at the tail of the linked list.
	if cache.List.Head == nil {
		cache.List.Head = node
		cache.List.Tail = node
		return
	}
	node.Prev = cache.List.Tail
	cache.List.Tail.Next = node
	cache.List.Tail = node
}

// Get retrieves the value associated with the given key from the cache.
// If the key doesn't exist in the cache, it returns nil.
// If the key exists, it promotes the corresponding node to the tail of the linked list (indicating recent use).
func (cache *LRUCache) Get(k domain.Key) domain.Key {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	node, ok := cache.Store[k]
	if !ok {
		return nil
	}

	// Move the accessed node to the tail of the linked list (indicating recent use).
	if node == cache.List.Head {
		cache.List.Head = cache.List.Head.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	}

	if node.Prev != nil {
		node.Prev.Next = node.Next
	}

	node.Prev = cache.List.Tail
	cache.List.Tail.Next = node
	cache.List.Tail = node

	return node.Val
}

// GetAllCacheData prints all keys along with their corresponding values in the cache.
func (cache *LRUCache) GetAllCacheData() {
	fmt.Println()
	for key, val := range cache.Store {
		fmt.Println(key, " , ", val.Val)
	}
}

// PrintLevelCacheData prints cached keys grouped by their levels.
// For LRUCache, since there are no levels, this function does nothing.
func (cache *LRUCache) PrintLevelCacheData() {
	// No levels to print for LRUCache.
}

// EvictKey evicts the least recently used key from the cache.
func (cache *LRUCache) EvictKey() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	headNode := cache.List.Head

	// Remove the least recently used node from the cache.
	delete(cache.Store, headNode.Key)

	// Update the head of the linked list.
	if cache.List.Head.Next != nil {
		cache.List.Head.Next.Prev = nil
		cache.List.Head.Next = nil
	}
	headNode = nil
}
func (cache *LRUCache) Delete(key domain.Key) {
	delete(cache.Store, key)
}
