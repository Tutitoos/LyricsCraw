package cache

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// entry represents a cached value with an expiration time.
type entry struct {
	value     string
	expiresAt time.Time
}

// LyricsCache is a tiny in-memory TTL cache safe for concurrent use.
type LyricsCache struct {
	mu         sync.RWMutex
	items      map[string]entry
	ttl        time.Duration
	maxEntries int
	stopCh     chan struct{}
}

// NewLyricsCache creates a new LyricsCache with the given TTL and capacity.
func NewLyricsCache(ttl time.Duration, maxEntries int) *LyricsCache {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	if maxEntries <= 0 {
		maxEntries = 1000
	}
	return &LyricsCache{
		items:      make(map[string]entry, maxEntries),
		ttl:        ttl,
		maxEntries: maxEntries,
		stopCh:     make(chan struct{}),
	}
}

// Get fetches a value if present and not expired.
func (c *LyricsCache) Get(key string) (string, bool) {
	k := normalizeKey(key)
	c.mu.RLock()
	e, ok := c.items[k]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	if time.Now().After(e.expiresAt) {
		// lazy eviction on read
		c.mu.Lock()
		delete(c.items, k)
		c.mu.Unlock()
		return "", false
	}
	return e.value, true
}

// Set stores a value with the cache's TTL. Applies simple capacity control.
func (c *LyricsCache) Set(key, value string) {
	k := normalizeKey(key)
	now := time.Now()
	exp := now.Add(c.ttl)

	c.mu.Lock()
	// If over capacity, try to purge expired items first
	if len(c.items) >= c.maxEntries {
		c.purgeExpiredLocked(now)
	}
	// If still over capacity, remove the item with earliest expiration (approx LRU-ish)
	if len(c.items) >= c.maxEntries {
		var oldestK string
		var oldestT time.Time
		first := true
		for kk, ee := range c.items {
			if first || ee.expiresAt.Before(oldestT) {
				first = false
				oldestT = ee.expiresAt
				oldestK = kk
			}
		}
		if oldestK != "" {
			delete(c.items, oldestK)
		}
	}
	c.items[k] = entry{value: value, expiresAt: exp}
	c.mu.Unlock()
}

// Stop stops the background janitor if started.
func (c *LyricsCache) Stop() {
	select {
	case <-c.stopCh:
		return
	default:
		close(c.stopCh)
	}
}

// StartJanitor starts a background goroutine that periodically purges expired entries.
func (c *LyricsCache) StartJanitor() {
	interval := c.ttl / 2
	if interval < 30*time.Second {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.mu.Lock()
				c.purgeExpiredLocked(time.Now())
				c.mu.Unlock()
			case <-c.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *LyricsCache) purgeExpiredLocked(now time.Time) {
	for k, e := range c.items {
		if now.After(e.expiresAt) {
			delete(c.items, k)
		}
	}
}

// normalizeKey applies a simple normalization so similar queries hit the same key.
func normalizeKey(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Default is the process-wide cache instance.
var Default *LyricsCache

// InitFromEnv initializes the default cache using env vars:
// APP_LYRICS_CACHE_TTL_SECONDS (default 1800), APP_LYRICS_CACHE_MAX_ENTRIES (default 1000)
func InitFromEnv() {
	ttlSeconds := 1800
	if v := os.Getenv("APP_LYRICS_CACHE_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttlSeconds = n
		}
	}
	maxEntries := 1000
	if v := os.Getenv("APP_LYRICS_CACHE_MAX_ENTRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxEntries = n
		}
	}
	Default = NewLyricsCache(time.Duration(ttlSeconds)*time.Second, maxEntries)
	Default.StartJanitor()
}
