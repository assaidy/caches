package cache

type Cache[Key comparable, Val any] interface {
	Get(Key) (Val, bool)
	Put(Key, Val)
	evict() Key
}
