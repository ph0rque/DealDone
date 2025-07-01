package main

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	t.Run("NewRateLimiter creates limiter with correct rate", func(t *testing.T) {
		requestsPerMinute := 60
		rl := NewRateLimiter(requestsPerMinute)

		if rl.maxTokens != float64(requestsPerMinute) {
			t.Errorf("Expected max tokens %d, got %f", requestsPerMinute, rl.maxTokens)
		}

		expectedRate := float64(requestsPerMinute) / 60.0
		if rl.refillRate != expectedRate {
			t.Errorf("Expected refill rate %f, got %f", expectedRate, rl.refillRate)
		}
	})

	t.Run("TryAcquire respects limit", func(t *testing.T) {
		rl := NewRateLimiter(2) // 2 requests per minute

		// Should succeed for first two
		if !rl.TryAcquire() {
			t.Error("First request should succeed")
		}

		if !rl.TryAcquire() {
			t.Error("Second request should succeed")
		}

		// Third should fail
		if rl.TryAcquire() {
			t.Error("Third request should fail")
		}
	})

	t.Run("Tokens refill over time", func(t *testing.T) {
		rl := NewRateLimiter(60) // 60 per minute = 1 per second

		// Use all tokens
		rl.tokens = 0

		// Wait for refill
		time.Sleep(time.Second + time.Millisecond*100) // Wait slightly more than 1 second

		// Should have at least 1 token
		if !rl.TryAcquire() {
			t.Error("Should have refilled at least 1 token")
		}
	})

	t.Run("Wait blocks until token available", func(t *testing.T) {
		rl := NewRateLimiter(60) // 1 per second
		rl.tokens = 0            // Start empty

		ctx := context.Background()
		start := time.Now()

		err := rl.Wait(ctx)
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		elapsed := time.Since(start)
		// Should have waited approximately 1 second
		if elapsed < time.Millisecond*900 {
			t.Errorf("Expected to wait ~1 second, but only waited %v", elapsed)
		}
	})

	t.Run("Wait respects context cancellation", func(t *testing.T) {
		rl := NewRateLimiter(1) // Very low rate
		rl.tokens = 0

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
		defer cancel()

		err := rl.Wait(ctx)
		if err == nil {
			t.Error("Expected error from cancelled context")
		}
	})

	t.Run("TryAcquireN acquires multiple tokens", func(t *testing.T) {
		rl := NewRateLimiter(10)

		if !rl.TryAcquireN(5) {
			t.Error("Should be able to acquire 5 tokens")
		}

		// Should have 5 left
		if !rl.TryAcquireN(5) {
			t.Error("Should be able to acquire another 5 tokens")
		}

		// Should not have 5 more
		if rl.TryAcquireN(5) {
			t.Error("Should not be able to acquire 5 more tokens")
		}

		// But should have 1
		if rl.TryAcquire() {
			t.Error("Should not have any tokens left")
		}
	})

	t.Run("Reset restores full capacity", func(t *testing.T) {
		rl := NewRateLimiter(10)
		rl.tokens = 0

		rl.Reset()

		if rl.tokens != rl.maxTokens {
			t.Errorf("Expected %f tokens after reset, got %f", rl.maxTokens, rl.tokens)
		}
	})

	t.Run("SetRate updates rate limit", func(t *testing.T) {
		rl := NewRateLimiter(10)

		rl.SetRate(20)

		if rl.maxTokens != 20 {
			t.Errorf("Expected max tokens 20, got %f", rl.maxTokens)
		}

		// Should not exceed new max
		rl.tokens = 30
		rl.SetRate(15)

		if rl.tokens > 15 {
			t.Errorf("Tokens should be capped at new max, got %f", rl.tokens)
		}
	})

	t.Run("GetStats returns correct information", func(t *testing.T) {
		rl := NewRateLimiter(60)
		rl.TryAcquire()
		rl.TryAcquire()

		stats := rl.GetStats()

		available := stats["availableTokens"].(float64)
		// Use approximate comparison due to timing
		if available < 57.9 || available > 58.1 {
			t.Errorf("Expected ~58 available tokens, got %f", available)
		}

		maxTokens := stats["maxTokens"].(float64)
		if maxTokens != 60 {
			t.Errorf("Expected max tokens 60, got %f", maxTokens)
		}

		refillRate := stats["refillRate"].(float64)
		if refillRate != 60 {
			t.Errorf("Expected refill rate 60/min, got %f", refillRate)
		}
	})
}
