# Cache System README

## Overview

This go-cache cache system provides a flexible and efficient way to store and manage key-value pairs in memory. It supports various cache algorithms, including Least Recently Used (LRU), Least Frequently Used (LFU), and Time-To-Live (TTL) caching.

## Features

- **Multiple Cache Types:** Supports different cache types, including LRU, LFU, and TTL.
- **Flexible Configuration:** Allows customization of cache capacity and TTL duration.
- **Thread-Safe Operations:** Ensures safe concurrent access to cache data structures.
- **Easy Integration:** Simple API for adding, retrieving, and removing items from the cache.
- **Customizable Eviction:** Provides flexibility in handling cache eviction policies.

## Components

The cache system consists of the following main components:

1. **Cache Types:**
    - **LRU Cache:** Implements the Least Recently Used caching algorithm.
    - **LFU Cache:** Implements the Least Frequently Used caching algorithm.
    - **TTL Cache:** Implements the Time-To-Live caching algorithm.

2. **Cache Factory:**
    - `CacheFactory` allows easy creation of different cache types with specified configurations.

3. **Cache Cleaner:**
    - `TTLCleaner` periodically cleans expired entries from the TTL cache.

4. **Data Structures:**
    - `DoubleLinkedList`: Represents a doubly linked list used in cache implementations.
    - `ListNode`: Represents a node in a linked list.
    - `FreqListNode`: Represents a node in a frequency list, extending `ListNode` with frequency information.

## Usage

1. **Creating Caches:**
    - Use the `CacheFactory` to create caches of desired types with specified capacities.

2. **Adding Items:**
    - Use the `Put` method to add key-value pairs to the cache.

3. **Retrieving Items:**
    - Use the `Get` method to retrieve the value associated with a key from the cache.

4. **Removing Items:**
    - Use the `Evict` method to remove an item from the cache based on its key.

5. **Cleaning TTL Cache:**
    - The TTL cache automatically cleans expired entries using the `TTLCleaner`.

## Example

```go
package main

import (
	"cache"
	_interface "cache/internal/domain/interface"
	"fmt"
	"time"
)

func main() {
	// Create an LRU cache with a capacity of 100.
	lruCache := cache.NewCache("LRU", 100)

	// Add items to the cache.
	lruCache.Put("key1", "value1")
	lruCache.Put("key2", "value2")
	lruCache.Put("key3", "value3")

	// Retrieve items from the cache.
	val1 := lruCache.Get("key1")
	val2 := lruCache.Get("key2")

	fmt.Println("Value 1:", val1)
	fmt.Println("Value 2:", val2)

	// Create a TTL cache with a capacity of 50 and TTL duration of 1 minute.
	ttlCache := cache.NewTTLCache(50, time.Minute)

	// Add items to the TTL cache.
	ttlCache.Put("key1", "value1", time.Minute)
	ttlCache.Put("key2", "value2", time.Minute)
	ttlCache.Put("key3", "value3", time.Minute)

	// Wait for some time to let TTL expire for items.
	time.Sleep(2 * time.Minute)

	// Retrieve items from the TTL cache (expired items should not be found).
	val1 = ttlCache.Get("key1")
	val2 = ttlCache.Get("key2")

	fmt.Println("Value 1 after expiry:", val1)
	fmt.Println("Value 2 after expiry:", val2)
}
