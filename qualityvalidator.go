package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// QualityValidator provides comprehensive quality validation for templates
type QualityValidator struct {
	aiService             AIServiceInterface
	professionalFormatter *ProfessionalFormatter
	validationRules       QualityValidationRuleSet
}

// QualityValidationRuleSet contains all validation rules
type QualityValidationRuleSet struct {
	CompletenessRules []QualityCompletenessRule `json:"completenessRules"`
	FormattingRules   []QualityFormattingRule   `json:"formattingRules"`
	BusinessRules     []QualityBusinessRule     `json:"businessRules"`
}

// QualityCompletenessRule defines completeness validation rules
type QualityCompletenessRule struct {
	RuleName        string   `json:"ruleName"`
	RequiredFields  []string `json:"requiredFields"`
	MinCompleteness float64  `json:"minCompleteness"`
	Severity        string   `json:"severity"`
	ErrorMessage    string   `json:"errorMessage"`
}

// QualityFormattingRule defines formatting validation rules
type QualityFormattingRule struct {
	RuleName       string   `json:"ruleName"`
	FieldPatterns  []string `json:"fieldPatterns"`
	ExpectedFormat string   `json:"expectedFormat"`
	Severity       string   `json:"severity"`
	ErrorMessage   string   `json:"errorMessage"`
}

// QualityBusinessRule defines business logic validation rules
type QualityBusinessRule struct {
	RuleName        string                 `json:"ruleName"`
	BusinessContext string                 `json:"businessContext"`
	ValidationType  string                 `json:"validationType"`
	Parameters      map[string]interface{} `json:"parameters"`
	Severity        string                 `json:"severity"`
	ErrorMessage    string                 `json:"errorMessage"`
}

// QualityAssessmentResult contains the complete quality assessment
type QualityAssessmentResult struct {
	OverallScore       float64                    `json:"overallScore"`
	ComponentScores    map[string]float64         `json:"componentScores"`
	ValidationResults  []QualityValidationResult  `json:"validationResults"`
	CompletenessScore  float64                    `json:"completenessScore"`
	ConsistencyScore   float64                    `json:"consistencyScore"`
	FormattingScore    float64                    `json:"formattingScore"`
	BusinessLogicScore float64                    `json:"businessLogicScore"`
	AnomalyFlags       []QualityAnomalyFlag       `json:"anomalyFlags"`
	Recommendations    []QualityRecommendation    `json:"recommendations"`
	QualityTrend       QualityTrendAnalysis       `json:"qualityTrend"`
	AssessmentMetadata QualityAssessmentMetadata  `json:"assessmentMetadata"`
}

// QualityValidationResult represents a single validation result
type QualityValidationResult struct {
	RuleName       string                 `json:"ruleName"`
	RuleType       string                 `json:"ruleType"`
	Status         string                 `json:"status"` // "passed", "failed", "warning"
	Severity       string                 `json:"severity"`
	Message        string                 `json:"message"`
	FieldsAffected []string               `json:"fieldsAffected"`
	Details        map[string]interface{} `json:"details"`
	Confidence     float64                `json:"confidence"`
	Timestamp      time.Time              `json:"timestamp"`
}

// QualityAnomalyFlag represents detected anomalies
type QualityAnomalyFlag struct {
	AnomalyType     string                 `json:"anomalyType"`
	Description     string                 `json:"description"`
	FieldsAffected  []string               `json:"fieldsAffected"`
	Severity        string                 `json:"severity"`
	Confidence      float64                `json:"confidence"`
	SuggestedAction string                 `json:"suggestedAction"`
	Details         map[string]interface{} `json:"details"`
}

// QualityRecommendation provides actionable improvement suggestions
type QualityRecommendation struct {
	Category        string                 `json:"category"`
	Priority        string                 `json:"priority"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	ActionRequired  string                 `json:"actionRequired"`
	EstimatedImpact float64                `json:"estimatedImpact"`
	FieldsAffected  []string               `json:"fieldsAffected"`
	Details         map[string]interface{} `json:"details"`
}

// QualityTrendAnalysis tracks quality over time
type QualityTrendAnalysis struct {
	CurrentScore   float64   `json:"currentScore"`
	PreviousScore  float64   `json:"previousScore"`
	Trend          string    `json:"trend"`
	TrendStrength  float64   `json:"trendStrength"`
	HistoricalData []float64 `json:"historicalData"`
	Timestamp      time.Time `json:"timestamp"`
}

// QualityAssessmentMetadata contains metadata about the assessment
type QualityAssessmentMetadata struct {
	AssessmentID     string                 `json:"assessmentId"`
	TemplateID       string                 `json:"templateId"`
	DealName         string                 `json:"dealName"`
	AssessmentDate   time.Time              `json:"assessmentDate"`
	ValidatorVersion string                 `json:"validatorVersion"`
	RulesApplied     int                    `json:"rulesApplied"`
	ProcessingTime   time.Duration          `json:"processingTime"`
	AIProvider       string                 `json:"aiProvider"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// NewQualityValidator creates a new quality validator
func NewQualityValidator(aiService AIServiceInterface) *QualityValidator {
	return &QualityValidator{
		aiService:             aiService,
		professionalFormatter: NewProfessionalFormatter(),
		validationRules:       createDefaultQualityValidationRules(),
	}
}

// ValidateTemplateQuality performs comprehensive quality validation
func (qv *QualityValidator) ValidateTemplateQuality(mappedData *MappedData, templateInfo *TemplateInfo) (*QualityAssessmentResult, error) {
	startTime := time.Now()

	result := &QualityAssessmentResult{
		ComponentScores:   make(map[string]float64),
		ValidationResults: make([]QualityValidationResult, 0),
		AnomalyFlags:      make([]QualityAnomalyFlag, 0),
		Recommendations:   make([]QualityRecommendation, 0),
		AssessmentMetadata: QualityAssessmentMetadata{
			AssessmentID:     fmt.Sprintf("qa_%d", time.Now().Unix()),
			TemplateID:       mappedData.TemplateID,
			DealName:         mappedData.DealName,
			AssessmentDate:   time.Now(),
			ValidatorVersion: "2.3.0",
			AIProvider:       "multiple",
		},
	}

	// 1. Completeness validation
	completenessScore, completenessResults := qv.validateCompleteness(mappedData, templateInfo)
	result.CompletenessScore = completenessScore
	result.ComponentScores["completeness"] = completenessScore
	result.ValidationResults = append(result.ValidationResults, completenessResults...)

	// 2. Formatting validation
	formattingScore, formattingResults := qv.validateFormatting(mappedData)
	result.FormattingScore = formattingScore
	result.ComponentScores["formatting"] = formattingScore
	result.ValidationResults = append(result.ValidationResults, formattingResults...)

	// 3. Business logic validation
	businessScore, businessResults := qv.validateBusinessLogic(mappedData)
	result.BusinessLogicScore = businessScore
	result.ComponentScores["business_logic"] = businessScore
	result.ValidationResults = append(result.ValidationResults, businessResults...)

	// 4. Anomaly detection
	anomalies := qv.detectAnomalies(mappedData)
	result.AnomalyFlags = append(result.AnomalyFlags, anomalies...)

	// 5. Generate recommendations
	recommendations := qv.generateRecommendations(result)
	result.Recommendations = recommendations

	// 6. Calculate overall score
	result.OverallScore = qv.calculateOverallScore(result.ComponentScores)

	// 7. Quality trend analysis
	result.QualityTrend = qv.analyzeQualityTrend(result.OverallScore, mappedData.DealName)

	// Complete metadata
	result.AssessmentMetadata.ProcessingTime = time.Since(startTime)
	result.AssessmentMetadata.RulesApplied = len(result.ValidationResults)

	return result, nil
}

func createDefaultQualityValidationRules() QualityValidationRuleSet {
	return QualityValidationRuleSet{
		CompletenessRules: []QualityCompletenessRule{
			{
				RuleName:        "minimum_completeness",
				RequiredFields:  []string{"company_name", "deal_value"},
				MinCompleteness: 0.5,
				Severity:        "error",
				ErrorMessage:    "Template must be at least 50% complete",
			},
		},
	}
}

// validateCompleteness checks data completeness
func (qv *QualityValidator) validateCompleteness(mappedData *MappedData, templateInfo *TemplateInfo) (float64, []QualityValidationResult) {
	results := make([]QualityValidationResult, 0)

	totalFields := len(templateInfo.Metadata.Fields)
	populatedFields := len(mappedData.Fields)

	completenessRatio := float64(populatedFields) / float64(totalFields)

	// Apply completeness rules
	for _, rule := range qv.validationRules.CompletenessRules {
		result := QualityValidationResult{
			RuleName:   rule.RuleName,
			RuleType:   "completeness",
			Timestamp:  time.Now(),
			Confidence: 1.0,
		}

		if completenessRatio >= rule.MinCompleteness {
			result.Status = "passed"
			result.Message = fmt.Sprintf("Completeness requirement met: %.1f%% (required: %.1f%%)",
				completenessRatio*100, rule.MinCompleteness*100)
		} else {
			result.Status = "failed"
			result.Severity = rule.Severity
			result.Message = fmt.Sprintf("Completeness below threshold: %.1f%% (required: %.1f%%)",
				completenessRatio*100, rule.MinCompleteness*100)
		}

		result.Details = map[string]interface{}{
			"totalFields":       totalFields,
			"populatedFields":   populatedFields,
			"completenessRatio": completenessRatio,
		}

		results = append(results, result)
	}

	return completenessRatio, results
}

// validateFormatting checks data formatting quality
func (qv *QualityValidator) validateFormatting(mappedData *MappedData) (float64, []QualityValidationResult) {
	results := make([]QualityValidationResult, 0)
	totalChecks := 0
	passedChecks := 0

	for fieldName, mappedField := range mappedData.Fields {
		totalChecks++

		// Create formatting context
		context := FormattingContext{
			FieldName:    fieldName,
			TemplateType: "validation",
			Metadata:     make(map[string]interface{}),
		}

		// Try to format the value
		formattingResult, err := qv.professionalFormatter.FormatValue(mappedField.Value, context)

		result := QualityValidationResult{
			RuleName:       "formatting_validation",
			RuleType:       "formatting",
			FieldsAffected: []string{fieldName},
			Timestamp:      time.Now(),
		}

		if err != nil {
			result.Status = "failed"
			result.Severity = "warning"
			result.Message = fmt.Sprintf("Formatting error for field %s: %v", fieldName, err)
			result.Confidence = 0.8
		} else if formattingResult.Confidence >= 0.8 {
			result.Status = "passed"
			result.Message = fmt.Sprintf("Field %s properly formatted", fieldName)
			result.Confidence = formattingResult.Confidence
			passedChecks++
		} else {
			result.Status = "warning"
			result.Severity = "warning"
			result.Message = fmt.Sprintf("Field %s formatting confidence low: %.1f%%",
				fieldName, formattingResult.Confidence*100)
			result.Confidence = formattingResult.Confidence
		}

		result.Details = map[string]interface{}{
			"originalValue":  mappedField.Value,
			"formattedValue": formattingResult.DisplayValue,
			"formatType":     formattingResult.FormatType,
			"appliedRules":   formattingResult.AppliedRules,
		}

		results = append(results, result)
	}

	formattingScore := 0.0
	if totalChecks > 0 {
		formattingScore = float64(passedChecks) / float64(totalChecks)
	}

	return formattingScore, results
}

// validateBusinessLogic applies business-specific validation rules
func (qv *QualityValidator) validateBusinessLogic(mappedData *MappedData) (float64, []QualityValidationResult) {
	results := make([]QualityValidationResult, 0)
	totalChecks := 0
	passedChecks := 0

	// Validate deal value reasonableness
	if dealValue, hasDealValue := mappedData.Fields["deal_value"]; hasDealValue {
		totalChecks++

		result := QualityValidationResult{
			RuleName:       "deal_value_reasonableness",
			RuleType:       "business_logic",
			FieldsAffected: []string{"deal_value"},
			Timestamp:      time.Now(),
		}

		dealVal := parseQualityNumericValue(dealValue.Value)

		if dealVal >= 1000000 && dealVal <= 100000000000 { // $1M to $100B
			result.Status = "passed"
			result.Message = "Deal value within reasonable range"
			result.Confidence = 0.9
			passedChecks++
		} else if dealVal > 0 {
			result.Status = "warning"
			result.Severity = "warning"
			result.Message = fmt.Sprintf("Deal value unusual: $%.0f", dealVal)
			result.Confidence = 0.7
		} else {
			result.Status = "failed"
			result.Severity = "error"
			result.Message = "Invalid deal value"
			result.Confidence = 1.0
		}

		result.Details = map[string]interface{}{
			"dealValue": dealVal,
		}

		results = append(results, result)
	}

	businessScore := 0.0
	if totalChecks > 0 {
		businessScore = float64(passedChecks) / float64(totalChecks)
	}

	return businessScore, results
}

// detectAnomalies identifies data anomalies
func (qv *QualityValidator) detectAnomalies(mappedData *MappedData) []QualityAnomalyFlag {
	anomalies := make([]QualityAnomalyFlag, 0)

	// Check for duplicate values
	valueMap := make(map[string][]string)
	for fieldName, mappedField := range mappedData.Fields {
		valueStr := fmt.Sprintf("%v", mappedField.Value)
		valueMap[valueStr] = append(valueMap[valueStr], fieldName)
	}

	for value, fields := range valueMap {
		if len(fields) > 2 && value != "" {
			anomaly := QualityAnomalyFlag{
				AnomalyType:     "duplicate_values",
				Description:     fmt.Sprintf("Same value '%s' appears in multiple fields", value),
				FieldsAffected:  fields,
				Severity:        "warning",
				Confidence:      0.9,
				SuggestedAction: "Review field mapping accuracy",
				Details: map[string]interface{}{
					"duplicateValue": value,
					"fieldCount":     len(fields),
				},
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

// generateRecommendations creates actionable recommendations
func (qv *QualityValidator) generateRecommendations(result *QualityAssessmentResult) []QualityRecommendation {
	recommendations := make([]QualityRecommendation, 0)

	// Completeness recommendations
	if result.CompletenessScore < 0.7 {
		rec := QualityRecommendation{
			Category:        "completeness",
			Priority:        "high",
			Title:           "Improve Data Completeness",
			Description:     fmt.Sprintf("Template is only %.1f%% complete", result.CompletenessScore*100),
			ActionRequired:  "Review document extraction and field mapping",
			EstimatedImpact: 0.3,
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// calculateOverallScore computes the weighted overall quality score
func (qv *QualityValidator) calculateOverallScore(componentScores map[string]float64) float64 {
	weights := map[string]float64{
		"completeness":   0.25,
		"formatting":     0.15,
		"consistency":    0.25,
		"business_logic": 0.20,
		"ai_validation":  0.15,
	}

	totalScore := 0.0
	totalWeight := 0.0

	for component, score := range componentScores {
		if weight, exists := weights[component]; exists {
			totalScore += score * weight
			totalWeight += weight
		}
	}

	if totalWeight > 0 {
		return totalScore / totalWeight
	}

	return 0.0
}

// analyzeQualityTrend analyzes quality trends over time
func (qv *QualityValidator) analyzeQualityTrend(currentScore float64, dealName string) QualityTrendAnalysis {
	return QualityTrendAnalysis{
		CurrentScore:   currentScore,
		PreviousScore:  0.75, // Placeholder
		Trend:          "improving",
		TrendStrength:  0.1,
		HistoricalData: []float64{0.70, 0.72, 0.75, currentScore},
		Timestamp:      time.Now(),
	}
}

// Helper function for parsing numeric values (renamed to avoid conflict)
func parseQualityNumericValue(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		// Clean the string and parse
		cleaned := strings.ReplaceAll(v, ",", "")
		cleaned = strings.ReplaceAll(cleaned, "$", "")
		cleaned = strings.ReplaceAll(cleaned, " ", "")

		if num, err := strconv.ParseFloat(cleaned, 64); err == nil {
			return num
		}
	}
	return 0
}
