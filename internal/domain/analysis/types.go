package analysis

import "time"

// FieldMatchResult represents the result of field matching operation
type FieldMatchResult struct {
	MatchedFields   map[string]string   `json:"matchedFields"`
	UnmatchedSource []string            `json:"unmatchedSource"`
	UnmatchedTarget []string            `json:"unmatchedTarget"`
	Confidence      map[string]float64  `json:"confidence"`
	Suggestions     map[string][]string `json:"suggestions"`
	Success         bool                `json:"success"`
	Error           string              `json:"error,omitempty"`
}

// FinancialAnalysis represents extracted financial data from documents
type FinancialAnalysis struct {
	Revenue          float64                `json:"revenue"`
	EBITDA           float64                `json:"ebitda"`
	NetIncome        float64                `json:"netIncome"`
	TotalAssets      float64                `json:"totalAssets"`
	TotalLiabilities float64                `json:"totalLiabilities"`
	CashFlow         float64                `json:"cashFlow"`
	GrossMargin      float64                `json:"grossMargin"`
	OperatingMargin  float64                `json:"operatingMargin"`
	DebtToEquity     float64                `json:"debtToEquity"`
	CurrentRatio     float64                `json:"currentRatio"`
	WorkingCapital   float64                `json:"workingCapital"`
	Period           string                 `json:"period"`
	Currency         string                 `json:"currency"`
	Metrics          map[string]interface{} `json:"metrics,omitempty"`
	ExtractedAt      time.Time              `json:"extractedAt"`
	Confidence       float64                `json:"confidence"`
}

// Additional types already defined in the individual files
// (ValuationResult, ValuationRange, CompetitiveAnalysis, TrendAnalysisResult,
// DataPoint, AnomalyDetectionResult, MappedData) are in their respective files
