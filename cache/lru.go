package cache

import (
	"fmt"
	"slices"
	"sync"
)

// TODO: use a LRUEntry struct with (key-val pair, and lastUsed)
// instead of order slice for beter performance

type lruCache[Key comparable, Val any] struct {
	capacity int
	store    map[Key]Val
	order    []Key
	mu       sync.Mutex
}

func NewLRU[Key comparable, Val any](cap int) (Cache[Key, Val], error) {
	if cap <= 0 {
		return nil, fmt.Errorf("capacity must be greater than zero")
	}
	return &lruCache[Key, Val]{
		capacity: cap,
		store:    make(map[Key]Val),
		order:    make([]Key, 0, cap),
	}, nil
}

func (c *lruCache[Key, Val]) Get(k Key) (Val, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if val, ok := c.store[k]; ok {
		c.recentify(k)
		return val, true
	}
	var z Val
	return z, false
}

func (c *lruCache[Key, Val]) Put(k Key, v Val) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.store[k]; ok {
		c.recentify(k)
		c.store[k] = v
	} else {
		if len(c.store) == c.capacity {
			evc := c.evict()
			delete(c.store, evc)
		}
		c.recentify(k)
		c.store[k] = v
	}
}

func (c *lruCache[Key, Val]) evict() Key {
	evc := c.order[0]
	c.order = c.order[1:]
	return evc
}

func (c *lruCache[Key, Val]) recentify(k Key) {
	// this func deleted all elements that satisfy the condition
	// but any key will appear once (keys are unique)
	// so it will do only one deletion
	c.order = slices.DeleteFunc(c.order, func(e Key) bool { return e == k })
	c.order = append(c.order, k)
}
