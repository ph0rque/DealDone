package templates

import (
	"time"
)

// TemplateResult represents the result of template processing
// This type is used in webhook payloads to communicate template processing results
type TemplateResult struct {
	TemplatePath       string                 `json:"templatePath" validate:"required"`
	TemplateName       string                 `json:"templateName" validate:"required"`
	Status             string                 `json:"status" validate:"required,oneof=updated unchanged failed"`
	FieldsUpdated      []string               `json:"fieldsUpdated"`
	FieldsConflicted   []string               `json:"fieldsConflicted"`
	AverageConfidence  float64                `json:"averageConfidence" validate:"min=0,max=1"`
	BackupCreated      bool                   `json:"backupCreated"`
	BackupPath         string                 `json:"backupPath,omitempty"`
	UpdatedFields      map[string]interface{} `json:"updatedFields,omitempty"`
	PreviousValues     map[string]interface{} `json:"previousValues,omitempty"`
	ConflictResolution map[string]string      `json:"conflictResolution,omitempty"`
}

// TemplateAnalysisResult represents the result of analyzing and populating templates
type TemplateAnalysisResult struct {
	DealName           string        `json:"dealName"`
	ProcessedDocuments []string      `json:"processedDocuments"`
	CopiedTemplates    []string      `json:"copiedTemplates"`
	PopulatedTemplates []string      `json:"populatedTemplates"`
	Errors             []string      `json:"errors"`
	Success            bool          `json:"success"`
	StartTime          time.Time     `json:"startTime"`
	EndTime            time.Time     `json:"endTime"`
	ProcessingTime     time.Duration `json:"processingTime"`
}

// TemplateType represents the type of template file
type TemplateType string

const (
	TemplateTypeExcel      TemplateType = "excel"
	TemplateTypeCSV        TemplateType = "csv"
	TemplateTypeWord       TemplateType = "word"
	TemplateTypePowerPoint TemplateType = "powerpoint"
	TemplateTypePDF        TemplateType = "pdf"
	TemplateTypeText       TemplateType = "text"
	TemplateTypeMarkdown   TemplateType = "markdown"
	TemplateTypeUnknown    TemplateType = "unknown"
)

// FieldType represents the data type of a template field
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeNumeric  FieldType = "numeric"
	FieldTypeCurrency FieldType = "currency"
	FieldTypeDate     FieldType = "date"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeFormula  FieldType = "formula"
	FieldTypeUnknown  FieldType = "unknown"
)

// TemplateCategory represents the category of a template
type TemplateCategory string

const (
	TemplateCategoryFinancial   TemplateCategory = "financial"
	TemplateCategoryLegal       TemplateCategory = "legal"
	TemplateCategoryOperational TemplateCategory = "operational"
	TemplateCategoryGeneral     TemplateCategory = "general"
	TemplateCategoryCustom      TemplateCategory = "custom"
)

// FieldValidator represents a validation rule for a template field
type FieldValidator struct {
	Type       string                 `json:"type"`       // "required", "range", "pattern", "custom"
	Parameters map[string]interface{} `json:"parameters"` // Validation parameters
	Message    string                 `json:"message"`    // Error message if validation fails
}

// MappedData represents data mapped to template fields
type MappedData struct {
	TemplateID   string                            `json:"templateId"`
	TemplatePath string                            `json:"templatePath"`
	DealName     string                            `json:"dealName"`
	Fields       map[string]MappedField            `json:"fields"`
	Sheets       map[string]map[string]interface{} `json:"sheets,omitempty"`
	Confidence   float64                           `json:"confidence"`
	Sources      []string                          `json:"sources"`
	Timestamp    time.Time                         `json:"timestamp"`
	Warnings     []string                          `json:"warnings"`
}

// MappedField represents data mapped to a specific field
type MappedField struct {
	FieldName    string      `json:"fieldName"`
	Value        interface{} `json:"value"`
	Source       string      `json:"source"`
	SourceType   string      `json:"sourceType"` // "ai", "ocr", "extracted", "calculated"
	Confidence   float64     `json:"confidence"`
	OriginalText string      `json:"originalText,omitempty"`
}

// MatchingResult represents the result of field matching
type MatchingResult struct {
	TemplateFields    []string              `json:"templateFields"`
	SourceFields      []string              `json:"sourceFields"`
	Matches           map[string]FieldMatch `json:"matches"`
	UnmatchedTemplate []string              `json:"unmatchedTemplate"`
	UnmatchedSource   []string              `json:"unmatchedSource"`
	OverallConfidence float64               `json:"overallConfidence"`
}

// FieldMatch represents a match between source and template fields
type FieldMatch struct {
	SourceField    string  `json:"sourceField"`
	TemplateField  string  `json:"templateField"`
	Confidence     float64 `json:"confidence"`
	MatchType      string  `json:"matchType"` // "exact", "semantic", "pattern", "fuzzy"
	Transformation string  `json:"transformation,omitempty"`
}

// FormattingContext provides context for value formatting
type FormattingContext struct {
	FieldName    string                 `json:"fieldName"`
	FieldType    string                 `json:"fieldType"`
	TemplateType string                 `json:"templateType"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// FieldCorrection represents a correction made to a field value
type FieldCorrection struct {
	FieldName      string      `json:"fieldName"`
	OriginalValue  interface{} `json:"originalValue"`
	CorrectedValue interface{} `json:"correctedValue"`
	Reason         string      `json:"reason"`
	Timestamp      time.Time   `json:"timestamp"`
	UserID         string      `json:"userId"`
}
