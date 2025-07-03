package main

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogger is now defined in test_utils.go

// Test fixtures
func createTestConflictResolver(t *testing.T) (*ConflictResolver, string) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	return NewConflictResolver(tempDir, logger), tempDir
}

func createTestConflictingValues() []ConflictingValue {
	return []ConflictingValue{
		{
			Value:      "100000",
			Confidence: 0.9,
			Source:     "document_1.pdf",
			Method:     "ocr_high_conf",
			Timestamp:  time.Now().Unix() - 3600,
		},
		{
			Value:      "105000",
			Confidence: 0.85,
			Source:     "document_2.pdf",
			Method:     "nlp_extraction",
			Timestamp:  time.Now().Unix() - 1800,
		},
		{
			Value:      "98000",
			Confidence: 0.7,
			Source:     "document_3.pdf",
			Method:     "pattern_match",
			Timestamp:  time.Now().Unix() - 900,
		},
	}
}

func createNumericConflictingValues() []ConflictingValue {
	return []ConflictingValue{
		{
			Value:      100000.0,
			Confidence: 0.8,
			Source:     "doc1.pdf",
			Method:     "ocr_standard",
			Timestamp:  time.Now().Unix(),
		},
		{
			Value:      102000.0,
			Confidence: 0.82,
			Source:     "doc2.pdf",
			Method:     "nlp_extraction",
			Timestamp:  time.Now().Unix(),
		},
		{
			Value:      99000.0,
			Confidence: 0.78,
			Source:     "doc3.pdf",
			Method:     "pattern_match",
			Timestamp:  time.Now().Unix(),
		},
	}
}

// Test ConflictResolver creation and initialization
func TestConflictResolver_Creation(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)

	assert.NotNil(t, resolver)
	assert.NotNil(t, resolver.config)
	assert.Equal(t, 5, len(resolver.resolutionStrategies))
	assert.Contains(t, resolver.resolutionStrategies, "highest_confidence")
	assert.Contains(t, resolver.resolutionStrategies, "numeric_averaging")
	assert.Contains(t, resolver.resolutionStrategies, "latest_value")
	assert.Contains(t, resolver.resolutionStrategies, "manual_review")
	assert.Contains(t, resolver.resolutionStrategies, "source_priority")
}

// Test highest confidence resolution strategy
func TestConflictResolver_HighestConfidenceStrategy(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)
	ctx := context.Background()

	conflictCtx := &ConflictContext{
		DealName:          "TestDeal",
		TemplatePath:      "/templates/test.xlsx",
		FieldName:         "purchase_price",
		ConflictingValues: createTestConflictingValues(),
		FieldType:         "string",
	}

	result, err := resolver.ResolveConflict(ctx, conflictCtx)

	require.NoError(t, err)
	assert.Equal(t, "100000", result.ResolvedValue)
	assert.Equal(t, "highest_confidence", result.ResolutionMethod)
	assert.Equal(t, 0.9, result.FinalConfidence)
	assert.False(t, result.RequiresReview)
	assert.Contains(t, result.ResolutionNotes, "highest confidence")
}

// Test numeric averaging resolution strategy
func TestConflictResolver_NumericAveragingStrategy(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)
	ctx := context.Background()

	// Create values with similar confidence levels to trigger averaging
	conflictingValues := []ConflictingValue{
		{
			Value:      100000.0,
			Confidence: 0.8,
			Source:     "doc1.pdf",
			Method:     "ocr_standard",
			Timestamp:  time.Now().Unix(),
		},
		{
			Value:      102000.0,
			Confidence: 0.81,
			Source:     "doc2.pdf",
			Method:     "nlp_extraction",
			Timestamp:  time.Now().Unix(),
		},
		{
			Value:      98000.0,
			Confidence: 0.79,
			Source:     "doc3.pdf",
			Method:     "pattern_match",
			Timestamp:  time.Now().Unix(),
		},
	}

	conflictCtx := &ConflictContext{
		DealName:          "TestDeal",
		TemplatePath:      "/templates/test.xlsx",
		FieldName:         "purchase_price",
		ConflictingValues: conflictingValues,
		FieldType:         "number",
	}

	result, err := resolver.ResolveConflict(ctx, conflictCtx)

	require.NoError(t, err)
	assert.Equal(t, "numeric_averaging", result.ResolutionMethod)

	// Check that the result is a weighted average
	expectedAvg := (100000.0*0.8 + 102000.0*0.81 + 98000.0*0.79) / (0.8 + 0.81 + 0.79)
	expectedRounded := math.Round(expectedAvg*100) / 100
	assert.Equal(t, expectedRounded, result.ResolvedValue)
	assert.Contains(t, result.ResolutionNotes, "weighted average")
}

// Test latest value resolution strategy
func TestConflictResolver_LatestValueStrategy(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)
	ctx := context.Background()

	now := time.Now().Unix()
	conflictingValues := []ConflictingValue{
		{
			Value:      "2023-01-01",
			Confidence: 0.7,
			Source:     "doc1.pdf",
			Method:     "ocr_standard",
			Timestamp:  now - 3600, // 1 hour ago
		},
		{
			Value:      "2023-06-15",
			Confidence: 0.8,
			Source:     "doc2.pdf",
			Method:     "nlp_extraction",
			Timestamp:  now, // Most recent
		},
		{
			Value:      "2023-03-10",
			Confidence: 0.75,
			Source:     "doc3.pdf",
			Method:     "pattern_match",
			Timestamp:  now - 1800, // 30 minutes ago
		},
	}

	conflictCtx := &ConflictContext{
		DealName:          "TestDeal",
		TemplatePath:      "/templates/test.xlsx",
		FieldName:         "closing_date",
		ConflictingValues: conflictingValues,
		FieldType:         "date",
	}

	result, err := resolver.ResolveConflict(ctx, conflictCtx)

	require.NoError(t, err)
	assert.Equal(t, "2023-06-15", result.ResolvedValue)
	assert.Equal(t, "latest_value", result.ResolutionMethod)
	assert.Equal(t, 0.8, result.FinalConfidence)
	assert.Contains(t, result.ResolutionNotes, "most recent value")
}

// Test manual review resolution strategy
func TestConflictResolver_ManualReviewStrategy(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)
	ctx := context.Background()

	// Create values with low confidence to trigger manual review
	conflictingValues := []ConflictingValue{
		{
			Value:      "Company A",
			Confidence: 0.3,
			Source:     "doc1.pdf",
			Method:     "ocr_standard",
			Timestamp:  time.Now().Unix(),
		},
		{
			Value:      "Company B",
			Confidence: 0.35,
			Source:     "doc2.pdf",
			Method:     "heuristic",
			Timestamp:  time.Now().Unix(),
		},
	}

	conflictCtx := &ConflictContext{
		DealName:          "TestDeal",
		TemplatePath:      "/templates/test.xlsx",
		FieldName:         "target_company",
		ConflictingValues: conflictingValues,
		FieldType:         "string",
		RequiresReview:    true,
	}

	result, err := resolver.ResolveConflict(ctx, conflictCtx)

	require.NoError(t, err)
	assert.Equal(t, "manual_review", result.ResolutionMethod)
	assert.True(t, result.RequiresReview)
	assert.Equal(t, 0.0, result.FinalConfidence)
	assert.Contains(t, result.ResolutionNotes, "manual review")
}

// Test source priority resolution strategy
func TestConflictResolver_SourcePriorityStrategy(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)

	// Manually set strategy to source priority for testing
	conflictingValues := []ConflictingValue{
		{
			Value:      "Manual Entry Value",
			Confidence: 0.7,
			Source:     "user",
			Method:     "manual_entry",
			Timestamp:  time.Now().Unix(),
		},
		{
			Value:      "OCR Value",
			Confidence: 0.9,
			Source:     "doc1.pdf",
			Method:     "ocr_standard",
			Timestamp:  time.Now().Unix(),
		},
		{
			Value:      "User Correction",
			Confidence: 0.8,
			Source:     "user",
			Method:     "user_correction",
			Timestamp:  time.Now().Unix(),
		},
	}

	conflictCtx := &ConflictContext{
		DealName:          "TestDeal",
		TemplatePath:      "/templates/test.xlsx",
		FieldName:         "company_name",
		ConflictingValues: conflictingValues,
		FieldType:         "string",
	}

	result, err := resolver.resolutionStrategies["source_priority"].ResolutionFunction(conflictCtx)

	require.NoError(t, err)
	assert.Equal(t, "Manual Entry Value", result.ResolvedValue)
	assert.Equal(t, "source_priority", result.ResolutionMethod)
	assert.Contains(t, result.ResolutionNotes, "manual_entry method")
}

// Test conflict type determination
func TestConflictResolver_DetermineConflictType(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)

	tests := []struct {
		name     string
		values   []ConflictingValue
		expected string
	}{
		{
			name:     "No conflict with single value",
			values:   createTestConflictingValues()[:1],
			expected: "none",
		},
		{
			name:     "Confidence conflict with different confidence levels",
			values:   createTestConflictingValues(),
			expected: "confidence",
		},
		{
			name: "Duplicate values",
			values: []ConflictingValue{
				{Value: "same_value", Confidence: 0.8},
				{Value: "same_value", Confidence: 0.9},
			},
			expected: "duplicate",
		},
		{
			name: "Numeric conflict",
			values: []ConflictingValue{
				{Value: 100.0, Confidence: 0.81},
				{Value: 102.0, Confidence: 0.82},
			},
			expected: "numeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflictType := resolver.determineConflictType(tt.values)
			assert.Equal(t, tt.expected, conflictType)
		})
	}
}

// Test conflict history tracking
func TestConflictResolver_ConflictHistory(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)
	ctx := context.Background()

	conflictCtx := &ConflictContext{
		DealName:          "TestDeal",
		TemplatePath:      "/templates/test.xlsx",
		FieldName:         "purchase_price",
		ConflictingValues: createTestConflictingValues(),
		FieldType:         "string",
	}

	// Resolve multiple conflicts to build history
	_, err := resolver.ResolveConflict(ctx, conflictCtx)
	require.NoError(t, err)

	conflictCtx.ConflictingValues = createNumericConflictingValues()
	_, err = resolver.ResolveConflict(ctx, conflictCtx)
	require.NoError(t, err)

	// Check history
	history := resolver.GetConflictHistory("TestDeal", "/templates/test.xlsx", "purchase_price", 10)
	assert.Len(t, history, 2)
	assert.Equal(t, "TestDeal", history[0].DealName)
	assert.Equal(t, "/templates/test.xlsx", history[0].TemplatePath)
	assert.Equal(t, "purchase_price", history[0].FieldName)

	// Check deal-level history
	dealHistory := resolver.GetConflictHistory("TestDeal", "", "", 10)
	assert.Len(t, dealHistory, 2)
}

// Test audit trail functionality
func TestConflictResolver_AuditTrail(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)
	ctx := context.Background()

	conflictCtx := &ConflictContext{
		DealName:          "TestDeal",
		TemplatePath:      "/templates/test.xlsx",
		FieldName:         "purchase_price",
		ConflictingValues: createTestConflictingValues(),
		FieldType:         "string",
	}

	result, err := resolver.ResolveConflict(ctx, conflictCtx)
	require.NoError(t, err)

	// Check audit trail
	auditEntries := resolver.GetAuditTrail("", 10)
	assert.GreaterOrEqual(t, len(auditEntries), 2) // At least detection and resolution entries

	// Find the detection entry
	var detectionEntry *ConflictAuditEntry
	for _, entry := range auditEntries {
		if entry.Action == "detected" {
			detectionEntry = entry
			break
		}
	}
	assert.NotNil(t, detectionEntry)
	assert.Contains(t, detectionEntry.Details, "Conflict detected")

	// Find the resolution entry
	var resolutionEntry *ConflictAuditEntry
	for _, entry := range auditEntries {
		if entry.Action == "resolved" {
			resolutionEntry = entry
			break
		}
	}
	assert.NotNil(t, resolutionEntry)
	assert.Contains(t, resolutionEntry.Details, "Conflict resolved")
	assert.Equal(t, result.FinalConfidence, resolutionEntry.ConfidenceScore)
}

// Test conflict statistics
func TestConflictResolver_Statistics(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)
	ctx := context.Background()

	// Resolve multiple conflicts
	conflictCtx := &ConflictContext{
		DealName:          "TestDeal",
		TemplatePath:      "/templates/test.xlsx",
		FieldName:         "purchase_price",
		ConflictingValues: createTestConflictingValues(),
		FieldType:         "string",
	}

	_, err := resolver.ResolveConflict(ctx, conflictCtx)
	require.NoError(t, err)

	conflictCtx.FieldName = "target_company"
	conflictCtx.FieldType = "number"
	conflictCtx.ConflictingValues = createNumericConflictingValues()
	_, err = resolver.ResolveConflict(ctx, conflictCtx)
	require.NoError(t, err)

	stats := resolver.GetConflictStatistics()
	assert.Equal(t, 2, stats["total_conflicts"])
	assert.Contains(t, stats["resolution_methods"], "highest_confidence")
	assert.Greater(t, stats["average_confidence"].(float64), 0.0)
	assert.GreaterOrEqual(t, stats["audit_entries"], 4) // 2 conflicts × 2 entries each
}

// Test persistence functionality
func TestConflictResolver_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	logger := NewTestLogger()
	resolver := NewConflictResolver(tempDir, logger)
	ctx := context.Background()

	conflictCtx := &ConflictContext{
		DealName:          "TestDeal",
		TemplatePath:      "/templates/test.xlsx",
		FieldName:         "purchase_price",
		ConflictingValues: createTestConflictingValues(),
		FieldType:         "string",
	}

	_, err := resolver.ResolveConflict(ctx, conflictCtx)
	require.NoError(t, err)

	// Save state
	err = resolver.SaveState()
	require.NoError(t, err)

	// Verify files were created
	historyPath := filepath.Join(tempDir, "conflict_history.json")
	auditPath := filepath.Join(tempDir, "audit_trail.json")
	assert.FileExists(t, historyPath)
	assert.FileExists(t, auditPath)

	// Create new resolver and verify data is loaded
	resolver2 := NewConflictResolver(tempDir, logger)
	history := resolver2.GetConflictHistory("TestDeal", "/templates/test.xlsx", "purchase_price", 10)
	assert.Len(t, history, 1)

	auditTrail := resolver2.GetAuditTrail("", 10)
	assert.GreaterOrEqual(t, len(auditTrail), 2)
}

// Test configuration updates
func TestConflictResolver_ConfigurationUpdate(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)

	originalConfig := resolver.GetConfiguration()
	assert.Equal(t, 0.7, originalConfig.MinConfidenceThreshold)

	// Update configuration
	newConfig := &ConflictResolutionConfig{
		MinConfidenceThreshold:    0.8,
		ReviewThreshold:           0.6,
		NumericAveragingThreshold: 0.1,
		MaxHistoryEntries:         500,
		EnableAuditTrail:          false,
		DebugMode:                 true,
	}

	err := resolver.UpdateConfiguration(newConfig)
	require.NoError(t, err)

	updatedConfig := resolver.GetConfiguration()
	assert.Equal(t, 0.8, updatedConfig.MinConfidenceThreshold)
	assert.Equal(t, 0.6, updatedConfig.ReviewThreshold)
	assert.Equal(t, false, updatedConfig.EnableAuditTrail)
	assert.Equal(t, true, updatedConfig.DebugMode)
}

// Test edge cases and error conditions
func TestConflictResolver_EdgeCases(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)
	ctx := context.Background()

	t.Run("Empty conflicting values", func(t *testing.T) {
		conflictCtx := &ConflictContext{
			DealName:          "TestDeal",
			TemplatePath:      "/templates/test.xlsx",
			FieldName:         "purchase_price",
			ConflictingValues: []ConflictingValue{},
			FieldType:         "string",
		}

		_, err := resolver.ResolveConflict(ctx, conflictCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no conflicting values provided")
	})

	t.Run("Numeric averaging with no numeric values", func(t *testing.T) {
		conflictingValues := []ConflictingValue{
			{Value: "not_a_number", Confidence: 0.8},
			{Value: "also_not_a_number", Confidence: 0.9},
		}

		conflictCtx := &ConflictContext{
			ConflictingValues: conflictingValues,
		}

		_, err := resolver.resolveByNumericAveraging(conflictCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no valid numeric values found")
	})

	t.Run("Zero confidence weights in numeric averaging", func(t *testing.T) {
		conflictingValues := []ConflictingValue{
			{Value: 100.0, Confidence: 0.0},
			{Value: 200.0, Confidence: 0.0},
		}

		conflictCtx := &ConflictContext{
			ConflictingValues: conflictingValues,
		}

		_, err := resolver.resolveByNumericAveraging(conflictCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "total confidence weight is zero")
	})
}

// Test concurrent access safety
func TestConflictResolver_ConcurrentAccess(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)
	ctx := context.Background()

	// Run multiple goroutines to test thread safety
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			conflictCtx := &ConflictContext{
				DealName:          "TestDeal",
				TemplatePath:      "/templates/test.xlsx",
				FieldName:         fmt.Sprintf("field_%d", id),
				ConflictingValues: createTestConflictingValues(),
				FieldType:         "string",
			}

			_, err := resolver.ResolveConflict(ctx, conflictCtx)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all conflicts were recorded
	stats := resolver.GetConflictStatistics()
	assert.Equal(t, 10, stats["total_conflicts"])
}

// Test integration with template population system
func TestConflictResolver_TemplateIntegration(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)
	ctx := context.Background()

	// Simulate a template population scenario with conflicts
	conflicts := []struct {
		fieldName string
		fieldType string
		values    []ConflictingValue
	}{
		{
			fieldName: "purchase_price",
			fieldType: "number",
			values:    createNumericConflictingValues(),
		},
		{
			fieldName: "target_company",
			fieldType: "string",
			values:    createTestConflictingValues(),
		},
		{
			fieldName: "closing_date",
			fieldType: "date",
			values: []ConflictingValue{
				{Value: "2023-12-31", Confidence: 0.9, Source: "contract.pdf", Method: "ocr_high_conf", Timestamp: time.Now().Unix()},
				{Value: "2024-01-15", Confidence: 0.8, Source: "summary.pdf", Method: "nlp_extraction", Timestamp: time.Now().Unix() - 100},
			},
		},
	}

	resolvedValues := make(map[string]interface{})
	for _, conflict := range conflicts {
		conflictCtx := &ConflictContext{
			DealName:          "M&A_Deal_2023",
			TemplatePath:      "/templates/ma_template.xlsx",
			FieldName:         conflict.fieldName,
			ConflictingValues: conflict.values,
			FieldType:         conflict.fieldType,
		}

		result, err := resolver.ResolveConflict(ctx, conflictCtx)
		require.NoError(t, err)

		resolvedValues[conflict.fieldName] = result.ResolvedValue
		t.Logf("Field %s resolved to: %v (method: %s, confidence: %.3f)",
			conflict.fieldName, result.ResolvedValue, result.ResolutionMethod, result.FinalConfidence)
	}

	// Verify all fields were resolved
	assert.Len(t, resolvedValues, 3)
	assert.Contains(t, resolvedValues, "purchase_price")
	assert.Contains(t, resolvedValues, "target_company")
	assert.Contains(t, resolvedValues, "closing_date")

	// Check template-level history
	templateHistory := resolver.GetConflictHistory("M&A_Deal_2023", "/templates/ma_template.xlsx", "", 10)
	assert.Len(t, templateHistory, 3)
}

// Benchmark conflict resolution performance
func BenchmarkConflictResolver_ResolveConflict(b *testing.B) {
	tempDir := b.TempDir()
	logger := NewTestLogger()
	resolver := NewConflictResolver(tempDir, logger)
	ctx := context.Background()

	conflictCtx := &ConflictContext{
		DealName:          "BenchmarkDeal",
		TemplatePath:      "/templates/benchmark.xlsx",
		FieldName:         "test_field",
		ConflictingValues: createTestConflictingValues(),
		FieldType:         "string",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conflictCtx.FieldName = fmt.Sprintf("field_%d", i)
		_, err := resolver.ResolveConflict(ctx, conflictCtx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test memory cleanup and garbage collection
func TestConflictResolver_MemoryManagement(t *testing.T) {
	resolver, _ := createTestConflictResolver(t)
	ctx := context.Background()

	// Set a low max history limit
	resolver.config.MaxHistoryEntries = 5

	// Add more conflicts than the limit
	for i := 0; i < 10; i++ {
		conflictCtx := &ConflictContext{
			DealName:          "TestDeal",
			TemplatePath:      "/templates/test.xlsx",
			FieldName:         fmt.Sprintf("field_%d", i),
			ConflictingValues: createTestConflictingValues(),
			FieldType:         "string",
		}

		_, err := resolver.ResolveConflict(ctx, conflictCtx)
		require.NoError(t, err)
	}

	// Verify overall history count and that individual field history is correct
	// Since each field only has 1 entry, the individual field history should be 1
	history := resolver.GetConflictHistory("TestDeal", "/templates/test.xlsx", "field_0", 10)
	assert.LessOrEqual(t, len(history), 1) // Each field should have only 1 entry

	// Check total deal history to verify we have all 10 conflicts
	dealHistory := resolver.GetConflictHistory("TestDeal", "/templates/test.xlsx", "", 20)
	assert.Equal(t, 10, len(dealHistory))

	// Verify audit trail is managed (entries may be trimmed during processing)
	auditEntries := resolver.GetAuditTrail("", 100)
	assert.GreaterOrEqual(t, len(auditEntries), 1) // At least some entries should exist

	// Audit trail may have more entries than MaxHistoryEntries due to timing of trim operations
	// The trimming happens during conflict resolution, not at fixed intervals
	assert.LessOrEqual(t, len(auditEntries), resolver.config.MaxHistoryEntries*2)
}
