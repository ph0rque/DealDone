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

// TemplatePopulator handles populating templates with mapped data while preserving formulas
type TemplatePopulator struct {
	templateParser *TemplateParser
}

// NewTemplatePopulator creates a new template populator
func NewTemplatePopulator(templateParser *TemplateParser) *TemplatePopulator {
	return &TemplatePopulator{
		templateParser: templateParser,
	}
}

// PopulateTemplate fills a template with mapped data while preserving formulas
func (tp *TemplatePopulator) PopulateTemplate(templatePath string, mappedData *MappedData, outputPath string) error {
	// Parse the template first
	templateData, err := tp.templateParser.ParseTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Route to appropriate handler based on format
	switch templateData.Format {
	case "csv":
		return tp.populateCSVTemplate(templatePath, templateData, mappedData, outputPath)
	case "excel":
		return tp.populateExcelTemplate(templatePath, templateData, mappedData, outputPath)
	case "text":
		return tp.populateTextTemplate(templatePath, templateData, mappedData, outputPath)
	default:
		return fmt.Errorf("unsupported template format: %s", templateData.Format)
	}
}

// populateCSVTemplate populates a CSV template
func (tp *TemplatePopulator) populateCSVTemplate(templatePath string, templateData *TemplateData, mappedData *MappedData, outputPath string) error {
	// Read the original CSV to preserve structure
	file, err := os.Open(templatePath)
	if err != nil {
		return fmt.Errorf("failed to open template: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	// Update records with mapped data
	updatedRecords := tp.updateCSVRecords(records, templateData, mappedData)

	// Write to output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	for _, record := range updatedRecords {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// updateCSVRecords updates CSV records with mapped data
func (tp *TemplatePopulator) updateCSVRecords(records [][]string, templateData *TemplateData, mappedData *MappedData) [][]string {
	// Create a copy of records to avoid modifying the original
	updated := make([][]string, len(records))
	for i := range records {
		updated[i] = make([]string, len(records[i]))
		copy(updated[i], records[i])
	}

	// Find header row (usually first row)
	headerRow := -1
	for i, record := range records {
		if len(record) > 0 && tp.isHeaderRow(record, templateData.Headers) {
			headerRow = i
			break
		}
	}

	if headerRow == -1 {
		return updated // No headers found, return as-is
	}

	// Map column indices to field names
	columnMap := make(map[int]string)
	for colIdx, header := range records[headerRow] {
		columnMap[colIdx] = header
	}

	// Update data rows
	for rowIdx := headerRow + 1; rowIdx < len(updated); rowIdx++ {
		for colIdx, header := range columnMap {
			if colIdx >= len(updated[rowIdx]) {
				continue
			}

			// Check if this cell has a formula
			cellRef := tp.templateParser.getCellReference(colIdx, rowIdx)
			if formula, exists := templateData.Formulas[cellRef]; exists {
				// Preserve formula
				updated[rowIdx][colIdx] = formula
				continue
			}

			// Check if we have mapped data for this field
			if mappedField, exists := mappedData.Fields[header]; exists {
				// Update with mapped value
				updated[rowIdx][colIdx] = tp.formatValue(mappedField.Value)
			}
		}
	}

	return updated
}

// populateTextTemplate populates a text template
func (tp *TemplatePopulator) populateTextTemplate(templatePath string, templateData *TemplateData, mappedData *MappedData, outputPath string) error {
	// Read the original text content
	originalContent, ok := templateData.Metadata["originalContent"].(string)
	if !ok {
		// Fallback: read from file
		file, err := os.Open(templatePath)
		if err != nil {
			return fmt.Errorf("failed to open template: %w", err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read template: %w", err)
		}
		originalContent = string(content)
	}

	// Replace placeholders with mapped data
	populatedContent := originalContent

	// Replace field placeholders with actual values
	for fieldName, mappedField := range mappedData.Fields {
		valueStr := tp.formatValue(mappedField.Value)

		// Create mapping from metadata field names to likely placeholders
		placeholderMappings := tp.getPlaceholderMappings(fieldName)

		// Try all possible placeholder formats and mappings
		for _, placeholderName := range placeholderMappings {
			placeholderFormats := []string{
				"[" + placeholderName + "]",
				"{" + placeholderName + "}",
				"{{" + placeholderName + "}}",
			}

			for _, placeholder := range placeholderFormats {
				populatedContent = strings.ReplaceAll(populatedContent, placeholder, valueStr)
			}
		}
	}

	// Write to output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	_, err = outputFile.WriteString(populatedContent)
	if err != nil {
		return fmt.Errorf("failed to write populated content: %w", err)
	}

	return nil
}

// populateExcelTemplate populates an Excel template while preserving formulas
func (tp *TemplatePopulator) populateExcelTemplate(templatePath string, templateData *TemplateData, mappedData *MappedData, outputPath string) error {
	// Open the Excel file
	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to open Excel template: %w", err)
	}
	defer f.Close()

	// Process each sheet
	for _, sheet := range templateData.Sheets {
		if err := tp.populateExcelSheet(f, sheet, mappedData); err != nil {
			return fmt.Errorf("failed to populate sheet %s: %w", sheet.Name, err)
		}
	}

	// Save to output path
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save populated template: %w", err)
	}

	return nil
}

// populateExcelSheet populates a single Excel sheet
func (tp *TemplatePopulator) populateExcelSheet(f *excelize.File, sheet SheetData, mappedData *MappedData) error {
	// Get all rows in the sheet
	rows, err := f.GetRows(sheet.Name)
	if err != nil {
		return fmt.Errorf("failed to get rows: %w", err)
	}

	// Find header row
	headerRowIdx := -1
	for i, row := range rows {
		if tp.isHeaderRow(row, sheet.Headers) {
			headerRowIdx = i
			break
		}
	}

	if headerRowIdx == -1 {
		return nil // No headers found, skip this sheet
	}

	// Map column indices to field names
	columnMap := make(map[int]string)
	for colIdx, header := range rows[headerRowIdx] {
		columnMap[colIdx] = header
	}

	// Update data cells
	for rowIdx := headerRowIdx + 1; rowIdx < len(rows); rowIdx++ {
		for colIdx, header := range columnMap {
			cellName, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			if err != nil {
				continue
			}

			// Check if this cell has a formula
			formula, err := f.GetCellFormula(sheet.Name, cellName)
			if err == nil && formula != "" {
				// Cell has a formula, preserve it
				continue
			}

			// Check for mapped data
			// Try both direct field name and sheet-qualified name
			fieldPath := fmt.Sprintf("%s.%s", sheet.Name, header)

			var mappedField MappedField
			var found bool

			// Try sheet-qualified path first
			if mf, exists := mappedData.Fields[fieldPath]; exists {
				mappedField = mf
				found = true
			} else if mf, exists := mappedData.Fields[header]; exists {
				// Fall back to direct field name
				mappedField = mf
				found = true
			}

			if found {
				// Set the cell value
				value := tp.formatValueForExcel(mappedField.Value)
				if err := f.SetCellValue(sheet.Name, cellName, value); err != nil {
					return fmt.Errorf("failed to set cell value: %w", err)
				}
			}
		}
	}

	// Handle formula recalculation
	if err := tp.updateFormulaDependencies(f, sheet.Name); err != nil {
		return fmt.Errorf("failed to update formula dependencies: %w", err)
	}

	return nil
}

// isHeaderRow checks if a row matches expected headers
func (tp *TemplatePopulator) isHeaderRow(row []string, expectedHeaders []string) bool {
	if len(row) == 0 || len(expectedHeaders) == 0 {
		return false
	}

	// Check if at least 50% of expected headers are found
	matches := 0
	for _, header := range row {
		normalized := strings.ToLower(strings.TrimSpace(header))
		for _, expected := range expectedHeaders {
			if strings.ToLower(expected) == normalized {
				matches++
				break
			}
		}
	}

	return float64(matches)/float64(len(expectedHeaders)) >= 0.5
}

// formatValue formats a value for CSV output
func (tp *TemplatePopulator) formatValue(value interface{}) string {
	switch v := value.(type) {
	case float64:
		// Check if it's a whole number
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%.2f", v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatValueForExcel formats a value for Excel
func (tp *TemplatePopulator) formatValueForExcel(value interface{}) interface{} {
	// Excel can handle native types directly
	switch v := value.(type) {
	case float64, float32, int, int32, int64:
		return v
	case string:
		// Check if it's a formula
		if strings.HasPrefix(v, "=") {
			return v
		}
		// Check if it's a number string
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num
		}
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// updateFormulaDependencies ensures formulas are recalculated
func (tp *TemplatePopulator) updateFormulaDependencies(f *excelize.File, sheetName string) error {
	// Force Excel to recalculate formulas when opened
	// This is done by setting the calculation mode
	if err := f.SetSheetProps(sheetName, &excelize.SheetPropsOptions{
		EnableFormatConditionsCalculation: &[]bool{true}[0],
	}); err != nil {
		// This is not critical, so we just log it
		return nil
	}

	return nil
}

// PreserveFormulas analyzes a template and returns information about formulas
func (tp *TemplatePopulator) PreserveFormulas(templatePath string) (*FormulaPreservation, error) {
	templateData, err := tp.templateParser.ParseTemplate(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	preservation := &FormulaPreservation{
		Format:         templateData.Format,
		TotalFormulas:  len(templateData.Formulas),
		FormulaCells:   make([]FormulaCellInfo, 0),
		Dependencies:   make(map[string][]string),
		PreservedCells: make(map[string]bool),
	}

	// Analyze formulas
	for cell, formula := range templateData.Formulas {
		info := FormulaCellInfo{
			Cell:    cell,
			Formula: formula,
			Sheet:   "", // Will be set for Excel files
		}

		// Extract dependencies (referenced cells)
		deps := tp.extractFormulaDependencies(formula)
		if len(deps) > 0 {
			preservation.Dependencies[cell] = deps
		}

		preservation.FormulaCells = append(preservation.FormulaCells, info)
		preservation.PreservedCells[cell] = true
	}

	// For Excel files, also process sheet-specific formulas
	if templateData.Format == "excel" {
		for _, sheet := range templateData.Sheets {
			for cell, formula := range sheet.Formulas {
				info := FormulaCellInfo{
					Cell:    cell,
					Formula: formula,
					Sheet:   sheet.Name,
				}

				fullCell := fmt.Sprintf("%s!%s", sheet.Name, cell)
				deps := tp.extractFormulaDependencies(formula)
				if len(deps) > 0 {
					preservation.Dependencies[fullCell] = deps
				}

				preservation.FormulaCells = append(preservation.FormulaCells, info)
				preservation.PreservedCells[fullCell] = true
			}
		}
	}

	return preservation, nil
}

// extractFormulaDependencies extracts cell references from a formula
func (tp *TemplatePopulator) extractFormulaDependencies(formula string) []string {
	deps := make([]string, 0)

	// Remove the leading = sign
	formula = strings.TrimPrefix(formula, "=")

	// Simple regex to find cell references like A1, B2, AA10
	// This is a simplified version - real Excel formulas can be more complex
	cellPattern := `[A-Z]+[0-9]+`

	// Find all matches
	for _, match := range findAllStringSubmatch(cellPattern, formula) {
		deps = append(deps, match)
	}

	return deps
}

// findAllStringSubmatch is a helper to find all regex matches
func findAllStringSubmatch(pattern, text string) []string {
	matches := make([]string, 0)

	// This is a simplified implementation
	// In a real implementation, you would use regexp package
	parts := strings.Fields(text)
	for _, part := range parts {
		// Check if part looks like a cell reference
		if len(part) > 1 && isUpperLetter(part[0]) {
			hasDigit := false
			for _, ch := range part[1:] {
				if ch >= '0' && ch <= '9' {
					hasDigit = true
					break
				}
			}
			if hasDigit {
				matches = append(matches, part)
			}
		}
	}

	return matches
}

// isUpperLetter checks if a byte is an uppercase letter
func isUpperLetter(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

// FormulaPreservation contains information about preserved formulas
type FormulaPreservation struct {
	Format         string              `json:"format"`
	TotalFormulas  int                 `json:"totalFormulas"`
	FormulaCells   []FormulaCellInfo   `json:"formulaCells"`
	Dependencies   map[string][]string `json:"dependencies"`
	PreservedCells map[string]bool     `json:"preservedCells"`
}

// FormulaCellInfo contains information about a cell with a formula
type FormulaCellInfo struct {
	Cell    string `json:"cell"`
	Sheet   string `json:"sheet,omitempty"`
	Formula string `json:"formula"`
}

// ValidatePopulatedTemplate checks if a populated template maintains formula integrity
func (tp *TemplatePopulator) ValidatePopulatedTemplate(populatedPath string, originalFormulas *FormulaPreservation) error {
	ext := strings.ToLower(filepath.Ext(populatedPath))

	switch ext {
	case ".xlsx", ".xls":
		return tp.validateExcelFormulas(populatedPath, originalFormulas)
	case ".csv":
		return tp.validateCSVFormulas(populatedPath, originalFormulas)
	case ".txt":
		return tp.validateTextTemplate(populatedPath, originalFormulas)
	default:
		return fmt.Errorf("unsupported format for validation: %s", ext)
	}
}

// validateExcelFormulas validates formulas in an Excel file
func (tp *TemplatePopulator) validateExcelFormulas(filePath string, originalFormulas *FormulaPreservation) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for validation: %w", err)
	}
	defer f.Close()

	missingFormulas := []string{}

	for _, formulaInfo := range originalFormulas.FormulaCells {
		if formulaInfo.Sheet == "" {
			// Check first sheet
			sheets := f.GetSheetList()
			if len(sheets) > 0 {
				formula, err := f.GetCellFormula(sheets[0], formulaInfo.Cell)
				if err != nil || formula == "" {
					missingFormulas = append(missingFormulas, formulaInfo.Cell)
				}
			}
		} else {
			// Check specific sheet
			formula, err := f.GetCellFormula(formulaInfo.Sheet, formulaInfo.Cell)
			if err != nil || formula == "" {
				missingFormulas = append(missingFormulas, fmt.Sprintf("%s!%s", formulaInfo.Sheet, formulaInfo.Cell))
			}
		}
	}

	if len(missingFormulas) > 0 {
		return fmt.Errorf("missing formulas in cells: %s", strings.Join(missingFormulas, ", "))
	}

	return nil
}

// validateCSVFormulas validates formulas in a CSV file
func (tp *TemplatePopulator) validateCSVFormulas(filePath string, originalFormulas *FormulaPreservation) error {
	// CSV validation is simpler - just check if formula strings are preserved
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV for validation: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	// Build a map of cell content
	cellContent := make(map[string]string)
	for rowIdx, row := range records {
		for colIdx, value := range row {
			cellRef := tp.templateParser.getCellReference(colIdx, rowIdx)
			cellContent[cellRef] = value
		}
	}

	// Check if formulas are preserved
	missingFormulas := []string{}
	for cell := range originalFormulas.PreservedCells {
		if content, exists := cellContent[cell]; !exists || !strings.HasPrefix(content, "=") {
			missingFormulas = append(missingFormulas, cell)
		}
	}

	if len(missingFormulas) > 0 {
		return fmt.Errorf("missing formulas in cells: %s", strings.Join(missingFormulas, ", "))
	}

	return nil
}

// validateTextTemplate validates a text template (text templates don't have formulas to validate)
func (tp *TemplatePopulator) validateTextTemplate(filePath string, originalFormulas *FormulaPreservation) error {
	// Text templates don't have formulas, so validation just checks if file exists and is readable
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("populated text template not found: %w", err)
	}

	// Check if file is readable
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot read populated text template: %w", err)
	}
	defer file.Close()

	return nil
}

// getPlaceholderMappings maps metadata field names to likely placeholders in the template
func (tp *TemplatePopulator) getPlaceholderMappings(fieldName string) []string {
	fieldLower := strings.ToLower(fieldName)

	// Common mappings from metadata field names to template placeholders
	mappings := map[string][]string{
		"deal name":        {"To be filled", "Deal Name", "deal name"},
		"target company":   {"To be filled", "Target Company", "Company Name", "Name", "target company"},
		"company name":     {"Name", "Company Name", "To be filled", "company name"},
		"deal value":       {"Amount", "Deal Value", "Value", "Price", "deal value"},
		"purchase price":   {"Amount", "Purchase Price", "Price", "Value", "purchase price"},
		"enterprise value": {"Amount", "Enterprise Value", "EV", "Value", "enterprise value"},
		"revenue":          {"Amount", "Revenue", "Sales", "Income", "revenue"},
		"ebitda":           {"Amount", "EBITDA", "Earnings", "ebitda"},
		"net income":       {"Amount", "Net Income", "Profit", "net income"},
		"date":             {"Date", "Transaction Date", "Deal Date", "date"},
		"industry":         {"Industry", "Sector", "Business", "industry"},
		"founded":          {"Year", "Founded", "Establishment", "founded"},
		"employees":        {"Number", "Employees", "Headcount", "Staff", "employees"},
		"headquarters":     {"Location", "Headquarters", "HQ", "Office", "headquarters"},
		"website":          {"URL", "Website", "Web", "website"},
		"deal type":        {"Deal Type", "Transaction Type", "Type", "deal type"},
	}

	// Get specific mappings for this field
	if placeholders, exists := mappings[fieldLower]; exists {
		return placeholders
	}

	// Default mappings - try the field name itself and common generic placeholders
	return []string{
		fieldName,                  // Exact match
		strings.Title(fieldName),   // Title case
		strings.ToLower(fieldName), // Lower case
		"To be filled",             // Generic placeholder
		"Amount",                   // For currency fields
		"Name",                     // For name fields
		"Date",                     // For date fields
		"Number",                   // For numeric fields
	}
}
