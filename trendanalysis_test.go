package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTrendAnalyzer(t *testing.T) {
	aiService := &AIService{}
	dataMapper := &DataMapper{}

	analyzer := NewTrendAnalyzer(aiService, dataMapper)

	assert.NotNil(t, analyzer)
	assert.Equal(t, aiService, analyzer.aiService)
	assert.Equal(t, dataMapper, analyzer.dataMapper)
	assert.Equal(t, 365*24*time.Hour, analyzer.timeWindow)
}

func TestDetermineTimeRange(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	now := time.Now()
	data := map[string][]DataPoint{
		"revenue": {
			{Timestamp: now.AddDate(0, -12, 0), Value: 1000000},
			{Timestamp: now.AddDate(0, -6, 0), Value: 1100000},
			{Timestamp: now, Value: 1200000},
		},
		"costs": {
			{Timestamp: now.AddDate(0, -11, 0), Value: 500000},
			{Timestamp: now.AddDate(0, -1, 0), Value: 550000},
		},
	}

	timeRange := analyzer.determineTimeRange(data)

	assert.NotNil(t, timeRange)
	assert.False(t, timeRange.Start.IsZero())
	assert.False(t, timeRange.End.IsZero())
	assert.Greater(t, timeRange.DataPoints, 0)
	assert.NotEmpty(t, timeRange.Granularity)

	// Check that start is the earliest timestamp
	assert.True(t, timeRange.Start.Equal(now.AddDate(0, -12, 0)) || timeRange.Start.Before(now.AddDate(0, -12, 0)))
	// Check that end is the latest timestamp
	assert.True(t, timeRange.End.Equal(now) || timeRange.End.After(now.AddDate(0, -1, 0)))
}

func TestCalculateTrendLine(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	// Create linear growth data
	data := make([]DataPoint, 10)
	for i := 0; i < 10; i++ {
		data[i] = DataPoint{
			Timestamp: time.Now().AddDate(0, -10+i, 0),
			Value:     1000 + float64(i)*100, // Linear growth: y = 100x + 1000
		}
	}

	trendLine := analyzer.calculateTrendLine(data)

	assert.Equal(t, "linear", trendLine.Type)
	assert.InDelta(t, 100, trendLine.Slope, 0.1)
	assert.InDelta(t, 1000, trendLine.Intercept, 0.1)
	assert.Contains(t, trendLine.Formula, "y =")
}

func TestCalculateRSquared(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	// Perfect linear data (R² should be 1)
	perfectData := make([]DataPoint, 10)
	for i := 0; i < 10; i++ {
		perfectData[i] = DataPoint{
			Timestamp: time.Now().AddDate(0, -10+i, 0),
			Value:     float64(i * 100),
		}
	}

	trendLine := analyzer.calculateTrendLine(perfectData)
	rSquared := analyzer.calculateRSquared(perfectData, trendLine)

	assert.InDelta(t, 1.0, rSquared, 0.01) // Should be very close to 1

	// Random data (R² should be lower)
	randomData := []DataPoint{
		{Value: 1000}, {Value: 800}, {Value: 1200}, {Value: 900}, {Value: 1100},
	}

	trendLine2 := analyzer.calculateTrendLine(randomData)
	rSquared2 := analyzer.calculateRSquared(randomData, trendLine2)

	assert.Less(t, rSquared2, 0.5) // Should be lower for random data
}

func TestCalculateCAGR(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	// 100% growth over 2 years = ~41.4% CAGR
	data := []DataPoint{
		{Timestamp: time.Now().AddDate(-2, 0, 0), Value: 1000000},
		{Timestamp: time.Now(), Value: 2000000},
	}

	cagr := analyzer.calculateCAGR(data)

	assert.InDelta(t, 0.414, cagr, 0.01) // √2 - 1 ≈ 0.414

	// Test with no growth
	dataNoGrowth := []DataPoint{
		{Timestamp: time.Now().AddDate(-1, 0, 0), Value: 1000000},
		{Timestamp: time.Now(), Value: 1000000},
	}

	cagrNoGrowth := analyzer.calculateCAGR(dataNoGrowth)
	assert.InDelta(t, 0.0, cagrNoGrowth, 0.001)
}

func TestCalculateVolatility(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	// Stable data (low volatility)
	stableData := []DataPoint{
		{Value: 1000}, {Value: 1010}, {Value: 1020}, {Value: 1030}, {Value: 1040},
	}

	volatilityStable := analyzer.calculateVolatility(stableData)
	assert.Less(t, volatilityStable, 0.02) // Low volatility

	// Volatile data (high volatility)
	volatileData := []DataPoint{
		{Value: 1000}, {Value: 1200}, {Value: 800}, {Value: 1300}, {Value: 700},
	}

	volatilityHigh := analyzer.calculateVolatility(volatileData)
	assert.Greater(t, volatilityHigh, 0.2) // High volatility
}

func TestAnalyzeSeasonality(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	// Create seasonal data (higher in summer months)
	data := make([]DataPoint, 24) // 2 years of monthly data
	baseValue := 1000000.0

	for i := 0; i < 24; i++ {
		month := i % 12
		seasonalFactor := 1.0

		// Summer months (June, July, August) have 20% higher values
		if month >= 5 && month <= 7 {
			seasonalFactor = 1.2
		}
		// Winter months (Dec, Jan, Feb) have 10% lower values
		if month == 11 || month <= 1 {
			seasonalFactor = 0.9
		}

		data[i] = DataPoint{
			Timestamp: time.Now().AddDate(0, -24+i, 0),
			Value:     baseValue * seasonalFactor,
		}
	}

	seasonality := analyzer.analyzeSeasonality(data)

	assert.NotNil(t, seasonality)
	assert.True(t, seasonality.HasSeasonality)
	assert.NotEmpty(t, seasonality.SeasonalFactors)
	assert.NotEmpty(t, seasonality.PeakPeriods)
	assert.NotEmpty(t, seasonality.LowPeriods)
	assert.Equal(t, "annual", seasonality.CycleDuration)

	// Check that summer months are identified as peaks
	foundSummerPeak := false
	for _, peak := range seasonality.PeakPeriods {
		if peak == "June" || peak == "July" || peak == "August" {
			foundSummerPeak = true
			break
		}
	}
	assert.True(t, foundSummerPeak)
}

func TestAnalyzeMetricTrend(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	// Increasing trend
	increasingData := make([]DataPoint, 12)
	for i := 0; i < 12; i++ {
		increasingData[i] = DataPoint{
			Timestamp: time.Now().AddDate(0, -12+i, 0),
			Value:     1000000 * (1 + float64(i)*0.05),
		}
	}

	trend := analyzer.analyzeMetricTrend("Revenue", increasingData)

	assert.NotNil(t, trend)
	assert.Equal(t, "Revenue", trend.MetricName)
	assert.Equal(t, "increasing", trend.Direction)
	assert.Greater(t, trend.Strength, 0.0)
	assert.Greater(t, trend.Correlation, 0.8) // Should have high correlation
	assert.NotEmpty(t, trend.Forecast)
	assert.Len(t, trend.Forecast, 3) // 3 forecast periods
}

func TestGenerateForecast(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	// Create historical data
	historicalData := make([]DataPoint, 12)
	for i := 0; i < 12; i++ {
		historicalData[i] = DataPoint{
			Timestamp: time.Now().AddDate(0, -12+i, 0),
			Value:     1000000 + float64(i)*50000,
		}
	}

	trendLine := analyzer.calculateTrendLine(historicalData)
	forecast := analyzer.generateForecast(trendLine, historicalData, 3)

	assert.Len(t, forecast, 3)

	// Check forecast structure
	for i, fc := range forecast {
		assert.False(t, fc.Timestamp.IsZero())
		assert.Greater(t, fc.Value, 0.0)
		assert.Less(t, fc.ConfidenceLow, fc.Value)
		assert.Greater(t, fc.ConfidenceHigh, fc.Value)
		assert.Equal(t, 0.95, fc.Probability)

		// Check that timestamps are in the future
		assert.True(t, fc.Timestamp.After(historicalData[len(historicalData)-1].Timestamp))

		// Check that values follow the trend
		if i > 0 {
			assert.Greater(t, fc.Value, forecast[i-1].Value)
		}
	}
}

func TestIdentifyKeyInsights(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	analysis := &TrendAnalysisResult{
		FinancialTrends: &FinancialTrends{
			GrowthRates: map[string]float64{
				"revenue": 0.25, // 25% growth
			},
			RevenueTrend: &MetricTrend{
				Direction: "increasing",
			},
			MarginTrends: map[string]*MetricTrend{
				"ebitda_margin": {
					Direction: "increasing",
				},
			},
		},
		RiskTrends: &RiskTrendAnalysis{
			EmergingRisks: []EmergingRisk{
				{
					Name:            "Cyber risk",
					Likelihood:      0.8,
					PotentialImpact: "high",
					GrowthRate:      0.3,
				},
			},
		},
	}

	insights := analyzer.identifyKeyInsights(analysis)

	assert.NotEmpty(t, insights)

	// Check for revenue growth insight
	foundRevenueInsight := false
	foundRiskInsight := false

	for _, insight := range insights {
		if insight.Type == "opportunity" && contains(insight.Description, "revenue growth") {
			foundRevenueInsight = true
			assert.Equal(t, "high", insight.Impact)
			assert.NotEmpty(t, insight.Evidence)
			assert.NotEmpty(t, insight.ActionItems)
		}
		if insight.Type == "risk" {
			foundRiskInsight = true
		}
	}

	assert.True(t, foundRevenueInsight)
	assert.True(t, foundRiskInsight)
}

func TestDetectAnomalies(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	// Create data with an anomaly
	data := map[string][]DataPoint{
		"revenue": make([]DataPoint, 10),
	}

	// Normal growth pattern
	for i := 0; i < 10; i++ {
		value := 1000000.0 + float64(i)*10000

		// Insert anomaly at position 7
		if i == 7 {
			value = 2000000 // Spike
		}

		data["revenue"][i] = DataPoint{
			Timestamp: time.Now().AddDate(0, -10+i, 0),
			Value:     value,
		}
	}

	anomalies := analyzer.detectAnomalies(data)

	assert.NotEmpty(t, anomalies)

	// Check anomaly detection
	foundAnomaly := false
	for _, anomaly := range anomalies {
		if anomaly.Metric == "revenue" {
			foundAnomaly = true
			assert.Greater(t, anomaly.Deviation, 2.0) // Should be 2+ standard deviations
			assert.NotEmpty(t, anomaly.Severity)
			assert.NotEmpty(t, anomaly.PossibleCauses)
		}
	}

	assert.True(t, foundAnomaly)
}

func TestCreateProjectionScenarios(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	trend := &MetricTrend{
		DataPoints: []DataPoint{
			{Value: 1000000},
			{Value: 1100000},
			{Value: 1200000},
		},
	}

	growthRate := 0.20 // 20% growth

	scenarios := analyzer.createProjectionScenarios(trend, growthRate)

	assert.NotNil(t, scenarios)

	// Check scenario relationships
	assert.Greater(t, scenarios.BestCase.Value, scenarios.BaseCase.Value)
	assert.Greater(t, scenarios.BaseCase.Value, scenarios.WorstCase.Value)
	assert.Greater(t, scenarios.BestCase.Growth, scenarios.BaseCase.Growth)
	assert.Greater(t, scenarios.BaseCase.Growth, scenarios.WorstCase.Growth)

	// Check probabilities sum to reasonable amount
	totalProb := scenarios.BestCase.Probability + scenarios.BaseCase.Probability + scenarios.WorstCase.Probability
	assert.InDelta(t, 1.0, totalProb, 0.1)

	// Check drivers are provided
	assert.NotEmpty(t, scenarios.BestCase.Drivers)
	assert.NotEmpty(t, scenarios.BaseCase.Drivers)
	assert.NotEmpty(t, scenarios.WorstCase.Drivers)
}

func TestQuickTrendAssessment(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	values := []float64{100, 110, 120, 130, 140, 150}

	assessment := analyzer.QuickTrendAssessment("Test Metric", values)

	assert.NotNil(t, assessment)
	assert.Contains(t, assessment, "direction")
	assert.Contains(t, assessment, "strength")
	assert.Contains(t, assessment, "correlation")
	assert.Contains(t, assessment, "cagr")
	assert.Contains(t, assessment, "volatility")

	// Check values
	assert.Equal(t, "increasing", assessment["direction"])
	assert.Greater(t, assessment["cagr"].(float64), 0.0)
	assert.NotNil(t, assessment["nextPeriodEstimate"])

	// Test with insufficient data
	shortValues := []float64{100}
	shortAssessment := analyzer.QuickTrendAssessment("Short", shortValues)
	assert.Contains(t, shortAssessment, "error")
}

func TestAnalyzeTrendsIntegration(t *testing.T) {
	aiService := &AIService{}
	dataMapper := &DataMapper{}
	analyzer := NewTrendAnalyzer(aiService, dataMapper)

	documents := []DocumentInfo{
		{
			Name: "financials_2023.xlsx",
			Type: DocTypeFinancial,
			Path: "/test/financials_2023.xlsx",
		},
		{
			Name: "metrics_report.pdf",
			Type: DocTypeGeneral,
			Path: "/test/metrics_report.pdf",
		},
	}

	historicalData := map[string]interface{}{
		"previousYearRevenue": 8000000.0,
		"marketGrowth":        0.15,
	}

	ctx := context.Background()
	result, err := analyzer.AnalyzeTrends(ctx, "Test Deal", documents, historicalData)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Deal", result.DealName)
	assert.NotNil(t, result.TimeRange)
	assert.NotNil(t, result.FinancialTrends)
	assert.NotNil(t, result.OperationalTrends)
	assert.NotNil(t, result.MarketTrends)
	assert.NotNil(t, result.RiskTrends)
	assert.NotNil(t, result.Projections)
	assert.NotEmpty(t, result.Summary)
}

func TestGenerateTrendSummary(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	analysis := &TrendAnalysisResult{
		DealName: "Summary Test Deal",
		TimeRange: TimeRange{
			Start:      time.Now().AddDate(-1, 0, 0),
			End:        time.Now(),
			DataPoints: 12,
		},
		FinancialTrends: &FinancialTrends{
			GrowthRates: map[string]float64{
				"revenue": 0.25,
			},
			RevenueTrend: &MetricTrend{
				Direction:   "increasing",
				Correlation: 0.95,
			},
			Seasonality: &SeasonalityAnalysis{
				HasSeasonality: true,
				PeakPeriods:    []string{"Q4"},
			},
		},
		KeyInsights: []TrendInsight{
			{
				Type:        "opportunity",
				Description: "Strong growth momentum",
				Impact:      "high",
			},
		},
		Projections: &TrendProjections{
			TimeHorizon: "1 year",
			RevenueProjection: &ProjectionScenarios{
				BaseCase: ProjectionResult{
					Value:  15000000,
					Growth: 0.25,
				},
				BestCase: ProjectionResult{
					Value:  18000000,
					Growth: 0.35,
				},
			},
		},
		Anomalies: []TrendAnomaly{
			{Severity: "high"},
			{Severity: "critical"},
		},
	}

	summary := analyzer.generateTrendSummary(analysis)

	assert.Contains(t, summary, "Summary Test Deal")
	assert.Contains(t, summary, "12 data points")
	assert.Contains(t, summary, "25.0%") // Revenue CAGR
	assert.Contains(t, summary, "increasing")
	assert.Contains(t, summary, "Seasonal patterns detected")
	assert.Contains(t, summary, "Strong growth momentum")
	assert.Contains(t, summary, "$15.0M") // Base case projection
	assert.Contains(t, summary, "2 anomalies")
	assert.Contains(t, summary, "2 high/critical severity")
}

// Helper functions

func TestCalculateMarginTrend(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	revenue := []DataPoint{
		{Timestamp: time.Now().AddDate(0, -3, 0), Value: 1000000},
		{Timestamp: time.Now().AddDate(0, -2, 0), Value: 1100000},
		{Timestamp: time.Now().AddDate(0, -1, 0), Value: 1200000},
	}

	costs := []DataPoint{
		{Timestamp: time.Now().AddDate(0, -3, 0), Value: 200000},
		{Timestamp: time.Now().AddDate(0, -2, 0), Value: 220000},
		{Timestamp: time.Now().AddDate(0, -1, 0), Value: 240000},
	}

	marginTrend := analyzer.calculateMarginTrend(revenue, costs)

	assert.Len(t, marginTrend, 3)

	// Check margin calculation (costs/revenue)
	for i := range marginTrend {
		expectedMargin := costs[i].Value / revenue[i].Value
		assert.InDelta(t, expectedMargin, marginTrend[i].Value, 0.001)
	}
}

func TestHelperMethods(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	// Test max function
	assert.Equal(t, 5, max(3, 5))
	assert.Equal(t, 10, max(10, 7))

	// Test anomaly severity
	assert.Equal(t, "critical", analyzer.getAnomalySeverity(5.0))
	assert.Equal(t, "high", analyzer.getAnomalySeverity(3.5))
	assert.Equal(t, "medium", analyzer.getAnomalySeverity(2.7))
	assert.Equal(t, "low", analyzer.getAnomalySeverity(2.1))

	// Test anomaly causes
	revenueCauses := analyzer.suggestAnomalyCauses("revenue", true)
	assert.Contains(t, revenueCauses[0], "Large deal")

	costCauses := analyzer.suggestAnomalyCauses("costs", false)
	assert.Contains(t, costCauses[0], "Cost reduction")
}

// Test standard deviation calculations
func TestStandardDeviationCalculations(t *testing.T) {
	analyzer := NewTrendAnalyzer(nil, nil)

	// Test with slice of floats
	values := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	stdDev := analyzer.calculateStdDev(values)
	assert.InDelta(t, 2.0, stdDev, 0.1) // Known standard deviation

	// Test with data points
	dataPoints := []DataPoint{
		{Value: 2}, {Value: 4}, {Value: 4}, {Value: 4},
		{Value: 5}, {Value: 5}, {Value: 7}, {Value: 9},
	}
	stdDevPoints := analyzer.calculateStdDevFromDataPoints(dataPoints)
	assert.InDelta(t, stdDev, stdDevPoints, 0.001)
}
