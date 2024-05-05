package domain

type CacheInf interface {
	Put(k Key, val Key)
	Get(K Key) Key
	evictKey()
}
