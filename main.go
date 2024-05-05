package main

import (
	cacheModel "cache/internal/domain"
	"fmt"
)

func main() {

	var cache = cacheModel.NewCache(3)
	cache.Put(2, 2)
	cache.Put("2", "ankit")
	cache.Put("hello", "mishra")
	cache.Get("2")
	cache.Get(2)
	cache.Put(4, 3)

	for key, val := range cache.CacheMap {
		fmt.Print(key, " - ", val.Val)
		fmt.Println()
	}
}
