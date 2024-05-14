package cache

import (
	"cache/internal/domain"
	"errors"
	"fmt"
	"sync"
)

type LFUCache struct {
	CacheMap map[domain.Key]*domain.FreqListNode
	FreqMap  map[int]*domain.FreqListNode
	minLevel int
	capacity int
	lock     sync.RWMutex
}

func NewCache(capacity int) domain.Cache {
	return &LFUCache{
		CacheMap: make(map[domain.Key]*domain.FreqListNode),
		FreqMap:  make(map[int]*domain.FreqListNode),
		minLevel: 0,
		capacity: capacity,
		lock:     sync.RWMutex{},
	}
}

func (cache *LFUCache) Put(k domain.Key, val domain.Key) {
	cache.lock.Lock()
	fmt.Print("adding cache entry ", k)
	if len(cache.CacheMap) == cache.capacity {
		cache.EvictKey()
	}
	cache.minLevel = 1
	cache.CacheMap[k] = cache.createNode(k, val)
	cache.lock.Unlock()
}

func (cache *LFUCache) Get(K domain.Key) domain.Key {
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	currNode, ok := cache.CacheMap[K]
	if !ok {
		return errors.New("keys doent exsists")
	}
	nodeFreq := cache.updateNode(currNode)
	return nodeFreq.Val
}

func (cache *LFUCache) GetAllCacheData() {
	fmt.Printf("Printing All Cache Keys and Value pairs\n")
	for key, val := range cache.CacheMap {
		fmt.Println(key, " , ", val.Freq)
	}
	cache.PrintFreqWiseCachedData()
}

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

func (cache *LFUCache) createNode(k domain.Key, val domain.Key) *domain.FreqListNode {
	node := &domain.FreqListNode{
		ListNode: &domain.ListNode{Val: val, Key: k},
		Freq:     1,
		Next:     nil,
		Prev:     nil,
	}
	prevNode, ok := cache.FreqMap[1]
	node.Next = prevNode
	if ok {
		prevNode.Prev = node
	}
	cache.FreqMap[1] = node
	return node
}

func (cache *LFUCache) updateNode(cacheNode *domain.FreqListNode) *domain.FreqListNode {
	fmt.Println("calling update node for node freq", cacheNode.Key)
	removedNode := removeNodeFromList(cache, cacheNode)
	if removedNode == nil {
		delete(cache.FreqMap, cacheNode.Freq)
		cache.minLevel = cache.minLevel + 1
	}
	nextFreqNode, ok := cache.FreqMap[cacheNode.Freq+1]
	cacheNode.Freq = cacheNode.Freq + 1
	if !ok {
		cache.FreqMap[cacheNode.Freq] = cacheNode
	} else {
		nextFreqNode.Prev = cacheNode
		cacheNode.Next = nextFreqNode
		cacheNode.Prev = nil
	}
	fmt.Println("update details for node", cacheNode)
	cache.FreqMap[cacheNode.Freq] = cacheNode
	cache.PrintFreqWiseCachedData()
	return cacheNode
}

func removeNodeFromList(cache *LFUCache, node *domain.FreqListNode) *domain.FreqListNode {
	prevNode := node.Prev
	nextNode := node.Next
	if prevNode != nil {
		prevNode.Next = node.Next
	}

	if nextNode != nil {
		nextNode.Prev = node.Prev
		if prevNode == nil {
			cache.FreqMap[node.Freq] = nextNode
		}
	}
	node.Prev = nil
	node.Next = nil

	if prevNode == nil {
		return nextNode
	}
	return prevNode
}

func (cache *LFUCache) EvictKey() {
	fmt.Println()
	fmt.Println("Evicting Key")
	fmt.Println()
	fmt.Println("min level ", cache.minLevel)

	minFreqList := cache.FreqMap[cache.minLevel]
	fmt.Println()
	newNode := removeNodeFromList(cache, minFreqList)
	delete(cache.CacheMap, minFreqList.Key)
	if newNode == nil {
		delete(cache.FreqMap, cache.minLevel)
		cache.minLevel++
	}
}
