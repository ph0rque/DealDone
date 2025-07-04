package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Logger interface for conflict resolver logging
type Logger interface {
	Info(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// ConflictResolver manages intelligent data merging and conflict resolution
type ConflictResolver struct {
	mutex                sync.RWMutex
	conflictHistory      map[string][]*ConflictResolutionRecord
	auditTrail           []*ConflictAuditEntry
	resolutionStrategies map[string]ResolutionStrategy
	config               *ConflictResolutionConfig
	persistencePath      string
	logger               Logger
}

// ConflictResolutionRecord tracks details of conflict resolution
type ConflictResolutionRecord struct {
	ID                 string                 `json:"id"`
	DealName           string                 `json:"dealName"`
	TemplatePath       string                 `json:"templatePath"`
	FieldName          string                 `json:"fieldName"`
	ConflictType       string                 `json:"conflictType"`
	ConflictingValues  []ConflictingValue     `json:"conflictingValues"`
	ResolvedValue      interface{}            `json:"resolvedValue"`
	ResolutionStrategy string                 `json:"resolutionStrategy"`
	FinalConfidence    float64                `json:"finalConfidence"`
	RequiresReview     bool                   `json:"requiresReview"`
	ResolvedAt         time.Time              `json:"resolvedAt"`
	ResolvedBy         string                 `json:"resolvedBy"`
	Notes              string                 `json:"notes,omitempty"`
	PreviousValues     []interface{}          `json:"previousValues,omitempty"`
	Context            map[string]interface{} `json:"context,omitempty"`
}

// ConflictAuditEntry represents a single audit trail entry
type ConflictAuditEntry struct {
	ID              string      `json:"id"`
	ConflictID      string      `json:"conflictId"`
	Action          string      `json:"action"` // "detected", "resolved", "reviewed", "overridden"
	Timestamp       time.Time   `json:"timestamp"`
	UserID          string      `json:"userId,omitempty"`
	Details         string      `json:"details"`
	BeforeState     interface{} `json:"beforeState,omitempty"`
	AfterState      interface{} `json:"afterState,omitempty"`
	ConfidenceScore float64     `json:"confidenceScore"`
}

// ResolutionStrategy defines how conflicts should be resolved
type ResolutionStrategy struct {
	Name               string                                          `json:"name"`
	Description        string                                          `json:"description"`
	Priority           int                                             `json:"priority"`
	MinConfidenceDiff  float64                                         `json:"minConfidenceDiff"`
	ApplicableTypes    []string                                        `json:"applicableTypes"`
	ResolutionFunction func(*ConflictContext) (*ConflictResult, error) `json:"-"`
}

// ConflictContext provides context for conflict resolution
type ConflictContext struct {
	DealName          string                 `json:"dealName"`
	TemplatePath      string                 `json:"templatePath"`
	FieldName         string                 `json:"fieldName"`
	ConflictingValues []ConflictingValue     `json:"conflictingValues"`
	FieldType         string                 `json:"fieldType"`
	PreviousValue     interface{}            `json:"previousValue,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	RequiresReview    bool                   `json:"requiresReview"`
}

// ConflictResolutionConfig contains configuration for the conflict resolver
type ConflictResolutionConfig struct {
	MinConfidenceThreshold    float64           `json:"minConfidenceThreshold"`
	ReviewThreshold           float64           `json:"reviewThreshold"`
	NumericAveragingThreshold float64           `json:"numericAveragingThreshold"`
	MaxHistoryEntries         int               `json:"maxHistoryEntries"`
	PersistenceInterval       time.Duration     `json:"persistenceInterval"`
	DefaultStrategies         []string          `json:"defaultStrategies"`
	TypeSpecificStrategies    map[string]string `json:"typeSpecificStrategies"`
	EnableAuditTrail          bool              `json:"enableAuditTrail"`
	DebugMode                 bool              `json:"debugMode"`
}

// NewConflictResolver creates a new conflict resolver with default configuration
func NewConflictResolver(persistencePath string, logger Logger) *ConflictResolver {
	resolver := &ConflictResolver{
		conflictHistory:      make(map[string][]*ConflictResolutionRecord),
		auditTrail:           make([]*ConflictAuditEntry, 0),
		resolutionStrategies: make(map[string]ResolutionStrategy),
		persistencePath:      persistencePath,
		logger:               logger,
		config: &ConflictResolutionConfig{
			MinConfidenceThreshold:    0.7,
			ReviewThreshold:           0.5,
			NumericAveragingThreshold: 0.05, // 5% confidence difference
			MaxHistoryEntries:         1000,
			PersistenceInterval:       5 * time.Minute,
			DefaultStrategies:         []string{"highest_confidence", "numeric_averaging", "manual_review"},
			TypeSpecificStrategies: map[string]string{
				"number":  "numeric_averaging",
				"date":    "latest_value",
				"string":  "highest_confidence",
				"boolean": "highest_confidence",
			},
			EnableAuditTrail: true,
			DebugMode:        false,
		},
	}

	// Initialize default resolution strategies
	resolver.initializeDefaultStrategies()

	// Load existing data
	resolver.loadPersistedData()

	return resolver
}

// initializeDefaultStrategies sets up the built-in conflict resolution strategies
func (cr *ConflictResolver) initializeDefaultStrategies() {
	// Strategy 1: Highest Confidence Wins
	cr.resolutionStrategies["highest_confidence"] = ResolutionStrategy{
		Name:               "Highest Confidence",
		Description:        "Select the value with the highest confidence score",
		Priority:           1,
		MinConfidenceDiff:  0.1,
		ApplicableTypes:    []string{"string", "number", "date", "boolean"},
		ResolutionFunction: cr.resolveByHighestConfidence,
	}

	// Strategy 2: Numeric Averaging
	cr.resolutionStrategies["numeric_averaging"] = ResolutionStrategy{
		Name:               "Numeric Averaging",
		Description:        "Average numeric values when confidence scores are similar",
		Priority:           2,
		MinConfidenceDiff:  0.05,
		ApplicableTypes:    []string{"number"},
		ResolutionFunction: cr.resolveByNumericAveraging,
	}

	// Strategy 3: Latest Value (for dates/timestamps)
	cr.resolutionStrategies["latest_value"] = ResolutionStrategy{
		Name:               "Latest Value",
		Description:        "Select the most recent value for date/timestamp fields",
		Priority:           3,
		MinConfidenceDiff:  0.0,
		ApplicableTypes:    []string{"date", "timestamp"},
		ResolutionFunction: cr.resolveByLatestValue,
	}

	// Strategy 4: Manual Review Required
	cr.resolutionStrategies["manual_review"] = ResolutionStrategy{
		Name:               "Manual Review",
		Description:        "Flag conflicts requiring manual review",
		Priority:           99,
		MinConfidenceDiff:  0.0,
		ApplicableTypes:    []string{"string", "number", "date", "boolean"},
		ResolutionFunction: cr.resolveByManualReview,
	}

	// Strategy 5: Source Priority (weighted by extraction method)
	cr.resolutionStrategies["source_priority"] = ResolutionStrategy{
		Name:               "Source Priority",
		Description:        "Prioritize values based on extraction method reliability",
		Priority:           4,
		MinConfidenceDiff:  0.0,
		ApplicableTypes:    []string{"string", "number", "date", "boolean"},
		ResolutionFunction: cr.resolveBySourcePriority,
	}
}

// ResolveConflict resolves a conflict between multiple values for a field
func (cr *ConflictResolver) ResolveConflict(ctx context.Context, conflictCtx *ConflictContext) (*ConflictResult, error) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	if cr.logger != nil {
		cr.logger.Info("Resolving conflict for field: %s in deal: %s", conflictCtx.FieldName, conflictCtx.DealName)
	}

	// Create audit entry for conflict detection
	conflictID := uuid.New().String()
	cr.addAuditEntry(conflictID, "detected", "", fmt.Sprintf("Conflict detected for field %s with %d values",
		conflictCtx.FieldName, len(conflictCtx.ConflictingValues)), nil, nil, 0.0)

	// Determine conflict type and appropriate strategy
	conflictType := cr.determineConflictType(conflictCtx.ConflictingValues)
	strategy := cr.selectResolutionStrategy(conflictType, conflictCtx.FieldType, conflictCtx)

	if cr.logger != nil {
		cr.logger.Debug("Using resolution strategy: %s for conflict type: %s", strategy.Name, conflictType)
	}

	// Apply resolution strategy
	result, err := strategy.ResolutionFunction(conflictCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve conflict using strategy %s: %w", strategy.Name, err)
	}

	// Enhance result with metadata
	result.FieldName = conflictCtx.FieldName
	result.ConflictType = conflictType
	result.ConflictingValues = conflictCtx.ConflictingValues
	// Use the strategy key, not the display name
	for key, strat := range cr.resolutionStrategies {
		if strat.Name == strategy.Name {
			result.ResolutionMethod = key
			break
		}
	}

	// Create resolution record
	strategyKey := result.ResolutionMethod
	record := &ConflictResolutionRecord{
		ID:                 conflictID,
		DealName:           conflictCtx.DealName,
		TemplatePath:       conflictCtx.TemplatePath,
		FieldName:          conflictCtx.FieldName,
		ConflictType:       conflictType,
		ConflictingValues:  conflictCtx.ConflictingValues,
		ResolvedValue:      result.ResolvedValue,
		ResolutionStrategy: strategyKey,
		FinalConfidence:    result.FinalConfidence,
		RequiresReview:     result.RequiresReview,
		ResolvedAt:         time.Now(),
		ResolvedBy:         "system",
		Notes:              result.ResolutionNotes,
		Context:            conflictCtx.Metadata,
	}

	// Store in history
	historyKey := fmt.Sprintf("%s:%s:%s", conflictCtx.DealName, conflictCtx.TemplatePath, conflictCtx.FieldName)
	cr.conflictHistory[historyKey] = append(cr.conflictHistory[historyKey], record)

	// Add audit entry for resolution
	cr.addAuditEntry(conflictID, "resolved", "system",
		fmt.Sprintf("Conflict resolved using %s strategy", strategy.Name),
		conflictCtx.ConflictingValues, result.ResolvedValue, result.FinalConfidence)

	// Trim history if needed
	if len(cr.conflictHistory[historyKey]) > cr.config.MaxHistoryEntries {
		cr.conflictHistory[historyKey] = cr.conflictHistory[historyKey][len(cr.conflictHistory[historyKey])-cr.config.MaxHistoryEntries:]
	}

	return result, nil
}

// determineConflictType analyzes conflicting values to determine the type of conflict
func (cr *ConflictResolver) determineConflictType(values []ConflictingValue) string {
	if len(values) <= 1 {
		return "none"
	}

	// Check value types and determine if they're similar first
	allNumeric := true
	allString := true
	allSame := true

	firstValue := values[0].Value
	for _, v := range values {
		if _, isNum := v.Value.(float64); !isNum {
			if _, isInt := v.Value.(int); !isInt {
				allNumeric = false
			}
		}
		if _, isStr := v.Value.(string); !isStr {
			allString = false
		}
		if !cr.valuesEqual(v.Value, firstValue) {
			allSame = false
		}
	}

	// Check if all values are the same (regardless of confidence)
	if allSame {
		return "duplicate"
	}

	// Check confidence levels for non-duplicate values
	confidences := make([]float64, len(values))
	for i, v := range values {
		confidences[i] = v.Confidence
	}
	sort.Float64s(confidences)

	confDiff := confidences[len(confidences)-1] - confidences[0]

	if confDiff > cr.config.NumericAveragingThreshold {
		return "confidence"
	}

	// Determine type based on value types
	if allNumeric {
		return "numeric"
	}
	if allString {
		return "textual"
	}

	return "mixed"
}

// selectResolutionStrategy chooses the appropriate strategy based on conflict and field type
func (cr *ConflictResolver) selectResolutionStrategy(conflictType, fieldType string, conflictCtx *ConflictContext) ResolutionStrategy {
	// Check if manual review is explicitly required or confidence is too low
	maxConfidence := 0.0
	for _, cv := range conflictCtx.ConflictingValues {
		if cv.Confidence > maxConfidence {
			maxConfidence = cv.Confidence
		}
	}

	if conflictCtx.RequiresReview || maxConfidence < cr.config.ReviewThreshold {
		if strategy, exists := cr.resolutionStrategies["manual_review"]; exists {
			return strategy
		}
	}

	// Check for type-specific strategy first
	if strategyName, exists := cr.config.TypeSpecificStrategies[fieldType]; exists {
		if strategy, found := cr.resolutionStrategies[strategyName]; found {
			return strategy
		}
	}

	// Select strategy based on conflict type
	switch conflictType {
	case "numeric":
		if strategy, exists := cr.resolutionStrategies["numeric_averaging"]; exists {
			return strategy
		}
	case "confidence":
		if strategy, exists := cr.resolutionStrategies["highest_confidence"]; exists {
			return strategy
		}
	case "duplicate":
		if strategy, exists := cr.resolutionStrategies["highest_confidence"]; exists {
			return strategy
		}
	case "textual":
		// For textual conflicts, default to highest confidence
		if strategy, exists := cr.resolutionStrategies["highest_confidence"]; exists {
			return strategy
		}
	}

	// Default to highest confidence strategy
	return cr.resolutionStrategies["highest_confidence"]
}

// Resolution strategy implementations

// resolveByHighestConfidence selects the value with the highest confidence score
func (cr *ConflictResolver) resolveByHighestConfidence(ctx *ConflictContext) (*ConflictResult, error) {
	if len(ctx.ConflictingValues) == 0 {
		return nil, fmt.Errorf("no conflicting values provided")
	}

	// Sort by confidence descending
	values := make([]ConflictingValue, len(ctx.ConflictingValues))
	copy(values, ctx.ConflictingValues)
	sort.Slice(values, func(i, j int) bool {
		return values[i].Confidence > values[j].Confidence
	})

	highest := values[0]
	requiresReview := highest.Confidence < cr.config.ReviewThreshold

	notes := fmt.Sprintf("Selected value with highest confidence (%.3f) from %s using %s method",
		highest.Confidence, highest.Source, highest.Method)

	return &ConflictResult{
		ResolvedValue:    highest.Value,
		ResolutionMethod: "highest_confidence",
		FinalConfidence:  highest.Confidence,
		RequiresReview:   requiresReview,
		ResolutionNotes:  notes,
		Metadata: map[string]interface{}{
			"selected_source":   highest.Source,
			"selected_method":   highest.Method,
			"confidence_spread": values[0].Confidence - values[len(values)-1].Confidence,
			"total_candidates":  len(values),
		},
	}, nil
}

// resolveByNumericAveraging averages numeric values when confidence scores are similar
func (cr *ConflictResolver) resolveByNumericAveraging(ctx *ConflictContext) (*ConflictResult, error) {
	if len(ctx.ConflictingValues) == 0 {
		return nil, fmt.Errorf("no conflicting values provided")
	}

	// Extract numeric values and calculate weighted average
	var weightedSum, totalWeight float64
	validValues := 0

	for _, cv := range ctx.ConflictingValues {
		var numVal float64
		switch v := cv.Value.(type) {
		case float64:
			numVal = v
		case int:
			numVal = float64(v)
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				numVal = parsed
			} else {
				continue // Skip non-numeric strings
			}
		default:
			continue // Skip non-numeric values
		}

		weight := cv.Confidence
		weightedSum += numVal * weight
		totalWeight += weight
		validValues++
	}

	if validValues == 0 {
		return nil, fmt.Errorf("no valid numeric values found for averaging")
	}

	if totalWeight == 0 {
		return nil, fmt.Errorf("total confidence weight is zero")
	}

	averageValue := weightedSum / totalWeight
	averageConfidence := totalWeight / float64(validValues)

	// Round to reasonable precision for display
	roundedValue := math.Round(averageValue*100) / 100

	requiresReview := averageConfidence < cr.config.ReviewThreshold

	notes := fmt.Sprintf("Calculated weighted average (%.2f) from %d numeric values with average confidence %.3f",
		roundedValue, validValues, averageConfidence)

	metadata := map[string]interface{}{
		"averaging_method":     "weighted",
		"source_values":        validValues,
		"raw_average":          averageValue,
		"rounded_average":      roundedValue,
		"average_confidence":   averageConfidence,
		"confidence_weighting": true,
	}

	return &ConflictResult{
		ResolvedValue:    roundedValue,
		ResolutionMethod: "numeric_averaging",
		FinalConfidence:  averageConfidence,
		RequiresReview:   requiresReview,
		ResolutionNotes:  notes,
		Metadata:         metadata,
	}, nil
}

// resolveByLatestValue selects the most recent value for date/timestamp fields
func (cr *ConflictResolver) resolveByLatestValue(ctx *ConflictContext) (*ConflictResult, error) {
	if len(ctx.ConflictingValues) == 0 {
		return nil, fmt.Errorf("no conflicting values provided")
	}

	// Sort by timestamp descending to get the latest
	values := make([]ConflictingValue, len(ctx.ConflictingValues))
	copy(values, ctx.ConflictingValues)
	sort.Slice(values, func(i, j int) bool {
		return values[i].Timestamp > values[j].Timestamp
	})

	latest := values[0]
	requiresReview := latest.Confidence < cr.config.ReviewThreshold

	notes := fmt.Sprintf("Selected most recent value (timestamp: %d) with confidence %.3f from %s",
		latest.Timestamp, latest.Confidence, latest.Source)

	return &ConflictResult{
		ResolvedValue:    latest.Value,
		ResolutionMethod: "latest_value",
		FinalConfidence:  latest.Confidence,
		RequiresReview:   requiresReview,
		ResolutionNotes:  notes,
		Metadata: map[string]interface{}{
			"selected_timestamp": latest.Timestamp,
			"selected_source":    latest.Source,
			"total_candidates":   len(values),
			"time_span_seconds":  values[0].Timestamp - values[len(values)-1].Timestamp,
		},
	}, nil
}

// resolveByManualReview flags conflicts for manual review
func (cr *ConflictResolver) resolveByManualReview(ctx *ConflictContext) (*ConflictResult, error) {
	if len(ctx.ConflictingValues) == 0 {
		return nil, fmt.Errorf("no conflicting values provided")
	}

	// Default to the first value but mark for review
	defaultValue := ctx.ConflictingValues[0]

	notes := fmt.Sprintf("Conflict requires manual review - %d values with varying confidence levels",
		len(ctx.ConflictingValues))

	return &ConflictResult{
		ResolvedValue:    defaultValue.Value,
		ResolutionMethod: "manual_review",
		FinalConfidence:  0.0, // Zero confidence to indicate manual review needed
		RequiresReview:   true,
		ResolutionNotes:  notes,
		Metadata: map[string]interface{}{
			"review_reason":    "confidence_levels_require_human_judgment",
			"candidate_count":  len(ctx.ConflictingValues),
			"default_selected": true,
			"review_priority":  "high",
		},
	}, nil
}

// resolveBySourcePriority prioritizes values based on extraction method reliability
func (cr *ConflictResolver) resolveBySourcePriority(ctx *ConflictContext) (*ConflictResult, error) {
	if len(ctx.ConflictingValues) == 0 {
		return nil, fmt.Errorf("no conflicting values provided")
	}

	// Define source priorities (higher number = higher priority)
	sourcePriorities := map[string]int{
		"manual_entry":    100,
		"user_correction": 90,
		"ocr_high_conf":   80,
		"nlp_extraction":  70,
		"pattern_match":   60,
		"ocr_standard":    50,
		"template_guess":  30,
		"heuristic":       20,
		"fallback":        10,
	}

	// Score each value based on source priority + confidence
	type scoredValue struct {
		value ConflictingValue
		score float64
	}

	var scored []scoredValue
	for _, cv := range ctx.ConflictingValues {
		priority := sourcePriorities[cv.Method]
		if priority == 0 {
			priority = 40 // Default priority for unknown methods
		}

		// Combine priority (0-100) with confidence (0-1) weighted 60/40
		compositeScore := (float64(priority) * 0.6) + (cv.Confidence * 40)
		scored = append(scored, scoredValue{cv, compositeScore})
	}

	// Sort by composite score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	best := scored[0]
	requiresReview := best.value.Confidence < cr.config.ReviewThreshold

	notes := fmt.Sprintf("Selected value from %s method (priority score: %.2f) with confidence %.3f",
		best.value.Method, best.score, best.value.Confidence)

	return &ConflictResult{
		ResolvedValue:    best.value.Value,
		ResolutionMethod: "source_priority",
		FinalConfidence:  best.value.Confidence,
		RequiresReview:   requiresReview,
		ResolutionNotes:  notes,
		Metadata: map[string]interface{}{
			"selected_method":    best.value.Method,
			"selected_source":    best.value.Source,
			"composite_score":    best.score,
			"priority_weighting": "60% method priority, 40% confidence",
			"total_candidates":   len(scored),
		},
	}, nil
}

// Helper methods

// valuesEqual compares two values for equality
func (cr *ConflictResolver) valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Convert to strings for comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.EqualFold(aStr, bStr)
}

// addAuditEntry adds an entry to the audit trail
func (cr *ConflictResolver) addAuditEntry(conflictID, action, userID, details string, beforeState, afterState interface{}, confidence float64) {
	if !cr.config.EnableAuditTrail {
		return
	}

	entry := &ConflictAuditEntry{
		ID:              uuid.New().String(),
		ConflictID:      conflictID,
		Action:          action,
		Timestamp:       time.Now(),
		UserID:          userID,
		Details:         details,
		BeforeState:     beforeState,
		AfterState:      afterState,
		ConfidenceScore: confidence,
	}

	cr.auditTrail = append(cr.auditTrail, entry)

	// Trim audit trail if it gets too long
	if len(cr.auditTrail) > cr.config.MaxHistoryEntries*2 {
		cr.auditTrail = cr.auditTrail[len(cr.auditTrail)-cr.config.MaxHistoryEntries:]
	}
}

// Query and History Methods

// GetConflictHistory returns conflict resolution history for a specific field or deal
func (cr *ConflictResolver) GetConflictHistory(dealName, templatePath, fieldName string, limit int) []*ConflictResolutionRecord {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	if fieldName != "" {
		// Specific field query
		historyKey := fmt.Sprintf("%s:%s:%s", dealName, templatePath, fieldName)
		if records, exists := cr.conflictHistory[historyKey]; exists {
			if limit > 0 && len(records) > limit {
				return records[len(records)-limit:]
			}
			return records
		}
		return []*ConflictResolutionRecord{}
	}

	// Deal or template level query
	var results []*ConflictResolutionRecord
	for key, records := range cr.conflictHistory {
		parts := strings.Split(key, ":")
		if len(parts) >= 2 {
			if dealName != "" && parts[0] != dealName {
				continue
			}
			if templatePath != "" && parts[1] != templatePath {
				continue
			}
			results = append(results, records...)
		}
	}

	// Sort by resolution time descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].ResolvedAt.After(results[j].ResolvedAt)
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// GetAuditTrail returns audit trail entries for debugging and compliance
func (cr *ConflictResolver) GetAuditTrail(conflictID string, limit int) []*ConflictAuditEntry {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	var results []*ConflictAuditEntry
	for _, entry := range cr.auditTrail {
		if conflictID == "" || entry.ConflictID == conflictID {
			results = append(results, entry)
		}
	}

	// Sort by timestamp descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// GetConflictStatistics returns statistics about conflict resolution
func (cr *ConflictResolver) GetConflictStatistics() map[string]interface{} {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	stats := make(map[string]interface{})

	totalConflicts := 0
	resolutionMethods := make(map[string]int)
	avgConfidence := 0.0
	reviewRequired := 0

	for _, records := range cr.conflictHistory {
		totalConflicts += len(records)
		for _, record := range records {
			resolutionMethods[record.ResolutionStrategy]++
			avgConfidence += record.FinalConfidence
			if record.RequiresReview {
				reviewRequired++
			}
		}
	}

	if totalConflicts > 0 {
		avgConfidence /= float64(totalConflicts)
	}

	stats["total_conflicts"] = totalConflicts
	stats["resolution_methods"] = resolutionMethods
	stats["average_confidence"] = avgConfidence
	stats["review_required"] = reviewRequired
	stats["review_percentage"] = float64(reviewRequired) / float64(totalConflicts) * 100
	stats["audit_entries"] = len(cr.auditTrail)

	return stats
}

// Persistence methods

// loadPersistedData loads conflict history and audit trail from disk
func (cr *ConflictResolver) loadPersistedData() error {
	if cr.persistencePath == "" {
		return nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(cr.persistencePath, 0755); err != nil {
		return fmt.Errorf("failed to create persistence directory: %w", err)
	}

	// Load conflict history
	historyPath := filepath.Join(cr.persistencePath, "conflict_history.json")
	if data, err := ioutil.ReadFile(historyPath); err == nil {
		if err := json.Unmarshal(data, &cr.conflictHistory); err != nil {
			if cr.logger != nil {
				cr.logger.Warn("Failed to load conflict history: %v", err)
			}
		}
	}

	// Load audit trail
	auditPath := filepath.Join(cr.persistencePath, "audit_trail.json")
	if data, err := ioutil.ReadFile(auditPath); err == nil {
		if err := json.Unmarshal(data, &cr.auditTrail); err != nil {
			if cr.logger != nil {
				cr.logger.Warn("Failed to load audit trail: %v", err)
			}
		}
	}

	return nil
}

// SaveState persists conflict resolution data to disk
func (cr *ConflictResolver) SaveState() error {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	if cr.persistencePath == "" {
		return nil
	}

	// Save conflict history
	if data, err := json.Marshal(cr.conflictHistory); err == nil {
		historyPath := filepath.Join(cr.persistencePath, "conflict_history.json")
		if err := ioutil.WriteFile(historyPath, data, 0644); err != nil {
			return fmt.Errorf("failed to save conflict history: %w", err)
		}
	}

	// Save audit trail
	if data, err := json.Marshal(cr.auditTrail); err == nil {
		auditPath := filepath.Join(cr.persistencePath, "audit_trail.json")
		if err := ioutil.WriteFile(auditPath, data, 0644); err != nil {
			return fmt.Errorf("failed to save audit trail: %w", err)
		}
	}

	return nil
}

// UpdateConfiguration updates the conflict resolver configuration
func (cr *ConflictResolver) UpdateConfiguration(config *ConflictResolutionConfig) error {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	cr.config = config
	return nil
}

// GetConfiguration returns the current configuration
func (cr *ConflictResolver) GetConfiguration() *ConflictResolutionConfig {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	// Return a copy to prevent external modification
	configCopy := *cr.config
	return &configCopy
}
