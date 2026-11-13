package cache

import (
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type Cache[T any] interface {
	Get(key string) (T, bool)
	Set(key string, value T)
}

type item[T any] struct {
	value     T
	expiresAt time.Time
}

type lruCache[T any] struct {
	entries *lru.Cache[string, item[T]]
	ttl     time.Duration
}

func New[T any](maxEntries int, ttl time.Duration) (Cache[T], error) {
	entries, err := lru.New[string, item[T]](maxEntries)
	if err != nil {
		return nil, err
	}

	return &lruCache[T]{
		entries: entries,
		ttl:     ttl,
	}, nil
}

func (c *lruCache[T]) Get(key string) (T, bool) {
	entry, exists := c.entries.Get(key)
	if !exists {
		var zero T
		return zero, false
	}

	if time.Now().After(entry.expiresAt) {
		c.entries.Remove(key)

		var zero T
		return zero, false
	}

	return entry.value, true
}

func (c *lruCache[T]) Set(key string, value T) {
	c.entries.Add(key, item[T]{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	})
}
