package _interface

import "cache/internal/domain"

type TTLCacheInf interface {
	Cache
	Evict(key domain.Key)
}
