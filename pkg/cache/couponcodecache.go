package cache

import (
	"container/list"
	"sync"
)

type LRUCache struct {
	mu        sync.Mutex
	maxSize   int
	items     map[string]*list.Element
	evictList *list.List // most recent → the least recent
}

type entry struct {
	key   string
	value bool
}

func NewLRUCache(maxSize int) *LRUCache {
	return &LRUCache{
		maxSize:   maxSize,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

// Get looks up a key in the cache.
//
// If the key exists, it returns the associated value and true.
// The entry is also marked as most recently used (moved to the front of the LRU list).
//
// If the key does not exist, it returns false and false.
func (c *LRUCache) Get(key string) (bool, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		c.evictList.MoveToFront(el)
		return el.Value.(*entry).value, true
	}
	return false, false
}

func (c *LRUCache) Set(key string, value bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		c.evictList.MoveToFront(el)
		el.Value.(*entry).value = value
		return
	}

	el := c.evictList.PushFront(&entry{key: key, value: value})
	c.items[key] = el

	if c.evictList.Len() > c.maxSize {
		c.removeOldest()
	}
}

func (c *LRUCache) removeOldest() {
	el := c.evictList.Back()
	if el != nil {
		c.evictList.Remove(el)
		ent := el.Value.(*entry)
		delete(c.items, ent.key)
	}
}
