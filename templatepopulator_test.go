package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplatePopulator(t *testing.T) {
	parser := &TemplateParser{}
	populator := NewTemplatePopulator(parser)

	assert.NotNil(t, populator)
	assert.Equal(t, parser, populator.templateParser)
}

func TestIsHeaderRow(t *testing.T) {
	populator := NewTemplatePopulator(nil)

	tests := []struct {
		name            string
		row             []string
		expectedHeaders []string
		expected        bool
	}{
		{
			name:            "Exact match",
			row:             []string{"Company", "Revenue", "Date"},
			expectedHeaders: []string{"Company", "Revenue", "Date"},
			expected:        true,
		},
		{
			name:            "Case insensitive match",
			row:             []string{"company", "REVENUE", "Date"},
			expectedHeaders: []string{"Company", "Revenue", "Date"},
			expected:        true,
		},
		{
			name:            "Partial match (>50%)",
			row:             []string{"Company", "Revenue", "Something Else"},
			expectedHeaders: []string{"Company", "Revenue", "Date"},
			expected:        true,
		},
		{
			name:            "Not enough matches",
			row:             []string{"Name", "Value", "Type"},
			expectedHeaders: []string{"Company", "Revenue", "Date"},
			expected:        false,
		},
		{
			name:            "Empty row",
			row:             []string{},
			expectedHeaders: []string{"Company", "Revenue"},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := populator.isHeaderRow(tt.row, tt.expectedHeaders)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatValue(t *testing.T) {
	populator := NewTemplatePopulator(nil)

	tests := []struct {
		value    interface{}
		expected string
	}{
		{1000000.0, "1000000"},
		{1234.56, "1234.56"},
		{"Test String", "Test String"},
		{true, "true"},
		{nil, "<nil>"},
	}

	for _, tt := range tests {
		result := populator.formatValue(tt.value)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFormatValueForExcel(t *testing.T) {
	populator := NewTemplatePopulator(nil)

	tests := []struct {
		value    interface{}
		expected interface{}
	}{
		{1000000.0, 1000000.0},
		{int64(42), int64(42)},
		{"=SUM(A1:A10)", "=SUM(A1:A10)"},
		{"1234.56", 1234.56},
		{"Not a number", "Not a number"},
		{true, "true"},
	}

	for _, tt := range tests {
		result := populator.formatValueForExcel(tt.value)
		assert.Equal(t, tt.expected, result)
	}
}

func TestExtractFormulaDependencies(t *testing.T) {
	populator := NewTemplatePopulator(nil)

	tests := []struct {
		formula  string
		expected []string
	}{
		{"=A1+B1", []string{"A1", "B1"}},
		{"=SUM(A1:A10)", []string{"A1", "A10"}},
		{"=B2*C2+D2", []string{"B2", "C2", "D2"}},
		{"=AVERAGE(AA1,BB2,CC3)", []string{"AA1", "BB2", "CC3"}},
		{"=100+200", []string{}},
		{"Just text", []string{}},
	}

	for _, tt := range tests {
		deps := populator.extractFormulaDependencies(tt.formula)
		if len(tt.expected) == 0 {
			assert.Empty(t, deps)
		} else {
			// Check that expected dependencies are found
			for _, exp := range tt.expected {
				found := false
				for _, dep := range deps {
					if dep == exp {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected dependency %s not found in %v", exp, deps)
			}
		}
	}
}

func TestUpdateCSVRecords(t *testing.T) {
	parser := &TemplateParser{}
	populator := NewTemplatePopulator(parser)

	// Original CSV records
	records := [][]string{
		{"Company", "Revenue", "EBITDA", "Total"},
		{"Test Corp", "1000", "200", "=B2+C2"},
		{"Example Inc", "2000", "400", "=B3+C3"},
	}

	// Template data with formulas
	templateData := &TemplateData{
		Headers: []string{"Company", "Revenue", "EBITDA", "Total"},
		Formulas: map[string]string{
			"D2": "=B2+C2",
			"D3": "=B3+C3",
		},
	}

	// Mapped data
	mappedData := &MappedData{
		Fields: map[string]MappedField{
			"Company": {Value: "Updated Corp"},
			"Revenue": {Value: 1500000.0},
			"EBITDA":  {Value: 300000.0},
		},
	}

	updated := populator.updateCSVRecords(records, templateData, mappedData)

	// Check headers are unchanged
	assert.Equal(t, records[0], updated[0])

	// Check data is updated
	assert.Equal(t, "Updated Corp", updated[1][0])
	assert.Equal(t, "1500000", updated[1][1])
	assert.Equal(t, "300000", updated[1][2])

	// Check formulas are preserved
	assert.Equal(t, "=B2+C2", updated[1][3])
	assert.Equal(t, "=B3+C3", updated[2][3])
}

func TestPreserveFormulas(t *testing.T) {
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "test_formulas.csv")

	// Create CSV with formulas
	csvContent := `Company,Q1,Q2,Q3,Q4,Total
Test Corp,1000,2000,3000,4000,=B2+C2+D2+E2
Costs,500,600,700,800,=B3+C3+D3+E3
Profit,=B2-B3,=C2-C3,=D2-D3,=E2-E3,=F2-F3`

	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	require.NoError(t, err)

	parser := NewTemplateParser(tempDir)
	populator := NewTemplatePopulator(parser)

	preservation, err := populator.PreserveFormulas(csvPath)

	require.NoError(t, err)
	assert.Equal(t, "csv", preservation.Format)
	assert.Greater(t, preservation.TotalFormulas, 5)
	assert.Greater(t, len(preservation.FormulaCells), 5)

	// Check specific formulas
	assert.Contains(t, preservation.PreservedCells, "F2")
	assert.Contains(t, preservation.PreservedCells, "B4")
	assert.Contains(t, preservation.PreservedCells, "F4")
}

func TestPopulateCSVTemplate(t *testing.T) {
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "template.csv")
	outputPath := filepath.Join(tempDir, "output.csv")

	// Create template CSV
	csvContent := `Company Name,Revenue,EBITDA,Margin
[To be filled],0,0,=C2/B2`

	err := os.WriteFile(templatePath, []byte(csvContent), 0644)
	require.NoError(t, err)

	parser := NewTemplateParser(tempDir)
	populator := NewTemplatePopulator(parser)

	// Create mapped data
	mappedData := &MappedData{
		Fields: map[string]MappedField{
			"Company Name": {Value: "Tech Corp"},
			"Revenue":      {Value: 1000000.0},
			"EBITDA":       {Value: 200000.0},
		},
	}

	// Populate template
	err = populator.PopulateTemplate(templatePath, mappedData, outputPath)
	require.NoError(t, err)

	// Read output file
	outputContent, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	// Check content
	output := string(outputContent)
	assert.Contains(t, output, "Tech Corp")
	assert.Contains(t, output, "1000000")
	assert.Contains(t, output, "200000")
	assert.Contains(t, output, "=C2/B2") // Formula preserved
}

func TestIsUpperLetter(t *testing.T) {
	tests := []struct {
		b        byte
		expected bool
	}{
		{'A', true},
		{'Z', true},
		{'M', true},
		{'a', false},
		{'z', false},
		{'1', false},
		{'!', false},
	}

	for _, tt := range tests {
		result := isUpperLetter(tt.b)
		assert.Equal(t, tt.expected, result, "Byte: %c", tt.b)
	}
}

func TestValidatePopulatedTemplate(t *testing.T) {
	tempDir := t.TempDir()

	// Create a populated CSV file
	csvPath := filepath.Join(tempDir, "populated.csv")
	csvContent := `Company,Revenue,Total
Tech Corp,1000,=B2*1.1`

	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	require.NoError(t, err)

	parser := NewTemplateParser(tempDir)
	populator := NewTemplatePopulator(parser)

	// Create formula preservation info
	preservation := &FormulaPreservation{
		Format: "csv",
		PreservedCells: map[string]bool{
			"C2": true,
		},
		FormulaCells: []FormulaCellInfo{
			{Cell: "C2", Formula: "=B2*1.1"},
		},
	}

	// Validate
	err = populator.ValidatePopulatedTemplate(csvPath, preservation)
	require.NoError(t, err)

	// Test with missing formula
	csvContent2 := `Company,Revenue,Total
Tech Corp,1000,1100`

	csvPath2 := filepath.Join(tempDir, "populated2.csv")
	err = os.WriteFile(csvPath2, []byte(csvContent2), 0644)
	require.NoError(t, err)

	err = populator.ValidatePopulatedTemplate(csvPath2, preservation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing formulas")
}
