package cache

import (
	"sync"
	"time"
)

// entry holds a cached value with its expiration time.
type entry struct {
	value     any
	expiresAt time.Time
}

func (e *entry) expired() bool {
	return time.Now().After(e.expiresAt)
}

// Cache is a thread-safe in-memory cache with TTL support.
type Cache struct {
	mu         sync.RWMutex
	items      map[string]*entry
	defaultTTL time.Duration
	hits       int64
	misses     int64
	stopClean  chan struct{}
	stopOnce   sync.Once
}

// New creates a cache with the given default TTL.
// A background goroutine cleans expired entries every cleanupInterval.
func New(defaultTTL, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		items:      make(map[string]*entry),
		defaultTTL: defaultTTL,
		stopClean:  make(chan struct{}),
	}
	go c.cleanupLoop(cleanupInterval)
	return c
}

// Get retrieves a value by key. Returns nil and false if not found or expired.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	if e.expired() {
		c.mu.Lock()
		delete(c.items, key)
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
	return e.value, true
}

// Set stores a value with the default TTL.
func (c *Cache) Set(key string, value any) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value with a custom TTL.
func (c *Cache) SetWithTTL(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	c.items[key] = &entry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

// Invalidate removes a single key from the cache.
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// InvalidatePrefix removes all keys that start with the given prefix.
func (c *Cache) InvalidatePrefix(prefix string) {
	c.mu.Lock()
	for k := range c.items {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}

// InvalidateAll removes all entries from the cache.
func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	c.items = make(map[string]*entry)
	c.mu.Unlock()
}

// Len returns the number of items in the cache (including expired ones not yet cleaned).
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Stats returns hit and miss counts.
func (c *Cache) Stats() (hits, misses int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses
}

// ResetStats resets hit and miss counters.
func (c *Cache) ResetStats() {
	c.mu.Lock()
	c.hits = 0
	c.misses = 0
	c.mu.Unlock()
}

// SetDefaultTTL updates the default TTL for new cache entries.
func (c *Cache) SetDefaultTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultTTL = ttl
}

// Stop terminates the background cleanup goroutine.
// Safe to call multiple times.
func (c *Cache) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopClean)
	})
}

func (c *Cache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopClean:
			return
		}
	}
}

func (c *Cache) cleanup() {
	c.mu.Lock()
	now := time.Now()
	for k, e := range c.items {
		if now.After(e.expiresAt) {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}
