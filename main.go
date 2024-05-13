package main

import (
	cacheFactory "cache/internal/domain/factory"
	"log"
)

func main() {

	var cache, err = cacheFactory.CreateCache("LRiU", 5)
	if err != nil {
		log.Fatal("Invalid Cache type")
	}
	cache.Put(2, 2)
	cache.Put("2", "ankit")
	cache.Put("hello", "mishra")
	cache.Get("2")
	cache.Get(2)
	cache.Put(4, 3)
	cache.GetAllCacheData()

}
