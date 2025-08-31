package cache

import (
	"container/list"
	"sync"
	"time"
)

// LRUCache is a thread-safe generic LRU cache.
type LRUCache[V any] struct {
	mu        sync.Mutex
	maxSize   int
	cacheTTL  time.Duration
	items     map[string]*list.Element
	evictList *list.List // most recent → the least recent
}

type entry[V any] struct {
	key       string
	value     V
	expiresAt time.Time
}

// NewLRUCache creates a new LRU cache with the given max size.
func NewLRUCache[V any](maxSize int, ttl time.Duration) *LRUCache[V] {
	return &LRUCache[V]{
		maxSize:   maxSize,
		cacheTTL:  ttl,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

// Get retrieves a value from the cache and marks it as recently used.
func (c *LRUCache[V]) Get(key string) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ele)
		return ele.Value.(*entry[V]).value, true
	}

	var zero V
	return zero, false
}

// Put inserts or updates a value in the cache.
func (c *LRUCache[V]) Put(key string, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ele)
		ele.Value.(*entry[V]).value = value
		return
	}

	// Add new entry
	ele := c.evictList.PushFront(&entry[V]{key, value, time.Now().Add(c.cacheTTL)})
	c.items[key] = ele

	if c.evictList.Len() > c.maxSize {
		c.removeOldest()
	}
}

// removeOldest evicts the least recently used item.
func (c *LRUCache[V]) removeOldest() {
	ele := c.evictList.Back()
	if ele != nil {
		c.evictList.Remove(ele)
		ent := ele.Value.(*entry[V])
		delete(c.items, ent.key)
	}
}
