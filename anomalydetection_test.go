package main

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnomalyDetector(t *testing.T) {
	aiService := &AIService{}
	dataMapper := &DataMapper{}

	detector := NewAnomalyDetector(aiService, dataMapper)

	assert.NotNil(t, detector)
	assert.Equal(t, aiService, detector.aiService)
	assert.Equal(t, dataMapper, detector.dataMapper)
	assert.Equal(t, 2.0, detector.sensitivity)
}

func TestSetSensitivity(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	detector.SetSensitivity(3.0)
	assert.Equal(t, 3.0, detector.sensitivity)

	// Should not set negative sensitivity
	detector.SetSensitivity(-1.0)
	assert.Equal(t, 3.0, detector.sensitivity) // Unchanged
}

func TestAssessDataQuality(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	// Good quality data
	goodData := map[string][]DataPoint{
		"revenue": generateConsistentData(12, 1000000, 0.05),
		"costs":   generateConsistentData(12, 500000, 0.05),
		"ebitda":  generateConsistentData(12, 200000, 0.05),
	}

	goodQuality := detector.assessDataQuality(goodData)

	assert.NotNil(t, goodQuality)
	assert.Greater(t, goodQuality.OverallScore, 0.7)
	assert.Greater(t, goodQuality.Completeness, 0.8)
	assert.Greater(t, goodQuality.Consistency, 0.8)
	assert.Empty(t, goodQuality.Issues) // Should have no major issues

	// Poor quality data (with missing values)
	poorData := map[string][]DataPoint{
		"revenue": {
			{Timestamp: time.Now().AddDate(0, -3, 0), Value: 1000000},
			{Timestamp: time.Now().AddDate(0, -2, 0), Value: 0}, // Missing
			{Timestamp: time.Now().AddDate(0, -1, 0), Value: 0}, // Missing
			{Timestamp: time.Now(), Value: 1100000},
		},
	}

	poorQuality := detector.assessDataQuality(poorData)

	assert.Less(t, poorQuality.OverallScore, 0.7)
	assert.NotEmpty(t, poorQuality.Issues)
}

func TestDetectStatisticalAnomalies(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	// Create data with an anomaly
	data := make([]DataPoint, 10)
	for i := 0; i < 10; i++ {
		value := 1000000.0 + float64(i)*10000 // Normal growth
		if i == 7 {
			value = 2000000 // Spike anomaly
		}
		data[i] = DataPoint{
			Timestamp: time.Now().AddDate(0, -10+i, 0),
			Value:     value,
		}
	}

	anomalies := detector.detectStatisticalAnomalies("test_metric", data)

	assert.NotEmpty(t, anomalies)

	// Find the spike anomaly
	foundSpike := false
	for _, anomaly := range anomalies {
		if anomaly.ActualValue == 2000000 {
			foundSpike = true
			assert.Equal(t, "above", anomaly.Direction)
			assert.Greater(t, anomaly.Deviation, 2.0)
			assert.NotEmpty(t, anomaly.Severity)
		}
	}
	assert.True(t, foundSpike)
}

func TestDetectFinancialAnomalies(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	// Create financial data with anomalies
	data := map[string][]DataPoint{
		"revenue": {
			{Timestamp: time.Now().AddDate(0, -6, 0), Value: 1000000},
			{Timestamp: time.Now().AddDate(0, -5, 0), Value: 1050000},
			{Timestamp: time.Now().AddDate(0, -4, 0), Value: 1100000},
			{Timestamp: time.Now().AddDate(0, -3, 0), Value: 2000000}, // Anomaly
			{Timestamp: time.Now().AddDate(0, -2, 0), Value: 1150000},
			{Timestamp: time.Now().AddDate(0, -1, 0), Value: 1200000},
		},
		"costs": {
			{Timestamp: time.Now().AddDate(0, -6, 0), Value: 500000},
			{Timestamp: time.Now().AddDate(0, -5, 0), Value: 520000},
			{Timestamp: time.Now().AddDate(0, -4, 0), Value: 540000},
			{Timestamp: time.Now().AddDate(0, -3, 0), Value: 560000},
			{Timestamp: time.Now().AddDate(0, -2, 0), Value: 580000},
			{Timestamp: time.Now().AddDate(0, -1, 0), Value: 800000}, // Anomaly
		},
	}

	anomalies := detector.detectFinancialAnomalies(data)

	assert.NotEmpty(t, anomalies)

	// Check for revenue spike
	foundRevenueSpike := false
	foundCostAnomaly := false

	for _, anomaly := range anomalies {
		if anomaly.Type == "revenue_spike" {
			foundRevenueSpike = true
			assert.Equal(t, "revenue", anomaly.Metric)
			assert.Greater(t, anomaly.ActualValue, anomaly.ExpectedValue)
			assert.NotEmpty(t, anomaly.PossibleCauses)
		}
		if anomaly.Type == "cost_overrun" {
			foundCostAnomaly = true
		}
	}

	assert.True(t, foundRevenueSpike)
	assert.True(t, foundCostAnomaly)
}

func TestDetectMarginAnomalies(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	// Create revenue and profit data
	revenue := []DataPoint{
		{Timestamp: time.Now().AddDate(0, -3, 0), Value: 1000000},
		{Timestamp: time.Now().AddDate(0, -2, 0), Value: 1100000},
		{Timestamp: time.Now().AddDate(0, -1, 0), Value: 1200000},
		{Timestamp: time.Now(), Value: 1300000},
	}

	profit := []DataPoint{
		{Timestamp: time.Now().AddDate(0, -3, 0), Value: 200000}, // 20% margin
		{Timestamp: time.Now().AddDate(0, -2, 0), Value: 220000}, // 20% margin
		{Timestamp: time.Now().AddDate(0, -1, 0), Value: 240000}, // 20% margin
		{Timestamp: time.Now(), Value: 100000},                   // 7.7% margin - anomaly
	}

	anomalies := detector.detectMarginAnomalies(revenue, profit, "Profit Margin")

	assert.NotEmpty(t, anomalies)

	// Check for margin deviation
	for _, anomaly := range anomalies {
		if anomaly.Type == "margin_deviation" {
			assert.Equal(t, "below", anomaly.Direction)
			assert.NotEmpty(t, anomaly.PossibleCauses)
		}
	}
}

func TestDetectOperationalAnomalies(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	data := map[string][]DataPoint{
		"customers": {
			{Timestamp: time.Now().AddDate(0, -3, 0), Value: 1000},
			{Timestamp: time.Now().AddDate(0, -2, 0), Value: 1050},
			{Timestamp: time.Now().AddDate(0, -1, 0), Value: 800}, // Drop
			{Timestamp: time.Now(), Value: 1100},
		},
		"productivity": {
			{Timestamp: time.Now().AddDate(0, -3, 0), Value: 100},
			{Timestamp: time.Now().AddDate(0, -2, 0), Value: 105},
			{Timestamp: time.Now().AddDate(0, -1, 0), Value: 70}, // Significant drop
			{Timestamp: time.Now(), Value: 110},
		},
	}

	anomalies := detector.detectOperationalAnomalies(data)

	assert.NotEmpty(t, anomalies)

	// Check for customer and productivity anomalies
	foundCustomerAnomaly := false
	foundProductivityDrop := false

	for _, anomaly := range anomalies {
		if anomaly.Type == "customer_anomaly" {
			foundCustomerAnomaly = true
			assert.NotEmpty(t, anomaly.Description)
			assert.NotEmpty(t, anomaly.Remediation)
		}
		if anomaly.Type == "efficiency_drop" {
			foundProductivityDrop = true
			assert.Equal(t, "high", anomaly.UrgencyLevel)
		}
	}

	assert.True(t, foundCustomerAnomaly || foundProductivityDrop)
}

func TestDetectPatternAnomalies(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	// Create data with pattern changes
	data := map[string][]DataPoint{
		"revenue": generateDataWithTrendReversal(12),
		"costs":   generateDataWithTrendReversal(12),
	}

	anomalies := detector.detectPatternAnomalies(data)

	// Should detect trend reversals
	foundTrendReversal := false
	for _, anomaly := range anomalies {
		if anomaly.Type == "trend_reversal" {
			foundTrendReversal = true
			assert.NotEmpty(t, anomaly.Description)
			assert.Greater(t, anomaly.PatternStrength, 0.0)
		}
	}

	assert.True(t, foundTrendReversal)
}

func TestCorrelateAnomalies(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	result := &AnomalyDetectionResult{
		FinancialAnomalies: []FinancialAnomaly{
			{
				ID:               "fin_1",
				Metric:           "revenue",
				Timestamp:        time.Now(),
				RelatedAnomalies: []string{},
			},
			{
				ID:               "fin_2",
				Metric:           "costs",
				Timestamp:        time.Now().AddDate(0, 0, -5), // 5 days apart
				RelatedAnomalies: []string{},
			},
			{
				ID:               "fin_3",
				Metric:           "ebitda",
				Timestamp:        time.Now().AddDate(0, 0, -2), // 2 days apart from revenue
				RelatedAnomalies: []string{},
			},
		},
	}

	detector.correlateAnomalies(result)

	// Check that related anomalies were identified
	// Revenue and EBITDA should be related (close in time and related metrics)
	found := false
	for _, anomaly := range result.FinancialAnomalies {
		if anomaly.ID == "fin_1" {
			for _, relatedID := range anomaly.RelatedAnomalies {
				if relatedID == "fin_3" {
					found = true
					break
				}
			}
		}
	}
	assert.True(t, found)
}

func TestCalculateRiskIndicators(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	result := &AnomalyDetectionResult{
		DataQuality: &DataQualityAssessment{
			OverallScore: 0.5, // Poor quality
		},
		FinancialAnomalies: []FinancialAnomaly{
			{Severity: "critical"},
			{Severity: "high"},
			{Severity: "high"},
		},
		OperationalAnomalies: []OperationalAnomaly{
			{ImpactScore: 0.8, UrgencyLevel: "immediate"},
		},
		PatternAnomalies: []PatternAnomaly{
			{PatternStrength: 0.7, StatisticalScore: 0.8},
		},
	}

	indicators := detector.calculateRiskIndicators(result)

	assert.NotEmpty(t, indicators)

	// Check for different risk categories
	foundFinancial := false
	foundOperational := false
	foundDataQuality := false

	for _, indicator := range indicators {
		switch indicator.Category {
		case "financial":
			foundFinancial = true
			assert.Contains(t, []string{"critical", "high"}, indicator.CurrentLevel)
			assert.NotEmpty(t, indicator.ContributingFactors)
		case "operational":
			foundOperational = true
		case "compliance":
			foundDataQuality = true
			assert.Greater(t, indicator.Score, 0.3) // High risk due to poor quality
		}
	}

	assert.True(t, foundFinancial)
	assert.True(t, foundOperational)
	assert.True(t, foundDataQuality)
}

func TestGenerateAnomalySummary(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	result := &AnomalyDetectionResult{
		FinancialAnomalies: []FinancialAnomaly{
			{Type: "revenue_spike", Severity: "critical"},
			{Type: "cost_overrun", Severity: "high"},
			{Type: "margin_deviation", Severity: "medium"},
		},
		OperationalAnomalies: []OperationalAnomaly{
			{Type: "efficiency_drop", UrgencyLevel: "high"},
		},
		PatternAnomalies: []PatternAnomaly{
			{Type: "trend_reversal"},
		},
	}

	summary := detector.generateAnomalySummary(result)

	assert.NotNil(t, summary)
	assert.Equal(t, 5, summary.TotalAnomalies)
	assert.Equal(t, 1, summary.CriticalCount)
	assert.Equal(t, 2, summary.HighCount)
	assert.Equal(t, 1, summary.MediumCount)
	assert.Equal(t, "critical", summary.RiskLevel)
	assert.NotEmpty(t, summary.KeyFindings)
	assert.NotEmpty(t, summary.ImpactAssessment)
	assert.NotEmpty(t, summary.DistributionByType)
}

func TestGenerateRecommendations(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	result := &AnomalyDetectionResult{
		DataQuality: &DataQualityAssessment{
			OverallScore: 0.6, // Poor quality
		},
		FinancialAnomalies: []FinancialAnomaly{
			{Severity: "critical"},
			{Severity: "critical"},
		},
		PatternAnomalies: []PatternAnomaly{
			{Type: "trend_reversal"},
			{Type: "correlation_break"},
			{Type: "seasonality_shift"},
		},
		RiskIndicators: []AnomalyRiskIndicator{
			{Name: "Financial Risk", CurrentLevel: "critical", Category: "financial"},
		},
	}

	recommendations := detector.generateRecommendations(result)

	assert.NotEmpty(t, recommendations)

	// Check for critical financial recommendation
	foundCritical := false
	foundDataQuality := false
	foundMonitoring := false

	for _, rec := range recommendations {
		if rec.Priority == "immediate" && rec.Type == "investigation" {
			foundCritical = true
			assert.NotEmpty(t, rec.Actions)
			assert.NotEmpty(t, rec.Timeline)
		}
		if rec.Type == "process_change" {
			foundDataQuality = true
		}
		if rec.Type == "monitoring" {
			foundMonitoring = true
		}
	}

	assert.True(t, foundCritical)
	assert.True(t, foundDataQuality)
	assert.True(t, foundMonitoring)

	// Check that recommendations are sorted by priority
	for i := 1; i < len(recommendations); i++ {
		priorityOrder := map[string]int{"immediate": 4, "high": 3, "medium": 2, "low": 1}
		assert.GreaterOrEqual(t,
			priorityOrder[recommendations[i-1].Priority],
			priorityOrder[recommendations[i].Priority])
	}
}

func TestQuickAnomalyCheck(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	// Normal case
	historicalValues := []float64{100, 105, 110, 115, 120}
	currentValue := 125.0

	result := detector.QuickAnomalyCheck("test_metric", currentValue, historicalValues)

	assert.NotNil(t, result)
	assert.Contains(t, result, "mean")
	assert.Contains(t, result, "stdDev")
	assert.Contains(t, result, "isAnomaly")
	assert.False(t, result["isAnomaly"].(bool)) // Should not be anomalous

	// Anomalous case
	anomalousValue := 200.0
	anomalyResult := detector.QuickAnomalyCheck("test_metric", anomalousValue, historicalValues)

	assert.True(t, anomalyResult["isAnomaly"].(bool))
	assert.Contains(t, anomalyResult, "severity")
	assert.Contains(t, anomalyResult, "direction")
	assert.Contains(t, anomalyResult, "deviationPercent")
	assert.Equal(t, "above", anomalyResult["direction"])

	// Insufficient data case
	shortHistory := []float64{100, 105}
	shortResult := detector.QuickAnomalyCheck("test_metric", 110, shortHistory)
	assert.Contains(t, shortResult, "error")
}

func TestDetectAnomaliesIntegration(t *testing.T) {
	aiService := &AIService{}
	dataMapper := &DataMapper{}
	detector := NewAnomalyDetector(aiService, dataMapper)

	// Create comprehensive time series data
	timeSeriesData := map[string][]DataPoint{
		"revenue":   generateDataWithAnomaly(12, 1000000, 0.05, 8, 2.0),
		"costs":     generateDataWithAnomaly(12, 500000, 0.05, 10, 1.5),
		"ebitda":    generateConsistentData(12, 200000, 0.1),
		"customers": generateDataWithAnomaly(12, 1000, 0.02, 6, -0.3),
	}

	documents := []DocumentInfo{
		{Name: "financials.xlsx", Type: DocTypeFinancial},
		{Name: "operations.pdf", Type: DocTypeGeneral},
	}

	ctx := context.Background()
	result, err := detector.DetectAnomalies(ctx, "Test Deal", documents, timeSeriesData)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Deal", result.DealName)
	assert.NotNil(t, result.DataQuality)
	assert.NotEmpty(t, result.FinancialAnomalies)
	assert.NotNil(t, result.Summary)
	assert.NotEmpty(t, result.Recommendations)
}

// Helper functions for test data generation

func generateConsistentData(points int, baseValue float64, growthRate float64) []DataPoint {
	data := make([]DataPoint, points)
	for i := 0; i < points; i++ {
		data[i] = DataPoint{
			Timestamp: time.Now().AddDate(0, -points+i, 0),
			Value:     baseValue * math.Pow(1+growthRate, float64(i)),
		}
	}
	return data
}

func generateDataWithAnomaly(points int, baseValue float64, growthRate float64, anomalyIndex int, anomalyFactor float64) []DataPoint {
	data := generateConsistentData(points, baseValue, growthRate)
	if anomalyIndex >= 0 && anomalyIndex < points {
		data[anomalyIndex].Value *= (1 + anomalyFactor)
	}
	return data
}

func generateDataWithTrendReversal(points int) []DataPoint {
	data := make([]DataPoint, points)
	midpoint := points / 2

	// First half: increasing trend
	for i := 0; i < midpoint; i++ {
		data[i] = DataPoint{
			Timestamp: time.Now().AddDate(0, -points+i, 0),
			Value:     1000000 + float64(i)*50000,
		}
	}

	// Second half: decreasing trend
	for i := midpoint; i < points; i++ {
		data[i] = DataPoint{
			Timestamp: time.Now().AddDate(0, -points+i, 0),
			Value:     data[midpoint-1].Value - float64(i-midpoint)*40000,
		}
	}

	return data
}

// Test helper methods

func TestCalculateCorrelation(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	// Perfect positive correlation
	series1 := []DataPoint{
		{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5},
	}
	series2 := []DataPoint{
		{Value: 2}, {Value: 4}, {Value: 6}, {Value: 8}, {Value: 10},
	}

	corr := detector.calculateCorrelation(series1, series2)
	assert.InDelta(t, 1.0, corr, 0.01) // Perfect positive correlation

	// Negative correlation
	series3 := []DataPoint{
		{Value: 5}, {Value: 4}, {Value: 3}, {Value: 2}, {Value: 1},
	}

	negCorr := detector.calculateCorrelation(series1, series3)
	assert.InDelta(t, -1.0, negCorr, 0.01) // Perfect negative correlation

	// No correlation
	series4 := []DataPoint{
		{Value: 3}, {Value: 1}, {Value: 4}, {Value: 1}, {Value: 5},
	}

	noCorr := detector.calculateCorrelation(series1, series4)
	assert.Less(t, math.Abs(noCorr), 0.5) // Weak correlation
}

func TestCalculateTrendDirection(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	// Increasing trend
	increasingData := []DataPoint{
		{Value: 100}, {Value: 110}, {Value: 120}, {Value: 130}, {Value: 140},
	}

	increasingTrend := detector.calculateTrendDirection(increasingData)
	assert.Greater(t, increasingTrend, 0.0)

	// Decreasing trend
	decreasingData := []DataPoint{
		{Value: 140}, {Value: 130}, {Value: 120}, {Value: 110}, {Value: 100},
	}

	decreasingTrend := detector.calculateTrendDirection(decreasingData)
	assert.Less(t, decreasingTrend, 0.0)

	// Flat trend
	flatData := []DataPoint{
		{Value: 100}, {Value: 100}, {Value: 100}, {Value: 100}, {Value: 100},
	}

	flatTrend := detector.calculateTrendDirection(flatData)
	assert.InDelta(t, 0.0, flatTrend, 0.01)
}

func TestExtractSeasonalPattern(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	// Create seasonal data
	data := []DataPoint{
		{Value: 100}, {Value: 110}, {Value: 120}, {Value: 130},
		{Value: 140}, {Value: 150}, {Value: 160}, {Value: 150},
		{Value: 140}, {Value: 130}, {Value: 120}, {Value: 110},
	}

	pattern := detector.extractSeasonalPattern(data)

	assert.Len(t, pattern, len(data))

	// Pattern should show relative values
	mean := 0.0
	for _, p := range pattern {
		mean += p
	}
	mean /= float64(len(pattern))

	// Average should be close to 1.0
	assert.InDelta(t, 1.0, mean, 0.1)
}

func TestComparePatterns(t *testing.T) {
	detector := NewAnomalyDetector(nil, nil)

	// Identical patterns
	pattern1 := []float64{1.0, 1.1, 1.2, 1.1, 1.0}
	pattern2 := []float64{1.0, 1.1, 1.2, 1.1, 1.0}

	diff := detector.comparePatterns(pattern1, pattern2)
	assert.Equal(t, 0.0, diff)

	// Different patterns
	pattern3 := []float64{1.0, 0.9, 0.8, 0.9, 1.0}

	diff2 := detector.comparePatterns(pattern1, pattern3)
	assert.Greater(t, diff2, 0.2)

	// Different lengths
	pattern4 := []float64{1.0, 1.1}

	diff3 := detector.comparePatterns(pattern1, pattern4)
	assert.Equal(t, 1.0, diff3) // Maximum difference
}
