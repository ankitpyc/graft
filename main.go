package main

import (
	cacheFactory "cache/internal/domain/factory"
	"log"
)

func main() {
	var cache, err = cacheFactory.CreateCache("LFU", 5)
	if err != nil {
		log.Fatal("Invalid Cache type")
	}
	go cache.Put(2, "TWO")
	go cache.Put(5, "ankit")
	go cache.Put("hello", "mishra")
	cache.Put(7, "snkit")
	cache.Get(7)
	go cache.Get(5)
	go cache.Get(2)
	cache.Put(9, 9)
	cache.Get(9)
	cache.Put("ankit", "data1")
	cache.Get("ankit")
	cache.Get("ankit")
	cache.GetAllCacheData()

}
