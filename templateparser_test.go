package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateParser(t *testing.T) {
	parser := NewTemplateParser("/test/path")
	assert.NotNil(t, parser)
	assert.Equal(t, "/test/path", parser.templatesPath)
}

func TestParseCSVTemplate(t *testing.T) {
	// Create a temporary CSV file
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "test_template.csv")

	csvContent := `Company Name,Revenue,EBITDA,Net Income
"Test Corp",1000000,200000,150000
"Example Inc",2000000,400000,300000
"Formula Test",=B2+B3,=C2+C3,=D2+D3`

	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	require.NoError(t, err)

	parser := NewTemplateParser(tempDir)
	data, err := parser.ParseTemplate(csvPath)

	require.NoError(t, err)
	assert.Equal(t, "csv", data.Format)
	assert.Equal(t, []string{"Company Name", "Revenue", "EBITDA", "Net Income"}, data.Headers)
	assert.Len(t, data.Data, 3)

	// Check formula detection
	assert.Contains(t, data.Formulas, "B4")
	assert.Equal(t, "=B2+B3", data.Formulas["B4"])
	assert.Contains(t, data.Formulas, "C4")
	assert.Equal(t, "=C2+C3", data.Formulas["C4"])
}

func TestExtractDataFields(t *testing.T) {
	parser := NewTemplateParser("")

	templateData := &TemplateData{
		Format:  "csv",
		Headers: []string{"Company Name", "Revenue", "Date", "Status"},
		Data: [][]string{
			{"Test Corp", "1000000", "2024-01-01", "Active"},
			{"Example Inc", "2000000", "2024-01-02", "Pending"},
		},
	}

	fields := parser.ExtractDataFields(templateData)

	assert.Len(t, fields, 4)

	// Check field properties
	companyField := fields[0]
	assert.Equal(t, "Company Name", companyField.Name)
	assert.Equal(t, "string", companyField.DataType)
	assert.True(t, companyField.IsRequired)

	revenueField := fields[1]
	assert.Equal(t, "Revenue", revenueField.Name)
	assert.Equal(t, "number", revenueField.DataType)
	assert.True(t, revenueField.IsRequired)

	statusField := fields[3]
	assert.Equal(t, "Status", statusField.Name)
	assert.Equal(t, "string", statusField.DataType)
	assert.False(t, statusField.IsRequired)
}

func TestInferDataType(t *testing.T) {
	parser := NewTemplateParser("")

	tests := []struct {
		name     string
		data     [][]string
		colIndex int
		expected string
	}{
		{
			name: "Currency detection",
			data: [][]string{
				{"$1,000"},
				{"$2,500"},
				{"$3,000"},
			},
			colIndex: 0,
			expected: "currency",
		},
		{
			name: "Number detection",
			data: [][]string{
				{"1000"},
				{"2500.50"},
				{"3000"},
			},
			colIndex: 0,
			expected: "number",
		},
		{
			name: "String detection",
			data: [][]string{
				{"Active"},
				{"Pending"},
				{"Completed"},
			},
			colIndex: 0,
			expected: "string",
		},
		{
			name: "Mixed with mostly numbers",
			data: [][]string{
				{"1000"},
				{"2000"},
				{""},
			},
			colIndex: 0,
			expected: "number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.inferDataType(tt.data, tt.colIndex)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTemplateStructure(t *testing.T) {
	parser := NewTemplateParser("")

	// Test empty template
	emptyTemplate := &TemplateData{
		Format:  "csv",
		Headers: []string{},
		Data:    [][]string{},
	}
	err := parser.ValidateTemplateStructure(emptyTemplate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no headers")

	// Test valid financial template
	financialTemplate := &TemplateData{
		Format:  "csv",
		Headers: []string{"Company", "Revenue", "EBITDA"},
		Data: [][]string{
			{"Test Corp", "1000000", "200000"},
		},
	}
	err = parser.ValidateTemplateStructure(financialTemplate)
	assert.NoError(t, err)

	// Test financial template without numeric fields
	invalidFinancial := &TemplateData{
		Format:  "csv",
		Headers: []string{"Revenue Analysis", "Status"},
		Data: [][]string{
			{"Good", "Active"},
		},
	}
	err = parser.ValidateTemplateStructure(invalidFinancial)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required fields")
}

func TestIsFinancialTemplate(t *testing.T) {
	parser := NewTemplateParser("")

	// Test with financial headers
	financialTemplate := &TemplateData{
		Headers: []string{"Company", "Revenue", "EBITDA"},
	}
	assert.True(t, parser.isFinancialTemplate(financialTemplate))

	// Test with non-financial headers
	nonFinancialTemplate := &TemplateData{
		Headers: []string{"Name", "Status", "Date"},
	}
	assert.False(t, parser.isFinancialTemplate(nonFinancialTemplate))

	// Test with sheet names
	excelTemplate := &TemplateData{
		Headers: []string{},
		Sheets: []SheetData{
			{Name: "Income Statement"},
			{Name: "Balance Sheet"},
		},
	}
	assert.True(t, parser.isFinancialTemplate(excelTemplate))
}

func TestParseTemplateWithFormulas(t *testing.T) {
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "formula_test.csv")

	csvContent := `Item,Q1,Q2,Q3,Q4,Total
Revenue,1000,2000,3000,4000,=B2+C2+D2+E2
Costs,500,600,700,800,=B3+C3+D3+E3
Profit,=B2-B3,=C2-C3,=D2-D3,=E2-E3,=F2-F3`

	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	require.NoError(t, err)

	parser := NewTemplateParser(tempDir)
	data, err := parser.ParseTemplate(csvPath)

	require.NoError(t, err)

	// Check multiple formulas were detected
	assert.Greater(t, len(data.Formulas), 5)

	// Check specific formulas
	assert.Equal(t, "=B2+C2+D2+E2", data.Formulas["F2"])
	assert.Equal(t, "=B2-B3", data.Formulas["B4"])
	assert.Equal(t, "=F2-F3", data.Formulas["F4"])
}

func TestGetCellReference(t *testing.T) {
	parser := NewTemplateParser("")

	tests := []struct {
		col      int
		row      int
		expected string
	}{
		{0, 0, "A1"},
		{1, 0, "B1"},
		{25, 0, "Z1"},
		{26, 0, "AA1"},
		{0, 9, "A10"},
		{5, 99, "F100"},
	}

	for _, tt := range tests {
		result := parser.getCellReference(tt.col, tt.row)
		assert.Equal(t, tt.expected, result)
	}
}

func TestExtractDataFieldsWithSheets(t *testing.T) {
	parser := NewTemplateParser("")

	templateData := &TemplateData{
		Format: "excel",
		Sheets: []SheetData{
			{
				Name:    "Revenue",
				Headers: []string{"Month", "Amount"},
				Data: [][]string{
					{"January", "100000"},
					{"February", "120000"},
				},
			},
			{
				Name:    "Expenses",
				Headers: []string{"Category", "Cost"},
				Data: [][]string{
					{"Marketing", "50000"},
					{"Operations", "80000"},
				},
			},
		},
	}

	fields := parser.ExtractDataFields(templateData)

	assert.Len(t, fields, 4)

	// Check sheet-specific paths
	assert.Equal(t, "Revenue.Month", fields[0].Path)
	assert.Equal(t, "Revenue", fields[0].Sheet)
	assert.Equal(t, "Revenue.Amount", fields[1].Path)
	assert.Equal(t, "number", fields[1].DataType)

	assert.Equal(t, "Expenses.Category", fields[2].Path)
	assert.Equal(t, "Expenses", fields[2].Sheet)
	assert.Equal(t, "Expenses.Cost", fields[3].Path)
}

func TestIsRequiredField(t *testing.T) {
	parser := NewTemplateParser("")

	tests := []struct {
		fieldName string
		expected  bool
	}{
		{"Company Name", true},
		{"Total Revenue", true},
		{"Deal Date", true},
		{"Amount", true},
		{"Purchase Price", true},
		{"Status", false},
		{"Notes", false},
		{"Description", false},
	}

	for _, tt := range tests {
		result := parser.isRequiredField(tt.fieldName)
		assert.Equal(t, tt.expected, result, "Field: %s", tt.fieldName)
	}
}
