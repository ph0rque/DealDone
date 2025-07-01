package main

import (
	"testing"
)

func TestAIService(t *testing.T) {
	t.Run("NewAIService creates service with config", func(t *testing.T) {
		apiKey := "test-key"
		endpoint := "https://api.example.com"
		model := "gpt-4"

		as := NewAIService(apiKey, endpoint, model)

		if as.apiKey != apiKey {
			t.Errorf("Expected apiKey %s, got %s", apiKey, as.apiKey)
		}

		if as.endpoint != endpoint {
			t.Errorf("Expected endpoint %s, got %s", endpoint, as.endpoint)
		}

		if as.model != model {
			t.Errorf("Expected model %s, got %s", model, as.model)
		}
	})

	t.Run("IsAvailable checks configuration", func(t *testing.T) {
		// Service with API key should be available
		as := NewAIService("test-key", "", "")
		if !as.IsAvailable() {
			t.Error("Expected service with API key to be available")
		}

		// Service without API key should not be available
		as2 := NewAIService("", "", "")
		if as2.IsAvailable() {
			t.Error("Expected service without API key to be unavailable")
		}

		// Nil service should not be available
		var as3 *AIService
		if as3.IsAvailable() {
			t.Error("Expected nil service to be unavailable")
		}
	})

	t.Run("ClassifyDocument returns placeholder result", func(t *testing.T) {
		as := NewAIService("test-key", "", "")

		result, err := as.ClassifyDocument("test content", nil)
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

	t.Run("ExtractTextFromImage returns not implemented", func(t *testing.T) {
		as := NewAIService("test-key", "", "")

		_, err := as.ExtractTextFromImage("test.jpg")
		if err == nil {
			t.Error("Expected not implemented error")
		}
	})

	t.Run("AnalyzeFinancialData returns placeholder result", func(t *testing.T) {
		as := NewAIService("test-key", "", "")

		result, err := as.AnalyzeFinancialData("revenue: $1M")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.DataPoints == nil {
			t.Error("Expected DataPoints map to be initialized")
		}
	})
}
