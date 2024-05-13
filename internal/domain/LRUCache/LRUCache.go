package domain

import (
	"cache/internal/domain"
	"fmt"
)

type LRUCache struct {
	CacheMap map[domain.Key]*domain.ListNode
	List     *domain.DoubleLinkedList
	capacity int
}

func NewCache(capacity int) LRUCache {
	return LRUCache{
		CacheMap: make(map[domain.Key]*domain.ListNode),
		capacity: capacity,
		List: &domain.DoubleLinkedList{
			Head: nil,
			Tail: nil,
		},
	}
}

func (cache LRUCache) Put(k domain.Key, val domain.Key) {
	fmt.Printf("Cache size %d cache Capacity %d", len(cache.CacheMap), cache.capacity)
	fmt.Println()
	if len(cache.CacheMap) == cache.capacity {
		cache.EvictKey()
	}
	node := &domain.ListNode{
		Val:  val,
		Key:  k,
		Prev: nil,
		Next: nil,
	}
	cache.CacheMap[k] = node
	if cache.List.Head == nil {
		cache.List.Head = node
		cache.List.Tail = node
		return
	}
	node.Prev = cache.List.Tail
	cache.List.Tail.Next = node
	cache.List.Tail = node
}

func (cache LRUCache) Get(k domain.Key) domain.Key {
	node, ok := cache.CacheMap[k]
	if !ok {
		return nil
	}
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

func (cache LRUCache) GetAllCacheData() {
	for key, val := range cache.CacheMap {
		fmt.Println(key, " , ", val.Val)
	}
}

func (cache LRUCache) PrintLevelCacheData() {

}

func (cache LRUCache) EvictKey() {
	headNode := cache.List.Head
	fmt.Print("evicting key ", headNode.Key)
	fmt.Println()
	delete(cache.CacheMap, headNode.Key)
	if cache.List.Head.Next != nil {
		cache.List.Head.Next.Prev = nil
		cache.List.Head.Next = nil
	}
	headNode = nil
}
