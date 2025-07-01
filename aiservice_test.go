package main

import (
	"context"
	"testing"
	"time"
)

func TestAIService(t *testing.T) {
	t.Run("NewAIService creates service with config", func(t *testing.T) {
		config := &AIConfig{
			OpenAIKey:   "test-key",
			OpenAIModel: "gpt-4",
			ClaudeKey:   "",
			ClaudeModel: "",
			CacheTTL:    time.Minute * 5,
			RateLimit:   60,
			MaxRetries:  3,
			RetryDelay:  time.Second,
		}

		as := NewAIService(config)

		if as == nil {
			t.Fatal("Expected AI service to be created")
		}

		// Should have OpenAI and Default providers
		providers := as.GetAvailableProviders()
		if len(providers) < 1 {
			t.Error("Expected at least default provider to be available")
		}
	})

	t.Run("IsAvailable checks configuration", func(t *testing.T) {
		// Service with no API keys should still have default provider
		config := &AIConfig{
			CacheTTL:  time.Minute,
			RateLimit: 60,
		}
		as := NewAIService(config)

		if !as.IsAvailable() {
			t.Error("Expected service to be available with default provider")
		}

		// Service with API key should be available
		config2 := &AIConfig{
			OpenAIKey: "test-key",
			CacheTTL:  time.Minute,
			RateLimit: 60,
		}
		as2 := NewAIService(config2)

		if !as2.IsAvailable() {
			t.Error("Expected service with API key to be available")
		}
	})

	t.Run("ClassifyDocument uses fallback when needed", func(t *testing.T) {
		config := &AIConfig{
			CacheTTL:  time.Minute,
			RateLimit: 60,
		}
		as := NewAIService(config)

		ctx := context.Background()
		result, err := as.ClassifyDocument(ctx, "test content", nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.DocumentType == "" {
			t.Error("Expected document type to be set")
		}

		if result.Confidence <= 0 || result.Confidence > 1 {
			t.Errorf("Expected confidence between 0 and 1, got %f", result.Confidence)
		}

		if result.Keywords == nil {
			t.Error("Expected keywords to be initialized")
		}
	})

	t.Run("Caching works correctly", func(t *testing.T) {
		config := &AIConfig{
			CacheTTL:  time.Second * 2,
			RateLimit: 60,
		}
		as := NewAIService(config)

		ctx := context.Background()
		content := "test financial content with revenue and profit"
		metadata := map[string]interface{}{"test": true}

		// First call
		result1, err := as.ClassifyDocument(ctx, content, metadata)
		if err != nil {
			t.Fatalf("First call failed: %v", err)
		}

		// Second call should hit cache
		result2, err := as.ClassifyDocument(ctx, content, metadata)
		if err != nil {
			t.Fatalf("Second call failed: %v", err)
		}

		// Results should be identical
		if result1.DocumentType != result2.DocumentType {
			t.Error("Cached result differs from original")
		}
	})

	t.Run("Rate limiting works", func(t *testing.T) {
		config := &AIConfig{
			CacheTTL:  time.Minute,
			RateLimit: 2, // Very low rate
		}
		as := NewAIService(config)

		ctx := context.Background()

		// First two should succeed
		_, err1 := as.ClassifyDocument(ctx, "content1", nil)
		if err1 != nil {
			t.Errorf("First request failed: %v", err1)
		}

		_, err2 := as.ClassifyDocument(ctx, "content2", nil)
		if err2 != nil {
			t.Errorf("Second request failed: %v", err2)
		}

		// Third should wait or timeout
		shortCtx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
		defer cancel()

		_, err3 := as.ClassifyDocument(shortCtx, "content3", nil)
		if err3 == nil {
			t.Error("Expected rate limit error for third request")
		}
	})

	t.Run("SetPrimaryProvider changes provider order", func(t *testing.T) {
		config := &AIConfig{
			OpenAIKey: "test-key",
			ClaudeKey: "test-key",
			CacheTTL:  time.Minute,
			RateLimit: 60,
		}
		as := NewAIService(config)

		// Default should be OpenAI (first configured)
		if as.primaryProvider != ProviderOpenAI {
			t.Errorf("Expected primary provider to be OpenAI, got %s", as.primaryProvider)
		}

		// Change to Claude
		err := as.SetPrimaryProvider(ProviderClaude)
		if err != nil {
			t.Errorf("Failed to set primary provider: %v", err)
		}

		if as.primaryProvider != ProviderClaude {
			t.Errorf("Expected primary provider to be Claude, got %s", as.primaryProvider)
		}

		// Try to set non-existent provider
		err = as.SetPrimaryProvider("nonexistent")
		if err == nil {
			t.Error("Expected error when setting non-existent provider")
		}
	})
}
