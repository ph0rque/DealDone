package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataMapper(t *testing.T) {
	aiService := &AIService{}
	templateParser := &TemplateParser{}

	mapper := NewDataMapper(aiService, templateParser)
	assert.NotNil(t, mapper)
	assert.Equal(t, aiService, mapper.aiService)
	assert.Equal(t, templateParser, mapper.templateParser)
}

func TestMapFinancialField(t *testing.T) {
	mapper := NewDataMapper(nil, nil)

	financial := &FinancialAnalysis{
		Revenue:         1000000,
		EBITDA:          200000,
		NetIncome:       150000,
		TotalAssets:     5000000,
		GrossMargin:     0.35,
		OperatingMargin: 0.20,
		Confidence:      0.85,
		DataPoints: map[string]float64{
			"Working Capital": 500000,
			"Debt Ratio":      0.45,
		},
	}

	tests := []struct {
		fieldName     string
		expectedValue interface{}
		expectedConf  float64
		shouldFind    bool
	}{
		{"Revenue", 1000000.0, 0.85, true},
		{"Total Revenue", 1000000.0, 0.85, true},
		{"EBITDA", 200000.0, 0.85, true},
		{"Net Income", 150000.0, 0.85, true},
		{"Total Assets", 5000000.0, 0.85, true},
		{"Gross Margin", 0.35, 0.85, true},
		{"Operating Margin", 0.20, 0.85, true},
		{"Working Capital", 500000.0, 0.765, true}, // 0.85 * 0.9
		{"Unknown Field", nil, 0, false},
	}

	for _, tt := range tests {
		value, confidence := mapper.mapFinancialField(tt.fieldName, financial)
		if tt.shouldFind {
			assert.Equal(t, tt.expectedValue, value, "Field: %s", tt.fieldName)
			assert.InDelta(t, tt.expectedConf, confidence, 0.001, "Field: %s", tt.fieldName)
		} else {
			assert.Nil(t, value, "Field: %s", tt.fieldName)
			assert.Equal(t, float64(0), confidence, "Field: %s", tt.fieldName)
		}
	}
}

func TestFindMatchingEntity(t *testing.T) {
	mapper := NewDataMapper(nil, nil)

	entities := &EntityExtraction{
		Organizations: []Entity{
			{Text: "Acme Corp", Type: "organization", Confidence: 0.9},
			{Text: "Target Inc", Type: "organization", Confidence: 0.95},
		},
		Dates: []Entity{
			{Text: "January 1, 2024", Type: "date", Confidence: 0.95},
		},
		MonetaryValues: []Entity{
			{Text: "$1,000,000", Type: "monetary", Confidence: 0.9},
		},
	}

	// Test company field
	entity := mapper.findMatchingEntity("Company Name", entities)
	assert.NotNil(t, entity)
	assert.Equal(t, "Target Inc", entity.Text) // Higher confidence

	// Test date field
	entity = mapper.findMatchingEntity("Deal Date", entities)
	assert.NotNil(t, entity)
	assert.Equal(t, "January 1, 2024", entity.Text)

	// Test amount field
	entity = mapper.findMatchingEntity("Purchase Price", entities)
	assert.NotNil(t, entity)
	assert.Equal(t, "$1,000,000", entity.Text)

	// Test unmatched field
	entity = mapper.findMatchingEntity("Status", entities)
	assert.Nil(t, entity)
}

func TestParseValue(t *testing.T) {
	mapper := NewDataMapper(nil, nil)

	tests := []struct {
		value    string
		dataType string
		expected interface{}
	}{
		{"1000000", "number", 1000000.0},
		{"$1,000,000", "currency", 1000000.0},
		{"1,234.56", "number", 1234.56},
		{"January 1, 2024", "date", "2024-01-01"},
		{"1/15/2024", "date", "2024-01-15"},
		{"2024-01-15", "date", "2024-01-15"},
		{"Some text", "string", "Some text"},
		{"invalid", "number", "invalid"}, // Falls back to string
	}

	for _, tt := range tests {
		result := mapper.parseValue(tt.value, tt.dataType)
		assert.Equal(t, tt.expected, result, "Value: %s, Type: %s", tt.value, tt.dataType)
	}
}

func TestGetDefaultValue(t *testing.T) {
	mapper := NewDataMapper(nil, nil)

	numberField := DataField{Name: "Amount", DataType: "number"}
	assert.Equal(t, 0.0, mapper.getDefaultValue(numberField))

	currencyField := DataField{Name: "Price", DataType: "currency"}
	assert.Equal(t, 0.0, mapper.getDefaultValue(currencyField))

	dateField := DataField{Name: "Date", DataType: "date"}
	defaultDate := mapper.getDefaultValue(dateField).(string)
	assert.Equal(t, time.Now().Format("2006-01-02"), defaultDate)

	stringField := DataField{Name: "Name", DataType: "string"}
	assert.Equal(t, "[To be filled]", mapper.getDefaultValue(stringField))
}

func TestExtractAndMapData(t *testing.T) {
	// Create a mock template parser
	templateParser := &TemplateParser{}
	mapper := NewDataMapper(nil, templateParser)

	// Create test template data
	templateData := &TemplateData{
		Format:  "csv",
		Headers: []string{"Company Name", "Revenue", "Date"},
		Data:    [][]string{},
		Metadata: map[string]interface{}{
			"fileName": "test_template.csv",
		},
	}

	// Create test documents
	documents := []DocumentInfo{
		{
			Name: "financial_report.pdf",
			Type: DocTypeFinancial,
		},
		{
			Name: "contract.pdf",
			Type: DocTypeLegal,
		},
	}

	// Test mapping
	mappedData, err := mapper.ExtractAndMapData(templateData, documents, "Test Deal")

	require.NoError(t, err)
	assert.Equal(t, "test_template.csv", mappedData.TemplateID)
	assert.Equal(t, "Test Deal", mappedData.DealName)
	assert.Len(t, mappedData.SourceFiles, 2)
	assert.Contains(t, mappedData.SourceFiles, "financial_report.pdf")
	assert.Contains(t, mappedData.SourceFiles, "contract.pdf")
	assert.NotZero(t, mappedData.MappingDate)
}

func TestValidateMappedData(t *testing.T) {
	templateParser := &TemplateParser{}
	mapper := NewDataMapper(nil, templateParser)

	// Create mapped data
	mappedData := &MappedData{
		Fields: map[string]MappedField{
			"Company Name": {
				FieldName: "Company Name",
				Value:     "Test Corp",
			},
			"Revenue": {
				FieldName: "Revenue",
				Value:     1000000.0,
			},
		},
	}

	// Create template with required fields
	templateData := &TemplateData{
		Headers: []string{"Company Name", "Revenue", "Date"},
	}

	// Test validation - should fail due to missing required Date field
	err := mapper.ValidateMappedData(mappedData, templateData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required fields")
}

func TestIsValidType(t *testing.T) {
	mapper := NewDataMapper(nil, nil)

	// Test number types
	assert.True(t, mapper.isValidType(42.0, "number"))
	assert.True(t, mapper.isValidType(42, "number"))
	assert.True(t, mapper.isValidType("42.5", "number"))
	assert.False(t, mapper.isValidType("not a number", "number"))

	// Test currency types
	assert.True(t, mapper.isValidType(1000.0, "currency"))
	assert.True(t, mapper.isValidType("1000.50", "currency"))

	// Test date types
	assert.True(t, mapper.isValidType(time.Now(), "date"))
	assert.True(t, mapper.isValidType("2024-01-01", "date"))
	assert.False(t, mapper.isValidType("invalid date", "date"))

	// Test string types
	assert.True(t, mapper.isValidType("anything", "string"))
	assert.True(t, mapper.isValidType(42, "string"))
	assert.True(t, mapper.isValidType(true, "string"))
}

func TestOrganizeBySheets(t *testing.T) {
	mapper := NewDataMapper(nil, nil)

	fields := map[string]MappedField{
		"Revenue.Month":     {Value: "January"},
		"Revenue.Amount":    {Value: 100000.0},
		"Expenses.Category": {Value: "Marketing"},
		"Expenses.Cost":     {Value: 50000.0},
	}

	fieldDefs := []DataField{
		{Path: "Revenue.Month", Sheet: "Revenue"},
		{Path: "Revenue.Amount", Sheet: "Revenue"},
		{Path: "Expenses.Category", Sheet: "Expenses"},
		{Path: "Expenses.Cost", Sheet: "Expenses"},
	}

	sheets := mapper.organizeBySheets(fields, fieldDefs)

	assert.Len(t, sheets, 2)
	assert.Contains(t, sheets, "Revenue")
	assert.Contains(t, sheets, "Expenses")

	assert.Equal(t, "January", sheets["Revenue"]["Month"])
	assert.Equal(t, 100000.0, sheets["Revenue"]["Amount"])
	assert.Equal(t, "Marketing", sheets["Expenses"]["Category"])
	assert.Equal(t, 50000.0, sheets["Expenses"]["Cost"])
}

func TestExportMappedData(t *testing.T) {
	mapper := NewDataMapper(nil, nil)

	mappedData := &MappedData{
		TemplateID: "test.csv",
		DealName:   "Test Deal",
		Fields: map[string]MappedField{
			"Revenue": {
				FieldName:  "Revenue",
				Value:      1000000.0,
				Source:     "financial_analysis",
				SourceType: "ai",
				Confidence: 0.85,
			},
		},
		Confidence: 0.85,
	}

	exported, err := mapper.ExportMappedData(mappedData)

	require.NoError(t, err)
	assert.Contains(t, string(exported), "\"templateId\": \"test.csv\"")
	assert.Contains(t, string(exported), "\"dealName\": \"Test Deal\"")
	assert.Contains(t, string(exported), "\"confidence\": 0.85")
}
