package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type cacheEntry[Key comparable, Val any] struct {
	key         Key
	val         Val
	lastVisited time.Time
}

type TTLCache[Key comparable, Val any] struct {
	store         map[Key]*cacheEntry[Key, Val]
	timeToLive    time.Duration
	resetOnAccess bool
	mu            sync.Mutex
}

func NewTTL[Key comparable, Val any](ttl time.Duration, roa bool) (*TTLCache[Key, Val], error) {
	if ttl <= 0 {
		return nil, fmt.Errorf("ttl must be greater than zero.")
	}

	return &TTLCache[Key, Val]{
		store:         make(map[Key]*cacheEntry[Key, Val]),
		timeToLive:    ttl,
		resetOnAccess: roa,
	}, nil
}

func (c *TTLCache[Key, Val]) Get(k Key) (Val, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.store[k]; ok {
		if c.resetOnAccess {
			e.lastVisited = time.Now()
		}
		return e.val, true
	}
	var z Val
	return z, false
}

func (c *TTLCache[Key, Val]) Put(k Key, v Val) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.store[k]; ok {
		e.lastVisited = time.Now()
		e.val = v
	} else {
		e := &cacheEntry[Key, Val]{
			key:         k,
			val:         v,
			lastVisited: time.Now(),
		}
		c.store[k] = e
	}
}

func (c *TTLCache[Key, Val]) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.store)
}

func (c *TTLCache[Key, Val]) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, e := range c.store {
		if time.Since(e.lastVisited) >= c.timeToLive {
			delete(c.store, k)
		}
	}
}

func (c *TTLCache[Key, Val]) ScheduleCleanup(ctx context.Context, e time.Duration) {
	go func() {
		ticker := time.NewTicker(e)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.Cleanup()
			case <-ctx.Done():
				return
			}
		}
	}()
}
