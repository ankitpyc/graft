package _interface

import "cache/internal/store"

type Cache interface {
	store.StoreInf
	GetAllCacheData()
	EvictKey()
}
