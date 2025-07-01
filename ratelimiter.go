package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	mu             sync.Mutex
	tokens         float64
	maxTokens      float64
	refillRate     float64 // tokens per second
	lastRefillTime time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	maxTokens := float64(requestsPerMinute)
	refillRate := maxTokens / 60.0 // Convert to per second

	return &RateLimiter{
		tokens:         maxTokens, // Start with full bucket
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	ticker := time.NewTicker(time.Millisecond * 100) // Check every 100ms
	defer ticker.Stop()

	for {
		if rl.TryAcquire() {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for rate limit")
		case <-ticker.C:
			// Continue checking
		}
	}
}

// TryAcquire attempts to acquire a token without blocking
func (rl *RateLimiter) TryAcquire() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on time elapsed
	rl.refill()

	if rl.tokens >= 1.0 {
		rl.tokens--
		return true
	}

	return false
}

// TryAcquireN attempts to acquire n tokens without blocking
func (rl *RateLimiter) TryAcquireN(n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= float64(n) {
		rl.tokens -= float64(n)
		return true
	}

	return false
}

// refill adds tokens based on time elapsed
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefillTime).Seconds()

	if elapsed > 0 {
		tokensToAdd := elapsed * rl.refillRate
		rl.tokens = min(rl.tokens+tokensToAdd, rl.maxTokens)
		rl.lastRefillTime = now
	}
}

// GetAvailableTokens returns the current number of available tokens
func (rl *RateLimiter) GetAvailableTokens() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()
	return rl.tokens
}

// Reset resets the rate limiter to full capacity
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.tokens = rl.maxTokens
	rl.lastRefillTime = time.Now()
}

// SetRate updates the rate limit
func (rl *RateLimiter) SetRate(requestsPerMinute int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	maxTokens := float64(requestsPerMinute)
	rl.maxTokens = maxTokens
	rl.refillRate = maxTokens / 60.0

	// Don't exceed new max
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	return map[string]interface{}{
		"availableTokens": rl.tokens,
		"maxTokens":       rl.maxTokens,
		"refillRate":      rl.refillRate * 60, // Convert back to per minute
		"percentFull":     (rl.tokens / rl.maxTokens) * 100,
	}
}
