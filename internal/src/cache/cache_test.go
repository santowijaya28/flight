package cache

import (
	"testing"
	"time"
)

func newTestCache(t *testing.T, maxEntries int, ttl time.Duration) (Cache[string], *time.Time) {
	t.Helper()

	now := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	cache, err := newCache[string](maxEntries, ttl, func() time.Time {
		return now
	})
	if err != nil {
		t.Fatal(err)
	}

	return cache, &now
}

type cacheStep struct {
	after    time.Duration
	setKey   string
	setValue string
	getKey   string
	wantOK   bool
	wantVal  string
}

func TestCache(t *testing.T) {
	tests := []struct {
		name       string
		maxEntries int
		ttl        time.Duration
		steps      []cacheStep
	}{
		{
			name:       "set and get",
			maxEntries: 10,
			ttl:        30 * time.Second,
			steps: []cacheStep{
				{setKey: "key", setValue: "value"},
				{getKey: "key", wantOK: true, wantVal: "value"},
			},
		},
		{
			name:       "missing key",
			maxEntries: 10,
			ttl:        30 * time.Second,
			steps: []cacheStep{
				{getKey: "missing"},
			},
		},
		{
			name:       "valid before ttl",
			maxEntries: 10,
			ttl:        30 * time.Second,
			steps: []cacheStep{
				{setKey: "key", setValue: "value"},
				{after: 29 * time.Second, getKey: "key", wantOK: true, wantVal: "value"},
			},
		},
		{
			name:       "expires after ttl",
			maxEntries: 10,
			ttl:        30 * time.Second,
			steps: []cacheStep{
				{setKey: "key", setValue: "value"},
				{after: 29 * time.Second, getKey: "key", wantOK: true, wantVal: "value"},
				{after: 2 * time.Second, getKey: "key"},
			},
		},
		{
			name:       "cache hit does not extend ttl",
			maxEntries: 10,
			ttl:        30 * time.Second,
			steps: []cacheStep{
				{setKey: "key", setValue: "value"},
				{after: 20 * time.Second, getKey: "key", wantOK: true, wantVal: "value"},
				{after: 11 * time.Second, getKey: "key"},
			},
		},
		{
			name:       "set refreshes ttl",
			maxEntries: 10,
			ttl:        30 * time.Second,
			steps: []cacheStep{
				{setKey: "key", setValue: "old"},
				{after: 20 * time.Second, setKey: "key", setValue: "new"},
				{after: 20 * time.Second, getKey: "key", wantOK: true, wantVal: "new"},
				{after: 11 * time.Second, getKey: "key"},
			},
		},
		{
			name:       "lru eviction",
			maxEntries: 2,
			ttl:        time.Minute,
			steps: []cacheStep{
				{setKey: "a", setValue: "A"},
				{setKey: "b", setValue: "B"},
				{getKey: "a", wantOK: true, wantVal: "A"},
				{setKey: "c", setValue: "C"},
				{getKey: "a", wantOK: true, wantVal: "A"},
				{getKey: "b"},
				{getKey: "c", wantOK: true, wantVal: "C"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, now := newTestCache(t, tt.maxEntries, tt.ttl)

			for i, step := range tt.steps {
				*now = now.Add(step.after)

				if step.setKey != "" {
					cache.Set(step.setKey, step.setValue)
					continue
				}

				if step.getKey == "" {
					t.Fatalf("step %d has no set or get action", i)
				}

				got, ok := cache.Get(step.getKey)
				if ok != step.wantOK {
					t.Fatalf("step %d: exists = %v, want %v", i, ok, step.wantOK)
				}
				if ok && got != step.wantVal {
					t.Fatalf("step %d: value = %q, want %q", i, got, step.wantVal)
				}
			}
		})
	}
}

func TestNewRejectsInvalidSize(t *testing.T) {
	tests := []struct {
		name       string
		maxEntries int
	}{
		{name: "zero", maxEntries: 0},
		{name: "negative", maxEntries: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New[string](tt.maxEntries, time.Second); err == nil {
				t.Fatal("expected error for invalid cache size")
			}
		})
	}
}
