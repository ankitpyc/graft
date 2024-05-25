package store

import "cache/internal/domain"

type StoreInf interface {
	Set(key string, value string)
	Get(key string) (domain.Key, bool)
	Delete(key string) bool
}
