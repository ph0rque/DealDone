package analysis

import (
	"DealDone/internal/core/documents"
)

// Service defines the interface for analysis operations
type Service interface {
	// Field matching operations
	MatchFields(sourceFields []string, targetFields []string) (*FieldMatchResult, error)
	GetFieldMappingSuggestions(unmatchedFields []string, templatePath string) (map[string][]string, error)

	// Data mapping operations
	MapDataToTemplate(templatePath string, documentPaths []string, dealName string) (*MappedData, error)
	MapDocumentData(docInfo *documents.DocumentInfo, templateData interface{}) (*MappedData, error)

	// Deal valuation operations
	CalculateDealValuation(dealName string, financialData *FinancialAnalysis, marketData map[string]interface{}) (*ValuationResult, error)
	CalculateQuickValuation(revenue, ebitda, netIncome float64) (*ValuationRange, error)
	GenerateValuationReport(result *ValuationResult) (string, error)

	// Competitive analysis operations
	AnalyzeCompetitiveLandscape(dealName string, targetCompany string, documents []documents.DocumentInfo, marketData map[string]interface{}) (*CompetitiveAnalysis, error)
	QuickCompetitiveAssessment(targetCompany string, revenue float64, marketShare float64) (map[string]interface{}, error)

	// Trend analysis operations
	AnalyzeTrends(dealName string, documents []documents.DocumentInfo, historicalData map[string]interface{}) (*TrendAnalysisResult, error)
	QuickTrendAssessment(metricName string, values []float64) (map[string]interface{}, error)

	// Anomaly detection operations
	DetectAnomalies(dealName string, documents []documents.DocumentInfo, timeSeriesData map[string][]DataPoint) (*AnomalyDetectionResult, error)
	QuickAnomalyCheck(metricName string, currentValue float64, historicalValues []float64) (map[string]interface{}, error)

	// Export operations
	ExportValuationToCSV(result *ValuationResult, outputPath string) error
	ExportValuationToJSON(result *ValuationResult, outputPath string) error
	ExportCompetitiveAnalysisToCSV(analysis *CompetitiveAnalysis, outputPath string) error
	ExportCompetitiveAnalysisToJSON(analysis *CompetitiveAnalysis, outputPath string) error
	ExportTrendAnalysisToCSV(analysis *TrendAnalysisResult, outputPath string) error
	ExportTrendAnalysisToJSON(analysis *TrendAnalysisResult, outputPath string) error
	ExportAnomalyDetectionToCSV(result *AnomalyDetectionResult, outputPath string) error
	ExportAnomalyDetectionToJSON(result *AnomalyDetectionResult, outputPath string) error
	ExportCompleteAnalysisReport(dealName string, outputPath string, format string) error
}
