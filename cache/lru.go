package cache

import (
	"fmt"
	dll "github.com/emirpasic/gods/lists/doublylinkedlist"
	"sync"
)

type lruCache[Key comparable, Val any] struct {
	capacity int
	store    map[Key]Val
	order    *dll.List
	mu       sync.Mutex
}

func NewLRU[Key comparable, Val any](cap int) (*lruCache[Key, Val], error) {
	if cap <= 0 {
		return nil, fmt.Errorf("capacity must be greater than zero")
	}
	return &lruCache[Key, Val]{
		capacity: cap,
		store:    make(map[Key]Val),
		order:    dll.New(),
	}, nil
}

func (c *lruCache[Key, Val]) Get(k Key) (Val, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.store[k]; ok {
		c.recentify(k)
		return v, true
	}
	var z Val
	return z, false
}

func (c *lruCache[Key, Val]) Put(k Key, v Val) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.store[k]; ok {
		c.store[k] = v
		c.recentify(k)
	} else {
		if len(c.store) == c.capacity {
			c.evict()
		}
		c.store[k] = v
		c.order.Add(k)
	}
}

func (c *lruCache[Key, Val]) evict() {
	t, _ := c.order.Get(0)
	c.order.Remove(0)
	delete(c.store, t.(Key))
}

func (c *lruCache[Key, Val]) recentify(k Key) {
	c.order.Remove(c.order.IndexOf(k))
	c.order.Add(k)
}
