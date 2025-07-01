package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// AnomalyDetector performs anomaly detection across financial and operational data
type AnomalyDetector struct {
	aiService   *AIService
	dataMapper  *DataMapper
	sensitivity float64 // Sensitivity threshold (default 2.0 standard deviations)
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(aiService *AIService, dataMapper *DataMapper) *AnomalyDetector {
	return &AnomalyDetector{
		aiService:   aiService,
		dataMapper:  dataMapper,
		sensitivity: 2.0, // 2 standard deviations by default
	}
}

// AnomalyDetectionResult contains the results of anomaly detection
type AnomalyDetectionResult struct {
	DealName             string                  `json:"dealName"`
	AnalysisDate         time.Time               `json:"analysisDate"`
	DataQuality          *DataQualityAssessment  `json:"dataQuality"`
	FinancialAnomalies   []FinancialAnomaly      `json:"financialAnomalies"`
	OperationalAnomalies []OperationalAnomaly    `json:"operationalAnomalies"`
	PatternAnomalies     []PatternAnomaly        `json:"patternAnomalies"`
	RiskIndicators       []AnomalyRiskIndicator  `json:"riskIndicators"`
	Summary              *AnomalySummary         `json:"summary"`
	Recommendations      []AnomalyRecommendation `json:"recommendations"`
}

// DataQualityAssessment assesses the quality of input data
type DataQualityAssessment struct {
	OverallScore      float64                     `json:"overallScore"` // 0-1
	Completeness      float64                     `json:"completeness"`
	Consistency       float64                     `json:"consistency"`
	Timeliness        float64                     `json:"timeliness"`
	Accuracy          float64                     `json:"accuracy"`
	Issues            []DataQualityIssue          `json:"issues"`
	MetricAssessments map[string]MetricAssessment `json:"metricAssessments"`
}

// DataQualityIssue represents a data quality problem
type DataQualityIssue struct {
	Type        string `json:"type"` // "missing", "inconsistent", "outlier", "stale"
	Description string `json:"description"`
	Severity    string `json:"severity"` // "critical", "high", "medium", "low"
	Location    string `json:"location"`
	Impact      string `json:"impact"`
}

// MetricAssessment assesses quality for a specific metric
type MetricAssessment struct {
	MetricName   string   `json:"metricName"`
	Completeness float64  `json:"completeness"`
	Reliability  float64  `json:"reliability"`
	Issues       []string `json:"issues"`
}

// FinancialAnomaly represents an anomaly in financial data
type FinancialAnomaly struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"` // "revenue_spike", "cost_anomaly", "margin_deviation", etc.
	Metric           string                 `json:"metric"`
	Timestamp        time.Time              `json:"timestamp"`
	ExpectedValue    float64                `json:"expectedValue"`
	ActualValue      float64                `json:"actualValue"`
	Deviation        float64                `json:"deviation"` // In standard deviations
	Direction        string                 `json:"direction"` // "above", "below"
	Severity         string                 `json:"severity"`  // "critical", "high", "medium", "low"
	ConfidenceScore  float64                `json:"confidenceScore"`
	Context          map[string]interface{} `json:"context"`
	PossibleCauses   []string               `json:"possibleCauses"`
	RelatedAnomalies []string               `json:"relatedAnomalies"`
}

// OperationalAnomaly represents an anomaly in operational data
type OperationalAnomaly struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"` // "efficiency_drop", "quality_issue", "process_deviation"
	Area            string                 `json:"area"` // "production", "sales", "customer_service", etc.
	Timestamp       time.Time              `json:"timestamp"`
	Description     string                 `json:"description"`
	ImpactScore     float64                `json:"impactScore"`  // 0-1
	UrgencyLevel    string                 `json:"urgencyLevel"` // "immediate", "high", "medium", "low"
	AffectedMetrics []string               `json:"affectedMetrics"`
	Context         map[string]interface{} `json:"context"`
	Remediation     string                 `json:"remediation"`
}

// PatternAnomaly represents an anomaly in patterns or relationships
type PatternAnomaly struct {
	ID               string    `json:"id"`
	Type             string    `json:"type"` // "correlation_break", "seasonality_shift", "trend_reversal"
	Description      string    `json:"description"`
	DetectedAt       time.Time `json:"detectedAt"`
	TimeRange        TimeRange `json:"timeRange"`
	AffectedMetrics  []string  `json:"affectedMetrics"`
	PatternStrength  float64   `json:"patternStrength"`  // 0-1, how strong the anomaly is
	StatisticalScore float64   `json:"statisticalScore"` // Statistical significance
	Visualization    string    `json:"visualization"`    // Reference to visual representation
}

// AnomalyRiskIndicator represents risk indicators based on anomalies
type AnomalyRiskIndicator struct {
	Name                string   `json:"name"`
	Category            string   `json:"category"`     // "financial", "operational", "market", "compliance"
	CurrentLevel        string   `json:"currentLevel"` // "critical", "high", "medium", "low"
	Trend               string   `json:"trend"`        // "increasing", "stable", "decreasing"
	Score               float64  `json:"score"`        // 0-1
	ContributingFactors []string `json:"contributingFactors"`
	MitigationStatus    string   `json:"mitigationStatus"`
}

// AnomalySummary provides a summary of all anomalies
type AnomalySummary struct {
	TotalAnomalies     int            `json:"totalAnomalies"`
	CriticalCount      int            `json:"criticalCount"`
	HighCount          int            `json:"highCount"`
	MediumCount        int            `json:"mediumCount"`
	LowCount           int            `json:"lowCount"`
	TrendDirection     string         `json:"trendDirection"` // Overall trend
	RiskLevel          string         `json:"riskLevel"`
	KeyFindings        []string       `json:"keyFindings"`
	ImpactAssessment   string         `json:"impactAssessment"`
	DistributionByType map[string]int `json:"distributionByType"`
	TimeDistribution   map[string]int `json:"timeDistribution"`
}

// AnomalyRecommendation provides actionable recommendations
type AnomalyRecommendation struct {
	Priority    string   `json:"priority"` // "immediate", "high", "medium", "low"
	Type        string   `json:"type"`     // "investigation", "mitigation", "monitoring", "process_change"
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
	Timeline    string   `json:"timeline"`
	Owner       string   `json:"owner"`
	Impact      string   `json:"impact"`
	Resources   []string `json:"resources"`
}

// DetectAnomalies performs comprehensive anomaly detection
func (ad *AnomalyDetector) DetectAnomalies(ctx context.Context, dealName string, documents []DocumentInfo, timeSeriesData map[string][]DataPoint) (*AnomalyDetectionResult, error) {
	result := &AnomalyDetectionResult{
		DealName:             dealName,
		AnalysisDate:         time.Now(),
		FinancialAnomalies:   make([]FinancialAnomaly, 0),
		OperationalAnomalies: make([]OperationalAnomaly, 0),
		PatternAnomalies:     make([]PatternAnomaly, 0),
		RiskIndicators:       make([]AnomalyRiskIndicator, 0),
		Recommendations:      make([]AnomalyRecommendation, 0),
	}

	// Assess data quality first
	result.DataQuality = ad.assessDataQuality(timeSeriesData)

	// Detect financial anomalies
	financialAnomalies := ad.detectFinancialAnomalies(timeSeriesData)
	result.FinancialAnomalies = financialAnomalies

	// Detect operational anomalies
	operationalAnomalies := ad.detectOperationalAnomalies(timeSeriesData)
	result.OperationalAnomalies = operationalAnomalies

	// Detect pattern anomalies
	patternAnomalies := ad.detectPatternAnomalies(timeSeriesData)
	result.PatternAnomalies = patternAnomalies

	// Correlate anomalies
	ad.correlateAnomalies(result)

	// Calculate risk indicators
	result.RiskIndicators = ad.calculateRiskIndicators(result)

	// Generate summary
	result.Summary = ad.generateAnomalySummary(result)

	// Generate recommendations
	result.Recommendations = ad.generateRecommendations(result)

	return result, nil
}

// assessDataQuality assesses the quality of input data
func (ad *AnomalyDetector) assessDataQuality(data map[string][]DataPoint) *DataQualityAssessment {
	assessment := &DataQualityAssessment{
		Issues:            make([]DataQualityIssue, 0),
		MetricAssessments: make(map[string]MetricAssessment),
	}

	totalMetrics := len(data)
	completeMetrics := 0
	consistentMetrics := 0
	timelyMetrics := 0
	accurateMetrics := 0

	for metric, points := range data {
		metricAssessment := ad.assessMetricQuality(metric, points)
		assessment.MetricAssessments[metric] = metricAssessment

		// Check completeness
		if metricAssessment.Completeness > 0.9 {
			completeMetrics++
		} else if metricAssessment.Completeness < 0.7 {
			assessment.Issues = append(assessment.Issues, DataQualityIssue{
				Type:        "missing",
				Description: fmt.Sprintf("Metric '%s' has significant missing data (%.0f%% complete)", metric, metricAssessment.Completeness*100),
				Severity:    ad.getDataQualitySeverity(metricAssessment.Completeness),
				Location:    metric,
				Impact:      "May affect accuracy of trend analysis and projections",
			})
		}

		// Check consistency
		if ad.isConsistent(points) {
			consistentMetrics++
		}

		// Check timeliness
		if ad.isTimely(points) {
			timelyMetrics++
		}

		// Check accuracy (simplified - would use more sophisticated methods in production)
		if metricAssessment.Reliability > 0.8 {
			accurateMetrics++
		}
	}

	// Calculate overall scores
	if totalMetrics > 0 {
		assessment.Completeness = float64(completeMetrics) / float64(totalMetrics)
		assessment.Consistency = float64(consistentMetrics) / float64(totalMetrics)
		assessment.Timeliness = float64(timelyMetrics) / float64(totalMetrics)
		assessment.Accuracy = float64(accurateMetrics) / float64(totalMetrics)

		// Overall score is weighted average
		assessment.OverallScore = (assessment.Completeness*0.3 +
			assessment.Consistency*0.25 +
			assessment.Timeliness*0.2 +
			assessment.Accuracy*0.25)
	}

	return assessment
}

// assessMetricQuality assesses quality for a specific metric
func (ad *AnomalyDetector) assessMetricQuality(metricName string, data []DataPoint) MetricAssessment {
	assessment := MetricAssessment{
		MetricName: metricName,
		Issues:     make([]string, 0),
	}

	if len(data) == 0 {
		assessment.Completeness = 0
		assessment.Reliability = 0
		assessment.Issues = append(assessment.Issues, "No data available")
		return assessment
	}

	// Check for missing values (simplified - assuming 0 means missing)
	nonZeroCount := 0
	for _, point := range data {
		if point.Value != 0 {
			nonZeroCount++
		}
	}
	assessment.Completeness = float64(nonZeroCount) / float64(len(data))

	// Check reliability based on variance and outliers
	if len(data) > 3 {
		mean := ad.calculateMean(data)
		stdDev := ad.calculateStdDev(data)

		if stdDev > 0 {
			cv := stdDev / mean // Coefficient of variation
			if cv < 0.5 {
				assessment.Reliability = 0.9
			} else if cv < 1.0 {
				assessment.Reliability = 0.7
			} else {
				assessment.Reliability = 0.5
				assessment.Issues = append(assessment.Issues, "High variability detected")
			}
		} else {
			assessment.Reliability = 1.0 // No variation
		}
	} else {
		assessment.Reliability = 0.5 // Not enough data
		assessment.Issues = append(assessment.Issues, "Insufficient data points")
	}

	return assessment
}

// detectFinancialAnomalies detects anomalies in financial metrics
func (ad *AnomalyDetector) detectFinancialAnomalies(data map[string][]DataPoint) []FinancialAnomaly {
	anomalies := make([]FinancialAnomaly, 0)

	// Define financial metrics to analyze
	financialMetrics := []string{"revenue", "costs", "ebitda", "cash_flow", "working_capital"}

	for _, metric := range financialMetrics {
		if series, exists := data[metric]; exists && len(series) > 3 {
			// Statistical anomaly detection
			metricAnomalies := ad.detectStatisticalAnomalies(metric, series)

			// Convert to financial anomalies
			for _, anomaly := range metricAnomalies {
				finAnomaly := FinancialAnomaly{
					ID:               fmt.Sprintf("fin_%s_%d", metric, len(anomalies)),
					Type:             ad.classifyFinancialAnomalyType(metric, anomaly),
					Metric:           metric,
					Timestamp:        anomaly.Timestamp,
					ExpectedValue:    anomaly.ExpectedValue,
					ActualValue:      anomaly.ActualValue,
					Deviation:        anomaly.Deviation,
					Direction:        anomaly.Direction,
					Severity:         anomaly.Severity,
					ConfidenceScore:  ad.calculateConfidenceScore(anomaly),
					Context:          ad.gatherFinancialContext(metric, anomaly, data),
					PossibleCauses:   ad.suggestFinancialCauses(metric, anomaly),
					RelatedAnomalies: make([]string, 0),
				}

				anomalies = append(anomalies, finAnomaly)
			}
		}
	}

	// Detect ratio anomalies
	ratioAnomalies := ad.detectRatioAnomalies(data)
	anomalies = append(anomalies, ratioAnomalies...)

	// Sort by severity and timestamp
	sort.Slice(anomalies, func(i, j int) bool {
		severityOrder := map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1}
		if severityOrder[anomalies[i].Severity] != severityOrder[anomalies[j].Severity] {
			return severityOrder[anomalies[i].Severity] > severityOrder[anomalies[j].Severity]
		}
		return anomalies[i].Timestamp.After(anomalies[j].Timestamp)
	})

	return anomalies
}

// detectStatisticalAnomalies uses statistical methods to detect anomalies
func (ad *AnomalyDetector) detectStatisticalAnomalies(metric string, data []DataPoint) []StatisticalAnomaly {
	anomalies := make([]StatisticalAnomaly, 0)

	if len(data) < 4 {
		return anomalies
	}

	// Use moving average and standard deviation
	windowSize := min(len(data)/4, 6) // Adaptive window size

	for i := windowSize; i < len(data); i++ {
		// Calculate statistics for the window
		window := data[max(0, i-windowSize):i]
		mean := ad.calculateMean(window)
		stdDev := ad.calculateStdDev(window)

		if stdDev == 0 {
			continue // Skip if no variation
		}

		// Check if current point is anomalous
		currentValue := data[i].Value
		zScore := (currentValue - mean) / stdDev

		if math.Abs(zScore) > ad.sensitivity {
			direction := "above"
			if zScore < 0 {
				direction = "below"
			}

			anomaly := StatisticalAnomaly{
				Timestamp:     data[i].Timestamp,
				ExpectedValue: mean,
				ActualValue:   currentValue,
				Deviation:     math.Abs(zScore),
				Direction:     direction,
				Severity:      ad.calculateAnomalySeverity(math.Abs(zScore)),
			}

			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

// StatisticalAnomaly represents a statistical anomaly (internal type)
type StatisticalAnomaly struct {
	Timestamp     time.Time
	ExpectedValue float64
	ActualValue   float64
	Deviation     float64
	Direction     string
	Severity      string
}

// detectRatioAnomalies detects anomalies in financial ratios
func (ad *AnomalyDetector) detectRatioAnomalies(data map[string][]DataPoint) []FinancialAnomaly {
	anomalies := make([]FinancialAnomaly, 0)

	// Check profit margin anomalies
	if revenue, revExists := data["revenue"]; revExists {
		if ebitda, ebitdaExists := data["ebitda"]; ebitdaExists {
			marginAnomalies := ad.detectMarginAnomalies(revenue, ebitda, "EBITDA Margin")
			anomalies = append(anomalies, marginAnomalies...)
		}
	}

	// Check efficiency ratios
	if revenue, revExists := data["revenue"]; revExists {
		if costs, costExists := data["costs"]; costExists {
			efficiencyAnomalies := ad.detectEfficiencyAnomalies(revenue, costs)
			anomalies = append(anomalies, efficiencyAnomalies...)
		}
	}

	return anomalies
}

// detectMarginAnomalies detects anomalies in margin calculations
func (ad *AnomalyDetector) detectMarginAnomalies(revenue, profit []DataPoint, marginType string) []FinancialAnomaly {
	anomalies := make([]FinancialAnomaly, 0)

	// Calculate margins
	margins := make([]DataPoint, 0)
	for i := 0; i < len(revenue) && i < len(profit); i++ {
		if revenue[i].Value > 0 {
			margin := profit[i].Value / revenue[i].Value
			margins = append(margins, DataPoint{
				Timestamp: revenue[i].Timestamp,
				Value:     margin,
			})
		}
	}

	// Detect anomalies in margins
	marginAnomalies := ad.detectStatisticalAnomalies(marginType, margins)

	for _, anomaly := range marginAnomalies {
		finAnomaly := FinancialAnomaly{
			ID:              fmt.Sprintf("margin_%s_%d", marginType, len(anomalies)),
			Type:            "margin_deviation",
			Metric:          marginType,
			Timestamp:       anomaly.Timestamp,
			ExpectedValue:   anomaly.ExpectedValue * 100, // Convert to percentage
			ActualValue:     anomaly.ActualValue * 100,
			Deviation:       anomaly.Deviation,
			Direction:       anomaly.Direction,
			Severity:        anomaly.Severity,
			ConfidenceScore: 0.85,
			Context:         map[string]interface{}{"marginType": marginType},
			PossibleCauses:  ad.suggestMarginCauses(marginType, anomaly),
		}
		anomalies = append(anomalies, finAnomaly)
	}

	return anomalies
}

// detectEfficiencyAnomalies detects anomalies in operational efficiency
func (ad *AnomalyDetector) detectEfficiencyAnomalies(revenue, costs []DataPoint) []FinancialAnomaly {
	anomalies := make([]FinancialAnomaly, 0)

	// Calculate cost-to-revenue ratio
	ratios := make([]DataPoint, 0)
	for i := 0; i < len(revenue) && i < len(costs); i++ {
		if revenue[i].Value > 0 {
			ratio := costs[i].Value / revenue[i].Value
			ratios = append(ratios, DataPoint{
				Timestamp: revenue[i].Timestamp,
				Value:     ratio,
			})
		}
	}

	// Detect anomalies in efficiency
	effAnomalies := ad.detectStatisticalAnomalies("efficiency", ratios)

	for _, anomaly := range effAnomalies {
		finAnomaly := FinancialAnomaly{
			ID:              fmt.Sprintf("eff_%d", len(anomalies)),
			Type:            "efficiency_anomaly",
			Metric:          "cost_efficiency",
			Timestamp:       anomaly.Timestamp,
			ExpectedValue:   anomaly.ExpectedValue,
			ActualValue:     anomaly.ActualValue,
			Deviation:       anomaly.Deviation,
			Direction:       anomaly.Direction,
			Severity:        anomaly.Severity,
			ConfidenceScore: 0.80,
			Context:         map[string]interface{}{"ratioType": "cost_to_revenue"},
			PossibleCauses: []string{
				"Unexpected cost increase",
				"Revenue decline",
				"Operational inefficiency",
				"One-time expenses",
			},
		}
		anomalies = append(anomalies, finAnomaly)
	}

	return anomalies
}

// detectOperationalAnomalies detects anomalies in operational metrics
func (ad *AnomalyDetector) detectOperationalAnomalies(data map[string][]DataPoint) []OperationalAnomaly {
	anomalies := make([]OperationalAnomaly, 0)

	// Check customer-related anomalies
	if customers, exists := data["customers"]; exists {
		customerAnomalies := ad.detectCustomerAnomalies(customers)
		anomalies = append(anomalies, customerAnomalies...)
	}

	// Check productivity anomalies
	if productivity, exists := data["productivity"]; exists {
		prodAnomalies := ad.detectProductivityAnomalies(productivity)
		anomalies = append(anomalies, prodAnomalies...)
	}

	// Check quality metrics
	qualityMetrics := []string{"defect_rate", "customer_satisfaction", "service_level"}
	for _, metric := range qualityMetrics {
		if series, exists := data[metric]; exists {
			qualityAnomalies := ad.detectQualityAnomalies(metric, series)
			anomalies = append(anomalies, qualityAnomalies...)
		}
	}

	return anomalies
}

// detectCustomerAnomalies detects anomalies in customer metrics
func (ad *AnomalyDetector) detectCustomerAnomalies(customers []DataPoint) []OperationalAnomaly {
	anomalies := make([]OperationalAnomaly, 0)

	// Detect statistical anomalies
	statAnomalies := ad.detectStatisticalAnomalies("customers", customers)

	for _, anomaly := range statAnomalies {
		opAnomaly := OperationalAnomaly{
			ID:              fmt.Sprintf("cust_%d", len(anomalies)),
			Type:            "customer_anomaly",
			Area:            "customer_management",
			Timestamp:       anomaly.Timestamp,
			Description:     ad.describeCustomerAnomaly(anomaly),
			ImpactScore:     ad.calculateImpactScore(anomaly),
			UrgencyLevel:    ad.determineUrgencyLevel(anomaly),
			AffectedMetrics: []string{"customer_count", "revenue_potential"},
			Context: map[string]interface{}{
				"deviation": anomaly.Deviation,
				"direction": anomaly.Direction,
			},
			Remediation: ad.suggestCustomerRemediation(anomaly),
		}
		anomalies = append(anomalies, opAnomaly)
	}

	return anomalies
}

// detectProductivityAnomalies detects anomalies in productivity metrics
func (ad *AnomalyDetector) detectProductivityAnomalies(productivity []DataPoint) []OperationalAnomaly {
	anomalies := make([]OperationalAnomaly, 0)

	// Detect drops in productivity
	for i := 1; i < len(productivity); i++ {
		if productivity[i].Value < productivity[i-1].Value*0.8 { // 20% drop
			anomaly := OperationalAnomaly{
				ID:        fmt.Sprintf("prod_%d", len(anomalies)),
				Type:      "efficiency_drop",
				Area:      "production",
				Timestamp: productivity[i].Timestamp,
				Description: fmt.Sprintf("Significant productivity drop: %.1f%% decrease",
					(1-productivity[i].Value/productivity[i-1].Value)*100),
				ImpactScore:     0.8,
				UrgencyLevel:    "high",
				AffectedMetrics: []string{"productivity", "output", "efficiency"},
				Context: map[string]interface{}{
					"previousValue": productivity[i-1].Value,
					"currentValue":  productivity[i].Value,
				},
				Remediation: "Investigate causes of productivity decline and implement corrective measures",
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

// detectQualityAnomalies detects anomalies in quality metrics
func (ad *AnomalyDetector) detectQualityAnomalies(metric string, data []DataPoint) []OperationalAnomaly {
	anomalies := make([]OperationalAnomaly, 0)

	statAnomalies := ad.detectStatisticalAnomalies(metric, data)

	for _, anomaly := range statAnomalies {
		// For quality metrics, deterioration is more concerning
		if (metric == "defect_rate" && anomaly.Direction == "above") ||
			(metric == "customer_satisfaction" && anomaly.Direction == "below") {

			opAnomaly := OperationalAnomaly{
				ID:              fmt.Sprintf("qual_%s_%d", metric, len(anomalies)),
				Type:            "quality_issue",
				Area:            "quality_control",
				Timestamp:       anomaly.Timestamp,
				Description:     fmt.Sprintf("Quality metric '%s' shows concerning deviation", metric),
				ImpactScore:     0.7,
				UrgencyLevel:    "high",
				AffectedMetrics: []string{metric, "customer_retention", "brand_reputation"},
				Context: map[string]interface{}{
					"metric":    metric,
					"deviation": anomaly.Deviation,
				},
				Remediation: "Review quality control processes and implement improvements",
			}
			anomalies = append(anomalies, opAnomaly)
		}
	}

	return anomalies
}

// detectPatternAnomalies detects anomalies in patterns and relationships
func (ad *AnomalyDetector) detectPatternAnomalies(data map[string][]DataPoint) []PatternAnomaly {
	anomalies := make([]PatternAnomaly, 0)

	// Detect correlation breaks
	correlationAnomalies := ad.detectCorrelationAnomalies(data)
	anomalies = append(anomalies, correlationAnomalies...)

	// Detect seasonality shifts
	seasonalityAnomalies := ad.detectSeasonalityAnomalies(data)
	anomalies = append(anomalies, seasonalityAnomalies...)

	// Detect trend reversals
	trendAnomalies := ad.detectTrendReversals(data)
	anomalies = append(anomalies, trendAnomalies...)

	return anomalies
}

// detectCorrelationAnomalies detects breaks in expected correlations
func (ad *AnomalyDetector) detectCorrelationAnomalies(data map[string][]DataPoint) []PatternAnomaly {
	anomalies := make([]PatternAnomaly, 0)

	// Check revenue-cost correlation
	if revenue, revExists := data["revenue"]; revExists {
		if costs, costExists := data["costs"]; costExists {
			if len(revenue) >= 6 && len(costs) >= 6 {
				// Calculate rolling correlation
				windowSize := 3
				for i := windowSize; i < len(revenue)-windowSize && i < len(costs)-windowSize; i++ {
					prevCorr := ad.calculateCorrelation(
						revenue[i-windowSize:i],
						costs[i-windowSize:i],
					)
					currCorr := ad.calculateCorrelation(
						revenue[i:i+windowSize],
						costs[i:i+windowSize],
					)

					// Detect significant correlation change
					if math.Abs(prevCorr-currCorr) > 0.5 {
						anomaly := PatternAnomaly{
							ID:          fmt.Sprintf("corr_%d", len(anomalies)),
							Type:        "correlation_break",
							Description: fmt.Sprintf("Revenue-Cost correlation changed from %.2f to %.2f", prevCorr, currCorr),
							DetectedAt:  revenue[i].Timestamp,
							TimeRange: TimeRange{
								Start: revenue[i-windowSize].Timestamp,
								End:   revenue[i+windowSize-1].Timestamp,
							},
							AffectedMetrics:  []string{"revenue", "costs"},
							PatternStrength:  math.Abs(prevCorr - currCorr),
							StatisticalScore: ad.calculateCorrelationSignificance(prevCorr, currCorr, windowSize),
						}
						anomalies = append(anomalies, anomaly)
					}
				}
			}
		}
	}

	return anomalies
}

// detectSeasonalityAnomalies detects shifts in seasonal patterns
func (ad *AnomalyDetector) detectSeasonalityAnomalies(data map[string][]DataPoint) []PatternAnomaly {
	anomalies := make([]PatternAnomaly, 0)

	// This is a simplified version - real implementation would use more sophisticated methods
	for metric, series := range data {
		if len(series) >= 24 { // Need at least 2 years of monthly data
			// Compare seasonal patterns year-over-year
			yearOnePattern := ad.extractSeasonalPattern(series[:12])
			yearTwoPattern := ad.extractSeasonalPattern(series[12:24])

			patternDiff := ad.comparePatterns(yearOnePattern, yearTwoPattern)
			if patternDiff > 0.3 { // Significant difference
				anomaly := PatternAnomaly{
					ID:          fmt.Sprintf("season_%s_%d", metric, len(anomalies)),
					Type:        "seasonality_shift",
					Description: fmt.Sprintf("Seasonal pattern shift detected in %s", metric),
					DetectedAt:  series[12].Timestamp,
					TimeRange: TimeRange{
						Start: series[0].Timestamp,
						End:   series[23].Timestamp,
					},
					AffectedMetrics:  []string{metric},
					PatternStrength:  patternDiff,
					StatisticalScore: 0.7, // Simplified
				}
				anomalies = append(anomalies, anomaly)
			}
		}
	}

	return anomalies
}

// detectTrendReversals detects reversals in trends
func (ad *AnomalyDetector) detectTrendReversals(data map[string][]DataPoint) []PatternAnomaly {
	anomalies := make([]PatternAnomaly, 0)

	for metric, series := range data {
		if len(series) >= 6 {
			// Calculate trend direction for first and second half
			midpoint := len(series) / 2
			firstHalfTrend := ad.calculateTrendDirection(series[:midpoint])
			secondHalfTrend := ad.calculateTrendDirection(series[midpoint:])

			// Detect reversal
			if (firstHalfTrend > 0.1 && secondHalfTrend < -0.1) ||
				(firstHalfTrend < -0.1 && secondHalfTrend > 0.1) {
				anomaly := PatternAnomaly{
					ID:          fmt.Sprintf("trend_%s_%d", metric, len(anomalies)),
					Type:        "trend_reversal",
					Description: fmt.Sprintf("Trend reversal detected in %s", metric),
					DetectedAt:  series[midpoint].Timestamp,
					TimeRange: TimeRange{
						Start: series[0].Timestamp,
						End:   series[len(series)-1].Timestamp,
					},
					AffectedMetrics:  []string{metric},
					PatternStrength:  math.Abs(firstHalfTrend - secondHalfTrend),
					StatisticalScore: 0.8,
				}
				anomalies = append(anomalies, anomaly)
			}
		}
	}

	return anomalies
}

// correlateAnomalies finds relationships between different anomalies
func (ad *AnomalyDetector) correlateAnomalies(result *AnomalyDetectionResult) {
	// Correlate financial anomalies
	for i := range result.FinancialAnomalies {
		for j := range result.FinancialAnomalies {
			if i != j && ad.areAnomaliesRelated(&result.FinancialAnomalies[i], &result.FinancialAnomalies[j]) {
				result.FinancialAnomalies[i].RelatedAnomalies = append(
					result.FinancialAnomalies[i].RelatedAnomalies,
					result.FinancialAnomalies[j].ID,
				)
			}
		}
	}
}

// areAnomaliesRelated determines if two anomalies are related
func (ad *AnomalyDetector) areAnomaliesRelated(a1, a2 *FinancialAnomaly) bool {
	// Check temporal proximity (within 30 days)
	timeDiff := math.Abs(a1.Timestamp.Sub(a2.Timestamp).Hours() / 24)
	if timeDiff > 30 {
		return false
	}

	// Check if metrics are related
	relatedMetrics := map[string][]string{
		"revenue": {"costs", "ebitda", "customers"},
		"costs":   {"revenue", "ebitda", "efficiency"},
		"ebitda":  {"revenue", "costs"},
	}

	if related, exists := relatedMetrics[a1.Metric]; exists {
		for _, metric := range related {
			if a2.Metric == metric {
				return true
			}
		}
	}

	return false
}

// calculateRiskIndicators calculates risk indicators based on anomalies
func (ad *AnomalyDetector) calculateRiskIndicators(result *AnomalyDetectionResult) []AnomalyRiskIndicator {
	indicators := make([]AnomalyRiskIndicator, 0)

	// Financial risk indicator
	financialRisk := ad.calculateFinancialRisk(result.FinancialAnomalies)
	indicators = append(indicators, financialRisk)

	// Operational risk indicator
	operationalRisk := ad.calculateOperationalRisk(result.OperationalAnomalies)
	indicators = append(indicators, operationalRisk)

	// Pattern risk indicator
	patternRisk := ad.calculatePatternRisk(result.PatternAnomalies)
	indicators = append(indicators, patternRisk)

	// Data quality risk indicator
	dataQualityRisk := ad.calculateDataQualityRisk(result.DataQuality)
	indicators = append(indicators, dataQualityRisk)

	return indicators
}

// calculateFinancialRisk calculates financial risk based on anomalies
func (ad *AnomalyDetector) calculateFinancialRisk(anomalies []FinancialAnomaly) AnomalyRiskIndicator {
	riskScore := 0.0
	criticalCount := 0
	highCount := 0

	for _, anomaly := range anomalies {
		switch anomaly.Severity {
		case "critical":
			riskScore += 1.0
			criticalCount++
		case "high":
			riskScore += 0.7
			highCount++
		case "medium":
			riskScore += 0.4
		case "low":
			riskScore += 0.1
		}
	}

	// Normalize score
	if len(anomalies) > 0 {
		riskScore = riskScore / float64(len(anomalies))
	}

	// Determine level and trend
	level := ad.getRiskLevel(riskScore)
	trend := "stable"
	if criticalCount > 1 || highCount > 3 {
		trend = "increasing"
	}

	contributingFactors := make([]string, 0)
	if criticalCount > 0 {
		contributingFactors = append(contributingFactors, fmt.Sprintf("%d critical anomalies", criticalCount))
	}
	if highCount > 0 {
		contributingFactors = append(contributingFactors, fmt.Sprintf("%d high severity anomalies", highCount))
	}

	return AnomalyRiskIndicator{
		Name:                "Financial Risk",
		Category:            "financial",
		CurrentLevel:        level,
		Trend:               trend,
		Score:               riskScore,
		ContributingFactors: contributingFactors,
		MitigationStatus:    "monitoring",
	}
}

// calculateOperationalRisk calculates operational risk
func (ad *AnomalyDetector) calculateOperationalRisk(anomalies []OperationalAnomaly) AnomalyRiskIndicator {
	riskScore := 0.0
	immediateCount := 0

	for _, anomaly := range anomalies {
		riskScore += anomaly.ImpactScore
		if anomaly.UrgencyLevel == "immediate" {
			immediateCount++
		}
	}

	// Normalize
	if len(anomalies) > 0 {
		riskScore = riskScore / float64(len(anomalies))
	}

	level := ad.getRiskLevel(riskScore)
	trend := "stable"
	if immediateCount > 0 {
		trend = "increasing"
	}

	return AnomalyRiskIndicator{
		Name:                "Operational Risk",
		Category:            "operational",
		CurrentLevel:        level,
		Trend:               trend,
		Score:               riskScore,
		ContributingFactors: []string{fmt.Sprintf("%d operational anomalies", len(anomalies))},
		MitigationStatus:    "in_progress",
	}
}

// calculatePatternRisk calculates risk from pattern anomalies
func (ad *AnomalyDetector) calculatePatternRisk(anomalies []PatternAnomaly) AnomalyRiskIndicator {
	riskScore := 0.0

	for _, anomaly := range anomalies {
		riskScore += anomaly.PatternStrength * anomaly.StatisticalScore
	}

	// Normalize
	if len(anomalies) > 0 {
		riskScore = riskScore / float64(len(anomalies))
	}

	level := ad.getRiskLevel(riskScore)

	types := make(map[string]int)
	for _, anomaly := range anomalies {
		types[anomaly.Type]++
	}

	contributingFactors := make([]string, 0)
	for typ, count := range types {
		contributingFactors = append(contributingFactors, fmt.Sprintf("%d %s", count, typ))
	}

	return AnomalyRiskIndicator{
		Name:                "Pattern Risk",
		Category:            "market",
		CurrentLevel:        level,
		Trend:               "stable",
		Score:               riskScore,
		ContributingFactors: contributingFactors,
		MitigationStatus:    "monitoring",
	}
}

// calculateDataQualityRisk calculates risk from data quality issues
func (ad *AnomalyDetector) calculateDataQualityRisk(quality *DataQualityAssessment) AnomalyRiskIndicator {
	// Invert quality score to get risk score
	riskScore := 1.0 - quality.OverallScore

	level := ad.getRiskLevel(riskScore)

	contributingFactors := make([]string, 0)
	if quality.Completeness < 0.8 {
		contributingFactors = append(contributingFactors, "Incomplete data")
	}
	if quality.Consistency < 0.8 {
		contributingFactors = append(contributingFactors, "Inconsistent data")
	}
	if quality.Accuracy < 0.8 {
		contributingFactors = append(contributingFactors, "Accuracy concerns")
	}

	return AnomalyRiskIndicator{
		Name:                "Data Quality Risk",
		Category:            "compliance",
		CurrentLevel:        level,
		Trend:               "stable",
		Score:               riskScore,
		ContributingFactors: contributingFactors,
		MitigationStatus:    "not_started",
	}
}

// generateAnomalySummary generates a summary of all anomalies
func (ad *AnomalyDetector) generateAnomalySummary(result *AnomalyDetectionResult) *AnomalySummary {
	summary := &AnomalySummary{
		DistributionByType: make(map[string]int),
		TimeDistribution:   make(map[string]int),
		KeyFindings:        make([]string, 0),
	}

	// Count anomalies by severity
	for _, anomaly := range result.FinancialAnomalies {
		summary.TotalAnomalies++
		summary.DistributionByType[anomaly.Type]++

		switch anomaly.Severity {
		case "critical":
			summary.CriticalCount++
		case "high":
			summary.HighCount++
		case "medium":
			summary.MediumCount++
		case "low":
			summary.LowCount++
		}
	}

	for _, anomaly := range result.OperationalAnomalies {
		summary.TotalAnomalies++
		summary.DistributionByType[anomaly.Type]++

		if anomaly.UrgencyLevel == "immediate" {
			summary.CriticalCount++
		} else if anomaly.UrgencyLevel == "high" {
			summary.HighCount++
		}
	}

	for _, anomaly := range result.PatternAnomalies {
		summary.TotalAnomalies++
		summary.DistributionByType[anomaly.Type]++
	}

	// Determine overall trend
	if summary.CriticalCount > 2 || summary.HighCount > 5 {
		summary.TrendDirection = "worsening"
	} else if summary.TotalAnomalies > 10 {
		summary.TrendDirection = "concerning"
	} else {
		summary.TrendDirection = "stable"
	}

	// Determine risk level
	if summary.CriticalCount > 0 {
		summary.RiskLevel = "critical"
	} else if summary.HighCount > 3 {
		summary.RiskLevel = "high"
	} else if summary.MediumCount > 5 {
		summary.RiskLevel = "medium"
	} else {
		summary.RiskLevel = "low"
	}

	// Key findings
	if summary.CriticalCount > 0 {
		summary.KeyFindings = append(summary.KeyFindings,
			fmt.Sprintf("%d critical anomalies require immediate attention", summary.CriticalCount))
	}

	if len(result.FinancialAnomalies) > 5 {
		summary.KeyFindings = append(summary.KeyFindings,
			"Multiple financial anomalies detected indicating potential systemic issues")
	}

	if result.DataQuality.OverallScore < 0.7 {
		summary.KeyFindings = append(summary.KeyFindings,
			"Data quality issues may affect anomaly detection accuracy")
	}

	// Impact assessment
	if summary.RiskLevel == "critical" || summary.RiskLevel == "high" {
		summary.ImpactAssessment = "High impact on deal valuation and risk assessment. Immediate investigation required."
	} else if summary.RiskLevel == "medium" {
		summary.ImpactAssessment = "Moderate impact detected. Further analysis recommended before proceeding."
	} else {
		summary.ImpactAssessment = "Low impact. Continue monitoring but no immediate action required."
	}

	return summary
}

// generateRecommendations generates actionable recommendations
func (ad *AnomalyDetector) generateRecommendations(result *AnomalyDetectionResult) []AnomalyRecommendation {
	recommendations := make([]AnomalyRecommendation, 0)

	// Critical financial anomalies
	criticalFinancial := 0
	for _, anomaly := range result.FinancialAnomalies {
		if anomaly.Severity == "critical" {
			criticalFinancial++
		}
	}

	if criticalFinancial > 0 {
		recommendations = append(recommendations, AnomalyRecommendation{
			Priority:    "immediate",
			Type:        "investigation",
			Description: fmt.Sprintf("Investigate %d critical financial anomalies", criticalFinancial),
			Actions: []string{
				"Review source documents for affected periods",
				"Verify calculations and data entry",
				"Consult with finance team",
				"Document findings and corrections",
			},
			Timeline:  "Within 24 hours",
			Owner:     "Due Diligence Lead",
			Impact:    "Critical for accurate valuation",
			Resources: []string{"Finance team", "Source documents", "Audit trail"},
		})
	}

	// Data quality issues
	if result.DataQuality.OverallScore < 0.7 {
		recommendations = append(recommendations, AnomalyRecommendation{
			Priority:    "high",
			Type:        "process_change",
			Description: "Improve data collection and validation processes",
			Actions: []string{
				"Implement data validation checks",
				"Standardize data collection procedures",
				"Train team on data quality standards",
				"Create data quality dashboard",
			},
			Timeline:  "Within 1 week",
			Owner:     "Data Team Lead",
			Impact:    "Improves reliability of all analyses",
			Resources: []string{"Data team", "Validation tools"},
		})
	}

	// Pattern anomalies
	if len(result.PatternAnomalies) > 2 {
		recommendations = append(recommendations, AnomalyRecommendation{
			Priority:    "medium",
			Type:        "monitoring",
			Description: "Enhanced monitoring of business patterns",
			Actions: []string{
				"Set up automated alerts for pattern changes",
				"Create pattern tracking dashboard",
				"Schedule regular pattern reviews",
				"Document pattern baselines",
			},
			Timeline:  "Within 2 weeks",
			Owner:     "Analytics Team",
			Impact:    "Early detection of business changes",
			Resources: []string{"Analytics tools", "Monitoring system"},
		})
	}

	// Risk mitigation
	for _, indicator := range result.RiskIndicators {
		if indicator.CurrentLevel == "critical" || indicator.CurrentLevel == "high" {
			recommendations = append(recommendations, AnomalyRecommendation{
				Priority:    "high",
				Type:        "mitigation",
				Description: fmt.Sprintf("Mitigate %s risk", indicator.Name),
				Actions:     ad.getRiskMitigationActions(indicator),
				Timeline:    "Ongoing",
				Owner:       "Risk Management Team",
				Impact:      fmt.Sprintf("Reduces %s exposure", indicator.Category),
				Resources:   []string{"Risk management framework", "Mitigation budget"},
			})
		}
	}

	// Sort recommendations by priority
	sort.Slice(recommendations, func(i, j int) bool {
		priorityOrder := map[string]int{"immediate": 4, "high": 3, "medium": 2, "low": 1}
		return priorityOrder[recommendations[i].Priority] > priorityOrder[recommendations[j].Priority]
	})

	return recommendations
}

// Helper methods

func (ad *AnomalyDetector) calculateMean(data []DataPoint) float64 {
	if len(data) == 0 {
		return 0
	}

	sum := 0.0
	for _, point := range data {
		sum += point.Value
	}
	return sum / float64(len(data))
}

func (ad *AnomalyDetector) calculateStdDev(data []DataPoint) float64 {
	if len(data) < 2 {
		return 0
	}

	mean := ad.calculateMean(data)
	variance := 0.0

	for _, point := range data {
		variance += math.Pow(point.Value-mean, 2)
	}

	return math.Sqrt(variance / float64(len(data)-1))
}

func (ad *AnomalyDetector) calculateCorrelation(series1, series2 []DataPoint) float64 {
	if len(series1) != len(series2) || len(series1) < 2 {
		return 0
	}

	mean1 := ad.calculateMean(series1)
	mean2 := ad.calculateMean(series2)

	covariance := 0.0
	var1 := 0.0
	var2 := 0.0

	for i := 0; i < len(series1); i++ {
		diff1 := series1[i].Value - mean1
		diff2 := series2[i].Value - mean2

		covariance += diff1 * diff2
		var1 += diff1 * diff1
		var2 += diff2 * diff2
	}

	if var1 == 0 || var2 == 0 {
		return 0
	}

	return covariance / math.Sqrt(var1*var2)
}

func (ad *AnomalyDetector) calculateTrendDirection(data []DataPoint) float64 {
	if len(data) < 2 {
		return 0
	}

	// Simple linear regression slope
	n := float64(len(data))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, point := range data {
		x := float64(i)
		y := point.Value

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Normalize by mean to get relative trend
	mean := sumY / n
	if mean != 0 {
		return slope / mean
	}

	return slope
}

func (ad *AnomalyDetector) extractSeasonalPattern(data []DataPoint) []float64 {
	pattern := make([]float64, len(data))
	mean := ad.calculateMean(data)

	for i, point := range data {
		if mean != 0 {
			pattern[i] = point.Value / mean
		} else {
			pattern[i] = 1.0
		}
	}

	return pattern
}

func (ad *AnomalyDetector) comparePatterns(pattern1, pattern2 []float64) float64 {
	if len(pattern1) != len(pattern2) {
		return 1.0 // Maximum difference
	}

	totalDiff := 0.0
	for i := range pattern1 {
		totalDiff += math.Abs(pattern1[i] - pattern2[i])
	}

	return totalDiff / float64(len(pattern1))
}

func (ad *AnomalyDetector) isConsistent(data []DataPoint) bool {
	if len(data) < 2 {
		return true
	}

	// Check for excessive jumps
	for i := 1; i < len(data); i++ {
		if data[i-1].Value != 0 {
			changeRate := math.Abs(data[i].Value-data[i-1].Value) / data[i-1].Value
			if changeRate > 5.0 { // 500% change
				return false
			}
		}
	}

	return true
}

func (ad *AnomalyDetector) isTimely(data []DataPoint) bool {
	if len(data) == 0 {
		return false
	}

	// Check if most recent data is within 90 days
	mostRecent := data[len(data)-1].Timestamp
	return time.Since(mostRecent) < 90*24*time.Hour
}

func (ad *AnomalyDetector) getDataQualitySeverity(completeness float64) string {
	if completeness < 0.5 {
		return "critical"
	} else if completeness < 0.7 {
		return "high"
	} else if completeness < 0.9 {
		return "medium"
	}
	return "low"
}

func (ad *AnomalyDetector) calculateAnomalySeverity(deviation float64) string {
	if deviation > 4 {
		return "critical"
	} else if deviation > 3 {
		return "high"
	} else if deviation > 2.5 {
		return "medium"
	}
	return "low"
}

func (ad *AnomalyDetector) classifyFinancialAnomalyType(metric string, anomaly StatisticalAnomaly) string {
	if metric == "revenue" {
		if anomaly.Direction == "above" {
			return "revenue_spike"
		}
		return "revenue_drop"
	} else if metric == "costs" {
		if anomaly.Direction == "above" {
			return "cost_overrun"
		}
		return "cost_reduction"
	} else if metric == "cash_flow" {
		if anomaly.Direction == "below" {
			return "cash_flow_shortage"
		}
		return "cash_flow_surplus"
	}
	return "general_anomaly"
}

func (ad *AnomalyDetector) calculateConfidenceScore(anomaly StatisticalAnomaly) float64 {
	// Base confidence on deviation strength
	baseConfidence := math.Min(anomaly.Deviation/5.0, 1.0)

	// Adjust based on severity
	severityMultiplier := map[string]float64{
		"critical": 0.95,
		"high":     0.85,
		"medium":   0.75,
		"low":      0.65,
	}

	if mult, exists := severityMultiplier[anomaly.Severity]; exists {
		return baseConfidence * mult
	}

	return baseConfidence * 0.7
}

func (ad *AnomalyDetector) gatherFinancialContext(metric string, anomaly StatisticalAnomaly, allData map[string][]DataPoint) map[string]interface{} {
	context := make(map[string]interface{})

	context["metric"] = metric
	context["deviation_sigma"] = anomaly.Deviation
	context["percentage_change"] = ((anomaly.ActualValue - anomaly.ExpectedValue) / anomaly.ExpectedValue) * 100

	// Add related metrics context
	relatedMetrics := map[string][]string{
		"revenue": {"costs", "ebitda", "customers"},
		"costs":   {"revenue", "ebitda"},
		"ebitda":  {"revenue", "costs"},
	}

	if related, exists := relatedMetrics[metric]; exists {
		for _, relMetric := range related {
			if data, exists := allData[relMetric]; exists {
				// Find corresponding value
				for _, point := range data {
					if point.Timestamp.Equal(anomaly.Timestamp) {
						context[relMetric+"_value"] = point.Value
						break
					}
				}
			}
		}
	}

	return context
}

func (ad *AnomalyDetector) suggestFinancialCauses(metric string, anomaly StatisticalAnomaly) []string {
	causes := make([]string, 0)

	switch metric {
	case "revenue":
		if anomaly.Direction == "above" {
			causes = []string{
				"Large one-time deal",
				"Seasonal peak",
				"New product launch success",
				"Market expansion",
				"Accounting adjustment",
			}
		} else {
			causes = []string{
				"Customer churn",
				"Market downturn",
				"Competitive pressure",
				"Product issues",
				"Seasonal low",
			}
		}
	case "costs":
		if anomaly.Direction == "above" {
			causes = []string{
				"One-time expense",
				"Increased material costs",
				"Expansion investments",
				"Regulatory compliance costs",
				"Inefficiency issues",
			}
		} else {
			causes = []string{
				"Cost reduction initiative",
				"Process improvements",
				"Vendor renegotiation",
				"Reduced operations",
				"Accounting adjustment",
			}
		}
	case "cash_flow":
		if anomaly.Direction == "below" {
			causes = []string{
				"Delayed collections",
				"Increased inventory",
				"Capital expenditures",
				"Working capital changes",
				"Payment timing",
			}
		} else {
			causes = []string{
				"Improved collections",
				"Asset sales",
				"Working capital optimization",
				"Deferred payments",
				"Financing activities",
			}
		}
	default:
		causes = []string{
			"Data anomaly",
			"Business model change",
			"External factors",
			"Internal process change",
			"Measurement error",
		}
	}

	return causes
}

func (ad *AnomalyDetector) suggestMarginCauses(marginType string, anomaly StatisticalAnomaly) []string {
	if anomaly.Direction == "above" {
		return []string{
			"Pricing power improvement",
			"Cost optimization success",
			"Product mix shift",
			"Operational efficiency gains",
			"Favorable market conditions",
		}
	}
	return []string{
		"Pricing pressure",
		"Cost inflation",
		"Inefficiency issues",
		"Competitive pressure",
		"Product mix deterioration",
	}
}

func (ad *AnomalyDetector) describeCustomerAnomaly(anomaly StatisticalAnomaly) string {
	changePercent := ((anomaly.ActualValue - anomaly.ExpectedValue) / anomaly.ExpectedValue) * 100

	if anomaly.Direction == "above" {
		return fmt.Sprintf("Unusual customer growth: %.1f%% above expected (%.1f sigma deviation)", changePercent, anomaly.Deviation)
	}
	return fmt.Sprintf("Concerning customer decline: %.1f%% below expected (%.1f sigma deviation)", math.Abs(changePercent), anomaly.Deviation)
}

func (ad *AnomalyDetector) calculateImpactScore(anomaly StatisticalAnomaly) float64 {
	// Base impact on deviation and direction
	baseImpact := math.Min(anomaly.Deviation/4.0, 1.0)

	// Negative anomalies generally have higher impact
	if anomaly.Direction == "below" {
		baseImpact *= 1.2
	}

	return math.Min(baseImpact, 1.0)
}

func (ad *AnomalyDetector) determineUrgencyLevel(anomaly StatisticalAnomaly) string {
	if anomaly.Deviation > 4 && anomaly.Direction == "below" {
		return "immediate"
	} else if anomaly.Deviation > 3 {
		return "high"
	} else if anomaly.Deviation > 2.5 {
		return "medium"
	}
	return "low"
}

func (ad *AnomalyDetector) suggestCustomerRemediation(anomaly StatisticalAnomaly) string {
	if anomaly.Direction == "below" {
		return "Implement customer retention program, analyze churn reasons, enhance customer experience"
	}
	return "Ensure infrastructure can handle growth, maintain service quality, optimize onboarding"
}

func (ad *AnomalyDetector) calculateCorrelationSignificance(corr1, corr2 float64, n int) float64 {
	// Fisher's z-transformation for correlation comparison
	z1 := 0.5 * math.Log((1+corr1)/(1-corr1))
	z2 := 0.5 * math.Log((1+corr2)/(1-corr2))

	// Standard error
	se := math.Sqrt(2.0 / (float64(n) - 3))

	// Z-score
	if se > 0 {
		zScore := math.Abs(z1-z2) / se
		// Convert to 0-1 scale (simplified)
		return math.Min(zScore/4.0, 1.0)
	}

	return 0.5
}

func (ad *AnomalyDetector) getRiskLevel(score float64) string {
	if score > 0.8 {
		return "critical"
	} else if score > 0.6 {
		return "high"
	} else if score > 0.4 {
		return "medium"
	}
	return "low"
}

func (ad *AnomalyDetector) getRiskMitigationActions(indicator AnomalyRiskIndicator) []string {
	baseActions := []string{
		"Conduct detailed root cause analysis",
		"Implement enhanced monitoring",
		"Review and update risk controls",
	}

	switch indicator.Category {
	case "financial":
		return append(baseActions,
			"Review financial controls and procedures",
			"Enhance financial reporting accuracy",
			"Implement variance analysis process")
	case "operational":
		return append(baseActions,
			"Review operational procedures",
			"Implement process improvements",
			"Enhance quality control measures")
	case "market":
		return append(baseActions,
			"Monitor market conditions closely",
			"Review competitive positioning",
			"Adjust strategy as needed")
	case "compliance":
		return append(baseActions,
			"Review data collection procedures",
			"Implement data quality controls",
			"Train team on compliance requirements")
	}

	return baseActions
}

// QuickAnomalyCheck performs a quick anomaly check on a single metric
func (ad *AnomalyDetector) QuickAnomalyCheck(metricName string, currentValue float64, historicalValues []float64) map[string]interface{} {
	result := make(map[string]interface{})

	if len(historicalValues) < 3 {
		result["error"] = "Insufficient historical data"
		return result
	}

	// Convert to DataPoints for consistency
	dataPoints := make([]DataPoint, len(historicalValues))
	for i, value := range historicalValues {
		dataPoints[i] = DataPoint{
			Timestamp: time.Now().AddDate(0, -len(historicalValues)+i, 0),
			Value:     value,
		}
	}

	// Calculate statistics
	mean := ad.calculateMean(dataPoints)
	stdDev := ad.calculateStdDev(dataPoints)

	result["mean"] = mean
	result["stdDev"] = stdDev

	// Check if current value is anomalous
	if stdDev > 0 {
		zScore := (currentValue - mean) / stdDev
		result["zScore"] = zScore
		result["isAnomaly"] = math.Abs(zScore) > ad.sensitivity

		if math.Abs(zScore) > ad.sensitivity {
			result["severity"] = ad.calculateAnomalySeverity(math.Abs(zScore))
			result["direction"] = "above"
			if zScore < 0 {
				result["direction"] = "below"
			}
			result["deviationPercent"] = ((currentValue - mean) / mean) * 100
		}
	} else {
		result["isAnomaly"] = false
		result["note"] = "No variation in historical data"
	}

	return result
}

// SetSensitivity adjusts the anomaly detection sensitivity
func (ad *AnomalyDetector) SetSensitivity(sensitivity float64) {
	if sensitivity > 0 {
		ad.sensitivity = sensitivity
	}
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
