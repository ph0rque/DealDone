package documents

import (
	"time"
)

// DocumentInfo represents information about a processed document
type DocumentInfo struct {
	Path           string                 `json:"path"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`
	Classification DocumentClassification `json:"classification"`
	Confidence     float64                `json:"confidence"`
	ExtractedText  string                 `json:"extractedText,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
	ProcessedAt    time.Time              `json:"processedAt"`
	Size           int64                  `json:"size"`
	Hash           string                 `json:"hash"`
	// Additional fields for compatibility
	Extension    string   `json:"extension,omitempty"`
	MimeType     string   `json:"mimeType,omitempty"`
	IsScanned    bool     `json:"isScanned,omitempty"`
	Keywords     []string `json:"keywords,omitempty"`
	ErrorMessage string   `json:"errorMessage,omitempty"`
}

// DocumentClassification represents the classification of a document
type DocumentClassification struct {
	Category    string   `json:"category"`
	SubCategory string   `json:"subCategory,omitempty"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords"`
	Topics      []string `json:"topics"`
}

// DocumentType represents the type of document
type DocumentType string

const (
	DocumentTypePDF        DocumentType = "pdf"
	DocumentTypeWord       DocumentType = "docx"
	DocumentTypeExcel      DocumentType = "xlsx"
	DocumentTypePowerPoint DocumentType = "pptx"
	DocumentTypeText       DocumentType = "txt"
	DocumentTypeCSV        DocumentType = "csv"
	DocumentTypeImage      DocumentType = "image"
	DocumentTypeUnknown    DocumentType = "unknown"
)

// DocumentCategory represents the category of a document
type DocumentCategory string

const (
	DocumentCategoryFinancial DocumentCategory = "financial"
	DocumentCategoryLegal     DocumentCategory = "legal"
	DocumentCategoryGeneral   DocumentCategory = "general"
	DocumentCategoryUnknown   DocumentCategory = "unknown"
)

// Legacy constants for backward compatibility
const (
	DocTypeLegal     DocumentType = "legal"
	DocTypeFinancial DocumentType = "financial"
	DocTypeGeneral   DocumentType = "general"
	DocTypeUnknown   DocumentType = "unknown"
)

// RoutingResult represents the result of document routing
type RoutingResult struct {
	Success         bool                   `json:"success"`
	SourcePath      string                 `json:"sourcePath"`
	DestinationPath string                 `json:"destinationPath"`
	Category        DocumentCategory       `json:"category"`
	Classification  DocumentClassification `json:"classification"`
	Error           string                 `json:"error,omitempty"`
	ProcessingTime  time.Duration          `json:"processingTime"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	// Additional fields for compatibility
	DocumentType     DocumentType `json:"documentType,omitempty"`
	AlreadyProcessed bool         `json:"alreadyProcessed,omitempty"`
}

// DocumentInsights represents AI-generated insights about a document
type DocumentInsights struct {
	Summary       string                 `json:"summary"`
	KeyPoints     []string               `json:"keyPoints"`
	ActionItems   []string               `json:"actionItems"`
	Risks         []string               `json:"risks"`
	Opportunities []string               `json:"opportunities"`
	RelatedTopics []string               `json:"relatedTopics"`
	Sentiment     string                 `json:"sentiment"`
	Importance    string                 `json:"importance"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// EntityExtraction represents extracted entities from a document
type EntityExtraction struct {
	People         []Entity            `json:"people"`
	Organizations  []Entity            `json:"organizations"`
	Locations      []Entity            `json:"locations"`
	Dates          []DateEntity        `json:"dates"`
	MonetaryValues []MonetaryEntity    `json:"monetaryValues"`
	Percentages    []PercentageEntity  `json:"percentages"`
	CustomEntities map[string][]Entity `json:"customEntities"`
}

// Entity represents a basic extracted entity
type Entity struct {
	Text       string  `json:"text"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
	Context    string  `json:"context,omitempty"`
}

// DateEntity represents an extracted date
type DateEntity struct {
	Entity
	ParsedDate *time.Time `json:"parsedDate,omitempty"`
	Format     string     `json:"format,omitempty"`
}

// MonetaryEntity represents an extracted monetary value
type MonetaryEntity struct {
	Entity
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

// PercentageEntity represents an extracted percentage
type PercentageEntity struct {
	Entity
	Value float64 `json:"value"`
}

// FinancialAnalysis represents financial data extracted from documents
type FinancialAnalysis struct {
	Revenue          *FinancialMetric            `json:"revenue,omitempty"`
	EBITDA           *FinancialMetric            `json:"ebitda,omitempty"`
	NetIncome        *FinancialMetric            `json:"netIncome,omitempty"`
	TotalAssets      *FinancialMetric            `json:"totalAssets,omitempty"`
	TotalLiabilities *FinancialMetric            `json:"totalLiabilities,omitempty"`
	CashFlow         *FinancialMetric            `json:"cashFlow,omitempty"`
	CustomMetrics    map[string]*FinancialMetric `json:"customMetrics,omitempty"`
	Period           string                      `json:"period"`
	Currency         string                      `json:"currency"`
	Confidence       float64                     `json:"confidence"`
}

// FinancialMetric represents a single financial metric
type FinancialMetric struct {
	Value      float64 `json:"value"`
	Currency   string  `json:"currency"`
	Period     string  `json:"period,omitempty"`
	Growth     float64 `json:"growth,omitempty"`
	Confidence float64 `json:"confidence"`
}

// RiskAnalysis represents risk analysis of a document or deal
type RiskAnalysis struct {
	OverallRiskLevel string       `json:"overallRiskLevel"`
	RiskScore        float64      `json:"riskScore"`
	RiskFactors      []RiskFactor `json:"riskFactors"`
	Mitigations      []Mitigation `json:"mitigations"`
	Recommendations  []string     `json:"recommendations"`
	Confidence       float64      `json:"confidence"`
}

// RiskFactor represents a single risk factor
type RiskFactor struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Severity    string  `json:"severity"`
	Likelihood  string  `json:"likelihood"`
	Impact      string  `json:"impact"`
	Score       float64 `json:"score"`
}

// Mitigation represents a risk mitigation strategy
type Mitigation struct {
	RiskFactor    string `json:"riskFactor"`
	Strategy      string `json:"strategy"`
	Description   string `json:"description"`
	Effort        string `json:"effort"`
	Effectiveness string `json:"effectiveness"`
}
