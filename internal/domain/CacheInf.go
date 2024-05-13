package domain

type Cache interface {
	GetAllCacheData()
	Put(k Key, val Key)
	Get(K Key) Key
	EvictKey()
}
