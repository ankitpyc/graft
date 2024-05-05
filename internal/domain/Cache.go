package domain

import "fmt"

type Cache struct {
	CacheMap map[Key]*ListNode
	List     *DoubleLinkedList
	capacity int
}

func NewCache(capacity int) Cache {
	return Cache{
		CacheMap: make(map[Key]*ListNode),
		capacity: capacity,
		List: &DoubleLinkedList{
			Head: nil,
			Tail: nil,
		},
	}
}

func (cache *Cache) Put(k Key, val Key) {
	fmt.Printf("Cache size %d cache Capacity %d", len(cache.CacheMap), cache.capacity)
	fmt.Println()
	if len(cache.CacheMap) == cache.capacity {
		cache.evictKey()
	}
	node := &ListNode{
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

func (cache *Cache) Get(k Key) Key {
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

func (cache *Cache) evictKey() {
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
