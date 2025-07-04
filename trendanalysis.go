package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"DealDone/internal/core/documents"
	"DealDone/internal/domain/analysis"
	"DealDone/internal/infrastructure/ai"
)

// TrendAnalyzer performs trend analysis across multiple documents
type TrendAnalyzer struct {
	aiService  *ai.AIService
	dataMapper *analysis.DataMapper
	timeWindow time.Duration // Default analysis window
}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer(aiService *ai.AIService, dataMapper *analysis.DataMapper) *TrendAnalyzer {
	return &TrendAnalyzer{
		aiService:  aiService,
		dataMapper: dataMapper,
		timeWindow: 365 * 24 * time.Hour, // Default 1 year
	}
}

// TrendAnalysisResult contains comprehensive trend analysis
type TrendAnalysisResult struct {
	DealName          string               `json:"dealName"`
	AnalysisDate      time.Time            `json:"analysisDate"`
	TimeRange         TimeRange            `json:"timeRange"`
	FinancialTrends   *FinancialTrends     `json:"financialTrends"`
	OperationalTrends *OperationalTrends   `json:"operationalTrends"`
	MarketTrends      *MarketTrendAnalysis `json:"marketTrends"`
	RiskTrends        *RiskTrendAnalysis   `json:"riskTrends"`
	Projections       *TrendProjections    `json:"projections"`
	KeyInsights       []TrendInsight       `json:"keyInsights"`
	Anomalies         []TrendAnomaly       `json:"anomalies"`
	Summary           string               `json:"summary"`
}

// TimeRange represents the time period analyzed
type TimeRange struct {
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	DataPoints  int       `json:"dataPoints"`
	Granularity string    `json:"granularity"` // "daily", "monthly", "quarterly", "yearly"
}

// FinancialTrends contains financial metric trends
type FinancialTrends struct {
	RevenueTrend       *MetricTrend            `json:"revenueTrend"`
	ProfitabilityTrend *MetricTrend            `json:"profitabilityTrend"`
	CashFlowTrend      *MetricTrend            `json:"cashFlowTrend"`
	MarginTrends       map[string]*MetricTrend `json:"marginTrends"`
	GrowthRates        map[string]float64      `json:"growthRates"`
	Seasonality        *SeasonalityAnalysis    `json:"seasonality"`
	Volatility         map[string]float64      `json:"volatility"`
}

// MetricTrend represents trend data for a specific metric
type MetricTrend struct {
	MetricName  string          `json:"metricName"`
	DataPoints  []DataPoint     `json:"dataPoints"`
	TrendLine   TrendLine       `json:"trendLine"`
	Direction   string          `json:"direction"`   // "increasing", "decreasing", "stable"
	Strength    float64         `json:"strength"`    // 0-1, trend strength
	Correlation float64         `json:"correlation"` // R-squared value
	Forecast    []ForecastPoint `json:"forecast,omitempty"`
}

// DataPoint represents a single data point in time
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label,omitempty"`
}

// TrendLine represents the mathematical trend line
type TrendLine struct {
	Type      string  `json:"type"` // "linear", "exponential", "polynomial"
	Slope     float64 `json:"slope"`
	Intercept float64 `json:"intercept"`
	Formula   string  `json:"formula"`
}

// ForecastPoint represents a forecasted data point
type ForecastPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	Value          float64   `json:"value"`
	ConfidenceLow  float64   `json:"confidenceLow"`
	ConfidenceHigh float64   `json:"confidenceHigh"`
	Probability    float64   `json:"probability"`
}

// SeasonalityAnalysis contains seasonal pattern analysis
type SeasonalityAnalysis struct {
	HasSeasonality  bool               `json:"hasSeasonality"`
	SeasonalFactors map[string]float64 `json:"seasonalFactors"` // month -> factor
	PeakPeriods     []string           `json:"peakPeriods"`
	LowPeriods      []string           `json:"lowPeriods"`
	CycleDuration   string             `json:"cycleDuration"`
}

// OperationalTrends contains operational metric trends
type OperationalTrends struct {
	EfficiencyTrends   map[string]*MetricTrend `json:"efficiencyTrends"`
	ProductivityTrends map[string]*MetricTrend `json:"productivityTrends"`
	QualityMetrics     map[string]*MetricTrend `json:"qualityMetrics"`
	CustomerMetrics    *CustomerTrends         `json:"customerMetrics"`
	EmployeeMetrics    *EmployeeTrends         `json:"employeeMetrics"`
}

// CustomerTrends contains customer-related trends
type CustomerTrends struct {
	AcquisitionTrend   *MetricTrend   `json:"acquisitionTrend"`
	RetentionTrend     *MetricTrend   `json:"retentionTrend"`
	SatisfactionTrend  *MetricTrend   `json:"satisfactionTrend"`
	LifetimeValueTrend *MetricTrend   `json:"lifetimeValueTrend"`
	ChurnAnalysis      *ChurnAnalysis `json:"churnAnalysis"`
}

// ChurnAnalysis contains customer churn analysis
type ChurnAnalysis struct {
	CurrentRate      float64  `json:"currentRate"`
	TrendDirection   string   `json:"trendDirection"`
	PredictedRate    float64  `json:"predictedRate"`
	RiskFactors      []string `json:"riskFactors"`
	RetentionDrivers []string `json:"retentionDrivers"`
}

// EmployeeTrends contains employee-related trends
type EmployeeTrends struct {
	HeadcountTrend    *MetricTrend `json:"headcountTrend"`
	TurnoverTrend     *MetricTrend `json:"turnoverTrend"`
	ProductivityTrend *MetricTrend `json:"productivityTrend"`
	EngagementTrend   *MetricTrend `json:"engagementTrend"`
}

// MarketTrendAnalysis contains market-related trends
type MarketTrendAnalysis struct {
	MarketShareTrend  *MetricTrend            `json:"marketShareTrend"`
	CompetitiveTrends map[string]*MetricTrend `json:"competitiveTrends"`
	IndustryGrowth    *MetricTrend            `json:"industryGrowth"`
	PricingTrends     *MetricTrend            `json:"pricingTrends"`
	DemandIndicators  []DemandIndicator       `json:"demandIndicators"`
}

// DemandIndicator represents a demand indicator
type DemandIndicator struct {
	Name         string  `json:"name"`
	CurrentValue float64 `json:"currentValue"`
	Trend        string  `json:"trend"`
	Impact       string  `json:"impact"` // "positive", "negative", "neutral"
}

// RiskTrendAnalysis contains risk-related trends
type RiskTrendAnalysis struct {
	OverallRiskTrend   *MetricTrend            `json:"overallRiskTrend"`
	RiskCategories     map[string]*MetricTrend `json:"riskCategories"`
	EmergingRisks      []EmergingRisk          `json:"emergingRisks"`
	MitigationProgress map[string]float64      `json:"mitigationProgress"`
}

// EmergingRisk represents a newly identified risk
type EmergingRisk struct {
	Name            string    `json:"name"`
	FirstIdentified time.Time `json:"firstIdentified"`
	GrowthRate      float64   `json:"growthRate"`
	PotentialImpact string    `json:"potentialImpact"`
	Likelihood      float64   `json:"likelihood"`
}

// TrendProjections contains future projections based on trends
type TrendProjections struct {
	TimeHorizon       string               `json:"timeHorizon"` // "3 months", "6 months", "1 year"
	RevenueProjection *ProjectionScenarios `json:"revenueProjection"`
	GrowthProjection  *ProjectionScenarios `json:"growthProjection"`
	RiskProjection    map[string]float64   `json:"riskProjection"`
	Assumptions       []string             `json:"assumptions"`
}

// ProjectionScenarios contains different projection scenarios
type ProjectionScenarios struct {
	BestCase   ProjectionResult `json:"bestCase"`
	BaseCase   ProjectionResult `json:"baseCase"`
	WorstCase  ProjectionResult `json:"worstCase"`
	MostLikely ProjectionResult `json:"mostLikely"`
}

// ProjectionResult represents a projection outcome
type ProjectionResult struct {
	Value       float64  `json:"value"`
	Growth      float64  `json:"growth"`
	Probability float64  `json:"probability"`
	Drivers     []string `json:"drivers"`
}

// TrendInsight represents a key insight from trend analysis
type TrendInsight struct {
	Type        string   `json:"type"` // "opportunity", "risk", "inflection", "correlation"
	Description string   `json:"description"`
	Impact      string   `json:"impact"` // "high", "medium", "low"
	Confidence  float64  `json:"confidence"`
	Evidence    []string `json:"evidence"`
	ActionItems []string `json:"actionItems"`
}

// TrendAnomaly represents an anomaly in the trends
type TrendAnomaly struct {
	Metric         string    `json:"metric"`
	Timestamp      time.Time `json:"timestamp"`
	ExpectedValue  float64   `json:"expectedValue"`
	ActualValue    float64   `json:"actualValue"`
	Deviation      float64   `json:"deviation"`
	Severity       string    `json:"severity"` // "critical", "high", "medium", "low"
	PossibleCauses []string  `json:"possibleCauses"`
}

// AnalyzeTrends performs comprehensive trend analysis across documents
func (ta *TrendAnalyzer) AnalyzeTrends(ctx context.Context, dealName string, documents []documents.DocumentInfo, historicalData map[string]interface{}) (*TrendAnalysisResult, error) {
	result := &TrendAnalysisResult{
		DealName:     dealName,
		AnalysisDate: time.Now(),
		KeyInsights:  make([]TrendInsight, 0),
		Anomalies:    make([]TrendAnomaly, 0),
	}

	// Extract time series data from documents
	timeSeriesData, err := ta.extractTimeSeriesData(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("failed to extract time series data: %w", err)
	}

	// Determine time range and granularity
	result.TimeRange = ta.determineTimeRange(timeSeriesData)

	// Analyze financial trends
	result.FinancialTrends = ta.analyzeFinancialTrends(timeSeriesData, historicalData)

	// Analyze operational trends
	result.OperationalTrends = ta.analyzeOperationalTrends(timeSeriesData)

	// Analyze market trends
	result.MarketTrends = ta.analyzeMarketTrends(timeSeriesData, historicalData)

	// Analyze risk trends
	result.RiskTrends = ta.analyzeRiskTrends(timeSeriesData)

	// Generate projections
	result.Projections = ta.generateProjections(result)

	// Identify key insights
	result.KeyInsights = ta.identifyKeyInsights(result)

	// Detect anomalies
	result.Anomalies = ta.detectAnomalies(timeSeriesData)

	// Generate summary
	result.Summary = ta.generateTrendSummary(result)

	return result, nil
}

// extractTimeSeriesData extracts time series data from documents
func (ta *TrendAnalyzer) extractTimeSeriesData(ctx context.Context, documents []documents.DocumentInfo) (map[string][]DataPoint, error) {
	timeSeriesData := make(map[string][]DataPoint)

	// In a real implementation, this would parse documents and extract time series
	// For now, generate sample data

	// Revenue data
	baseRevenue := 10000000.0
	revenueData := make([]DataPoint, 0)
	for i := 0; i < 12; i++ {
		timestamp := time.Now().AddDate(0, -12+i, 0)
		growth := 1.0 + (0.02 * float64(i)) + (0.01 * math.Sin(float64(i)))
		value := baseRevenue * growth
		revenueData = append(revenueData, DataPoint{
			Timestamp: timestamp,
			Value:     value,
		})
	}
	timeSeriesData["revenue"] = revenueData

	// EBITDA data
	ebitdaData := make([]DataPoint, 0)
	for i := 0; i < 12; i++ {
		timestamp := time.Now().AddDate(0, -12+i, 0)
		margin := 0.20 + (0.005 * float64(i))
		value := revenueData[i].Value * margin
		ebitdaData = append(ebitdaData, DataPoint{
			Timestamp: timestamp,
			Value:     value,
		})
	}
	timeSeriesData["ebitda"] = ebitdaData

	// Customer data
	baseCustomers := 1000.0
	customerData := make([]DataPoint, 0)
	for i := 0; i < 12; i++ {
		timestamp := time.Now().AddDate(0, -12+i, 0)
		growth := 1.0 + (0.03 * float64(i))
		value := baseCustomers * growth
		customerData = append(customerData, DataPoint{
			Timestamp: timestamp,
			Value:     value,
		})
	}
	timeSeriesData["customers"] = customerData

	return timeSeriesData, nil
}

// determineTimeRange determines the time range of the data
func (ta *TrendAnalyzer) determineTimeRange(data map[string][]DataPoint) TimeRange {
	var start, end time.Time
	dataPoints := 0

	for _, series := range data {
		if len(series) > 0 {
			if start.IsZero() || series[0].Timestamp.Before(start) {
				start = series[0].Timestamp
			}
			if end.IsZero() || series[len(series)-1].Timestamp.After(end) {
				end = series[len(series)-1].Timestamp
			}
			if len(series) > dataPoints {
				dataPoints = len(series)
			}
		}
	}

	// Determine granularity
	granularity := "monthly"
	if dataPoints > 365 {
		granularity = "daily"
	} else if dataPoints < 12 {
		granularity = "quarterly"
	}

	return TimeRange{
		Start:       start,
		End:         end,
		DataPoints:  dataPoints,
		Granularity: granularity,
	}
}

// analyzeFinancialTrends analyzes financial metric trends
func (ta *TrendAnalyzer) analyzeFinancialTrends(data map[string][]DataPoint, historicalData map[string]interface{}) *FinancialTrends {
	trends := &FinancialTrends{
		MarginTrends: make(map[string]*MetricTrend),
		GrowthRates:  make(map[string]float64),
		Volatility:   make(map[string]float64),
	}

	// Analyze revenue trend
	if revenueData, exists := data["revenue"]; exists {
		trends.RevenueTrend = ta.analyzeMetricTrend("Revenue", revenueData)
		trends.GrowthRates["revenue"] = ta.calculateCAGR(revenueData)
		trends.Volatility["revenue"] = ta.calculateVolatility(revenueData)
	}

	// Analyze profitability trend
	if ebitdaData, exists := data["ebitda"]; exists {
		trends.ProfitabilityTrend = ta.analyzeMetricTrend("EBITDA", ebitdaData)
		trends.GrowthRates["ebitda"] = ta.calculateCAGR(ebitdaData)
		trends.Volatility["ebitda"] = ta.calculateVolatility(ebitdaData)
	}

	// Calculate margin trends
	if revenueData, revExists := data["revenue"]; revExists {
		if ebitdaData, ebitdaExists := data["ebitda"]; ebitdaExists {
			marginData := ta.calculateMarginTrend(revenueData, ebitdaData)
			trends.MarginTrends["ebitda_margin"] = ta.analyzeMetricTrend("EBITDA Margin", marginData)
		}
	}

	// Analyze seasonality
	trends.Seasonality = ta.analyzeSeasonality(data["revenue"])

	return trends
}

// analyzeMetricTrend analyzes a single metric trend
func (ta *TrendAnalyzer) analyzeMetricTrend(metricName string, data []DataPoint) *MetricTrend {
	if len(data) < 2 {
		return nil
	}

	trend := &MetricTrend{
		MetricName: metricName,
		DataPoints: data,
	}

	// Calculate trend line (simple linear regression)
	trend.TrendLine = ta.calculateTrendLine(data)

	// Determine direction
	if trend.TrendLine.Slope > 0.01 {
		trend.Direction = "increasing"
	} else if trend.TrendLine.Slope < -0.01 {
		trend.Direction = "decreasing"
	} else {
		trend.Direction = "stable"
	}

	// Calculate strength and correlation
	trend.Strength = math.Abs(trend.TrendLine.Slope) / ta.calculateMean(data)
	trend.Correlation = ta.calculateRSquared(data, trend.TrendLine)

	// Generate forecast
	trend.Forecast = ta.generateForecast(trend.TrendLine, data, 3) // 3 periods ahead

	return trend
}

// calculateTrendLine calculates linear regression trend line
func (ta *TrendAnalyzer) calculateTrendLine(data []DataPoint) TrendLine {
	n := float64(len(data))

	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, point := range data {
		x := float64(i) // Use index as X for simplicity
		y := point.Value

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope and intercept
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	return TrendLine{
		Type:      "linear",
		Slope:     slope,
		Intercept: intercept,
		Formula:   fmt.Sprintf("y = %.2fx + %.2f", slope, intercept),
	}
}

// calculateRSquared calculates R-squared value for trend fit
func (ta *TrendAnalyzer) calculateRSquared(data []DataPoint, trendLine TrendLine) float64 {
	mean := ta.calculateMean(data)

	ssTotal := 0.0
	ssResidual := 0.0

	for i, point := range data {
		// Predicted value
		predicted := trendLine.Slope*float64(i) + trendLine.Intercept

		// Total sum of squares
		ssTotal += math.Pow(point.Value-mean, 2)

		// Residual sum of squares
		ssResidual += math.Pow(point.Value-predicted, 2)
	}

	if ssTotal == 0 {
		return 0
	}

	return 1 - (ssResidual / ssTotal)
}

// calculateCAGR calculates Compound Annual Growth Rate
func (ta *TrendAnalyzer) calculateCAGR(data []DataPoint) float64 {
	if len(data) < 2 {
		return 0
	}

	startValue := data[0].Value
	endValue := data[len(data)-1].Value

	if startValue <= 0 {
		return 0
	}

	years := data[len(data)-1].Timestamp.Sub(data[0].Timestamp).Hours() / (24 * 365)
	if years <= 0 {
		return 0
	}

	return math.Pow(endValue/startValue, 1/years) - 1
}

// calculateVolatility calculates volatility (standard deviation of returns)
func (ta *TrendAnalyzer) calculateVolatility(data []DataPoint) float64 {
	if len(data) < 2 {
		return 0
	}

	returns := make([]float64, len(data)-1)
	for i := 1; i < len(data); i++ {
		if data[i-1].Value != 0 {
			returns[i-1] = (data[i].Value - data[i-1].Value) / data[i-1].Value
		}
	}

	return ta.calculateStdDev(returns)
}

// analyzeSeasonality analyzes seasonal patterns
func (ta *TrendAnalyzer) analyzeSeasonality(data []DataPoint) *SeasonalityAnalysis {
	if len(data) < 12 {
		return &SeasonalityAnalysis{HasSeasonality: false}
	}

	seasonality := &SeasonalityAnalysis{
		SeasonalFactors: make(map[string]float64),
		PeakPeriods:     make([]string, 0),
		LowPeriods:      make([]string, 0),
	}

	// Calculate monthly averages
	monthlyAvg := make(map[int][]float64)
	for _, point := range data {
		month := int(point.Timestamp.Month())
		monthlyAvg[month] = append(monthlyAvg[month], point.Value)
	}

	// Calculate seasonal factors
	overallAvg := ta.calculateMean(data)
	monthNames := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	for month := 1; month <= 12; month++ {
		if values, exists := monthlyAvg[month]; exists && len(values) > 0 {
			monthAvg := ta.calculateMeanFromSlice(values)
			factor := monthAvg / overallAvg
			seasonality.SeasonalFactors[monthNames[month]] = factor

			if factor > 1.1 {
				seasonality.PeakPeriods = append(seasonality.PeakPeriods, monthNames[month])
			} else if factor < 0.9 {
				seasonality.LowPeriods = append(seasonality.LowPeriods, monthNames[month])
			}
		}
	}

	// Determine if seasonality exists
	maxFactor := 0.0
	minFactor := 2.0
	for _, factor := range seasonality.SeasonalFactors {
		if factor > maxFactor {
			maxFactor = factor
		}
		if factor < minFactor {
			minFactor = factor
		}
	}

	seasonality.HasSeasonality = (maxFactor - minFactor) > 0.2
	seasonality.CycleDuration = "annual"

	return seasonality
}

// generateForecast generates forecast points
func (ta *TrendAnalyzer) generateForecast(trendLine TrendLine, historicalData []DataPoint, periods int) []ForecastPoint {
	forecast := make([]ForecastPoint, periods)

	lastIndex := float64(len(historicalData) - 1)
	lastTimestamp := historicalData[len(historicalData)-1].Timestamp

	// Calculate standard error for confidence intervals
	stdError := ta.calculateStandardError(historicalData, trendLine)

	for i := 0; i < periods; i++ {
		forecastIndex := lastIndex + float64(i+1)
		forecastValue := trendLine.Slope*forecastIndex + trendLine.Intercept

		// Add time based on data granularity (assume monthly)
		forecastTime := lastTimestamp.AddDate(0, i+1, 0)

		// Calculate confidence intervals (95%)
		margin := 1.96 * stdError

		forecast[i] = ForecastPoint{
			Timestamp:      forecastTime,
			Value:          forecastValue,
			ConfidenceLow:  forecastValue - margin,
			ConfidenceHigh: forecastValue + margin,
			Probability:    0.95,
		}
	}

	return forecast
}

// analyzeOperationalTrends analyzes operational metrics
func (ta *TrendAnalyzer) analyzeOperationalTrends(data map[string][]DataPoint) *OperationalTrends {
	trends := &OperationalTrends{
		EfficiencyTrends:   make(map[string]*MetricTrend),
		ProductivityTrends: make(map[string]*MetricTrend),
		QualityMetrics:     make(map[string]*MetricTrend),
	}

	// Analyze customer metrics
	trends.CustomerMetrics = ta.analyzeCustomerTrends(data)

	// Analyze employee metrics
	trends.EmployeeMetrics = ta.analyzeEmployeeTrends(data)

	// Add other operational metrics as available

	return trends
}

// analyzeCustomerTrends analyzes customer-related trends
func (ta *TrendAnalyzer) analyzeCustomerTrends(data map[string][]DataPoint) *CustomerTrends {
	customerTrends := &CustomerTrends{}

	// Analyze customer acquisition
	if custData, exists := data["customers"]; exists {
		customerTrends.AcquisitionTrend = ta.analyzeMetricTrend("Customer Acquisition", custData)

		// Calculate churn analysis
		customerTrends.ChurnAnalysis = &ChurnAnalysis{
			CurrentRate:      0.10, // 10% churn
			TrendDirection:   "decreasing",
			PredictedRate:    0.08,
			RiskFactors:      []string{"Price sensitivity", "Competition"},
			RetentionDrivers: []string{"Product quality", "Customer service"},
		}
	}

	return customerTrends
}

// analyzeEmployeeTrends analyzes employee-related trends
func (ta *TrendAnalyzer) analyzeEmployeeTrends(data map[string][]DataPoint) *EmployeeTrends {
	// In a real implementation, this would analyze actual employee data
	return &EmployeeTrends{
		// Placeholder for employee trends
	}
}

// analyzeMarketTrends analyzes market-related trends
func (ta *TrendAnalyzer) analyzeMarketTrends(data map[string][]DataPoint, historicalData map[string]interface{}) *MarketTrendAnalysis {
	marketTrends := &MarketTrendAnalysis{
		CompetitiveTrends: make(map[string]*MetricTrend),
		DemandIndicators:  make([]DemandIndicator, 0),
	}

	// Add demand indicators
	marketTrends.DemandIndicators = []DemandIndicator{
		{
			Name:         "Market Growth",
			CurrentValue: 0.15,
			Trend:        "increasing",
			Impact:       "positive",
		},
		{
			Name:         "Customer Inquiries",
			CurrentValue: 150,
			Trend:        "increasing",
			Impact:       "positive",
		},
	}

	return marketTrends
}

// analyzeRiskTrends analyzes risk-related trends
func (ta *TrendAnalyzer) analyzeRiskTrends(data map[string][]DataPoint) *RiskTrendAnalysis {
	riskTrends := &RiskTrendAnalysis{
		RiskCategories:     make(map[string]*MetricTrend),
		EmergingRisks:      make([]EmergingRisk, 0),
		MitigationProgress: make(map[string]float64),
	}

	// Add emerging risks
	riskTrends.EmergingRisks = []EmergingRisk{
		{
			Name:            "Cybersecurity threats",
			FirstIdentified: time.Now().AddDate(0, -3, 0),
			GrowthRate:      0.25,
			PotentialImpact: "high",
			Likelihood:      0.7,
		},
		{
			Name:            "Supply chain disruption",
			FirstIdentified: time.Now().AddDate(0, -6, 0),
			GrowthRate:      0.15,
			PotentialImpact: "medium",
			Likelihood:      0.5,
		},
	}

	// Mitigation progress
	riskTrends.MitigationProgress["cybersecurity"] = 0.65
	riskTrends.MitigationProgress["supply_chain"] = 0.45

	return riskTrends
}

// generateProjections generates future projections
func (ta *TrendAnalyzer) generateProjections(analysis *TrendAnalysisResult) *TrendProjections {
	projections := &TrendProjections{
		TimeHorizon:    "1 year",
		RiskProjection: make(map[string]float64),
		Assumptions:    make([]string, 0),
	}

	// Revenue projections
	if analysis.FinancialTrends != nil && analysis.FinancialTrends.RevenueTrend != nil {
		projections.RevenueProjection = ta.createProjectionScenarios(
			analysis.FinancialTrends.RevenueTrend,
			analysis.FinancialTrends.GrowthRates["revenue"],
		)
	}

	// Add assumptions
	projections.Assumptions = []string{
		"Market conditions remain stable",
		"No major regulatory changes",
		"Competitive landscape unchanged",
		"Current management team remains",
	}

	// Risk projections
	projections.RiskProjection["market_risk"] = 0.3
	projections.RiskProjection["operational_risk"] = 0.25
	projections.RiskProjection["financial_risk"] = 0.2

	return projections
}

// createProjectionScenarios creates different projection scenarios
func (ta *TrendAnalyzer) createProjectionScenarios(trend *MetricTrend, growthRate float64) *ProjectionScenarios {
	currentValue := trend.DataPoints[len(trend.DataPoints)-1].Value

	return &ProjectionScenarios{
		BestCase: ProjectionResult{
			Value:       currentValue * (1 + growthRate*1.5),
			Growth:      growthRate * 1.5,
			Probability: 0.20,
			Drivers:     []string{"Market expansion", "Product innovation"},
		},
		BaseCase: ProjectionResult{
			Value:       currentValue * (1 + growthRate),
			Growth:      growthRate,
			Probability: 0.60,
			Drivers:     []string{"Organic growth", "Market trends"},
		},
		WorstCase: ProjectionResult{
			Value:       currentValue * (1 + growthRate*0.5),
			Growth:      growthRate * 0.5,
			Probability: 0.20,
			Drivers:     []string{"Economic downturn", "Increased competition"},
		},
		MostLikely: ProjectionResult{
			Value:       currentValue * (1 + growthRate*1.1),
			Growth:      growthRate * 1.1,
			Probability: 0.70,
			Drivers:     []string{"Current trajectory", "Market conditions"},
		},
	}
}

// identifyKeyInsights identifies key insights from the analysis
func (ta *TrendAnalyzer) identifyKeyInsights(analysis *TrendAnalysisResult) []TrendInsight {
	insights := make([]TrendInsight, 0)

	// Check revenue growth
	if analysis.FinancialTrends != nil && analysis.FinancialTrends.RevenueTrend != nil {
		if analysis.FinancialTrends.GrowthRates["revenue"] > 0.20 {
			insights = append(insights, TrendInsight{
				Type:        "opportunity",
				Description: "Strong revenue growth trajectory indicates market momentum",
				Impact:      "high",
				Confidence:  0.85,
				Evidence:    []string{"20%+ revenue CAGR", "Consistent monthly growth"},
				ActionItems: []string{"Accelerate expansion", "Increase investment"},
			})
		}
	}

	// Check margin trends
	if analysis.FinancialTrends != nil && analysis.FinancialTrends.MarginTrends["ebitda_margin"] != nil {
		marginTrend := analysis.FinancialTrends.MarginTrends["ebitda_margin"]
		if marginTrend.Direction == "increasing" {
			insights = append(insights, TrendInsight{
				Type:        "opportunity",
				Description: "Improving margins indicate operational efficiency gains",
				Impact:      "medium",
				Confidence:  0.80,
				Evidence:    []string{"Margin expansion", "Cost optimization"},
				ActionItems: []string{"Continue efficiency programs", "Scale operations"},
			})
		}
	}

	// Check for risks
	if analysis.RiskTrends != nil && len(analysis.RiskTrends.EmergingRisks) > 0 {
		for _, risk := range analysis.RiskTrends.EmergingRisks {
			if risk.Likelihood > 0.6 && risk.PotentialImpact == "high" {
				insights = append(insights, TrendInsight{
					Type:        "risk",
					Description: fmt.Sprintf("Emerging risk: %s requires immediate attention", risk.Name),
					Impact:      "high",
					Confidence:  0.75,
					Evidence:    []string{fmt.Sprintf("%.0f%% growth rate", risk.GrowthRate*100)},
					ActionItems: []string{"Develop mitigation plan", "Allocate resources"},
				})
			}
		}
	}

	return insights
}

// detectAnomalies detects anomalies in the data
func (ta *TrendAnalyzer) detectAnomalies(data map[string][]DataPoint) []TrendAnomaly {
	anomalies := make([]TrendAnomaly, 0)

	for metric, series := range data {
		if len(series) < 3 {
			continue
		}

		// Calculate moving average and standard deviation
		for i := 2; i < len(series); i++ {
			// Simple anomaly detection: check if value deviates significantly
			window := series[max(0, i-3):i]
			mean := ta.calculateMeanFromDataPoints(window)
			stdDev := ta.calculateStdDevFromDataPoints(window)

			// Check for 2+ standard deviation
			deviation := math.Abs(series[i].Value - mean)
			if deviation > 2*stdDev && stdDev > 0 {
				anomalies = append(anomalies, TrendAnomaly{
					Metric:         metric,
					Timestamp:      series[i].Timestamp,
					ExpectedValue:  mean,
					ActualValue:    series[i].Value,
					Deviation:      deviation / stdDev,
					Severity:       ta.getAnomalySeverity(deviation / stdDev),
					PossibleCauses: ta.suggestAnomalyCauses(metric, series[i].Value > mean),
				})
			}
		}
	}

	return anomalies
}

// Helper methods

func (ta *TrendAnalyzer) calculateMean(data []DataPoint) float64 {
	if len(data) == 0 {
		return 0
	}

	sum := 0.0
	for _, point := range data {
		sum += point.Value
	}
	return sum / float64(len(data))
}

func (ta *TrendAnalyzer) calculateMeanFromSlice(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	sum := 0.0
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
}

func (ta *TrendAnalyzer) calculateMeanFromDataPoints(data []DataPoint) float64 {
	if len(data) == 0 {
		return 0
	}

	sum := 0.0
	for _, point := range data {
		sum += point.Value
	}
	return sum / float64(len(data))
}

func (ta *TrendAnalyzer) calculateStdDev(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}

	mean := ta.calculateMeanFromSlice(data)
	variance := 0.0

	for _, value := range data {
		variance += math.Pow(value-mean, 2)
	}

	return math.Sqrt(variance / float64(len(data)-1))
}

func (ta *TrendAnalyzer) calculateStdDevFromDataPoints(data []DataPoint) float64 {
	if len(data) < 2 {
		return 0
	}

	mean := ta.calculateMeanFromDataPoints(data)
	variance := 0.0

	for _, point := range data {
		variance += math.Pow(point.Value-mean, 2)
	}

	return math.Sqrt(variance / float64(len(data)-1))
}

func (ta *TrendAnalyzer) calculateStandardError(data []DataPoint, trendLine TrendLine) float64 {
	if len(data) < 3 {
		return 0
	}

	sumSquaredErrors := 0.0
	for i, point := range data {
		predicted := trendLine.Slope*float64(i) + trendLine.Intercept
		error := point.Value - predicted
		sumSquaredErrors += error * error
	}

	mse := sumSquaredErrors / float64(len(data)-2) // degrees of freedom
	return math.Sqrt(mse)
}

func (ta *TrendAnalyzer) calculateMarginTrend(revenue, cost []DataPoint) []DataPoint {
	marginData := make([]DataPoint, 0)

	for i := 0; i < len(revenue) && i < len(cost); i++ {
		if revenue[i].Value > 0 {
			margin := cost[i].Value / revenue[i].Value
			marginData = append(marginData, DataPoint{
				Timestamp: revenue[i].Timestamp,
				Value:     margin,
			})
		}
	}

	return marginData
}

func (ta *TrendAnalyzer) getAnomalySeverity(deviations float64) string {
	if deviations > 4 {
		return "critical"
	} else if deviations > 3 {
		return "high"
	} else if deviations > 2.5 {
		return "medium"
	}
	return "low"
}

func (ta *TrendAnalyzer) suggestAnomalyCauses(metric string, isHigher bool) []string {
	causes := make([]string, 0)

	direction := "spike"
	if !isHigher {
		direction = "drop"
	}

	switch metric {
	case "revenue":
		if isHigher {
			causes = []string{"Large deal closure", "Seasonal peak", "Marketing campaign success"}
		} else {
			causes = []string{"Customer churn", "Seasonal low", "Market disruption"}
		}
	case "costs":
		if isHigher {
			causes = []string{"One-time expense", "Investment in growth", "Supplier price increase"}
		} else {
			causes = []string{"Cost reduction initiative", "Efficiency improvement", "Vendor negotiation"}
		}
	default:
		causes = []string{fmt.Sprintf("Unusual %s in %s", direction, metric), "Data quality issue", "External factor"}
	}

	return causes
}

// generateTrendSummary generates an executive summary
func (ta *TrendAnalyzer) generateTrendSummary(analysis *TrendAnalysisResult) string {
	summary := fmt.Sprintf("Trend Analysis Summary for %s\n", analysis.DealName)
	summary += fmt.Sprintf("Analysis Period: %s to %s (%d data points)\n\n",
		analysis.TimeRange.Start.Format("Jan 2006"),
		analysis.TimeRange.End.Format("Jan 2006"),
		analysis.TimeRange.DataPoints)

	// Financial trends
	if analysis.FinancialTrends != nil {
		summary += "Financial Trends:\n"
		if rate, exists := analysis.FinancialTrends.GrowthRates["revenue"]; exists {
			summary += fmt.Sprintf("- Revenue CAGR: %.1f%%\n", rate*100)
		}
		if trend := analysis.FinancialTrends.RevenueTrend; trend != nil {
			summary += fmt.Sprintf("- Revenue trend: %s (R² = %.2f)\n", trend.Direction, trend.Correlation)
		}
		if analysis.FinancialTrends.Seasonality != nil && analysis.FinancialTrends.Seasonality.HasSeasonality {
			summary += fmt.Sprintf("- Seasonal patterns detected with peaks in: %v\n",
				analysis.FinancialTrends.Seasonality.PeakPeriods)
		}
		summary += "\n"
	}

	// Key insights
	if len(analysis.KeyInsights) > 0 {
		summary += "Key Insights:\n"
		for i, insight := range analysis.KeyInsights {
			if i >= 3 {
				break
			}
			summary += fmt.Sprintf("- %s: %s (%s impact)\n", insight.Type, insight.Description, insight.Impact)
		}
		summary += "\n"
	}

	// Projections
	if analysis.Projections != nil && analysis.Projections.RevenueProjection != nil {
		summary += fmt.Sprintf("Revenue Projections (%s):\n", analysis.Projections.TimeHorizon)
		proj := analysis.Projections.RevenueProjection
		summary += fmt.Sprintf("- Base case: $%.1fM (%.1f%% growth)\n",
			proj.BaseCase.Value/1000000,
			proj.BaseCase.Growth*100)
		summary += fmt.Sprintf("- Best case: $%.1fM (%.1f%% growth)\n",
			proj.BestCase.Value/1000000,
			proj.BestCase.Growth*100)
		summary += "\n"
	}

	// Anomalies
	if len(analysis.Anomalies) > 0 {
		summary += fmt.Sprintf("Detected %d anomalies requiring attention\n", len(analysis.Anomalies))
		criticalCount := 0
		for _, anomaly := range analysis.Anomalies {
			if anomaly.Severity == "critical" || anomaly.Severity == "high" {
				criticalCount++
			}
		}
		if criticalCount > 0 {
			summary += fmt.Sprintf("- %d high/critical severity anomalies\n", criticalCount)
		}
	}

	return summary
}

// QuickTrendAssessment performs a quick trend assessment
func (ta *TrendAnalyzer) QuickTrendAssessment(metricName string, values []float64) map[string]interface{} {
	assessment := make(map[string]interface{})

	if len(values) < 2 {
		assessment["error"] = "Insufficient data points"
		return assessment
	}

	// Convert to data points
	dataPoints := make([]DataPoint, len(values))
	now := time.Now()
	for i, value := range values {
		dataPoints[i] = DataPoint{
			Timestamp: now.AddDate(0, -len(values)+i+1, 0),
			Value:     value,
		}
	}

	// Analyze trend
	trend := ta.analyzeMetricTrend(metricName, dataPoints)

	assessment["direction"] = trend.Direction
	assessment["strength"] = trend.Strength
	assessment["correlation"] = trend.Correlation
	assessment["cagr"] = ta.calculateCAGR(dataPoints)
	assessment["volatility"] = ta.calculateVolatility(dataPoints)

	// Simple projection
	if trend.TrendLine.Slope != 0 {
		nextValue := trend.TrendLine.Slope*float64(len(values)) + trend.TrendLine.Intercept
		assessment["nextPeriodEstimate"] = nextValue
	}

	return assessment
}
