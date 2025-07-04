package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// AICache provides caching for AI responses
type AICache struct {
	mu      sync.RWMutex
	items   map[string]*cacheItem
	ttl     time.Duration
	maxSize int
}

// cacheItem represents a cached item
type cacheItem struct {
	value      interface{}
	expiry     time.Time
	accessTime time.Time
	hits       int
}

// NewAICache creates a new AI cache
func NewAICache(ttl time.Duration) *AICache {
	cache := &AICache{
		items:   make(map[string]*cacheItem),
		ttl:     ttl,
		maxSize: 1000, // Default max size
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// GenerateKey creates a cache key from operation, content, and metadata
func (ac *AICache) GenerateKey(operation string, content string, metadata map[string]interface{}) string {
	h := sha256.New()
	h.Write([]byte(operation))
	h.Write([]byte(content))

	if metadata != nil {
		metaBytes, _ := json.Marshal(metadata)
		h.Write(metaBytes)
	}

	return hex.EncodeToString(h.Sum(nil))
}

// Get retrieves an item from cache
func (ac *AICache) Get(key string) interface{} {
	ac.mu.RLock()
	item, exists := ac.items[key]
	ac.mu.RUnlock()

	if !exists {
		return nil
	}

	// Check if expired
	if time.Now().After(item.expiry) {
		ac.Delete(key)
		return nil
	}

	// Update access time and hits
	ac.mu.Lock()
	item.accessTime = time.Now()
	item.hits++
	ac.mu.Unlock()

	return item.value
}

// Set stores an item in cache
func (ac *AICache) Set(key string, value interface{}) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Check size limit and evict if necessary
	if len(ac.items) >= ac.maxSize {
		ac.evictLRU()
	}

	ac.items[key] = &cacheItem{
		value:      value,
		expiry:     time.Now().Add(ac.ttl),
		accessTime: time.Now(),
		hits:       0,
	}
}

// Delete removes an item from cache
func (ac *AICache) Delete(key string) {
	ac.mu.Lock()
	delete(ac.items, key)
	ac.mu.Unlock()
}

// Clear removes all items from cache
func (ac *AICache) Clear() {
	ac.mu.Lock()
	ac.items = make(map[string]*cacheItem)
	ac.mu.Unlock()
}

// GetStats returns cache statistics
func (ac *AICache) GetStats() map[string]interface{} {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	totalHits := 0
	for _, item := range ac.items {
		totalHits += item.hits
	}

	return map[string]interface{}{
		"size":      len(ac.items),
		"maxSize":   ac.maxSize,
		"totalHits": totalHits,
		"ttl":       ac.ttl.String(),
	}
}

// evictLRU removes the least recently used item
func (ac *AICache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range ac.items {
		if oldestTime.IsZero() || item.accessTime.Before(oldestTime) {
			oldestTime = item.accessTime
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(ac.items, oldestKey)
	}
}

// cleanupExpired runs periodically to remove expired items
func (ac *AICache) cleanupExpired() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ac.mu.Lock()
		now := time.Now()
		for key, item := range ac.items {
			if now.After(item.expiry) {
				delete(ac.items, key)
			}
		}
		ac.mu.Unlock()
	}
}

// SetTTL updates the TTL for future cache entries
func (ac *AICache) SetTTL(ttl time.Duration) {
	ac.mu.Lock()
	ac.ttl = ttl
	ac.mu.Unlock()
}

// SetMaxSize updates the maximum cache size
func (ac *AICache) SetMaxSize(size int) {
	ac.mu.Lock()
	ac.maxSize = size
	ac.mu.Unlock()

	// Evict items if over new limit
	for len(ac.items) > ac.maxSize {
		ac.evictLRU()
	}
}
