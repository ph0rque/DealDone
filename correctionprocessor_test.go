package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCorrectionLogger implements Logger interface for testing
type TestCorrectionLogger struct {
	logs []string
}

func (l *TestCorrectionLogger) Info(format string, args ...interface{}) {
	l.logs = append(l.logs, fmt.Sprintf("INFO: "+format, args...))
}

func (l *TestCorrectionLogger) Debug(format string, args ...interface{}) {
	l.logs = append(l.logs, fmt.Sprintf("DEBUG: "+format, args...))
}

func (l *TestCorrectionLogger) Warn(format string, args ...interface{}) {
	l.logs = append(l.logs, fmt.Sprintf("WARN: "+format, args...))
}

func (l *TestCorrectionLogger) Error(format string, args ...interface{}) {
	l.logs = append(l.logs, fmt.Sprintf("ERROR: "+format, args...))
}

func (l *TestCorrectionLogger) GetLogs() []string {
	return l.logs
}

func (l *TestCorrectionLogger) ClearLogs() {
	l.logs = []string{}
}

func createTestCorrectionProcessor(t *testing.T) (*CorrectionProcessor, *TestCorrectionLogger, string) {
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("correction_test_%d", time.Now().UnixNano()))
	require.NoError(t, os.MkdirAll(tempDir, 0755))

	logger := &TestCorrectionLogger{}
	config := CorrectionDetectionConfig{
		MonitoringInterval:         1 * time.Second,
		MinLearningWeight:          0.1,
		PatternConfidenceThreshold: 0.5,
		MaxPatternAge:              30 * 24 * time.Hour,
		MinFrequencyForPattern:     2,
		EnableRAGLearning:          true,
		LearningRateDecay:          0.95,
		ValidationThreshold:        0.7,
		StoragePath:                tempDir,
		BackupInterval:             5 * time.Minute,
		MaxCorrectionHistory:       100,
	}

	processor := NewCorrectionProcessor(config, logger)
	return processor, logger, tempDir
}

func TestCorrectionProcessor_Creation(t *testing.T) {
	processor, logger, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	assert.NotNil(t, processor)
	assert.NotNil(t, processor.corrections)
	assert.NotNil(t, processor.learningModel)
	assert.NotNil(t, processor.patternDetector)
	assert.NotNil(t, processor.ragLearning)
	assert.Equal(t, "1.0.0", processor.learningModel.Version)
	assert.Equal(t, 0, processor.learningModel.TotalCorrections)

	// Check if storage directory was created
	assert.DirExists(t, tempDir)

	// Check initial log
	logs := logger.GetLogs()
	assert.True(t, len(logs) > 0)
}

func TestCorrectionProcessor_DetectCorrection(t *testing.T) {
	processor, logger, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	correction := &CorrectionEntry{
		DealID:             "deal_001",
		DocumentID:         "doc_001",
		TemplateID:         "template_001",
		FieldName:          "purchase_price",
		OriginalValue:      "1000000",
		CorrectedValue:     "1500000",
		CorrectionType:     FieldValueCorrection,
		UserID:             "user_001",
		ProcessingMethod:   "OCR",
		OriginalConfidence: 0.6,
		ValidationStatus:   "pending",
	}

	err := processor.DetectCorrection(correction)
	assert.NoError(t, err)

	// Verify correction was stored
	assert.Len(t, processor.corrections, 1)
	assert.NotEmpty(t, correction.ID)
	assert.False(t, correction.Timestamp.IsZero())
	assert.Greater(t, correction.LearningWeight, 0.0)

	// Verify learning model was updated
	assert.Equal(t, 1, processor.learningModel.TotalCorrections)
	assert.Equal(t, 1, processor.learningModel.PerformanceMetrics.TotalCorrections)

	// Check logs
	logs := logger.GetLogs()
	found := false
	for _, log := range logs {
		if correctionStringContains(log, "Correction detected and processed") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected correction processing log")
}

func TestCorrectionProcessor_CorrectionTypes(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	testCases := []struct {
		name           string
		correctionType CorrectionType
		expectedWeight float64
	}{
		{"Field Value", FieldValueCorrection, 0.8},
		{"Field Mapping", FieldMappingCorrection, 1.2},
		{"Template Selection", TemplateCorrection, 1.5},
		{"Formula", FormulaCorrection, 1.3},
		{"Validation", ValidationCorrection, 0.6},
		{"Category", CategoryCorrection, 1.1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			correction := &CorrectionEntry{
				DealID:         fmt.Sprintf("deal_%s", tc.name),
				FieldName:      "test_field",
				OriginalValue:  "original",
				CorrectedValue: "corrected",
				CorrectionType: tc.correctionType,
				UserID:         "user_001",
			}

			err := processor.DetectCorrection(correction)
			assert.NoError(t, err)
			assert.InDelta(t, tc.expectedWeight, correction.LearningWeight, 0.1)
		})
	}
}

func TestCorrectionProcessor_MonitorTemplateChanges(t *testing.T) {
	processor, logger, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	beforeData := map[string]interface{}{
		"purchase_price": "1000000",
		"company_name":   "Test Corp",
		"deal_date":      "2024-01-01",
	}

	afterData := map[string]interface{}{
		"purchase_price": "1500000",    // Changed
		"company_name":   "Test Corp",  // Same
		"deal_date":      "2024-01-15", // Changed
		"new_field":      "new_value",  // Added
	}

	err := processor.MonitorTemplateChanges("deal_001", "template_001", beforeData, afterData, "user_001")
	assert.NoError(t, err)

	// Should detect 3 changes: purchase_price, deal_date, new_field
	assert.Equal(t, 3, len(processor.corrections))

	// Verify change types
	changeTypes := make(map[string]string)
	for _, correction := range processor.corrections {
		changeTypes[correction.FieldName] = correction.Context["change_type"].(string)
	}

	assert.Equal(t, "replacement", changeTypes["purchase_price"])
	assert.Equal(t, "replacement", changeTypes["deal_date"])
	assert.Equal(t, "addition", changeTypes["new_field"])

	// Check logs
	logs := logger.GetLogs()
	assert.True(t, len(logs) > 0)
}

func TestCorrectionProcessor_PatternDetection(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	// Create multiple similar corrections to trigger pattern detection
	corrections := []*CorrectionEntry{
		{
			DealID:         "deal_001",
			FieldName:      "purchase_price",
			OriginalValue:  "1000000",
			CorrectedValue: "1,000,000",
			CorrectionType: FieldValueCorrection,
			UserID:         "user_001",
		},
		{
			DealID:         "deal_002",
			FieldName:      "purchase_price",
			OriginalValue:  "2000000",
			CorrectedValue: "2,000,000",
			CorrectionType: FieldValueCorrection,
			UserID:         "user_001",
		},
		{
			DealID:         "deal_003",
			FieldName:      "purchase_price",
			OriginalValue:  "3000000",
			CorrectedValue: "3,000,000",
			CorrectionType: FieldValueCorrection,
			UserID:         "user_001",
		},
	}

	for _, correction := range corrections {
		err := processor.DetectCorrection(correction)
		assert.NoError(t, err)
	}

	// Check that pattern was detected
	patternKey := fmt.Sprintf("%s_%s", FieldValueCorrection, "purchase_price")
	pattern := processor.patternDetector.patterns[patternKey]
	assert.NotNil(t, pattern)
	assert.Equal(t, 3, pattern.FrequencyCount)
	assert.Equal(t, ConfidenceMedium, pattern.Confidence)
	assert.True(t, pattern.IsActive)
}

func TestCorrectionProcessor_RAGLearning(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	// Enable RAG learning
	processor.config.EnableRAGLearning = true

	correction := &CorrectionEntry{
		DealID:         "deal_001",
		FieldName:      "company_valuation",
		OriginalValue:  "10000000",
		CorrectedValue: "12000000",
		CorrectionType: FieldValueCorrection,
		UserID:         "user_001",
	}

	err := processor.DetectCorrection(correction)
	assert.NoError(t, err)

	// Check RAG learning storage
	key := fmt.Sprintf("correction_%s_%s", correction.CorrectionType, correction.FieldName)
	knowledgeData := processor.ragLearning.knowledgeBase[key]
	assert.NotNil(t, knowledgeData)

	// Check vector store
	embedding := processor.ragLearning.vectorStore[correction.ID]
	assert.NotNil(t, embedding)
	assert.Equal(t, 64, len(embedding))
}

func TestCorrectionProcessor_ApplyLearning(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	// First, train the system with a correction
	correction := &CorrectionEntry{
		DealID:         "deal_001",
		FieldName:      "purchase_price",
		OriginalValue:  "1000000",
		CorrectedValue: "1,000,000",
		CorrectionType: FieldValueCorrection,
		UserID:         "user_001",
	}

	err := processor.DetectCorrection(correction)
	assert.NoError(t, err)

	// Manually add pattern to learning model for testing
	patternID := "test_pattern_001"
	processor.learningModel.ActivePatterns[patternID] = &LearningPattern{
		ID:               patternID,
		PatternType:      string(FieldValueCorrection),
		FieldName:        "purchase_price",
		OriginalPattern:  "1000000",
		CorrectedPattern: "1,000,000",
		Confidence:       ConfidenceHigh,
		IsActive:         true,
		SuccessRate:      0.8,
	}

	// Apply learning to new document
	documentData := map[string]interface{}{
		"purchase_price": "2000000",
		"company_name":   "Test Corp",
	}

	context := ProcessingContext{
		DocumentCategory: "financial",
		DealType:         "acquisition",
		ProcessingMethod: "OCR",
	}

	result, err := processor.ApplyLearning(documentData, context)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.LearningApplied)
	assert.Contains(t, result.AppliedPatterns, patternID)
	assert.Contains(t, result.ConfidenceAdjustments, "purchase_price")
}

func TestCorrectionProcessor_LearningInsights(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	// Add some test corrections
	corrections := []*CorrectionEntry{
		{
			DealID:         "deal_001",
			FieldName:      "purchase_price",
			CorrectionType: FieldValueCorrection,
			UserID:         "user_001",
		},
		{
			DealID:         "deal_002",
			FieldName:      "company_name",
			CorrectionType: FieldMappingCorrection,
			UserID:         "user_001",
		},
		{
			DealID:         "deal_003",
			FieldName:      "deal_date",
			CorrectionType: ValidationCorrection,
			UserID:         "user_001",
		},
	}

	for _, correction := range corrections {
		err := processor.DetectCorrection(correction)
		assert.NoError(t, err)
	}

	insights, err := processor.GetLearningInsights()
	assert.NoError(t, err)
	assert.NotNil(t, insights)
	assert.Equal(t, 3, insights.TotalCorrections)
	assert.Equal(t, "1.0.0", insights.LearningVersion)
	assert.Equal(t, 3, insights.PerformanceMetrics.TotalCorrections)
	assert.Contains(t, insights.TopCorrections, "field_value")
	assert.Contains(t, insights.TopCorrections, "field_mapping")
	assert.Contains(t, insights.TopCorrections, "validation_override")
}

func TestCorrectionProcessor_ValidationErrors(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	testCases := []struct {
		name        string
		correction  *CorrectionEntry
		expectError bool
	}{
		{
			name: "Missing Deal ID",
			correction: &CorrectionEntry{
				FieldName:      "test_field",
				CorrectionType: FieldValueCorrection,
				UserID:         "user_001",
			},
			expectError: true,
		},
		{
			name: "Missing User ID",
			correction: &CorrectionEntry{
				DealID:         "deal_001",
				FieldName:      "test_field",
				CorrectionType: FieldValueCorrection,
			},
			expectError: true,
		},
		{
			name: "Missing Field Name for non-category correction",
			correction: &CorrectionEntry{
				DealID:         "deal_001",
				CorrectionType: FieldValueCorrection,
				UserID:         "user_001",
			},
			expectError: true,
		},
		{
			name: "Valid Category Correction without Field Name",
			correction: &CorrectionEntry{
				DealID:         "deal_001",
				CorrectionType: CategoryCorrection,
				UserID:         "user_001",
			},
			expectError: false,
		},
		{
			name: "Valid Correction",
			correction: &CorrectionEntry{
				DealID:         "deal_001",
				FieldName:      "test_field",
				CorrectionType: FieldValueCorrection,
				UserID:         "user_001",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := processor.DetectCorrection(tc.correction)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCorrectionProcessor_Persistence(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)

	// Add some corrections
	correction := &CorrectionEntry{
		DealID:         "deal_001",
		FieldName:      "purchase_price",
		OriginalValue:  "1000000",
		CorrectedValue: "1,000,000",
		CorrectionType: FieldValueCorrection,
		UserID:         "user_001",
	}

	err := processor.DetectCorrection(correction)
	assert.NoError(t, err)

	// Save state
	err = processor.saveState()
	assert.NoError(t, err)

	// Verify state file exists
	statePath := filepath.Join(tempDir, "correction_processor_state.json")
	assert.FileExists(t, statePath)

	// Shutdown processor
	err = processor.Shutdown()
	assert.NoError(t, err)

	// Create new processor with same config
	logger := &TestCorrectionLogger{}
	config := CorrectionDetectionConfig{
		StoragePath: tempDir,
	}
	newProcessor := NewCorrectionProcessor(config, logger)

	// Verify state was loaded
	assert.Equal(t, 1, len(newProcessor.corrections))
	assert.Equal(t, 1, newProcessor.learningModel.TotalCorrections)

	// Cleanup
	newProcessor.Shutdown()
	os.RemoveAll(tempDir)
}

func TestCorrectionProcessor_BackgroundProcessing(t *testing.T) {
	processor, logger, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	// Set very short monitoring interval for testing
	processor.config.MonitoringInterval = 100 * time.Millisecond

	// Add some corrections
	for i := 0; i < 5; i++ {
		correction := &CorrectionEntry{
			DealID:         fmt.Sprintf("deal_%03d", i),
			FieldName:      "test_field",
			CorrectionType: FieldValueCorrection,
			UserID:         "user_001",
		}
		err := processor.DetectCorrection(correction)
		assert.NoError(t, err)
	}

	// Wait for background processing
	time.Sleep(200 * time.Millisecond)

	// Check that background tasks ran (should see save state logs)
	logs := logger.GetLogs()
	assert.True(t, len(logs) > 0)
}

func TestCorrectionProcessor_HistoryCleanup(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	// Set small history limit
	processor.config.MaxCorrectionHistory = 3

	// Add more corrections than the limit
	for i := 0; i < 5; i++ {
		correction := &CorrectionEntry{
			DealID:         fmt.Sprintf("deal_%03d", i),
			FieldName:      "test_field",
			CorrectionType: FieldValueCorrection,
			UserID:         "user_001",
			Timestamp:      time.Now().Add(time.Duration(i) * time.Minute),
		}
		err := processor.DetectCorrection(correction)
		assert.NoError(t, err)
	}

	// Manually trigger cleanup
	processor.cleanupOldCorrections()

	// Should have only 3 corrections (the most recent ones)
	assert.Equal(t, 3, len(processor.corrections))
}

func TestCorrectionProcessor_ConcurrentAccess(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	// Test concurrent correction detection
	numRoutines := 10
	correctionsPerRoutine := 5
	done := make(chan bool, numRoutines)

	for i := 0; i < numRoutines; i++ {
		go func(routineID int) {
			for j := 0; j < correctionsPerRoutine; j++ {
				correction := &CorrectionEntry{
					DealID:         fmt.Sprintf("deal_%d_%d", routineID, j),
					FieldName:      "test_field",
					CorrectionType: FieldValueCorrection,
					UserID:         fmt.Sprintf("user_%d", routineID),
				}
				err := processor.DetectCorrection(correction)
				assert.NoError(t, err)
			}
			done <- true
		}(i)
	}

	// Wait for all routines to complete
	for i := 0; i < numRoutines; i++ {
		<-done
	}

	// Verify all corrections were processed
	expectedTotal := numRoutines * correctionsPerRoutine
	assert.Equal(t, expectedTotal, len(processor.corrections))
	assert.Equal(t, expectedTotal, processor.learningModel.TotalCorrections)
}

func TestCorrectionProcessor_LearningWeightCalculation(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	testCases := []struct {
		name               string
		correctionType     CorrectionType
		originalConfidence float64
		expectedMinWeight  float64
		expectedMaxWeight  float64
	}{
		{"Template Correction High Confidence", TemplateCorrection, 0.9, 0.1, 0.2},
		{"Template Correction Low Confidence", TemplateCorrection, 0.1, 1.0, 1.5},
		{"Field Value Correction", FieldValueCorrection, 0.5, 0.3, 0.5},
		{"Validation Correction", ValidationCorrection, 0.7, 0.1, 0.3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			correction := &CorrectionEntry{
				DealID:             fmt.Sprintf("deal_%s", tc.name),
				FieldName:          "test_field",
				CorrectionType:     tc.correctionType,
				OriginalConfidence: tc.originalConfidence,
				UserID:             "user_001",
			}

			err := processor.DetectCorrection(correction)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, correction.LearningWeight, tc.expectedMinWeight)
			assert.LessOrEqual(t, correction.LearningWeight, tc.expectedMaxWeight)
		})
	}
}

func TestCorrectionProcessor_PerformanceMetrics(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	// Add corrections of different types
	corrections := []*CorrectionEntry{
		{
			DealID:         "deal_001",
			FieldName:      "field1",
			CorrectionType: FieldValueCorrection,
			UserID:         "user_001",
		},
		{
			DealID:         "deal_002",
			FieldName:      "field1",
			CorrectionType: FieldValueCorrection,
			UserID:         "user_001",
		},
		{
			DealID:         "deal_003",
			FieldName:      "field2",
			CorrectionType: FieldMappingCorrection,
			UserID:         "user_001",
		},
	}

	for _, correction := range corrections {
		err := processor.DetectCorrection(correction)
		assert.NoError(t, err)
	}

	// Check performance metrics
	metrics := processor.learningModel.PerformanceMetrics
	assert.Equal(t, 3, metrics.TotalCorrections)
	assert.Equal(t, 2, metrics.CorrectionFrequency["field_value"])
	assert.Equal(t, 1, metrics.CorrectionFrequency["field_mapping"])
	assert.NotZero(t, metrics.LastEvaluationDate)

	// Check field accuracy tracking
	assert.Contains(t, metrics.FieldAccuracy, "field1")
	assert.Contains(t, metrics.FieldAccuracy, "field2")
}

func TestCorrectionProcessor_RecommendationGeneration(t *testing.T) {
	processor, _, tempDir := createTestCorrectionProcessor(t)
	defer func() {
		processor.Shutdown()
		os.RemoveAll(tempDir)
	}()

	// Set metrics to trigger recommendations
	processor.learningModel.PerformanceMetrics.LearningEffectiveness = 0.5 // Low effectiveness
	processor.learningModel.PerformanceMetrics.TotalProcessedDocs = 10
	processor.learningModel.PerformanceMetrics.TotalCorrections = 4 // 40% correction rate

	insights, err := processor.GetLearningInsights()
	assert.NoError(t, err)
	assert.NotEmpty(t, insights.RecommendedActions)

	// Should recommend training data quality improvement
	found := false
	for _, recommendation := range insights.RecommendedActions {
		if correctionStringContains(recommendation, "training data quality") {
			found = true
			break
		}
	}
	assert.True(t, found)
}

// Helper function for string contains
func correctionStringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
