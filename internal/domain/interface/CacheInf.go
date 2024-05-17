package _interface

import "cache/internal/domain"

type Cache interface {
	GetAllCacheData()
	Put(k domain.Key, val domain.Key)
	Get(K domain.Key) domain.Key
	EvictKey()
	Delete(k domain.Key)
}
