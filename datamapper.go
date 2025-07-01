package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DataMapper handles mapping extracted data to template fields
type DataMapper struct {
	aiService      *AIService
	templateParser *TemplateParser
}

// NewDataMapper creates a new data mapper
func NewDataMapper(aiService *AIService, templateParser *TemplateParser) *DataMapper {
	return &DataMapper{
		aiService:      aiService,
		templateParser: templateParser,
	}
}

// MappedData represents data mapped to template fields
type MappedData struct {
	TemplateID  string                            `json:"templateId"`
	DealName    string                            `json:"dealName"`
	SourceFiles []string                          `json:"sourceFiles"`
	Fields      map[string]MappedField            `json:"fields"`
	Sheets      map[string]map[string]interface{} `json:"sheets,omitempty"`
	Confidence  float64                           `json:"confidence"`
	MappingDate time.Time                         `json:"mappingDate"`
	Warnings    []string                          `json:"warnings"`
}

// MappedField represents a single mapped field
type MappedField struct {
	FieldName    string      `json:"fieldName"`
	Value        interface{} `json:"value"`
	Source       string      `json:"source"`
	SourceType   string      `json:"sourceType"` // "ai", "ocr", "extracted", "calculated"
	Confidence   float64     `json:"confidence"`
	OriginalText string      `json:"originalText,omitempty"`
}

// ExtractAndMapData extracts data from documents and maps to template
func (dm *DataMapper) ExtractAndMapData(templateData *TemplateData, documents []DocumentInfo, dealName string) (*MappedData, error) {
	mappedData := &MappedData{
		TemplateID:  templateData.Metadata["fileName"].(string),
		DealName:    dealName,
		SourceFiles: make([]string, 0),
		Fields:      make(map[string]MappedField),
		MappingDate: time.Now(),
		Warnings:    make([]string, 0),
	}

	// Extract source file names
	for _, doc := range documents {
		mappedData.SourceFiles = append(mappedData.SourceFiles, doc.Name)
	}

	// Get template fields
	fields := dm.templateParser.ExtractDataFields(templateData)

	// Create a data extraction context
	extractionContext := dm.createExtractionContext(documents)

	// Map each field
	totalConfidence := 0.0
	mappedCount := 0

	for _, field := range fields {
		mappedField, err := dm.mapField(field, extractionContext)
		if err != nil {
			mappedData.Warnings = append(mappedData.Warnings, fmt.Sprintf("Failed to map field %s: %v", field.Name, err))
			continue
		}

		if mappedField != nil {
			mappedData.Fields[field.Path] = *mappedField
			totalConfidence += mappedField.Confidence
			mappedCount++
		}
	}

	// Calculate overall confidence
	if mappedCount > 0 {
		mappedData.Confidence = totalConfidence / float64(mappedCount)
	}

	// Handle multi-sheet Excel templates
	if len(templateData.Sheets) > 0 {
		mappedData.Sheets = dm.organizeBySheets(mappedData.Fields, fields)
	}

	return mappedData, nil
}

// ExtractionContext holds all available data for extraction
type ExtractionContext struct {
	AIAnalysis      map[string]interface{}
	ExtractedText   map[string]string
	FinancialData   *FinancialAnalysis
	Entities        *EntityExtraction
	DocumentsByType map[string][]DocumentInfo
}

// createExtractionContext aggregates all available data sources
func (dm *DataMapper) createExtractionContext(documents []DocumentInfo) *ExtractionContext {
	context := &ExtractionContext{
		AIAnalysis:      make(map[string]interface{}),
		ExtractedText:   make(map[string]string),
		DocumentsByType: make(map[string][]DocumentInfo),
	}

	// Group documents by type
	for _, doc := range documents {
		docType := string(doc.Type)
		context.DocumentsByType[docType] = append(context.DocumentsByType[docType], doc)
	}

	// Aggregate AI analysis results
	// In a real implementation, this would collect actual AI analysis results
	// For now, we'll use placeholder data
	context.FinancialData = &FinancialAnalysis{
		Revenue:    1000000,
		EBITDA:     200000,
		NetIncome:  150000,
		Currency:   "USD",
		Confidence: 0.85,
	}

	context.Entities = &EntityExtraction{
		Organizations: []Entity{
			{Text: "Example Corp", Type: "organization", Confidence: 0.9},
			{Text: "Target Company Inc", Type: "organization", Confidence: 0.95},
		},
		MonetaryValues: []Entity{
			{Text: "$1,000,000", Type: "monetary", Confidence: 0.9},
			{Text: "$200,000", Type: "monetary", Confidence: 0.85},
		},
		Dates: []Entity{
			{Text: "January 1, 2024", Type: "date", Confidence: 0.95},
			{Text: "December 31, 2024", Type: "date", Confidence: 0.95},
		},
	}

	return context
}

// mapField maps a single field using available data sources
func (dm *DataMapper) mapField(field DataField, context *ExtractionContext) (*MappedField, error) {
	// Try different mapping strategies based on field type and name

	// Strategy 1: Direct financial data mapping
	if field.DataType == "number" || field.DataType == "currency" {
		if value, confidence := dm.mapFinancialField(field.Name, context.FinancialData); value != nil {
			return &MappedField{
				FieldName:  field.Name,
				Value:      value,
				Source:     "financial_analysis",
				SourceType: "ai",
				Confidence: confidence,
			}, nil
		}
	}

	// Strategy 2: Entity extraction mapping
	if entity := dm.findMatchingEntity(field.Name, context.Entities); entity != nil {
		return &MappedField{
			FieldName:    field.Name,
			Value:        entity.Text,
			Source:       "entity_extraction",
			SourceType:   "ai",
			Confidence:   entity.Confidence,
			OriginalText: entity.Text,
		}, nil
	}

	// Strategy 3: Pattern-based extraction
	if value, confidence := dm.extractByPattern(field, context.ExtractedText); value != nil {
		return &MappedField{
			FieldName:  field.Name,
			Value:      value,
			Source:     "pattern_extraction",
			SourceType: "extracted",
			Confidence: confidence,
		}, nil
	}

	// Strategy 4: Default values for required fields
	if field.IsRequired {
		defaultValue := dm.getDefaultValue(field)
		return &MappedField{
			FieldName:  field.Name,
			Value:      defaultValue,
			Source:     "default",
			SourceType: "calculated",
			Confidence: 0.5,
		}, nil
	}

	return nil, nil
}

// mapFinancialField maps financial data to field
func (dm *DataMapper) mapFinancialField(fieldName string, financial *FinancialAnalysis) (interface{}, float64) {
	if financial == nil {
		return nil, 0
	}

	fieldLower := strings.ToLower(fieldName)

	// Direct field mapping
	fieldMap := map[string]interface{}{
		"revenue":           financial.Revenue,
		"total revenue":     financial.Revenue,
		"ebitda":            financial.EBITDA,
		"net income":        financial.NetIncome,
		"total assets":      financial.TotalAssets,
		"total liabilities": financial.TotalLiabilities,
		"cash flow":         financial.CashFlow,
		"gross margin":      financial.GrossMargin,
		"operating margin":  financial.OperatingMargin,
	}

	for key, value := range fieldMap {
		if strings.Contains(fieldLower, key) {
			return value, financial.Confidence
		}
	}

	// Check additional data points
	if financial.DataPoints != nil {
		for key, value := range financial.DataPoints {
			if strings.Contains(fieldLower, strings.ToLower(key)) {
				return value, financial.Confidence * 0.9
			}
		}
	}

	return nil, 0
}

// findMatchingEntity finds an entity that matches the field
func (dm *DataMapper) findMatchingEntity(fieldName string, entities *EntityExtraction) *Entity {
	if entities == nil {
		return nil
	}

	fieldLower := strings.ToLower(fieldName)

	// Check for company/organization fields
	if strings.Contains(fieldLower, "company") || strings.Contains(fieldLower, "organization") {
		if len(entities.Organizations) > 0 {
			// Return the highest confidence organization
			best := entities.Organizations[0]
			for _, org := range entities.Organizations {
				if org.Confidence > best.Confidence {
					best = org
				}
			}
			return &best
		}
	}

	// Check for date fields
	if strings.Contains(fieldLower, "date") {
		if len(entities.Dates) > 0 {
			return &entities.Dates[0]
		}
	}

	// Check for monetary fields
	if strings.Contains(fieldLower, "amount") || strings.Contains(fieldLower, "price") || strings.Contains(fieldLower, "value") {
		if len(entities.MonetaryValues) > 0 {
			return &entities.MonetaryValues[0]
		}
	}

	return nil
}

// extractByPattern uses regex patterns to extract field values
func (dm *DataMapper) extractByPattern(field DataField, textContent map[string]string) (interface{}, float64) {
	patterns := dm.getPatternsForField(field)

	for source, text := range textContent {
		for _, pattern := range patterns {
			if matches := pattern.FindStringSubmatch(text); len(matches) > 1 {
				value := dm.parseValue(matches[1], field.DataType)
				if value != nil {
					// Higher confidence if source is more relevant
					confidence := 0.7
					if strings.Contains(source, "financial") {
						confidence = 0.8
					}
					return value, confidence
				}
			}
		}
	}

	return nil, 0
}

// getPatternsForField returns regex patterns for field extraction
func (dm *DataMapper) getPatternsForField(field DataField) []*regexp.Regexp {
	fieldLower := strings.ToLower(field.Name)
	patterns := make([]*regexp.Regexp, 0)

	// Revenue patterns
	if strings.Contains(fieldLower, "revenue") {
		patterns = append(patterns,
			regexp.MustCompile(`(?i)revenue[:\s]+\$?([\d,]+\.?\d*)`),
			regexp.MustCompile(`(?i)total revenue[:\s]+\$?([\d,]+\.?\d*)`),
		)
	}

	// Date patterns
	if strings.Contains(fieldLower, "date") {
		patterns = append(patterns,
			regexp.MustCompile(`(?i)date[:\s]+(\d{1,2}/\d{1,2}/\d{4})`),
			regexp.MustCompile(`(?i)date[:\s]+(\w+\s+\d{1,2},\s+\d{4})`),
		)
	}

	// Company name patterns
	if strings.Contains(fieldLower, "company") {
		patterns = append(patterns,
			regexp.MustCompile(`(?i)company name[:\s]+([A-Za-z0-9\s&.,]+)`),
			regexp.MustCompile(`(?i)target[:\s]+([A-Za-z0-9\s&.,]+)`),
		)
	}

	// Generic amount pattern
	if field.DataType == "number" || field.DataType == "currency" {
		patterns = append(patterns,
			regexp.MustCompile(fmt.Sprintf(`(?i)%s[:\s]+\$?([\d,]+\.?\d*)`, regexp.QuoteMeta(field.Name))),
		)
	}

	return patterns
}

// parseValue converts string value to appropriate type
func (dm *DataMapper) parseValue(value string, dataType string) interface{} {
	value = strings.TrimSpace(value)

	switch dataType {
	case "number", "currency":
		// Remove currency symbols and commas
		cleaned := strings.ReplaceAll(value, "$", "")
		cleaned = strings.ReplaceAll(cleaned, ",", "")
		if num, err := strconv.ParseFloat(cleaned, 64); err == nil {
			return num
		}
	case "date":
		// Try parsing common date formats
		formats := []string{
			"1/2/2006",
			"01/02/2006",
			"January 2, 2006",
			"Jan 2, 2006",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, value); err == nil {
				return t.Format("2006-01-02")
			}
		}
	}

	// Default to string
	return value
}

// getDefaultValue returns a default value for required fields
func (dm *DataMapper) getDefaultValue(field DataField) interface{} {
	switch field.DataType {
	case "number", "currency":
		return 0.0
	case "date":
		return time.Now().Format("2006-01-02")
	default:
		return "[To be filled]"
	}
}

// organizeBySheets groups mapped fields by their sheet
func (dm *DataMapper) organizeBySheets(fields map[string]MappedField, fieldDefs []DataField) map[string]map[string]interface{} {
	sheets := make(map[string]map[string]interface{})

	// Create a map of field paths to sheet names
	fieldToSheet := make(map[string]string)
	for _, field := range fieldDefs {
		if field.Sheet != "" {
			fieldToSheet[field.Path] = field.Sheet
		}
	}

	// Group fields by sheet
	for path, mappedField := range fields {
		if sheet, ok := fieldToSheet[path]; ok {
			if sheets[sheet] == nil {
				sheets[sheet] = make(map[string]interface{})
			}
			// Extract field name from path (remove sheet prefix)
			fieldName := strings.TrimPrefix(path, sheet+".")
			sheets[sheet][fieldName] = mappedField.Value
		}
	}

	return sheets
}

// ValidateMappedData validates the mapped data against template requirements
func (dm *DataMapper) ValidateMappedData(mappedData *MappedData, templateData *TemplateData) error {
	fields := dm.templateParser.ExtractDataFields(templateData)

	// Check required fields
	missingRequired := []string{}
	for _, field := range fields {
		if field.IsRequired {
			if _, exists := mappedData.Fields[field.Path]; !exists {
				missingRequired = append(missingRequired, field.Name)
			}
		}
	}

	if len(missingRequired) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missingRequired, ", "))
	}

	// Validate data types
	for path, mappedField := range mappedData.Fields {
		// Find field definition
		var fieldDef *DataField
		for _, f := range fields {
			if f.Path == path {
				fieldDef = &f
				break
			}
		}

		if fieldDef != nil && !dm.isValidType(mappedField.Value, fieldDef.DataType) {
			return fmt.Errorf("invalid type for field %s: expected %s", fieldDef.Name, fieldDef.DataType)
		}
	}

	return nil
}

// isValidType checks if a value matches the expected data type
func (dm *DataMapper) isValidType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "number", "currency":
		switch v := value.(type) {
		case float64, float32, int, int32, int64:
			return true
		case string:
			_, err := strconv.ParseFloat(v, 64)
			return err == nil
		}
	case "date":
		switch v := value.(type) {
		case time.Time:
			return true
		case string:
			_, err := time.Parse("2006-01-02", v)
			return err == nil
		}
	case "string":
		return true // Everything can be a string
	}
	return false
}

// ExportMappedData exports mapped data to JSON
func (dm *DataMapper) ExportMappedData(mappedData *MappedData) ([]byte, error) {
	return json.MarshalIndent(mappedData, "", "  ")
}
