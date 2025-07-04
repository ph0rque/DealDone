package ai

import (
	"testing"
	"time"
)

func TestAICache(t *testing.T) {
	t.Run("NewAICache creates cache with TTL", func(t *testing.T) {
		ttl := time.Minute * 5
		cache := NewAICache(ttl)

		if cache.ttl != ttl {
			t.Errorf("Expected TTL %v, got %v", ttl, cache.ttl)
		}

		if cache.maxSize != 1000 {
			t.Errorf("Expected default max size 1000, got %d", cache.maxSize)
		}
	})

	t.Run("GenerateKey creates consistent keys", func(t *testing.T) {
		cache := NewAICache(time.Minute)

		metadata := map[string]interface{}{"key": "value"}

		key1 := cache.GenerateKey("classify", "content", metadata)
		key2 := cache.GenerateKey("classify", "content", metadata)

		if key1 != key2 {
			t.Error("Expected same key for same inputs")
		}

		key3 := cache.GenerateKey("classify", "different", metadata)
		if key1 == key3 {
			t.Error("Expected different key for different content")
		}

		key4 := cache.GenerateKey("extract", "content", metadata)
		if key1 == key4 {
			t.Error("Expected different key for different operation")
		}
	})

	t.Run("Set and Get work correctly", func(t *testing.T) {
		cache := NewAICache(time.Second * 2)

		key := "test-key"
		value := &AIClassificationResult{
			DocumentType: "legal",
			Confidence:   0.9,
		}

		cache.Set(key, value)

		retrieved := cache.Get(key)
		if retrieved == nil {
			t.Fatal("Expected to retrieve cached value")
		}

		result, ok := retrieved.(*AIClassificationResult)
		if !ok {
			t.Fatal("Failed to cast retrieved value")
		}

		if result.DocumentType != "legal" {
			t.Errorf("Expected document type 'legal', got %s", result.DocumentType)
		}
	})

	t.Run("Cache expires items", func(t *testing.T) {
		cache := NewAICache(time.Millisecond * 100)

		key := "expiring-key"
		cache.Set(key, "value")

		// Should exist immediately
		if cache.Get(key) == nil {
			t.Error("Expected value to exist immediately after setting")
		}

		// Wait for expiration
		time.Sleep(time.Millisecond * 150)

		// Should be expired
		if cache.Get(key) != nil {
			t.Error("Expected value to be expired")
		}
	})

	t.Run("Cache respects max size", func(t *testing.T) {
		cache := NewAICache(time.Hour)
		cache.SetMaxSize(3)

		// Add 4 items
		for i := 0; i < 4; i++ {
			cache.Set(string(rune(i)), i)
			time.Sleep(time.Millisecond * 10) // Ensure different access times
		}

		stats := cache.GetStats()
		if stats["size"].(int) > 3 {
			t.Errorf("Expected cache size <= 3, got %d", stats["size"])
		}
	})

	t.Run("Clear removes all items", func(t *testing.T) {
		cache := NewAICache(time.Hour)

		cache.Set("key1", "value1")
		cache.Set("key2", "value2")

		cache.Clear()

		stats := cache.GetStats()
		if stats["size"].(int) != 0 {
			t.Errorf("Expected empty cache after clear, got size %d", stats["size"])
		}
	})

	t.Run("GetStats returns correct information", func(t *testing.T) {
		cache := NewAICache(time.Minute * 5)

		cache.Set("key1", "value1")
		cache.Get("key1")
		cache.Get("key1")

		stats := cache.GetStats()

		if stats["size"].(int) != 1 {
			t.Errorf("Expected size 1, got %d", stats["size"])
		}

		if stats["totalHits"].(int) != 2 {
			t.Errorf("Expected 2 total hits, got %d", stats["totalHits"])
		}

		if stats["maxSize"].(int) != 1000 {
			t.Errorf("Expected max size 1000, got %d", stats["maxSize"])
		}
	})
}
