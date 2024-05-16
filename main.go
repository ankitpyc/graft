package main

import (
	cacheFactory "cache/internal/domain/factory"
	"fmt"
	"log"
	"time"
)

func main() {
	var cache, err = cacheFactory.CreateCache("TTL", 5)
	if err != nil {
		log.Fatal("Invalid Cache type")
	}
	cache.Put(2, "TWO")
	cache.Put(5, "ankit")
	cache.Put("hello", "mishra")
	cache.Put(7, "snkit")
	time.Sleep(12 * time.Second)
	fmt.Println("Thread woke up")
	cache.Get(7)
	cache.Get(5)
	cache.Get(2)
	cache.Put(9, 9)
	cache.Get(9)
	cache.Put("ankit", "data1")
	cache.Get("ankit")
	cache.Get("ankit")
	cache.GetAllCacheData()
	time.Sleep(1 * time.Hour)
}
