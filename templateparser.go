package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// TemplateParser handles parsing of Excel and CSV templates
type TemplateParser struct {
	templatesPath string
}

// NewTemplateParser creates a new template parser
func NewTemplateParser(templatesPath string) *TemplateParser {
	return &TemplateParser{
		templatesPath: templatesPath,
	}
}

// TemplateData represents parsed template data
type TemplateData struct {
	Format   string                 `json:"format"` // "excel", "csv", or "text"
	Headers  []string               `json:"headers"`
	Data     [][]string             `json:"data"`
	Formulas map[string]string      `json:"formulas"` // Cell -> Formula mapping
	Sheets   []SheetData            `json:"sheets"`   // For Excel files with multiple sheets
	Metadata map[string]interface{} `json:"metadata"`
}

// SheetData represents data from a single sheet
type SheetData struct {
	Name     string            `json:"name"`
	Headers  []string          `json:"headers"`
	Data     [][]string        `json:"data"`
	Formulas map[string]string `json:"formulas"`
}

// CellReference represents a cell location
type CellReference struct {
	Sheet  string `json:"sheet"`
	Column string `json:"column"`
	Row    int    `json:"row"`
}

// ParseTemplate parses a template file and returns structured data
func (tp *TemplateParser) ParseTemplate(templatePath string) (*TemplateData, error) {
	ext := strings.ToLower(filepath.Ext(templatePath))

	switch ext {
	case ".csv":
		return tp.parseCSVTemplate(templatePath)
	case ".xlsx", ".xls":
		return tp.parseExcelTemplate(templatePath)
	case ".txt", ".md":
		return tp.parseTextTemplate(templatePath)
	default:
		return nil, fmt.Errorf("unsupported template format: %s", ext)
	}
}

// parseTextTemplate parses a text template file
func (tp *TemplateParser) parseTextTemplate(filePath string) (*TemplateData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open text template: %w", err)
	}
	defer file.Close()

	// Read the entire file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read text template: %w", err)
	}

	// Split content into lines
	lines := strings.Split(string(content), "\n")

	// For text templates, we treat each line as a potential field placeholder
	// Look for patterns like [Field Name] or {Field Name} or {{Field Name}}
	headers := []string{}
	data := [][]string{}

	// Extract field placeholders from content
	for _, line := range lines {
		// Look for common placeholder patterns
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			// Extract text between brackets
			start := strings.Index(line, "[")
			end := strings.Index(line, "]")
			if start >= 0 && end > start {
				placeholder := strings.TrimSpace(line[start+1 : end])
				if placeholder != "" && !tp.containsString(headers, placeholder) {
					headers = append(headers, placeholder)
				}
			}
		}
	}

	// If no placeholders found, create basic structure
	if len(headers) == 0 {
		headers = []string{"Content"}
		data = append(data, []string{string(content)})
	} else {
		// Create empty data row for placeholders
		emptyRow := make([]string, len(headers))
		for i := range emptyRow {
			emptyRow[i] = ""
		}
		data = append(data, emptyRow)
	}

	return &TemplateData{
		Format:   "text",
		Headers:  headers,
		Data:     data,
		Formulas: make(map[string]string), // Text templates don't have formulas
		Sheets:   []SheetData{},
		Metadata: map[string]interface{}{
			"originalContent": string(content),
			"lineCount":       len(lines),
		},
	}, nil
}

// containsString checks if a string slice contains a specific string
func (tp *TemplateParser) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// parseCSVTemplate parses a CSV template file
func (tp *TemplateParser) parseCSVTemplate(filePath string) (*TemplateData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	var headers []string
	var data [][]string
	formulas := make(map[string]string)

	rowIndex := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV: %w", err)
		}

		// Process formulas in cells
		for colIndex, cell := range record {
			if strings.HasPrefix(cell, "=") {
				// This is a formula
				cellRef := tp.getCellReference(colIndex, rowIndex)
				formulas[cellRef] = cell
			}
		}

		if rowIndex == 0 && len(headers) == 0 {
			// First non-empty row is usually headers
			headers = record
		} else {
			data = append(data, record)
		}
		rowIndex++
	}

	return &TemplateData{
		Format:   "csv",
		Headers:  headers,
		Data:     data,
		Formulas: formulas,
		Metadata: map[string]interface{}{
			"fileName": filepath.Base(filePath),
			"rows":     len(data),
			"columns":  len(headers),
		},
	}, nil
}

// parseExcelTemplate parses an Excel template file
func (tp *TemplateParser) parseExcelTemplate(filePath string) (*TemplateData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	templateData := &TemplateData{
		Format:   "excel",
		Sheets:   []SheetData{},
		Formulas: make(map[string]string),
		Metadata: map[string]interface{}{
			"fileName": filepath.Base(filePath),
		},
	}

	// Get all sheet names
	sheets := f.GetSheetList()

	for _, sheetName := range sheets {
		sheetData := SheetData{
			Name:     sheetName,
			Headers:  []string{},
			Data:     [][]string{},
			Formulas: make(map[string]string),
		}

		// Get all rows in the sheet
		rows, err := f.GetRows(sheetName)
		if err != nil {
			continue
		}

		// Extract formulas for this sheet
		tp.extractFormulas(f, sheetName, &sheetData)

		// Process rows
		for rowIndex, row := range rows {
			if rowIndex == 0 && len(sheetData.Headers) == 0 {
				// Assume first row contains headers
				sheetData.Headers = row
			} else {
				sheetData.Data = append(sheetData.Data, row)
			}
		}

		templateData.Sheets = append(templateData.Sheets, sheetData)

		// For backward compatibility, expose first sheet data at top level
		if len(templateData.Sheets) == 1 {
			templateData.Headers = sheetData.Headers
			templateData.Data = sheetData.Data
			for cell, formula := range sheetData.Formulas {
				templateData.Formulas[cell] = formula
			}
		}
	}

	templateData.Metadata["sheetCount"] = len(sheets)

	return templateData, nil
}

// extractFormulas extracts formulas from an Excel sheet
func (tp *TemplateParser) extractFormulas(f *excelize.File, sheetName string, sheetData *SheetData) {
	// Get the sheet's dimension
	dimension, err := f.GetSheetDimension(sheetName)
	if err != nil {
		return
	}

	// Parse dimension (e.g., "A1:F30")
	coords := strings.Split(dimension, ":")
	if len(coords) != 2 {
		return
	}

	startCol, startRow, _ := excelize.CellNameToCoordinates(coords[0])
	endCol, endRow, _ := excelize.CellNameToCoordinates(coords[1])

	// Check each cell for formulas
	for row := startRow; row <= endRow; row++ {
		for col := startCol; col <= endCol; col++ {
			cellName, _ := excelize.CoordinatesToCellName(col, row)
			formula, err := f.GetCellFormula(sheetName, cellName)
			if err == nil && formula != "" {
				sheetData.Formulas[cellName] = formula
			}
		}
	}
}

// ExtractDataFields extracts all data fields from a template
func (tp *TemplateParser) ExtractDataFields(templateData *TemplateData) []DataField {
	fields := []DataField{}

	if len(templateData.Sheets) > 0 {
		// Multi-sheet Excel file
		for _, sheet := range templateData.Sheets {
			for colIndex, header := range sheet.Headers {
				field := DataField{
					Name:       header,
					Path:       fmt.Sprintf("%s.%s", sheet.Name, header),
					DataType:   tp.inferDataType(sheet.Data, colIndex),
					IsRequired: tp.isRequiredField(header),
					Sheet:      sheet.Name,
				}
				fields = append(fields, field)
			}
		}
	} else {
		// Single sheet or CSV
		for colIndex, header := range templateData.Headers {
			field := DataField{
				Name:       header,
				Path:       header,
				DataType:   tp.inferDataType(templateData.Data, colIndex),
				IsRequired: tp.isRequiredField(header),
			}
			fields = append(fields, field)
		}
	}

	return fields
}

// DataField represents a field in the template
type DataField struct {
	Name       string `json:"name"`
	Path       string `json:"path"`     // Full path including sheet name
	DataType   string `json:"dataType"` // "string", "number", "date", "currency"
	IsRequired bool   `json:"isRequired"`
	Sheet      string `json:"sheet,omitempty"`
}

// inferDataType attempts to infer the data type from sample data
func (tp *TemplateParser) inferDataType(data [][]string, colIndex int) string {
	hasNumbers := false
	hasCurrency := false
	hasText := false

	for _, row := range data {
		if colIndex >= len(row) {
			continue
		}

		value := strings.TrimSpace(row[colIndex])
		if value == "" {
			continue
		}

		// Check for currency
		if strings.Contains(value, "$") || strings.Contains(value, "€") || strings.Contains(value, "£") {
			hasCurrency = true
			continue
		}

		// Check for number
		if _, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", ""), 64); err == nil {
			hasNumbers = true
		} else {
			hasText = true
		}
	}

	if hasCurrency {
		return "currency"
	}
	if hasNumbers && !hasText {
		return "number"
	}

	return "string"
}

// isRequiredField determines if a field should be required based on its name
func (tp *TemplateParser) isRequiredField(fieldName string) bool {
	requiredKeywords := []string{
		"name", "date", "amount", "total", "company", "revenue", "price",
	}

	fieldLower := strings.ToLower(fieldName)
	for _, keyword := range requiredKeywords {
		if strings.Contains(fieldLower, keyword) {
			return true
		}
	}

	return false
}

// getCellReference converts column and row indices to Excel-style cell reference
func (tp *TemplateParser) getCellReference(col, row int) string {
	cellName, _ := excelize.CoordinatesToCellName(col+1, row+1)
	return cellName
}

// ValidateTemplateStructure checks if a template has valid structure
func (tp *TemplateParser) ValidateTemplateStructure(templateData *TemplateData) error {
	if len(templateData.Headers) == 0 && len(templateData.Sheets) == 0 {
		return fmt.Errorf("template has no headers or sheets")
	}

	if templateData.Format == "excel" && len(templateData.Sheets) == 0 {
		return fmt.Errorf("Excel template has no sheets")
	}

	// Check for required fields in financial templates
	if tp.isFinancialTemplate(templateData) {
		if !tp.hasRequiredFinancialFields(templateData) {
			return fmt.Errorf("financial template missing required fields")
		}
	}

	return nil
}

// isFinancialTemplate checks if this is a financial template
func (tp *TemplateParser) isFinancialTemplate(templateData *TemplateData) bool {
	financialKeywords := []string{"revenue", "ebitda", "income", "expense", "cash flow"}

	// Check headers
	for _, header := range templateData.Headers {
		headerLower := strings.ToLower(header)
		for _, keyword := range financialKeywords {
			if strings.Contains(headerLower, keyword) {
				return true
			}
		}
	}

	// Check sheet names
	for _, sheet := range templateData.Sheets {
		sheetLower := strings.ToLower(sheet.Name)
		for _, keyword := range financialKeywords {
			if strings.Contains(sheetLower, keyword) {
				return true
			}
		}
	}

	return false
}

// hasRequiredFinancialFields checks for essential financial fields
func (tp *TemplateParser) hasRequiredFinancialFields(templateData *TemplateData) bool {
	// For now, just check that we have some numeric fields
	fields := tp.ExtractDataFields(templateData)

	for _, field := range fields {
		if field.DataType == "number" || field.DataType == "currency" {
			return true
		}
	}

	return false
}
