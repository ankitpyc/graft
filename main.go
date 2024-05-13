package main

import (
	cacheFactory "cache/internal/domain/factory"
	"log"
)

func main() {
	var cache, err = cacheFactory.CreateCache("LFU", 4)
	if err != nil {
		log.Fatal("Invalid Cache type")
	}
	cache.Put(2, "TWO")
	cache.Put(5, "ankit")
	cache.Put("hello", "mishra")
	cache.Put(7, "snkit")
	cache.Get(7)
	cache.Get(5)
	cache.Get(2)
	cache.Put(9, 9)
	cache.Get(9)
	cache.Put("ankit", "data1")
	cache.Put("monica", "data2")
	cache.Put("monisha", "data3")
	cache.PrintLevelCacheData()

}
