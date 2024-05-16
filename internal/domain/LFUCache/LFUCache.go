package cache

import (
	"cache/internal/domain"
	"cache/internal/domain/interface"
	"errors"
	"fmt"
	"sync"
)

// LFUCache implements a Least Frequently Used (LFU) cache algorithm.
type LFUCache struct {
	// CacheMap stores key-value pairs along with their associated frequency nodes.
	CacheMap map[domain.Key]*domain.FreqListNode

	// FreqMap maps frequency levels to their corresponding frequency nodes.
	FreqMap map[int]*domain.FreqListNode

	// minLevel tracks the minimum frequency level in the cache.
	minLevel int

	// capacity represents the maximum number of items the cache can hold.
	capacity int

	// lock provides synchronization for concurrent access to the cache.
	lock sync.RWMutex
}

// NewCache creates a new instance of LFUCache with the specified capacity.
func NewCache(capacity int) _interface.Cache {
	return &LFUCache{
		CacheMap: make(map[domain.Key]*domain.FreqListNode),
		FreqMap:  make(map[int]*domain.FreqListNode),
		minLevel: 0,
		capacity: capacity,
		lock:     sync.RWMutex{},
	}
}

// Put adds a new key-value pair to the cache.
// If the cache is at full capacity, it evicts the least frequently used item before adding the new one.
func (cache *LFUCache) Put(k domain.Key, val domain.Key) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	fmt.Print("Adding cache entry ", k)

	// Evict the least frequently used item if the cache is at full capacity.
	if len(cache.CacheMap) == cache.capacity {
		cache.EvictKey()
	}

	// Set the minimum frequency level to 1 when adding a new item.
	cache.minLevel = 1

	// Create a new frequency node for the key-value pair and insert it into the cache.
	cache.CacheMap[k] = cache.createNode(k, val)
}

// Get retrieves the value associated with the given key from the cache.
// If the key doesn't exist in the cache, it returns an error.
func (cache *LFUCache) Get(K domain.Key) domain.Key {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	currNode, ok := cache.CacheMap[K]
	if !ok {
		return errors.New("key doesn't exist")
	}

	// Update the frequency of the accessed node and adjust its position in the frequency list.
	nodeFreq := cache.updateNode(currNode)
	return nodeFreq.Val
}

// GetAllCacheData prints all keys along with their corresponding frequency levels.
func (cache *LFUCache) GetAllCacheData() {
	fmt.Printf("Printing All Cache Keys and Value pairs\n")
	for key, val := range cache.CacheMap {
		fmt.Println(key, " , ", val.Freq)
	}
	cache.PrintFreqWiseCachedData()
}

// PrintFreqWiseCachedData prints cached keys grouped by their frequency levels.
func (cache *LFUCache) PrintFreqWiseCachedData() {
	fmt.Println("-------------------- Printing Freq Wise Data -----------------------------")
	fmt.Println()
	for level, nodeList := range cache.FreqMap {
		temp := nodeList
		fmt.Printf("-------------------- Freq %d -----------------------------", level)
		for temp != nil {
			fmt.Print(" ", temp.Key)
			temp = temp.Next
		}
		fmt.Println()
		fmt.Printf("-------------------- Freq %d Finished  ---------------------", level)
	}
}

// createNode creates a new frequency node for the given key-value pair and inserts it into the cache.
func (cache *LFUCache) createNode(k domain.Key, val domain.Key) *domain.FreqListNode {
	node := &domain.FreqListNode{
		ListNode: &domain.ListNode{Val: val, Key: k},
		Freq:     1,
		Next:     nil,
		Prev:     nil,
	}

	// Insert the new node into the frequency list corresponding to frequency level 1.
	prevNode, ok := cache.FreqMap[1]
	node.Next = prevNode
	if ok {
		prevNode.Prev = node
	}
	cache.FreqMap[1] = node

	return node
}

// updateNode updates the frequency of the given cache node and adjusts its position in the frequency list.
func (cache *LFUCache) updateNode(cacheNode *domain.FreqListNode) *domain.FreqListNode {
	fmt.Println("calling update node for node freq", cacheNode.Key)

	// Remove the node from its current position in the frequency list.
	removedNode := removeNodeFromList(cache, cacheNode)

	// If the node was the only one at its previous frequency level, update the minimum frequency level.
	if removedNode == nil {
		delete(cache.FreqMap, cacheNode.Freq)
		cache.minLevel = cache.minLevel + 1
	}

	// Increase the frequency of the accessed node.
	nextFreqNode, ok := cache.FreqMap[cacheNode.Freq+1]
	cacheNode.Freq = cacheNode.Freq + 1
	if !ok {
		cache.FreqMap[cacheNode.Freq] = cacheNode
	} else {
		nextFreqNode.Prev = cacheNode
		cacheNode.Next = nextFreqNode
		cacheNode.Prev = nil
	}

	// Update the cache node in the FreqMap.
	cache.FreqMap[cacheNode.Freq] = cacheNode
	cache.PrintFreqWiseCachedData()

	return cacheNode
}

// removeNodeFromList removes the given node from its current position in the frequency list.
func removeNodeFromList(cache *LFUCache, node *domain.FreqListNode) *domain.FreqListNode {
	prevNode := node.Prev
	nextNode := node.Next

	// Adjust the pointers of adjacent nodes to remove the current node from the list.
	if prevNode != nil {
		prevNode.Next = node.Next
	}

	if nextNode != nil {
		nextNode.Prev = node.Prev
		if prevNode == nil {
			cache.FreqMap[node.Freq] = nextNode
		}
	}

	// Clear the previous and next pointers of the current node.
	node.Prev = nil
	node.Next = nil

	// If the previous node is nil, return the next node as the new head of the list.
	if prevNode == nil {
		return nextNode
	}

	return prevNode
}

// EvictKey evicts the least frequently used key from the cache.
func (cache *LFUCache) EvictKey() {
	fmt.Println()
	fmt.Println("Evicting Key")
	fmt.Println()
	fmt.Println("min level ", cache.minLevel)

	// Find the list of nodes with the minimum frequency level.
	minFreqList := cache.FreqMap[cache.minLevel]
	fmt.Println()

	// Remove the least frequently used node from the cache.
	newNode := removeNodeFromList(cache, minFreqList)
	delete(cache.CacheMap, minFreqList.Key)

	// If the removed node was the only one at its frequency level, update the minimum frequency level.
	if newNode == nil {
		delete(cache.FreqMap, cache.minLevel)
		cache.minLevel++
	}
}
