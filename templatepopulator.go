package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// TemplatePopulator handles populating templates with mapped data while preserving formulas
type TemplatePopulator struct {
	templateParser        *TemplateParser
	professionalFormatter *ProfessionalFormatter
}

// NewTemplatePopulator creates a new template populator
func NewTemplatePopulator(templateParser *TemplateParser) *TemplatePopulator {
	return &TemplatePopulator{
		templateParser:        templateParser,
		professionalFormatter: NewProfessionalFormatter(),
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

	// First, handle placeholder replacement in all cells
	for rowIdx := range updated {
		for colIdx := range updated[rowIdx] {
			cellValue := updated[rowIdx][colIdx]

			// Replace placeholders in CSV cells
			if strings.Contains(cellValue, "[To be filled]") {
				// Find a suitable replacement value
				for fieldName, mappedField := range mappedData.Fields {
					fieldLower := strings.ToLower(fieldName)
					if strings.Contains(fieldLower, "company") || strings.Contains(fieldLower, "name") || strings.Contains(fieldLower, "deal") {
						valueStr := fmt.Sprintf("%v", mappedField.Value)
						if valueStr != "" {
							updated[rowIdx][colIdx] = strings.ReplaceAll(cellValue, "[To be filled]", valueStr)
							fmt.Printf("DEBUG: CSV placeholder replacement: '[To be filled]' -> '%s' in cell [%d,%d]\n", valueStr, rowIdx, colIdx)
							break
						}
					}
				}
			}
		}
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
		return updated // No headers found, return updated with placeholder replacements
	}

	// Map column indices to field names
	columnMap := make(map[int]string)
	for colIdx, header := range records[headerRow] {
		columnMap[colIdx] = header
	}

	// Update data rows with mapped field data
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
				// Create context for professional formatting
				context := FormattingContext{
					FieldName:    header,
					FieldType:    "",
					TemplateType: "csv",
					Metadata:     make(map[string]interface{}),
				}
				// Update with professionally formatted value
				updated[rowIdx][colIdx] = tp.formatValueWithContext(mappedField.Value, context)
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

	// First, try the standard field mapping approach
	for fieldName, mappedField := range mappedData.Fields {
		// Create context for professional formatting
		context := FormattingContext{
			FieldName:    fieldName,
			FieldType:    "",
			TemplateType: "text",
			Metadata:     make(map[string]interface{}),
		}
		valueStr := tp.formatValueWithContext(mappedField.Value, context)

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
				if strings.Contains(populatedContent, placeholder) {
					fmt.Printf("DEBUG: Replacing placeholder '%s' with value '%s' for field '%s'\n", placeholder, valueStr, fieldName)
					populatedContent = strings.ReplaceAll(populatedContent, placeholder, valueStr)
				}
			}
		}
	}

	// Enhanced direct replacement logic - more comprehensive mapping
	// Create a comprehensive mapping of placeholders to values
	placeholderValues := make(map[string]string)

	// Initialize with empty strings to handle missing values gracefully
	placeholderValues["[To be filled]"] = ""
	placeholderValues["[Amount]"] = ""
	placeholderValues["[Name]"] = ""
	placeholderValues["[Date]"] = ""
	placeholderValues["[Industry]"] = ""
	placeholderValues["[Year]"] = ""
	placeholderValues["[Location]"] = ""
	placeholderValues["[Number]"] = ""
	placeholderValues["[URL]"] = ""
	placeholderValues["[%]"] = ""
	placeholderValues["[Type]"] = ""
	placeholderValues["[Acquisition/Merger/Investment]"] = ""

	// Map field values to the appropriate placeholders based on field names and content
	for fieldName, mappedField := range mappedData.Fields {
		valueStr := fmt.Sprintf("%v", mappedField.Value)
		fieldLower := strings.ToLower(fieldName)

		// Enhanced field mapping logic
		switch {
		// Deal name and company name mapping
		case strings.Contains(fieldLower, "deal") && strings.Contains(fieldLower, "name"),
			strings.Contains(fieldLower, "deal_name"):
			placeholderValues["[To be filled]"] = valueStr

		case strings.Contains(fieldLower, "company") && strings.Contains(fieldLower, "name"),
			strings.Contains(fieldLower, "target") && strings.Contains(fieldLower, "company"),
			strings.Contains(fieldLower, "company_name"),
			strings.Contains(fieldLower, "target_company"):
			if placeholderValues["[To be filled]"] == "" {
				placeholderValues["[To be filled]"] = valueStr
			}
			placeholderValues["[Name]"] = valueStr

		// Deal type mapping
		case strings.Contains(fieldLower, "deal") && strings.Contains(fieldLower, "type"),
			strings.Contains(fieldLower, "deal_type"),
			strings.Contains(fieldLower, "transaction") && strings.Contains(fieldLower, "type"):
			placeholderValues["[Acquisition/Merger/Investment]"] = valueStr
			placeholderValues["[Type]"] = valueStr

		// Financial values mapping
		case strings.Contains(fieldLower, "deal") && strings.Contains(fieldLower, "value"),
			strings.Contains(fieldLower, "deal_value"),
			strings.Contains(fieldLower, "purchase") && strings.Contains(fieldLower, "price"),
			strings.Contains(fieldLower, "enterprise") && strings.Contains(fieldLower, "value"),
			strings.Contains(fieldLower, "transaction") && strings.Contains(fieldLower, "value"):
			placeholderValues["[Amount]"] = valueStr

		case strings.Contains(fieldLower, "revenue"),
			strings.Contains(fieldLower, "ebitda"),
			strings.Contains(fieldLower, "income"),
			strings.Contains(fieldLower, "price"),
			strings.Contains(fieldLower, "value") && !strings.Contains(fieldLower, "deal"):
			if placeholderValues["[Amount]"] == "" {
				placeholderValues["[Amount]"] = valueStr
			}

		// Industry mapping
		case strings.Contains(fieldLower, "industry"),
			strings.Contains(fieldLower, "sector"),
			strings.Contains(fieldLower, "business"):
			placeholderValues["[Industry]"] = valueStr

		// Date mapping
		case strings.Contains(fieldLower, "date"),
			strings.Contains(fieldLower, "transaction") && strings.Contains(fieldLower, "date"),
			strings.Contains(fieldLower, "deal") && strings.Contains(fieldLower, "date"):
			placeholderValues["[Date]"] = valueStr

		// Year mapping
		case strings.Contains(fieldLower, "year"),
			strings.Contains(fieldLower, "founded"),
			strings.Contains(fieldLower, "established"):
			placeholderValues["[Year]"] = valueStr

		// Location mapping
		case strings.Contains(fieldLower, "location"),
			strings.Contains(fieldLower, "headquarters"),
			strings.Contains(fieldLower, "address"),
			strings.Contains(fieldLower, "office"):
			placeholderValues["[Location]"] = valueStr

		// Employee count mapping
		case strings.Contains(fieldLower, "employees"),
			strings.Contains(fieldLower, "headcount"),
			strings.Contains(fieldLower, "staff"):
			placeholderValues["[Number]"] = valueStr

		// Website mapping
		case strings.Contains(fieldLower, "website"),
			strings.Contains(fieldLower, "url"),
			strings.Contains(fieldLower, "web"):
			placeholderValues["[URL]"] = valueStr

		// Percentage mapping
		case strings.Contains(fieldLower, "margin"),
			strings.Contains(fieldLower, "growth"),
			strings.Contains(fieldLower, "percent"),
			strings.Contains(fieldLower, "%"):
			placeholderValues["[%]"] = valueStr
		}
	}

	// Apply all placeholder replacements
	for placeholder, replacement := range placeholderValues {
		if replacement != "" && strings.Contains(populatedContent, placeholder) {
			fmt.Printf("DEBUG: Enhanced replacement: '%s' -> '%s'\n", placeholder, replacement)
			populatedContent = strings.ReplaceAll(populatedContent, placeholder, replacement)
		}
	}

	// Special handling for common template patterns
	// Replace any remaining [To be filled] with first available company/deal name
	if strings.Contains(populatedContent, "[To be filled]") {
		for fieldName, mappedField := range mappedData.Fields {
			fieldLower := strings.ToLower(fieldName)
			if strings.Contains(fieldLower, "name") || strings.Contains(fieldLower, "company") || strings.Contains(fieldLower, "deal") {
				valueStr := fmt.Sprintf("%v", mappedField.Value)
				if valueStr != "" {
					fmt.Printf("DEBUG: Final fallback replacement for [To be filled]: '%s'\n", valueStr)
					populatedContent = strings.ReplaceAll(populatedContent, "[To be filled]", valueStr)
					break
				}
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
				// Create context for professional formatting
				context := FormattingContext{
					FieldName:    header,
					FieldType:    "",
					TemplateType: "excel",
					Metadata:     make(map[string]interface{}),
				}
				// Set the cell value with professional formatting
				value := tp.formatValueForExcelWithContext(mappedField.Value, context)
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

// formatValue formats a value for CSV output using professional formatting
func (tp *TemplatePopulator) formatValue(value interface{}) string {
	return tp.formatValueWithContext(value, FormattingContext{
		FieldName:    "",
		FieldType:    "",
		TemplateType: "csv",
		Metadata:     make(map[string]interface{}),
	})
}

// formatValueWithContext formats a value with context information
func (tp *TemplatePopulator) formatValueWithContext(value interface{}, context FormattingContext) string {
	result, err := tp.professionalFormatter.FormatValue(value, context)
	if err != nil {
		// Fallback to simple formatting
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
	return result.DisplayValue
}

// formatValueForExcel formats a value for Excel using professional formatting
func (tp *TemplatePopulator) formatValueForExcel(value interface{}) interface{} {
	return tp.formatValueForExcelWithContext(value, FormattingContext{
		FieldName:    "",
		FieldType:    "",
		TemplateType: "excel",
		Metadata:     make(map[string]interface{}),
	})
}

// formatValueForExcelWithContext formats a value for Excel with context information
func (tp *TemplatePopulator) formatValueForExcelWithContext(value interface{}, context FormattingContext) interface{} {
	// Check if it's a formula first
	if str, ok := value.(string); ok && strings.HasPrefix(str, "=") {
		return str
	}

	// Use professional formatting
	result, err := tp.professionalFormatter.FormatValue(value, context)
	if err != nil {
		// Fallback to Excel native types
		switch v := value.(type) {
		case float64, float32, int, int32, int64:
			return v
		case string:
			// Check if it's a number string
			if num, err := strconv.ParseFloat(v, 64); err == nil {
				return num
			}
			return v
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	// For Excel, return the formatted value or display value based on type
	switch result.FormatType {
	case "currency", "number":
		// Return numeric value for Excel to handle with cell formatting
		if num, ok := result.FormattedValue.(float64); ok {
			return num
		}
		return result.DisplayValue
	case "date", "date_financial":
		// Return time value for Excel to handle with cell formatting
		if t, ok := result.FormattedValue.(time.Time); ok {
			return t
		}
		return result.DisplayValue
	default:
		// Return display value for text
		return result.DisplayValue
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
	case ".txt", ".md":
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

	// Enhanced mappings from metadata field names to template placeholders
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
		// Additional specific field mappings for our template
		"deal_name":         {"To be filled", "Deal Name"},
		"target_company":    {"To be filled", "Name", "Company Name"},
		"deal_type":         {"Acquisition/Merger/Investment", "Type"},
		"deal_value":        {"Amount", "Value"},
		"company_name":      {"Name", "Company Name"},
		"revenue_last_year": {"Amount"},
		"ebitda_last_year":  {"Amount"},
		"ebitda_margin":     {"%"},
		"revenue_growth":    {"%"},
	}

	// Get specific mappings for this field
	if placeholders, exists := mappings[fieldLower]; exists {
		return placeholders
	}

	// Default mappings - try the field name itself and common generic placeholders
	placeholders := []string{
		fieldName,                  // Exact match
		strings.Title(fieldName),   // Title case
		strings.ToLower(fieldName), // Lower case
	}

	// Add context-specific placeholders
	if strings.Contains(fieldLower, "name") || strings.Contains(fieldLower, "company") {
		placeholders = append(placeholders, "To be filled", "Name", "Company Name")
	} else if strings.Contains(fieldLower, "value") || strings.Contains(fieldLower, "price") || strings.Contains(fieldLower, "revenue") || strings.Contains(fieldLower, "ebitda") || strings.Contains(fieldLower, "income") {
		placeholders = append(placeholders, "Amount", "Value", "Price")
	} else if strings.Contains(fieldLower, "date") {
		placeholders = append(placeholders, "Date")
	} else if strings.Contains(fieldLower, "type") {
		placeholders = append(placeholders, "Type", "Acquisition/Merger/Investment")
	} else if strings.Contains(fieldLower, "industry") {
		placeholders = append(placeholders, "Industry")
	} else if strings.Contains(fieldLower, "year") || strings.Contains(fieldLower, "founded") {
		placeholders = append(placeholders, "Year")
	} else if strings.Contains(fieldLower, "location") || strings.Contains(fieldLower, "headquarters") {
		placeholders = append(placeholders, "Location")
	} else if strings.Contains(fieldLower, "employees") || strings.Contains(fieldLower, "number") {
		placeholders = append(placeholders, "Number")
	} else if strings.Contains(fieldLower, "website") || strings.Contains(fieldLower, "url") {
		placeholders = append(placeholders, "URL")
	} else if strings.Contains(fieldLower, "margin") || strings.Contains(fieldLower, "growth") || strings.Contains(fieldLower, "percent") {
		placeholders = append(placeholders, "%")
	}

	// Always try "To be filled" as a last resort
	placeholders = append(placeholders, "To be filled")

	return placeholders
}

// EnhancedFormulaPreservation contains advanced formula preservation information
type EnhancedFormulaPreservation struct {
	*FormulaPreservation
	UpdatedReferences map[string]string `json:"updatedReferences"`
	ValidationResults map[string]bool   `json:"validationResults"`
	PreservationStats FormulaStats      `json:"preservationStats"`
	QualityScore      float64           `json:"qualityScore"`
}

// FormulaStats contains statistics about formula preservation
type FormulaStats struct {
	TotalFormulas     int `json:"totalFormulas"`
	PreservedFormulas int `json:"preservedFormulas"`
	UpdatedReferences int `json:"updatedReferences"`
	BrokenFormulas    int `json:"brokenFormulas"`
	ValidationsPassed int `json:"validationsPassed"`
	ValidationsFailed int `json:"validationsFailed"`
}

// EnhanceFormulaPreservation creates an enhanced formula preservation with validation
func (tp *TemplatePopulator) EnhanceFormulaPreservation(templatePath string, mappedData *MappedData) (*EnhancedFormulaPreservation, error) {
	// Get basic formula preservation
	basicPreservation, err := tp.PreserveFormulas(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze formulas: %w", err)
	}

	enhanced := &EnhancedFormulaPreservation{
		FormulaPreservation: basicPreservation,
		UpdatedReferences:   make(map[string]string),
		ValidationResults:   make(map[string]bool),
		PreservationStats: FormulaStats{
			TotalFormulas: basicPreservation.TotalFormulas,
		},
	}

	// Analyze formula dependencies and update requirements
	for _, formulaInfo := range basicPreservation.FormulaCells {
		// Check if formula references need updating based on mapped data
		updatedFormula, needsUpdate := tp.analyzeFormulaReferences(formulaInfo.Formula, mappedData)
		if needsUpdate {
			enhanced.UpdatedReferences[formulaInfo.Cell] = updatedFormula
			enhanced.PreservationStats.UpdatedReferences++
		}

		// Validate formula syntax and dependencies
		isValid := tp.validateFormulaSyntax(formulaInfo.Formula)
		enhanced.ValidationResults[formulaInfo.Cell] = isValid

		if isValid {
			enhanced.PreservationStats.ValidationsPassed++
			enhanced.PreservationStats.PreservedFormulas++
		} else {
			enhanced.PreservationStats.ValidationsFailed++
			enhanced.PreservationStats.BrokenFormulas++
		}
	}

	// Calculate quality score
	if enhanced.PreservationStats.TotalFormulas > 0 {
		preservationRate := float64(enhanced.PreservationStats.PreservedFormulas) / float64(enhanced.PreservationStats.TotalFormulas)
		validationRate := float64(enhanced.PreservationStats.ValidationsPassed) / float64(enhanced.PreservationStats.TotalFormulas)
		enhanced.QualityScore = (preservationRate * 0.7) + (validationRate * 0.3)
	} else {
		enhanced.QualityScore = 1.0 // No formulas to preserve
	}

	return enhanced, nil
}

// analyzeFormulaReferences checks if formula references need updating
func (tp *TemplatePopulator) analyzeFormulaReferences(formula string, mappedData *MappedData) (string, bool) {
	// Remove the leading = sign for analysis
	formulaContent := strings.TrimPrefix(formula, "=")
	needsUpdate := false

	// Look for cell references that might need updating
	cellRefPattern := regexp.MustCompile(`[A-Z]+[0-9]+`)
	matches := cellRefPattern.FindAllString(formulaContent, -1)

	for _, cellRef := range matches {
		// Check if this cell reference corresponds to a field that was updated
		// This is a simplified analysis - in practice, you'd need more sophisticated logic
		for fieldName := range mappedData.Fields {
			// If the field name suggests it might affect this cell reference
			if tp.mightAffectCellReference(fieldName, cellRef) {
				// For now, we don't automatically update references
				// but we flag that the formula might need attention
				needsUpdate = true
				break
			}
		}
	}

	return "=" + formulaContent, needsUpdate
}

// mightAffectCellReference determines if a field might affect a cell reference
func (tp *TemplatePopulator) mightAffectCellReference(fieldName, cellRef string) bool {
	// This is a simplified heuristic - in practice, you'd need template-specific logic
	fieldLower := strings.ToLower(fieldName)

	// If it's a financial field and the cell reference is in a typical calculation area
	if strings.Contains(fieldLower, "revenue") || strings.Contains(fieldLower, "ebitda") ||
		strings.Contains(fieldLower, "value") || strings.Contains(fieldLower, "amount") {
		return true
	}

	return false
}

// validateFormulaSyntax performs basic formula syntax validation
func (tp *TemplatePopulator) validateFormulaSyntax(formula string) bool {
	// Remove the leading = sign
	formulaContent := strings.TrimPrefix(formula, "=")

	// Basic syntax checks
	if formulaContent == "" {
		return false
	}

	// Check for balanced parentheses
	if !tp.hasBalancedParentheses(formulaContent) {
		return false
	}

	// Check for valid function names (basic check)
	if strings.Contains(formulaContent, "(") {
		// Extract function names and validate them
		funcPattern := regexp.MustCompile(`[A-Z]+\(`)
		functions := funcPattern.FindAllString(formulaContent, -1)

		for _, fn := range functions {
			funcName := strings.TrimSuffix(fn, "(")
			if !tp.isValidExcelFunction(funcName) {
				return false
			}
		}
	}

	return true
}

// hasBalancedParentheses checks if parentheses are balanced in a formula
func (tp *TemplatePopulator) hasBalancedParentheses(formula string) bool {
	count := 0
	for _, char := range formula {
		switch char {
		case '(':
			count++
		case ')':
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

// isValidExcelFunction checks if a function name is a valid Excel function
func (tp *TemplatePopulator) isValidExcelFunction(funcName string) bool {
	validFunctions := map[string]bool{
		"SUM": true, "AVERAGE": true, "COUNT": true, "MIN": true, "MAX": true,
		"IF": true, "AND": true, "OR": true, "NOT": true,
		"VLOOKUP": true, "HLOOKUP": true, "INDEX": true, "MATCH": true,
		"TODAY": true, "NOW": true, "DATE": true, "TIME": true,
		"CONCATENATE": true, "LEFT": true, "RIGHT": true, "MID": true,
		"NPV": true, "IRR": true, "PMT": true, "PV": true, "FV": true,
		"ROUND": true, "ROUNDUP": true, "ROUNDDOWN": true,
		"ABS": true, "SQRT": true, "POWER": true, "EXP": true, "LN": true,
		"COUNTA": true, "COUNTIF": true, "SUMIF": true, "AVERAGEIF": true,
	}

	return validFunctions[strings.ToUpper(funcName)]
}

// PopulateTemplateWithEnhancedFormulas populates a template with enhanced formula preservation
func (tp *TemplatePopulator) PopulateTemplateWithEnhancedFormulas(templatePath string, mappedData *MappedData, outputPath string) (*EnhancedFormulaPreservation, error) {
	// First, analyze formulas
	enhanced, err := tp.EnhanceFormulaPreservation(templatePath, mappedData)
	if err != nil {
		return nil, fmt.Errorf("failed to enhance formula preservation: %w", err)
	}

	// Populate the template normally
	err = tp.PopulateTemplate(templatePath, mappedData, outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to populate template: %w", err)
	}

	// Validate that formulas were preserved correctly
	err = tp.ValidatePopulatedTemplate(outputPath, enhanced.FormulaPreservation)
	if err != nil {
		// Update preservation stats
		enhanced.PreservationStats.BrokenFormulas++
		enhanced.PreservationStats.PreservedFormulas--

		// Recalculate quality score
		if enhanced.PreservationStats.TotalFormulas > 0 {
			preservationRate := float64(enhanced.PreservationStats.PreservedFormulas) / float64(enhanced.PreservationStats.TotalFormulas)
			validationRate := float64(enhanced.PreservationStats.ValidationsPassed) / float64(enhanced.PreservationStats.TotalFormulas)
			enhanced.QualityScore = (preservationRate * 0.7) + (validationRate * 0.3)
		}

		return enhanced, fmt.Errorf("formula preservation validation failed: %w", err)
	}

	return enhanced, nil
}
