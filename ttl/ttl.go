package ttl

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// cacheEntry represents an individual entry in the cache.
// It stores the value and metadata about when it was last accessed.
type cacheEntry[Key comparable, Val any] struct {
	val         Val       // The cached value.
	lastVisited time.Time // Timestamp of the last access, used for TTL calculations.
}

// TTLCache is a generic Time-To-Live cache that stores key-value pairs
// with optional reset-on-access behavior and supports automatic cleanup.
type TTLCache[Key comparable, Val any] struct {
	storege        map[Key]*cacheEntry[Key, Val] // Underlying storage for the cache.
	timeToLive     time.Duration                 // Duration for which an entry remains valid.
	resetOnRead    bool                          // If true, reading a value resets its TTL.
	mu             sync.RWMutex                  // RWMutex for concurrent reads and exclusive writes.
	cleanupRunning bool                          // Indicates if a cleanup routine is running.
	cleanupMu      sync.Mutex                    // Mutex to synchronize cleanup scheduling.
}

// NewTTL creates a new TTLCache with the specified TTL duration and reset-on-access behavior.
// If ror is true, last-access-time of an entry will be reset after when entry is read.
// Returns an error if the TTL duration is non-positive.
func NewTTL[Key comparable, Val any](ttl time.Duration, ror bool) (*TTLCache[Key, Val], error) {
	if ttl <= 0 {
		return nil, fmt.Errorf("ttl must be greater than zero.")
	}

	return &TTLCache[Key, Val]{
		storege:     make(map[Key]*cacheEntry[Key, Val]),
		timeToLive:  ttl,
		resetOnRead: ror,
	}, nil
}

// Get retrieves the value associated with the given key from the cache.
// If the key exists and is valid, it returns the value and true.
// If the key is not found or has expired, it returns the zero value of Val and false.
func (c *TTLCache[Key, Val]) Get(k Key) (Val, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.storege[k]
	if !ok {
		c.mu.RUnlock()
		var z Val
		return z, false
	}

	if time.Since(e.lastVisited) > c.timeToLive {
		c.mu.RUnlock()

		c.mu.Lock()
		delete(c.storege, k)
		c.mu.Unlock()

		var z Val
		return z, false
	}

	if c.resetOnRead {
		c.mu.RUnlock()

		c.mu.Lock()
		e.lastVisited = time.Now()
		c.mu.Unlock()

		c.mu.RLock()
	}

	defer c.mu.RUnlock()
	return e.val, true
}

// Put inserts or updates the value associated with the given key in the cache.
// If the key already exists, its value and last access time are updated.
// If it is a new key, it is added to the cache.
func (c *TTLCache[Key, Val]) Put(k Key, v Val) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.storege[k]; ok {
		e.lastVisited = time.Now()
		e.val = v
	} else {
		e := &cacheEntry[Key, Val]{
			val:         v,
			lastVisited: time.Now(),
		}
		c.storege[k] = e
	}
}

// Size returns the current number of entries in the cache.
func (c *TTLCache[Key, Val]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.storege)
}

// Cleanup removes expired entries from the cache.
// An entry is considered expired if its last access time exceeds the TTL duration.
func (c *TTLCache[Key, Val]) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, e := range c.storege {
		if time.Since(e.lastVisited) >= c.timeToLive {
			delete(c.storege, k)
		}
	}
}

// ScheduleCleanup starts a periodic cleanup routine in a separate goroutine.
// The cleanup runs at the specified interval and stops when the given context is canceled.
// Only one cleanup routine can run at a time;
// additional calls to this method will be ignored until the current routine is canceled.
func (c *TTLCache[Key, Val]) ScheduleCleanup(ctx context.Context, e time.Duration) {
	c.cleanupMu.Lock()
	defer c.cleanupMu.Unlock()

	if c.cleanupRunning {
		return
	}

	c.cleanupRunning = true

	go func() {
		ticker := time.NewTicker(e)
		defer func() {
			ticker.Stop()
			c.cleanupMu.Lock()
			c.cleanupRunning = false
			c.cleanupMu.Unlock()
		}()

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
