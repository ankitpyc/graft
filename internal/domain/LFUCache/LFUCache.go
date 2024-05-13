package cache

import (
	"cache/internal/domain"
	"errors"
	"fmt"
)

type LFUCache struct {
	CacheMap map[domain.Key]*domain.FreqListNode
	FreqMap  map[int]*domain.FreqListNode
	capacity int
}

func CreateCache(capacity int) domain.Cache {
	return &LFUCache{
		CacheMap: make(map[domain.Key]*domain.FreqListNode),
		FreqMap:  make(map[int]*domain.FreqListNode),
		capacity: capacity,
	}
}

func (cache LFUCache) Put(k domain.Key, val domain.Key) {
	if len(cache.CacheMap) == cache.capacity {
		cache.EvictKey()
	}
	cache.CacheMap[k] = cache.createNode(k, val)
}

func (cache LFUCache) Get(K domain.Key) domain.Key {
	currNode, ok := cache.CacheMap[K]
	if !ok {
		return errors.New("keys doent exsists")
	}
	nodeFreq := cache.updateNode(currNode)
	return nodeFreq.Val
}

func (cache LFUCache) GetAllCacheData() {
	for key, val := range cache.CacheMap {
		fmt.Println(key, " , ", val)
	}
}

func (cache LFUCache) createNode(k domain.Key, val domain.Key) *domain.FreqListNode {
	node := &domain.FreqListNode{
		ListNode: &domain.ListNode{
			Val: val,
			Key: k,
		},
		Freq: 1,
		Next: nil,
		Prev: nil,
	}
	prevNode, ok := cache.FreqMap[1]
	node.Next = prevNode
	if ok {
		prevNode.Prev = node
	}
	cache.FreqMap[1] = node
	return node
}

func (cache LFUCache) updateNode(cacheNode *domain.FreqListNode) *domain.FreqListNode {
	node, _ := cache.CacheMap[cacheNode.Freq]
	removeNodeFromList(&cache, node)
	headNode, _ := cache.FreqMap[node.Freq+1]
	node.Freq = node.Freq + 1
	if headNode == nil {
		cache.FreqMap[node.Freq+1] = node
	} else {
		headNode.Prev = node
		node.Next = headNode
	}
	cache.FreqMap[1] = node
	return node
}

func removeNodeFromList(cache *LFUCache, node *domain.FreqListNode) {
	prevNode := node.Prev
	nextNode := node.Next
	if prevNode != nil {
		prevNode.Next = node.Next
	}

	if nextNode != nil {
		nextNode.Prev = node.Prev
	}

}

func (cache LFUCache) EvictKey() {}
