package cache

import "sync"

type lfuCache[Key comparable, Val any] struct {
	store map[Key]Val
	mu    sync.Mutex
    // TODO: 
}

func (c *lfuCache[Key, Val]) Get(k Key) (Val, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // TODO: 

	var z Val
	return z, false
}

func (c *lfuCache[Key, Val]) Put(k Key, v Val) {
    c.mu.Lock()
    defer c.mu.Unlock()
    // TODO:
}

func (c *lfuCache[Key, Val]) evict() {
    // TODO:
}
