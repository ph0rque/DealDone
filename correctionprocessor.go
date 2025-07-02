package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CorrectionType defines the type of correction made
type CorrectionType string

const (
	FieldValueCorrection   CorrectionType = "field_value"
	FieldMappingCorrection CorrectionType = "field_mapping"
	TemplateCorrection     CorrectionType = "template_selection"
	FormulaCorrection      CorrectionType = "formula_correction"
	ValidationCorrection   CorrectionType = "validation_override"
	CategoryCorrection     CorrectionType = "document_category"
)

// LearningConfidence represents the confidence level of learned patterns
type LearningConfidence string

const (
	ConfidenceLow      LearningConfidence = "low"
	ConfidenceMedium   LearningConfidence = "medium"
	ConfidenceHigh     LearningConfidence = "high"
	ConfidenceVeryHigh LearningConfidence = "very_high"
)

// CorrectionEntry represents a single user correction
type CorrectionEntry struct {
	ID                 string                 `json:"id"`
	DealID             string                 `json:"deal_id"`
	DocumentID         string                 `json:"document_id"`
	TemplateID         string                 `json:"template_id"`
	FieldName          string                 `json:"field_name"`
	OriginalValue      interface{}            `json:"original_value"`
	CorrectedValue     interface{}            `json:"corrected_value"`
	CorrectionType     CorrectionType         `json:"correction_type"`
	UserID             string                 `json:"user_id"`
	Timestamp          time.Time              `json:"timestamp"`
	Context            map[string]interface{} `json:"context"`
	ProcessingMethod   string                 `json:"processing_method"`
	OriginalConfidence float64                `json:"original_confidence"`
	CorrectionReason   string                 `json:"correction_reason,omitempty"`
	ValidationStatus   string                 `json:"validation_status"`
	LearningWeight     float64                `json:"learning_weight"`
	AppliedToModel     bool                   `json:"applied_to_model"`
	EffectivenessScore *float64               `json:"effectiveness_score,omitempty"`
}

// LearningPattern represents a learned pattern from corrections
type LearningPattern struct {
	ID                 string                 `json:"id"`
	PatternType        string                 `json:"pattern_type"`
	DocumentCategory   string                 `json:"document_category"`
	FieldName          string                 `json:"field_name"`
	OriginalPattern    string                 `json:"original_pattern"`
	CorrectedPattern   string                 `json:"corrected_pattern"`
	Confidence         LearningConfidence     `json:"confidence"`
	SupportingExamples []string               `json:"supporting_examples"`
	FrequencyCount     int                    `json:"frequency_count"`
	LastSeen           time.Time              `json:"last_seen"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	SuccessRate        float64                `json:"success_rate"`
	Context            map[string]interface{} `json:"context"`
	IsActive           bool                   `json:"is_active"`
}

// LearningModel represents the overall learning state
type LearningModel struct {
	Version               string                      `json:"version"`
	LastUpdated           time.Time                   `json:"last_updated"`
	TotalCorrections      int                         `json:"total_corrections"`
	ActivePatterns        map[string]*LearningPattern `json:"active_patterns"`
	ConfidenceAdjustments map[string]float64          `json:"confidence_adjustments"`
	FieldMappingRules     map[string][]string         `json:"field_mapping_rules"`
	ValidationRules       []ValidationRule            `json:"validation_rules"`
	PerformanceMetrics    PerformanceMetrics          `json:"performance_metrics"`
}

// ValidationRule represents a learned validation rule
type ValidationRule struct {
	ID           string    `json:"id"`
	FieldName    string    `json:"field_name"`
	RuleType     string    `json:"rule_type"`
	Pattern      string    `json:"pattern"`
	ErrorMessage string    `json:"error_message"`
	Confidence   float64   `json:"confidence"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// PerformanceMetrics tracks the effectiveness of learning
type PerformanceMetrics struct {
	AccuracyImprovement   float64            `json:"accuracy_improvement"`
	CorrectionFrequency   map[string]int     `json:"correction_frequency"`
	FieldAccuracy         map[string]float64 `json:"field_accuracy"`
	CategoryAccuracy      map[string]float64 `json:"category_accuracy"`
	LearningEffectiveness float64            `json:"learning_effectiveness"`
	TotalProcessedDocs    int                `json:"total_processed_docs"`
	TotalCorrections      int                `json:"total_corrections"`
	PatternsLearned       int                `json:"patterns_learned"`
	LastEvaluationDate    time.Time          `json:"last_evaluation_date"`
}

// CorrectionDetectionConfig holds configuration for correction detection
type CorrectionDetectionConfig struct {
	MonitoringInterval         time.Duration `json:"monitoring_interval"`
	MinLearningWeight          float64       `json:"min_learning_weight"`
	PatternConfidenceThreshold float64       `json:"pattern_confidence_threshold"`
	MaxPatternAge              time.Duration `json:"max_pattern_age"`
	MinFrequencyForPattern     int           `json:"min_frequency_for_pattern"`
	EnableRAGLearning          bool          `json:"enable_rag_learning"`
	LearningRateDecay          float64       `json:"learning_rate_decay"`
	ValidationThreshold        float64       `json:"validation_threshold"`
	StoragePath                string        `json:"storage_path"`
	BackupInterval             time.Duration `json:"backup_interval"`
	MaxCorrectionHistory       int           `json:"max_correction_history"`
}

// CorrectionProcessor handles user correction detection and learning
type CorrectionProcessor struct {
	config          CorrectionDetectionConfig
	corrections     map[string]*CorrectionEntry
	learningModel   *LearningModel
	patternDetector *PatternDetector
	ragLearning     *RAGLearningEngine
	mutex           sync.RWMutex
	logger          Logger
	ctx             context.Context
	cancel          context.CancelFunc
	lastModelUpdate time.Time
}

// PatternDetector identifies patterns in corrections
type PatternDetector struct {
	patterns     map[string]*LearningPattern
	textAnalyzer *TextAnalyzer
	mutex        sync.RWMutex
}

// RAGLearningEngine implements Retrieval-Augmented Generation learning
type RAGLearningEngine struct {
	knowledgeBase  map[string]interface{}
	vectorStore    map[string][]float64
	contextWindow  int
	embeddingCache map[string][]float64
	learningRate   float64
	mutex          sync.RWMutex
}

// TextAnalyzer provides text analysis capabilities
type TextAnalyzer struct {
	stopWords map[string]bool
	tokenizer *SimpleTokenizer
}

// SimpleTokenizer for basic text tokenization
type SimpleTokenizer struct {
	separators []string
}

// NewCorrectionProcessor creates a new correction processor service
func NewCorrectionProcessor(config CorrectionDetectionConfig, logger Logger) *CorrectionProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	processor := &CorrectionProcessor{
		config:          config,
		corrections:     make(map[string]*CorrectionEntry),
		learningModel:   initializeLearningModel(),
		patternDetector: NewPatternDetector(),
		ragLearning:     NewRAGLearningEngine(),
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		lastModelUpdate: time.Now(),
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
		logger.Error("Failed to create storage directory: %v", err)
	}

	// Load existing state
	if err := processor.loadState(); err != nil {
		logger.Warn("Failed to load existing state: %v", err)
	}

	// Start background monitoring
	go processor.startBackgroundProcessing()

	return processor
}

// DetectCorrection identifies and records a user correction
func (cp *CorrectionProcessor) DetectCorrection(correction *CorrectionEntry) error {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// Generate unique ID if not provided
	if correction.ID == "" {
		correction.ID = cp.generateCorrectionID(correction)
	}

	// Set timestamp if not provided
	if correction.Timestamp.IsZero() {
		correction.Timestamp = time.Now()
	}

	// Calculate learning weight based on context
	correction.LearningWeight = cp.calculateLearningWeight(correction)

	// Validate correction
	if err := cp.validateCorrection(correction); err != nil {
		return fmt.Errorf("invalid correction: %v", err)
	}

	// Store correction
	cp.corrections[correction.ID] = correction

	// Update learning model
	if err := cp.updateLearningModel(correction); err != nil {
		cp.logger.Error("Failed to update learning model: %v", err)
	}

	// Detect patterns
	if err := cp.patternDetector.AnalyzeCorrection(correction); err != nil {
		cp.logger.Error("Failed to analyze correction patterns: %v", err)
	}

	// Apply RAG learning if enabled
	if cp.config.EnableRAGLearning {
		if err := cp.ragLearning.ProcessCorrection(correction); err != nil {
			cp.logger.Error("Failed to process RAG learning: %v", err)
		}
	}

	cp.logger.Info("Correction detected and processed: %s (type: %s, field: %s)",
		correction.ID, correction.CorrectionType, correction.FieldName)

	return nil
}

// MonitorTemplateChanges monitors for changes in template data
func (cp *CorrectionProcessor) MonitorTemplateChanges(dealID, templateID string, beforeData, afterData map[string]interface{}, userID string) error {
	changes := cp.detectDataChanges(beforeData, afterData)

	for fieldName, change := range changes {
		correction := &CorrectionEntry{
			DealID:           dealID,
			TemplateID:       templateID,
			FieldName:        fieldName,
			OriginalValue:    change["original"],
			CorrectedValue:   change["corrected"],
			CorrectionType:   FieldValueCorrection,
			UserID:           userID,
			Timestamp:        time.Now(),
			Context:          make(map[string]interface{}),
			ProcessingMethod: "user_input",
			ValidationStatus: "pending",
		}

		// Add context information
		correction.Context["data_size"] = len(fmt.Sprintf("%v", change["corrected"]))
		correction.Context["change_type"] = cp.classifyChange(change["original"], change["corrected"])

		if err := cp.DetectCorrection(correction); err != nil {
			cp.logger.Error("Failed to process template change for field %s: %v", fieldName, err)
		}
	}

	return nil
}

// GetLearningInsights returns insights from the learning model
func (cp *CorrectionProcessor) GetLearningInsights() (*LearningInsights, error) {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	insights := &LearningInsights{
		TotalCorrections:   len(cp.corrections),
		ActivePatterns:     len(cp.learningModel.ActivePatterns),
		LearningVersion:    cp.learningModel.Version,
		LastUpdate:         cp.learningModel.LastUpdated,
		PerformanceMetrics: cp.learningModel.PerformanceMetrics,
		TopCorrections:     cp.getTopCorrectionTypes(),
		ImprovementTrends:  cp.calculateImprovementTrends(),
		RecommendedActions: cp.generateRecommendations(),
	}

	return insights, nil
}

// ApplyLearning applies learned patterns to new document processing
func (cp *CorrectionProcessor) ApplyLearning(documentData map[string]interface{}, context ProcessingContext) (*ProcessingResult, error) {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	result := &ProcessingResult{
		EnhancedData:          make(map[string]interface{}),
		ConfidenceAdjustments: make(map[string]float64),
		AppliedPatterns:       make([]string, 0),
		LearningApplied:       false,
	}

	// Copy original data
	for k, v := range documentData {
		result.EnhancedData[k] = v
	}

	// Apply learned patterns
	for patternID, pattern := range cp.learningModel.ActivePatterns {
		if cp.patternApplies(pattern, context) {
			if enhancement := cp.applyPattern(pattern, result.EnhancedData); enhancement != nil {
				result.LearningApplied = true
				result.AppliedPatterns = append(result.AppliedPatterns, patternID)

				// Update confidence adjustments
				if pattern.FieldName != "" {
					result.ConfidenceAdjustments[pattern.FieldName] = cp.calculateConfidenceAdjustment(pattern)
				}
			}
		}
	}

	// Apply RAG-based enhancements
	if cp.config.EnableRAGLearning {
		if ragResult := cp.ragLearning.EnhanceProcessing(result.EnhancedData, context); ragResult != nil {
			result.LearningApplied = true
			for k, v := range ragResult {
				result.EnhancedData[k] = v
			}
		}
	}

	return result, nil
}

// Helper functions

func (cp *CorrectionProcessor) detectDataChanges(before, after map[string]interface{}) map[string]map[string]interface{} {
	changes := make(map[string]map[string]interface{})

	// Check for modified fields
	for key, afterValue := range after {
		if beforeValue, exists := before[key]; exists {
			if !cp.valuesEqual(beforeValue, afterValue) {
				changes[key] = map[string]interface{}{
					"original":  beforeValue,
					"corrected": afterValue,
				}
			}
		} else {
			// New field added
			changes[key] = map[string]interface{}{
				"original":  nil,
				"corrected": afterValue,
			}
		}
	}

	// Check for removed fields
	for key, beforeValue := range before {
		if _, exists := after[key]; !exists {
			changes[key] = map[string]interface{}{
				"original":  beforeValue,
				"corrected": nil,
			}
		}
	}

	return changes
}

func (cp *CorrectionProcessor) valuesEqual(a, b interface{}) bool {
	// Simple equality check - could be enhanced with deep comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return aStr == bStr
}

func (cp *CorrectionProcessor) classifyChange(original, corrected interface{}) string {
	if original == nil {
		return "addition"
	}
	if corrected == nil {
		return "deletion"
	}

	origStr := strings.ToLower(fmt.Sprintf("%v", original))
	corrStr := strings.ToLower(fmt.Sprintf("%v", corrected))

	if strings.Contains(corrStr, origStr) || strings.Contains(origStr, corrStr) {
		return "refinement"
	}

	return "replacement"
}

func (cp *CorrectionProcessor) calculateLearningWeight(correction *CorrectionEntry) float64 {
	weight := 1.0

	// Adjust based on correction type
	switch correction.CorrectionType {
	case FieldValueCorrection:
		weight = 0.8
	case FieldMappingCorrection:
		weight = 1.2
	case TemplateCorrection:
		weight = 1.5
	case FormulaCorrection:
		weight = 1.3
	case ValidationCorrection:
		weight = 0.6
	case CategoryCorrection:
		weight = 1.1
	}

	// Adjust based on original confidence
	if correction.OriginalConfidence > 0 {
		weight *= (1.0 - correction.OriginalConfidence)
	}

	// Minimum weight
	if weight < cp.config.MinLearningWeight {
		weight = cp.config.MinLearningWeight
	}

	return weight
}

func (cp *CorrectionProcessor) validateCorrection(correction *CorrectionEntry) error {
	if correction.DealID == "" {
		return fmt.Errorf("deal ID is required")
	}
	if correction.FieldName == "" && correction.CorrectionType != CategoryCorrection {
		return fmt.Errorf("field name is required for most correction types")
	}
	if correction.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	return nil
}

func (cp *CorrectionProcessor) updateLearningModel(correction *CorrectionEntry) error {
	// Update model statistics
	cp.learningModel.TotalCorrections++
	cp.learningModel.LastUpdated = time.Now()

	// Update performance metrics
	cp.updatePerformanceMetrics(correction)

	// Update confidence adjustments if applicable
	if correction.FieldName != "" {
		if _, exists := cp.learningModel.ConfidenceAdjustments[correction.FieldName]; !exists {
			cp.learningModel.ConfidenceAdjustments[correction.FieldName] = 0.0
		}
		cp.learningModel.ConfidenceAdjustments[correction.FieldName] += correction.LearningWeight * 0.1
	}

	return nil
}

func (cp *CorrectionProcessor) updatePerformanceMetrics(correction *CorrectionEntry) {
	metrics := &cp.learningModel.PerformanceMetrics
	metrics.TotalCorrections++

	// Update correction frequency
	if metrics.CorrectionFrequency == nil {
		metrics.CorrectionFrequency = make(map[string]int)
	}
	metrics.CorrectionFrequency[string(correction.CorrectionType)]++

	// Update field accuracy tracking
	if metrics.FieldAccuracy == nil {
		metrics.FieldAccuracy = make(map[string]float64)
	}

	if correction.FieldName != "" {
		currentAccuracy := metrics.FieldAccuracy[correction.FieldName]
		// Decrease accuracy when correction is needed
		newAccuracy := currentAccuracy * 0.95
		metrics.FieldAccuracy[correction.FieldName] = newAccuracy
	}

	metrics.LastEvaluationDate = time.Now()
}

func (cp *CorrectionProcessor) generateCorrectionID(correction *CorrectionEntry) string {
	data := fmt.Sprintf("%s:%s:%s:%s:%d",
		correction.DealID, correction.TemplateID, correction.FieldName,
		correction.UserID, correction.Timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// Background processing

func (cp *CorrectionProcessor) startBackgroundProcessing() {
	ticker := time.NewTicker(cp.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cp.performBackgroundTasks()
		case <-cp.ctx.Done():
			cp.logger.Info("Stopping correction processor background processing")
			return
		}
	}
}

func (cp *CorrectionProcessor) performBackgroundTasks() {
	// Update patterns
	cp.patternDetector.UpdatePatterns()

	// Evaluate learning effectiveness
	cp.evaluateLearningEffectiveness()

	// Cleanup old corrections
	cp.cleanupOldCorrections()

	// Save state
	if err := cp.saveState(); err != nil {
		cp.logger.Error("Failed to save state: %v", err)
	}
}

func (cp *CorrectionProcessor) evaluateLearningEffectiveness() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// Calculate overall learning effectiveness
	totalPatterns := len(cp.learningModel.ActivePatterns)
	if totalPatterns == 0 {
		return
	}

	totalEffectiveness := 0.0
	activePatterns := 0

	for _, pattern := range cp.learningModel.ActivePatterns {
		if pattern.IsActive && pattern.SuccessRate > 0 {
			totalEffectiveness += pattern.SuccessRate
			activePatterns++
		}
	}

	if activePatterns > 0 {
		cp.learningModel.PerformanceMetrics.LearningEffectiveness = totalEffectiveness / float64(activePatterns)
	}
}

func (cp *CorrectionProcessor) cleanupOldCorrections() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	if len(cp.corrections) <= cp.config.MaxCorrectionHistory {
		return
	}

	// Sort corrections by timestamp and keep only the most recent
	type correctionWithTime struct {
		id   string
		time time.Time
	}

	corrections := make([]correctionWithTime, 0, len(cp.corrections))
	for id, correction := range cp.corrections {
		corrections = append(corrections, correctionWithTime{id: id, time: correction.Timestamp})
	}

	// Sort by timestamp (most recent first)
	for i := 0; i < len(corrections)-1; i++ {
		for j := i + 1; j < len(corrections); j++ {
			if corrections[i].time.Before(corrections[j].time) {
				corrections[i], corrections[j] = corrections[j], corrections[i]
			}
		}
	}

	// Remove old corrections
	toRemove := len(corrections) - cp.config.MaxCorrectionHistory
	for i := cp.config.MaxCorrectionHistory; i < len(corrections); i++ {
		delete(cp.corrections, corrections[i].id)
	}

	cp.logger.Info("Cleaned up %d old corrections", toRemove)
}

// Persistence methods

func (cp *CorrectionProcessor) saveState() error {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	state := struct {
		Corrections   map[string]*CorrectionEntry `json:"corrections"`
		LearningModel *LearningModel              `json:"learning_model"`
		LastSaved     time.Time                   `json:"last_saved"`
	}{
		Corrections:   cp.corrections,
		LearningModel: cp.learningModel,
		LastSaved:     time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %v", err)
	}

	statePath := filepath.Join(cp.config.StoragePath, "correction_processor_state.json")
	tempPath := statePath + ".tmp"

	if err := ioutil.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp state file: %v", err)
	}

	if err := os.Rename(tempPath, statePath); err != nil {
		return fmt.Errorf("failed to rename temp state file: %v", err)
	}

	return nil
}

func (cp *CorrectionProcessor) loadState() error {
	statePath := filepath.Join(cp.config.StoragePath, "correction_processor_state.json")

	data, err := ioutil.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			cp.logger.Info("No existing correction processor state found, starting fresh")
			return nil
		}
		return fmt.Errorf("failed to read state file: %v", err)
	}

	var state struct {
		Corrections   map[string]*CorrectionEntry `json:"corrections"`
		LearningModel *LearningModel              `json:"learning_model"`
		LastSaved     time.Time                   `json:"last_saved"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %v", err)
	}

	cp.corrections = state.Corrections
	if cp.corrections == nil {
		cp.corrections = make(map[string]*CorrectionEntry)
	}

	if state.LearningModel != nil {
		cp.learningModel = state.LearningModel
	}

	cp.logger.Info("Correction processor state loaded successfully (last saved: %v)", state.LastSaved)
	return nil
}

// Shutdown gracefully shuts down the service
func (cp *CorrectionProcessor) Shutdown() error {
	cp.logger.Info("Shutting down CorrectionProcessor")

	cp.cancel()

	// Final state save
	if err := cp.saveState(); err != nil {
		cp.logger.Error("Failed to save final state during shutdown: %v", err)
		return err
	}

	cp.logger.Info("CorrectionProcessor shutdown complete")
	return nil
}

// Additional types for API responses

type LearningInsights struct {
	TotalCorrections   int                `json:"total_corrections"`
	ActivePatterns     int                `json:"active_patterns"`
	LearningVersion    string             `json:"learning_version"`
	LastUpdate         time.Time          `json:"last_update"`
	PerformanceMetrics PerformanceMetrics `json:"performance_metrics"`
	TopCorrections     map[string]int     `json:"top_corrections"`
	ImprovementTrends  map[string]float64 `json:"improvement_trends"`
	RecommendedActions []string           `json:"recommended_actions"`
}

type ProcessingContext struct {
	DocumentCategory string                 `json:"document_category"`
	DealType         string                 `json:"deal_type"`
	ProcessingMethod string                 `json:"processing_method"`
	UserContext      map[string]interface{} `json:"user_context"`
}

type ProcessingResult struct {
	EnhancedData          map[string]interface{} `json:"enhanced_data"`
	ConfidenceAdjustments map[string]float64     `json:"confidence_adjustments"`
	AppliedPatterns       []string               `json:"applied_patterns"`
	LearningApplied       bool                   `json:"learning_applied"`
}

// Initialize learning model
func initializeLearningModel() *LearningModel {
	return &LearningModel{
		Version:               "1.0.0",
		LastUpdated:           time.Now(),
		TotalCorrections:      0,
		ActivePatterns:        make(map[string]*LearningPattern),
		ConfidenceAdjustments: make(map[string]float64),
		FieldMappingRules:     make(map[string][]string),
		ValidationRules:       make([]ValidationRule, 0),
		PerformanceMetrics: PerformanceMetrics{
			AccuracyImprovement:   0.0,
			CorrectionFrequency:   make(map[string]int),
			FieldAccuracy:         make(map[string]float64),
			CategoryAccuracy:      make(map[string]float64),
			LearningEffectiveness: 0.0,
			TotalProcessedDocs:    0,
			TotalCorrections:      0,
			PatternsLearned:       0,
			LastEvaluationDate:    time.Now(),
		},
	}
}

// Helper methods for insights
func (cp *CorrectionProcessor) getTopCorrectionTypes() map[string]int {
	counts := make(map[string]int)
	for _, correction := range cp.corrections {
		counts[string(correction.CorrectionType)]++
	}
	return counts
}

func (cp *CorrectionProcessor) calculateImprovementTrends() map[string]float64 {
	// Simple trend calculation - could be enhanced
	trends := make(map[string]float64)
	trends["overall_accuracy"] = cp.learningModel.PerformanceMetrics.LearningEffectiveness
	trends["correction_rate"] = float64(cp.learningModel.PerformanceMetrics.TotalCorrections) /
		float64(max(cp.learningModel.PerformanceMetrics.TotalProcessedDocs, 1))
	return trends
}

func (cp *CorrectionProcessor) generateRecommendations() []string {
	recommendations := make([]string, 0)

	metrics := cp.learningModel.PerformanceMetrics

	if metrics.LearningEffectiveness < 0.7 {
		recommendations = append(recommendations, "Consider increasing training data quality")
	}

	if len(cp.learningModel.ActivePatterns) < 5 {
		recommendations = append(recommendations, "More correction data needed for better pattern recognition")
	}

	if float64(metrics.TotalCorrections) > float64(metrics.TotalProcessedDocs)*0.3 {
		recommendations = append(recommendations, "High correction rate indicates need for model retraining")
	}

	return recommendations
}

func (cp *CorrectionProcessor) patternApplies(pattern *LearningPattern, context ProcessingContext) bool {
	if !pattern.IsActive {
		return false
	}

	if pattern.DocumentCategory != "" && pattern.DocumentCategory != context.DocumentCategory {
		return false
	}

	if pattern.Confidence == ConfidenceLow {
		return false
	}

	return true
}

func (cp *CorrectionProcessor) applyPattern(pattern *LearningPattern, data map[string]interface{}) map[string]interface{} {
	// Simple pattern application - would be enhanced with actual ML logic
	if pattern.FieldName != "" {
		if value, exists := data[pattern.FieldName]; exists {
			// Apply pattern transformation
			if strings.Contains(fmt.Sprintf("%v", value), pattern.OriginalPattern) {
				newValue := strings.ReplaceAll(fmt.Sprintf("%v", value), pattern.OriginalPattern, pattern.CorrectedPattern)
				return map[string]interface{}{pattern.FieldName: newValue}
			}
		}
	}

	return nil
}

func (cp *CorrectionProcessor) calculateConfidenceAdjustment(pattern *LearningPattern) float64 {
	switch pattern.Confidence {
	case ConfidenceVeryHigh:
		return 0.15
	case ConfidenceHigh:
		return 0.10
	case ConfidenceMedium:
		return 0.05
	default:
		return 0.0
	}
}

// NewPatternDetector creates a new pattern detector
func NewPatternDetector() *PatternDetector {
	return &PatternDetector{
		patterns:     make(map[string]*LearningPattern),
		textAnalyzer: NewTextAnalyzer(),
		mutex:        sync.RWMutex{},
	}
}

// NewRAGLearningEngine creates a new RAG learning engine
func NewRAGLearningEngine() *RAGLearningEngine {
	return &RAGLearningEngine{
		knowledgeBase:  make(map[string]interface{}),
		vectorStore:    make(map[string][]float64),
		contextWindow:  512,
		embeddingCache: make(map[string][]float64),
		learningRate:   0.01,
		mutex:          sync.RWMutex{},
	}
}

// NewTextAnalyzer creates a new text analyzer
func NewTextAnalyzer() *TextAnalyzer {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
	}

	return &TextAnalyzer{
		stopWords: stopWords,
		tokenizer: &SimpleTokenizer{
			separators: []string{" ", "\t", "\n", ",", ".", ";", ":", "!", "?"},
		},
	}
}

// PatternDetector methods

func (pd *PatternDetector) AnalyzeCorrection(correction *CorrectionEntry) error {
	pd.mutex.Lock()
	defer pd.mutex.Unlock()

	// Simple pattern analysis - detect common correction patterns
	patternKey := fmt.Sprintf("%s_%s", correction.CorrectionType, correction.FieldName)

	pattern, exists := pd.patterns[patternKey]
	if !exists {
		// Create new pattern
		pattern = &LearningPattern{
			ID:                 generatePatternID(correction),
			PatternType:        string(correction.CorrectionType),
			FieldName:          correction.FieldName,
			OriginalPattern:    fmt.Sprintf("%v", correction.OriginalValue),
			CorrectedPattern:   fmt.Sprintf("%v", correction.CorrectedValue),
			Confidence:         ConfidenceLow,
			SupportingExamples: []string{correction.ID},
			FrequencyCount:     1,
			LastSeen:           correction.Timestamp,
			CreatedAt:          correction.Timestamp,
			UpdatedAt:          correction.Timestamp,
			SuccessRate:        0.0,
			Context:            make(map[string]interface{}),
			IsActive:           true,
		}
		pd.patterns[patternKey] = pattern
	} else {
		// Update existing pattern
		pattern.FrequencyCount++
		pattern.LastSeen = correction.Timestamp
		pattern.UpdatedAt = correction.Timestamp
		pattern.SupportingExamples = append(pattern.SupportingExamples, correction.ID)

		// Update confidence based on frequency
		if pattern.FrequencyCount >= 10 {
			pattern.Confidence = ConfidenceVeryHigh
		} else if pattern.FrequencyCount >= 5 {
			pattern.Confidence = ConfidenceHigh
		} else if pattern.FrequencyCount >= 3 {
			pattern.Confidence = ConfidenceMedium
		}
	}

	return nil
}

func (pd *PatternDetector) UpdatePatterns() {
	pd.mutex.Lock()
	defer pd.mutex.Unlock()

	// Clean up old or ineffective patterns
	for key, pattern := range pd.patterns {
		// Deactivate patterns that haven't been seen recently
		if time.Since(pattern.LastSeen) > 30*24*time.Hour { // 30 days
			pattern.IsActive = false
		}

		// Remove patterns with very low success rates
		if pattern.SuccessRate < 0.1 && pattern.FrequencyCount > 10 {
			delete(pd.patterns, key)
		}
	}
}

// RAGLearningEngine methods

func (rag *RAGLearningEngine) ProcessCorrection(correction *CorrectionEntry) error {
	rag.mutex.Lock()
	defer rag.mutex.Unlock()

	// Store correction in knowledge base
	key := fmt.Sprintf("correction_%s_%s", correction.CorrectionType, correction.FieldName)

	if existingData, exists := rag.knowledgeBase[key]; exists {
		// Update existing knowledge
		if dataList, ok := existingData.([]interface{}); ok {
			dataList = append(dataList, correction)
			rag.knowledgeBase[key] = dataList
		}
	} else {
		// Create new knowledge entry
		rag.knowledgeBase[key] = []interface{}{correction}
	}

	// Generate simple embedding (in real implementation, would use proper embedding model)
	embedding := rag.generateSimpleEmbedding(correction)
	rag.vectorStore[correction.ID] = embedding

	return nil
}

func (rag *RAGLearningEngine) EnhanceProcessing(data map[string]interface{}, context ProcessingContext) map[string]interface{} {
	rag.mutex.RLock()
	defer rag.mutex.RUnlock()

	enhanced := make(map[string]interface{})

	// Simple enhancement based on stored knowledge
	for field, value := range data {
		key := fmt.Sprintf("correction_%s_%s", FieldValueCorrection, field)
		if knowledgeData, exists := rag.knowledgeBase[key]; exists {
			if corrections, ok := knowledgeData.([]interface{}); ok && len(corrections) > 0 {
				// Apply most recent correction pattern
				if correction, ok := corrections[len(corrections)-1].(*CorrectionEntry); ok {
					if strings.Contains(fmt.Sprintf("%v", value), fmt.Sprintf("%v", correction.OriginalValue)) {
						enhanced[field] = correction.CorrectedValue
					}
				}
			}
		}
	}

	return enhanced
}

func (rag *RAGLearningEngine) generateSimpleEmbedding(correction *CorrectionEntry) []float64 {
	// Simple embedding generation - in practice would use proper embedding model
	text := fmt.Sprintf("%s %s %v %v", correction.CorrectionType, correction.FieldName,
		correction.OriginalValue, correction.CorrectedValue)

	// Create a simple hash-based embedding
	hash := sha256.Sum256([]byte(text))
	embedding := make([]float64, 64)

	for i := 0; i < 64; i++ {
		embedding[i] = float64(hash[i%32]) / 255.0
	}

	return embedding
}

func generatePatternID(correction *CorrectionEntry) string {
	data := fmt.Sprintf("pattern_%s_%s_%d", correction.CorrectionType, correction.FieldName, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:12]
}
