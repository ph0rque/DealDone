package app

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ProfessionalFormatter handles advanced formatting for template population
type ProfessionalFormatter struct {
	currencyConfig CurrencyConfig
	dateConfig     DateConfig
	textConfig     TextConfig
}

// CurrencyConfig defines currency formatting options
type CurrencyConfig struct {
	DefaultCurrency    string            `json:"defaultCurrency"`
	Locale             string            `json:"locale"`
	UseSymbol          bool              `json:"useSymbol"`
	DecimalPlaces      int               `json:"decimalPlaces"`
	ThousandSeparator  string            `json:"thousandSeparator"`
	DecimalSeparator   string            `json:"decimalSeparator"`
	CurrencySymbols    map[string]string `json:"currencySymbols"`
	CurrencyPlacements map[string]string `json:"currencyPlacements"` // "before" or "after"
}

// DateConfig defines date formatting options
type DateConfig struct {
	DefaultFormat   string            `json:"defaultFormat"`
	Locale          string            `json:"locale"`
	BusinessFormats map[string]string `json:"businessFormats"`
	TimeZone        string            `json:"timeZone"`
}

// TextConfig defines business text formatting options
type TextConfig struct {
	CompanyNameFormat   string            `json:"companyNameFormat"`
	PersonNameFormat    string            `json:"personNameFormat"`
	AbbreviationRules   map[string]string `json:"abbreviationRules"`
	CapitalizationRules map[string]string `json:"capitalizationRules"`
	BusinessTerms       map[string]string `json:"businessTerms"`
	IndustryTerms       map[string]string `json:"industryTerms"`
}

// FormattingContext provides context for formatting decisions
type FormattingContext struct {
	FieldName    string                 `json:"fieldName"`
	FieldType    string                 `json:"fieldType"`
	TemplateType string                 `json:"templateType"`
	BusinessArea string                 `json:"businessArea"`
	DealType     string                 `json:"dealType"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// FormattingResult contains the formatted value and metadata
type FormattingResult struct {
	FormattedValue interface{}            `json:"formattedValue"`
	DisplayValue   string                 `json:"displayValue"`
	FormatType     string                 `json:"formatType"`
	Confidence     float64                `json:"confidence"`
	Metadata       map[string]interface{} `json:"metadata"`
	AppliedRules   []string               `json:"appliedRules"`
}

// NewProfessionalFormatter creates a new professional formatter with default settings
func NewProfessionalFormatter() *ProfessionalFormatter {
	return &ProfessionalFormatter{
		currencyConfig: CurrencyConfig{
			DefaultCurrency:   "USD",
			Locale:            "en-US",
			UseSymbol:         true,
			DecimalPlaces:     0, // Default to whole numbers for M&A
			ThousandSeparator: ",",
			DecimalSeparator:  ".",
			CurrencySymbols: map[string]string{
				"USD": "$", "EUR": "€", "GBP": "£", "JPY": "¥", "CAD": "C$", "AUD": "A$",
				"CHF": "CHF", "CNY": "¥", "INR": "₹", "KRW": "₩", "BRL": "R$", "MXN": "$",
			},
			CurrencyPlacements: map[string]string{
				"USD": "before", "EUR": "before", "GBP": "before", "JPY": "before",
				"CAD": "before", "AUD": "before", "CHF": "after", "CNY": "before",
				"INR": "before", "KRW": "before", "BRL": "before", "MXN": "before",
			},
		},
		dateConfig: DateConfig{
			DefaultFormat: "January 2, 2006",
			Locale:        "en-US",
			BusinessFormats: map[string]string{
				"contract":  "January 2, 2006",
				"financial": "Q1 2006",
				"brief":     "Jan 2006",
				"iso":       "2006-01-02",
				"us":        "01/02/2006",
				"european":  "02/01/2006",
			},
			TimeZone: "UTC",
		},
		textConfig: TextConfig{
			CompanyNameFormat: "title",
			PersonNameFormat:  "title",
			AbbreviationRules: map[string]string{
				"Incorporated":              "Inc.",
				"Corporation":               "Corp.",
				"Limited":                   "Ltd.",
				"Limited Liability Company": "LLC",
				"Chief Executive Officer":   "CEO",
				"Chief Financial Officer":   "CFO",
				"Chief Operating Officer":   "COO",
				"Chief Technology Officer":  "CTO",
				"Mergers and Acquisitions":  "M&A",
				"Private Equity":            "PE",
				"Venture Capital":           "VC",
			},
			CapitalizationRules: map[string]string{
				"company_name": "title",
				"person_name":  "title",
				"industry":     "title",
				"deal_type":    "title",
			},
			BusinessTerms: map[string]string{
				"ebitda": "EBITDA",
				"roi":    "ROI",
				"irr":    "IRR",
				"npv":    "NPV",
				"dcf":    "DCF",
				"lbo":    "LBO",
				"ipo":    "IPO",
				"saas":   "SaaS",
				"b2b":    "B2B",
				"b2c":    "B2C",
			},
			IndustryTerms: map[string]string{
				"fintech":    "FinTech",
				"healthtech": "HealthTech",
				"edtech":     "EdTech",
				"proptech":   "PropTech",
				"insurtech":  "InsurTech",
			},
		},
	}
}

// FormatValue formats a value based on context and type detection
func (pf *ProfessionalFormatter) FormatValue(value interface{}, context FormattingContext) (*FormattingResult, error) {
	if value == nil {
		return &FormattingResult{
			FormattedValue: "",
			DisplayValue:   "",
			FormatType:     "empty",
			Confidence:     1.0,
		}, nil
	}

	// Detect value type and apply appropriate formatting
	valueStr := fmt.Sprintf("%v", value)

	// Try currency formatting first
	if isCurrencyField(context.FieldName) || isCurrencyValue(valueStr) {
		return pf.FormatCurrency(value, context)
	}

	// Try date formatting
	if isDateField(context.FieldName) || isDateValue(valueStr) {
		return pf.FormatDate(value, context)
	}

	// Try business text formatting
	if isTextValue(value) {
		return pf.FormatBusinessText(value, context)
	}

	// Default numeric formatting
	if isNumericValue(valueStr) {
		return pf.FormatNumber(value, context)
	}

	// Fallback to string formatting
	return &FormattingResult{
		FormattedValue: value,
		DisplayValue:   valueStr,
		FormatType:     "string",
		Confidence:     0.5,
	}, nil
}

// FormatCurrency formats currency values with proper symbols and separators
func (pf *ProfessionalFormatter) FormatCurrency(value interface{}, context FormattingContext) (*FormattingResult, error) {
	// Convert to float64
	var amount float64
	var err error

	switch v := value.(type) {
	case float64:
		amount = v
	case float32:
		amount = float64(v)
	case int:
		amount = float64(v)
	case int64:
		amount = float64(v)
	case string:
		// Clean the string first
		cleaned := cleanCurrencyString(v)
		amount, err = strconv.ParseFloat(cleaned, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse currency value: %s", v)
		}
	default:
		return nil, fmt.Errorf("unsupported currency value type: %T", value)
	}

	// Determine currency from context or default
	currency := pf.currencyConfig.DefaultCurrency
	if contextCurrency, exists := context.Metadata["currency"].(string); exists {
		currency = contextCurrency
	}

	// Get currency symbol
	symbol := pf.currencyConfig.CurrencySymbols[currency]
	if symbol == "" {
		symbol = currency + " "
	}

	// Format the number with separators
	formattedNumber := pf.formatNumberWithSeparators(amount, pf.currencyConfig.DecimalPlaces)

	// Apply currency symbol placement
	var displayValue string
	placement := pf.currencyConfig.CurrencyPlacements[currency]
	if placement == "after" {
		displayValue = formattedNumber + " " + symbol
	} else {
		displayValue = symbol + formattedNumber
	}

	appliedRules := []string{"currency_formatting", "thousand_separators"}
	if pf.currencyConfig.UseSymbol {
		appliedRules = append(appliedRules, "currency_symbol")
	}

	return &FormattingResult{
		FormattedValue: amount,
		DisplayValue:   displayValue,
		FormatType:     "currency",
		Confidence:     0.95,
		Metadata: map[string]interface{}{
			"currency":      currency,
			"symbol":        symbol,
			"originalValue": value,
		},
		AppliedRules: appliedRules,
	}, nil
}

// FormatDate formats date values in business-appropriate formats
func (pf *ProfessionalFormatter) FormatDate(value interface{}, context FormattingContext) (*FormattingResult, error) {
	var t time.Time
	var err error

	switch v := value.(type) {
	case time.Time:
		t = v
	case string:
		t, err = parseFlexibleDate(v)
		if err != nil {
			return nil, fmt.Errorf("cannot parse date value: %s", v)
		}
	default:
		return nil, fmt.Errorf("unsupported date value type: %T", value)
	}

	// Determine format from context
	formatType := "default"
	if contextFormat, exists := context.Metadata["dateFormat"].(string); exists {
		formatType = contextFormat
	} else if strings.Contains(strings.ToLower(context.FieldName), "quarter") {
		formatType = "financial"
	} else if strings.Contains(strings.ToLower(context.FieldName), "contract") {
		formatType = "contract"
	}

	// Get format string
	formatStr := pf.dateConfig.DefaultFormat
	if businessFormat, exists := pf.dateConfig.BusinessFormats[formatType]; exists {
		formatStr = businessFormat
	}

	// Handle special financial formats
	if formatType == "financial" {
		displayValue := formatFinancialDate(t)
		return &FormattingResult{
			FormattedValue: t,
			DisplayValue:   displayValue,
			FormatType:     "date_financial",
			Confidence:     0.9,
			Metadata: map[string]interface{}{
				"originalValue": value,
				"formatType":    formatType,
			},
			AppliedRules: []string{"date_formatting", "financial_quarters"},
		}, nil
	}

	displayValue := t.Format(formatStr)

	return &FormattingResult{
		FormattedValue: t,
		DisplayValue:   displayValue,
		FormatType:     "date",
		Confidence:     0.9,
		Metadata: map[string]interface{}{
			"originalValue": value,
			"formatType":    formatType,
		},
		AppliedRules: []string{"date_formatting"},
	}, nil
}

// FormatBusinessText formats text with proper capitalization and business terminology
func (pf *ProfessionalFormatter) FormatBusinessText(value interface{}, context FormattingContext) (*FormattingResult, error) {
	text := strings.TrimSpace(fmt.Sprintf("%v", value))
	if text == "" {
		return &FormattingResult{
			FormattedValue: "",
			DisplayValue:   "",
			FormatType:     "text",
			Confidence:     1.0,
		}, nil
	}

	appliedRules := []string{}

	// Apply business term standardization
	for term, standardized := range pf.textConfig.BusinessTerms {
		if strings.Contains(strings.ToLower(text), term) {
			text = replaceWordIgnoreCase(text, term, standardized)
			appliedRules = append(appliedRules, "business_terms")
		}
	}

	// Apply industry term standardization
	for term, standardized := range pf.textConfig.IndustryTerms {
		if strings.Contains(strings.ToLower(text), term) {
			text = replaceWordIgnoreCase(text, term, standardized)
			appliedRules = append(appliedRules, "industry_terms")
		}
	}

	// Apply capitalization rules based on field type
	fieldType := detectFieldType(context.FieldName)
	if rule, exists := pf.textConfig.CapitalizationRules[fieldType]; exists {
		switch rule {
		case "title":
			text = strings.Title(strings.ToLower(text))
			appliedRules = append(appliedRules, "title_case")
		case "upper":
			text = strings.ToUpper(text)
			appliedRules = append(appliedRules, "upper_case")
		case "lower":
			text = strings.ToLower(text)
			appliedRules = append(appliedRules, "lower_case")
		}
	}

	// Apply abbreviation rules for company names
	if fieldType == "company_name" {
		for full, abbrev := range pf.textConfig.AbbreviationRules {
			if strings.Contains(text, full) {
				text = strings.ReplaceAll(text, full, abbrev)
				appliedRules = append(appliedRules, "abbreviations")
			}
		}
	}

	// Clean up extra whitespace
	text = cleanWhitespace(text)

	return &FormattingResult{
		FormattedValue: text,
		DisplayValue:   text,
		FormatType:     "text_business",
		Confidence:     0.8,
		Metadata: map[string]interface{}{
			"originalValue": value,
			"fieldType":     fieldType,
		},
		AppliedRules: appliedRules,
	}, nil
}

// FormatNumber formats numeric values with appropriate precision
func (pf *ProfessionalFormatter) FormatNumber(value interface{}, context FormattingContext) (*FormattingResult, error) {
	var num float64
	var err error

	switch v := value.(type) {
	case float64:
		num = v
	case float32:
		num = float64(v)
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case string:
		num, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse numeric value: %s", v)
		}
	default:
		return nil, fmt.Errorf("unsupported numeric value type: %T", value)
	}

	// Determine decimal places based on context
	decimalPlaces := 0
	if strings.Contains(strings.ToLower(context.FieldName), "percentage") ||
		strings.Contains(strings.ToLower(context.FieldName), "rate") {
		decimalPlaces = 2
	} else if math.Abs(num) < 10 {
		decimalPlaces = 2
	}

	displayValue := pf.formatNumberWithSeparators(num, decimalPlaces)

	return &FormattingResult{
		FormattedValue: num,
		DisplayValue:   displayValue,
		FormatType:     "number",
		Confidence:     0.9,
		Metadata: map[string]interface{}{
			"originalValue": value,
			"decimalPlaces": decimalPlaces,
		},
		AppliedRules: []string{"number_formatting"},
	}, nil
}

// Helper functions

func (pf *ProfessionalFormatter) formatNumberWithSeparators(num float64, decimalPlaces int) string {
	// Handle negative numbers
	isNegative := num < 0
	if isNegative {
		num = -num
	}

	// Format with specified decimal places
	formatStr := fmt.Sprintf("%%.%df", decimalPlaces)
	formatted := fmt.Sprintf(formatStr, num)

	// Split into integer and decimal parts
	parts := strings.Split(formatted, ".")
	integerPart := parts[0]

	// Add thousand separators
	if len(integerPart) > 3 {
		integerPart = addThousandSeparators(integerPart, pf.currencyConfig.ThousandSeparator)
	}

	// Reconstruct the number
	result := integerPart
	if decimalPlaces > 0 && len(parts) > 1 {
		result += pf.currencyConfig.DecimalSeparator + parts[1]
	}

	// Add negative sign if needed
	if isNegative {
		result = "-" + result
	}

	return result
}

func addThousandSeparators(s, separator string) string {
	n := len(s)
	if n <= 3 {
		return s
	}

	result := ""
	for i, digit := range s {
		if i > 0 && (n-i)%3 == 0 {
			result += separator
		}
		result += string(digit)
	}

	return result
}

func cleanCurrencyString(s string) string {
	// Remove currency symbols and clean up
	cleaned := s

	// Remove common currency symbols
	symbols := []string{"$", "€", "£", "¥", "₹", "₩", "CHF", "USD", "EUR", "GBP", "JPY", "CAD", "AUD"}
	for _, symbol := range symbols {
		cleaned = strings.ReplaceAll(cleaned, symbol, "")
	}

	// Remove spaces and commas
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")

	return cleaned
}

func parseFlexibleDate(s string) (time.Time, error) {
	// Try multiple date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"January 2, 2006",
		"Jan 2, 2006",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"Q1 2006",
		"2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	// Handle quarter formats
	if strings.HasPrefix(s, "Q") && len(s) >= 6 {
		quarter := s[1:2]
		year := s[3:]
		if q, err := strconv.Atoi(quarter); err == nil && q >= 1 && q <= 4 {
			if y, err := strconv.Atoi(year); err == nil {
				month := (q-1)*3 + 1
				return time.Date(y, time.Month(month), 1, 0, 0, 0, 0, time.UTC), nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}

func formatFinancialDate(t time.Time) string {
	year := t.Year()
	month := t.Month()
	quarter := ((int(month) - 1) / 3) + 1
	return fmt.Sprintf("Q%d %d", quarter, year)
}

func replaceWordIgnoreCase(text, old, new string) string {
	re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(old) + `\b`)
	return re.ReplaceAllString(text, new)
}

func cleanWhitespace(s string) string {
	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

func detectFieldType(fieldName string) string {
	fieldLower := strings.ToLower(fieldName)

	if strings.Contains(fieldLower, "company") || strings.Contains(fieldLower, "target") {
		return "company_name"
	}
	if strings.Contains(fieldLower, "name") && (strings.Contains(fieldLower, "ceo") ||
		strings.Contains(fieldLower, "cfo") || strings.Contains(fieldLower, "president")) {
		return "person_name"
	}
	if strings.Contains(fieldLower, "industry") || strings.Contains(fieldLower, "sector") {
		return "industry"
	}
	if strings.Contains(fieldLower, "deal") && strings.Contains(fieldLower, "type") {
		return "deal_type"
	}

	return "general"
}

// Type detection functions

func isCurrencyField(fieldName string) bool {
	fieldLower := strings.ToLower(fieldName)
	currencyKeywords := []string{"price", "value", "amount", "revenue", "ebitda", "income", "cost", "fee", "payment"}

	for _, keyword := range currencyKeywords {
		if strings.Contains(fieldLower, keyword) {
			return true
		}
	}
	return false
}

func isCurrencyValue(value string) bool {
	// Check for currency symbols or patterns
	currencyPatterns := []string{"$", "€", "£", "¥", "USD", "EUR", "GBP", "JPY"}

	for _, pattern := range currencyPatterns {
		if strings.Contains(value, pattern) {
			return true
		}
	}

	// Check for large numbers that might be currency
	if num, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", ""), 64); err == nil {
		return num > 1000 // Assume large numbers are currency
	}

	return false
}

func isDateField(fieldName string) bool {
	fieldLower := strings.ToLower(fieldName)
	dateKeywords := []string{"date", "time", "year", "month", "quarter", "founded", "established", "closing"}

	for _, keyword := range dateKeywords {
		if strings.Contains(fieldLower, keyword) {
			return true
		}
	}
	return false
}

func isDateValue(value string) bool {
	// Try to parse as date
	_, err := parseFlexibleDate(value)
	return err == nil
}

func isTextValue(value interface{}) bool {
	_, ok := value.(string)
	return ok
}

func isNumericValue(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}
