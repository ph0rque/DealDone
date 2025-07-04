package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// ClaudeProvider implements AI service using Anthropic's Claude API
type ClaudeProvider struct {
	apiKey     string
	model      string
	endpoint   string
	httpClient *http.Client
	stats      *AIUsageStats
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(apiKey, model string) AIServiceInterface {
	if model == "" {
		model = "claude-3-opus-20240229"
	}

	return &ClaudeProvider{
		apiKey:   apiKey,
		model:    model,
		endpoint: "https://api.anthropic.com/v1",
		httpClient: &http.Client{
			Timeout: time.Minute * 2,
		},
		stats: &AIUsageStats{
			LastReset: time.Now(),
		},
	}
}

// Claude API types
type claudeRequest struct {
	Model       string          `json:"model"`
	Messages    []claudeMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	System      string          `json:"system,omitempty"`
	Temperature float64         `json:"temperature"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Role       string          `json:"role"`
	Content    []claudeContent `json:"content"`
	Model      string          `json:"model"`
	StopReason string          `json:"stop_reason"`
	Usage      claudeUsage     `json:"usage"`
	Error      *claudeError    `json:"error,omitempty"`
}

type claudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type claudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ClassifyDocument uses Claude to classify a document
func (cp *ClaudeProvider) ClassifyDocument(ctx context.Context, content string, metadata map[string]interface{}) (*AIClassificationResult, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	// Truncate content if too long
	if len(content) > 10000 {
		content = content[:10000] + "..."
	}

	systemPrompt := `You are an expert document analyst specializing in M&A due diligence. 
Analyze the provided document and classify it into one of these categories: legal, financial, or general.
Provide your response in JSON format with the following structure:
{
  "documentType": "legal|financial|general",
  "confidence": 0.0-1.0,
  "keywords": ["keyword1", "keyword2", ...],
  "categories": ["category1", "category2"],
  "language": "en",
  "summary": "Brief summary of the document"
}`

	userPrompt := fmt.Sprintf("Analyze this document and respond with JSON:\n\n%s", content)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	// Parse JSON response
	var result AIClassificationResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	result.Metadata = metadata
	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)

	return &result, nil
}

// ExtractFinancialData extracts financial information using Claude
func (cp *ClaudeProvider) ExtractFinancialData(ctx context.Context, content string) (*FinancialAnalysis, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	systemPrompt := `You are a financial analyst expert. Extract financial data from the provided document.
Provide your response in JSON format with the following structure:
{
  "revenue": 0.0,
  "ebitda": 0.0,
  "netIncome": 0.0,
  "totalAssets": 0.0,
  "totalLiabilities": 0.0,
  "cashFlow": 0.0,
  "grossMargin": 0.0,
  "operatingMargin": 0.0,
  "confidence": 0.0-1.0,
  "period": "Q1 2024|FY 2023|etc",
  "currency": "USD|EUR|etc",
  "dataPoints": {"key": value},
  "warnings": ["warning1", "warning2"]
}
Use 0 for any values not found. Include warnings about missing or unclear data.`

	userPrompt := fmt.Sprintf("Extract financial data from this document and respond with JSON:\n\n%s", content)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result FinancialAnalysis
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse financial data: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// AnalyzeRisks performs risk assessment using Claude
func (cp *ClaudeProvider) AnalyzeRisks(ctx context.Context, content string, docType string) (*RiskAnalysis, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	systemPrompt := fmt.Sprintf(`You are a risk assessment expert for M&A due diligence. 
Analyze the %s document for potential risks.
Provide your response in JSON format with the following structure:
{
  "overallRiskScore": 0.0-1.0,
  "riskCategories": [
    {
      "category": "legal|financial|operational|regulatory|market|technical",
      "description": "Description of the risk",
      "severity": "low|medium|high|critical",
      "score": 0.0-1.0,
      "mitigation": "Suggested mitigation"
    }
  ],
  "recommendations": ["recommendation1", "recommendation2"],
  "criticalIssues": ["issue1", "issue2"],
  "confidence": 0.0-1.0
}`, docType)

	userPrompt := fmt.Sprintf("Analyze risks in this document and respond with JSON:\n\n%s", content)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result RiskAnalysis
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse risk analysis: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// GenerateInsights creates insights using Claude
func (cp *ClaudeProvider) GenerateInsights(ctx context.Context, content string, docType string) (*DocumentInsights, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	systemPrompt := `You are an M&A expert providing strategic insights. 
Analyze the document and provide actionable insights.
Provide your response in JSON format with the following structure:
{
  "keyPoints": ["point1", "point2"],
  "opportunities": ["opportunity1", "opportunity2"],
  "concerns": ["concern1", "concern2"],
  "actionItems": ["action1", "action2"],
  "marketContext": "Brief market context",
  "competitiveInfo": {"key": "value"},
  "confidence": 0.0-1.0
}`

	userPrompt := fmt.Sprintf("Generate insights for this %s document and respond with JSON:\n\n%s", docType, content)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result DocumentInsights
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse insights: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// ExtractEntities extracts named entities using Claude
func (cp *ClaudeProvider) ExtractEntities(ctx context.Context, content string) (*EntityExtraction, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	systemPrompt := `Extract named entities from the document.
Provide your response in JSON format with the following structure:
{
  "people": [{"text": "name", "type": "person", "confidence": 0.9, "context": "CEO"}],
  "organizations": [{"text": "name", "type": "organization", "confidence": 0.9, "context": "buyer"}],
  "locations": [{"text": "name", "type": "location", "confidence": 0.9, "context": "headquarters"}],
  "dates": [{"text": "date", "type": "date", "confidence": 0.9, "context": "closing date"}],
  "monetaryValues": [{"text": "$1M", "type": "money", "confidence": 0.9, "context": "purchase price"}],
  "percentages": [{"text": "15%", "type": "percentage", "confidence": 0.9, "context": "stake"}],
  "products": [{"text": "name", "type": "product", "confidence": 0.9, "context": "main product"}]
}`

	userPrompt := fmt.Sprintf("Extract entities from this text and respond with JSON:\n\n%s", content)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result EntityExtraction
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse entities: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// makeRequest makes an API request to Claude
func (cp *ClaudeProvider) makeRequest(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := []claudeMessage{
		{Role: "user", Content: userPrompt},
	}

	reqBody := claudeRequest{
		Model:       cp.model,
		Messages:    messages,
		MaxTokens:   4096,
		Temperature: 0.3, // Lower temperature for more consistent results
		System:      systemPrompt,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cp.endpoint+"/messages", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cp.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := cp.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp claudeResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
			return "", fmt.Errorf("Claude API error: %s", errResp.Error.Message)
		}
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp claudeResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("no response from Claude")
	}

	// Update token usage
	atomic.AddInt64(&cp.stats.TotalTokens, int64(apiResp.Usage.InputTokens+apiResp.Usage.OutputTokens))

	// Claude returns content as array, get the first text content
	for _, content := range apiResp.Content {
		if content.Type == "text" {
			return content.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in Claude response")
}

// GetProvider returns the provider name
func (cp *ClaudeProvider) GetProvider() Provider {
	return ProviderClaude
}

// IsAvailable checks if the service is configured
func (cp *ClaudeProvider) IsAvailable() bool {
	return cp.apiKey != ""
}

// GetUsage returns usage statistics
func (cp *ClaudeProvider) GetUsage() *AIUsageStats {
	return cp.stats
}

// NEW METHODS FOR ENHANCED TEMPLATE PROCESSING

// ExtractDocumentFields extracts structured field data from documents for template mapping
func (cp *ClaudeProvider) ExtractDocumentFields(ctx context.Context, content string, documentType string, templateContext map[string]interface{}) (*DocumentFieldExtraction, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	// Truncate content if too long
	if len(content) > 15000 {
		content = content[:15000] + "..."
	}

	systemPrompt := `You are an expert document field extraction specialist for M&A due diligence. 
Extract structured field data from the provided document that would be useful for populating templates.
Focus on key financial metrics, dates, names, monetary values, percentages, and other important data points.

Provide your response in JSON format with the following structure:
{
  "fields": {
    "fieldName": "extractedValue",
    "company_name": "AquaFlow Technologies",
    "revenue": 25000000,
    "ebitda": 8500000,
    "closing_date": "2024-12-31",
    "purchase_price": 125000000
  },
  "fieldTypes": {
    "fieldName": "type",
    "company_name": "text",
    "revenue": "currency",
    "ebitda": "currency", 
    "closing_date": "date",
    "purchase_price": "currency"
  },
  "confidence": 0.85,
  "warnings": ["Any extraction warnings"],
  "metadata": {
    "extraction_method": "ai_analysis",
    "document_sections_analyzed": ["financial_summary", "terms"]
  },
  "source": "document_content"
}

Extract as many relevant fields as possible. Use descriptive field names. For monetary values, extract the raw number (without currency symbols). For dates, use ISO format when possible.`

	userPrompt := fmt.Sprintf("Extract structured fields from this %s document:\n\n%s", documentType, content)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result DocumentFieldExtraction
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse field extraction response: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// MapFieldsToTemplate maps extracted document fields to template field requirements
func (cp *ClaudeProvider) MapFieldsToTemplate(ctx context.Context, extractedFields map[string]interface{}, templateFields []TemplateField, mappingContext map[string]interface{}) (*FieldMappingResult, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	// Prepare template fields info for the prompt
	templateFieldsJSON, _ := json.Marshal(templateFields)
	extractedFieldsJSON, _ := json.Marshal(extractedFields)

	systemPrompt := `You are an expert field mapping specialist for M&A document processing.
Map extracted document fields to template field requirements based on semantic similarity and data type compatibility.

Provide your response in JSON format with the following structure:
{
  "mappings": [
    {
      "documentField": "company_name",
      "templateField": "target_company",
      "value": "AquaFlow Technologies",
      "confidence": 0.95,
      "transformApplied": "none"
    },
    {
      "documentField": "revenue",
      "templateField": "annual_revenue",
      "value": 25000000,
      "confidence": 0.90,
      "transformApplied": "currency_formatting"
    }
  ],
  "unmappedFields": ["field_not_mapped"],
  "missingFields": ["required_template_field_not_found"],
  "confidence": 0.85,
  "suggestions": [
    {
      "documentField": "alternative_field",
      "templateField": "target_field",
      "confidence": 0.70,
      "reason": "Semantic similarity but lower confidence"
    }
  ],
  "metadata": {
    "mapping_method": "ai_semantic_analysis",
    "total_mappings": 5
  }
}

Focus on creating high-confidence mappings. Consider field names, data types, and semantic meaning.`

	userPrompt := fmt.Sprintf("Map these extracted fields:\n%s\n\nTo these template fields:\n%s",
		string(extractedFieldsJSON), string(templateFieldsJSON))

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result FieldMappingResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse field mapping response: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// FormatFieldValue formats a raw field value according to template requirements
func (cp *ClaudeProvider) FormatFieldValue(ctx context.Context, rawValue interface{}, fieldType string, formatRequirements map[string]interface{}) (*FormattedFieldValue, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	formatReqJSON, _ := json.Marshal(formatRequirements)

	systemPrompt := `You are an expert data formatter for M&A templates.
Format the provided raw value according to the specified field type and format requirements.

Provide your response in JSON format with the following structure:
{
  "formattedValue": "$25,000,000",
  "originalValue": 25000000,
  "formatApplied": "currency_usd_with_commas",
  "confidence": 0.95,
  "warnings": ["Any formatting warnings"],
  "metadata": {
    "format_method": "ai_formatting",
    "locale": "en_US"
  }
}

Common formatting patterns:
- currency: Add currency symbol, commas, proper decimal places
- date: Convert to readable format (e.g., "December 31, 2024")
- percentage: Add % symbol, proper decimal places
- number: Add commas for thousands separator
- text: Clean and capitalize appropriately`

	userPrompt := fmt.Sprintf("Format this value: %v\nField type: %s\nFormat requirements: %s",
		rawValue, fieldType, string(formatReqJSON))

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result FormattedFieldValue
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse field formatting response: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// ValidateTemplateData validates that mapped data meets template requirements
func (cp *ClaudeProvider) ValidateTemplateData(ctx context.Context, templateData map[string]interface{}, validationRules []ValidationRule) (*ValidationResult, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	templateDataJSON, _ := json.Marshal(templateData)
	validationRulesJSON, _ := json.Marshal(validationRules)

	systemPrompt := `You are an expert data validation specialist for M&A templates.
Validate the provided template data against the specified validation rules.

Provide your response in JSON format with the following structure:
{
  "isValid": true,
  "errors": [
    {
      "field": "field_name",
      "rule": "required",
      "message": "Field is required but missing",
      "value": null
    }
  ],
  "warnings": [
    {
      "field": "field_name", 
      "message": "Value seems unusually high",
      "value": 1000000000
    }
  ],
  "summary": "Validation completed with 2 errors and 1 warning",
  "metadata": {
    "validation_method": "ai_analysis",
    "total_fields_validated": 15,
    "validation_time": "2024-01-01T12:00:00Z"
  }
}

Validation types:
- required: Field must have a value
- format: Value must match expected format
- range: Numeric value must be within specified range
- pattern: Text must match regex pattern
- type: Value must be of correct data type`

	userPrompt := fmt.Sprintf("Validate this template data:\n%s\n\nUsing these validation rules:\n%s",
		string(templateDataJSON), string(validationRulesJSON))

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result ValidationResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse validation response: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// ENHANCED ENTITY EXTRACTION METHODS FOR TASK 1.3

// ExtractCompanyAndDealNames extracts company names and deal names with enhanced confidence scoring
func (cp *ClaudeProvider) ExtractCompanyAndDealNames(ctx context.Context, content string, documentType string) (*CompanyDealExtraction, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	// Truncate content if too long
	if len(content) > 15000 {
		content = content[:15000] + "..."
	}

	systemPrompt := `You are an expert M&A document analyst specializing in company and deal identification.
Extract company names and deal names from the provided document with enhanced confidence scoring and validation.

Focus on:
- Company names with their roles (buyer, seller, target, advisor)
- Deal names and types (acquisition, merger, investment)
- Industry classification and location when available
- Confidence scoring based on context clarity

Provide your response in JSON format with the following structure:
{
  "companies": [
    {
      "name": "AquaFlow Technologies Inc.",
      "role": "target",
      "confidence": 0.95,
      "context": "target company being acquired",
      "industry": "water treatment technology",
      "location": "California, USA",
      "metadata": {"source_section": "executive_summary"},
      "validated": false
    }
  ],
  "dealNames": [
    {
      "name": "Project Neptune",
      "type": "acquisition",
      "status": "pending",
      "confidence": 0.90,
      "context": "code name for the acquisition",
      "metadata": {"source_section": "deal_overview"}
    }
  ],
  "confidence": 0.92,
  "metadata": {
    "extraction_method": "ai_analysis",
    "document_sections_analyzed": ["executive_summary", "deal_overview"]
  },
  "warnings": ["Any extraction warnings or ambiguities"]
}`

	userPrompt := fmt.Sprintf("Extract company and deal names from this %s document and respond with JSON:\n\n%s", documentType, content)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result CompanyDealExtraction
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse company and deal extraction: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// ExtractFinancialMetrics extracts financial metrics with enhanced validation
func (cp *ClaudeProvider) ExtractFinancialMetrics(ctx context.Context, content string, documentType string) (*FinancialMetricsExtraction, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	// Truncate content if too long
	if len(content) > 15000 {
		content = content[:15000] + "..."
	}

	systemPrompt := `You are a financial analyst expert specializing in M&A document analysis.
Extract comprehensive financial metrics with enhanced validation and confidence scoring.

Focus on:
- Revenue, EBITDA, Net Income, Assets, Liabilities, Cash Flow
- Deal value and financial multiples
- Currency identification and period specification
- Validation of financial consistency
- Confidence scoring based on data clarity

Provide your response in JSON format with the following structure:
{
  "revenue": {
    "value": 25000000,
    "confidence": 0.95,
    "source": "income_statement",
    "context": "annual revenue for 2023",
    "unit": "dollars",
    "validated": true
  },
  "ebitda": {
    "value": 8500000,
    "confidence": 0.90,
    "source": "financial_summary",
    "context": "adjusted EBITDA 2023",
    "unit": "dollars",
    "validated": true
  },
  "dealValue": {
    "value": 125000000,
    "confidence": 0.85,
    "source": "deal_terms",
    "context": "total enterprise value",
    "unit": "dollars",
    "validated": false
  },
  "multiples": {
    "ev_revenue": {
      "value": 5.0,
      "confidence": 0.80,
      "source": "calculated",
      "context": "EV/Revenue multiple",
      "unit": "ratio",
      "validated": true
    }
  },
  "ratios": {
    "ebitda_margin": {
      "value": 0.34,
      "confidence": 0.90,
      "source": "calculated", 
      "context": "EBITDA margin percentage",
      "unit": "percentage",
      "validated": true
    }
  },
  "currency": "USD",
  "period": "FY 2023",
  "confidence": 0.88,
  "validated": true,
  "warnings": ["Any validation warnings"],
  "metadata": {
    "extraction_method": "ai_analysis",
    "validation_checks": ["consistency", "reasonableness"]
  }
}`

	userPrompt := fmt.Sprintf("Extract financial metrics from this %s document and respond with JSON:\n\n%s", documentType, content)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result FinancialMetricsExtraction
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse financial metrics: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// ExtractPersonnelAndRoles extracts personnel information and roles
func (cp *ClaudeProvider) ExtractPersonnelAndRoles(ctx context.Context, content string, documentType string) (*PersonnelRoleExtraction, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	// Truncate content if too long
	if len(content) > 15000 {
		content = content[:15000] + "..."
	}

	systemPrompt := `You are an expert in organizational analysis for M&A due diligence.
Extract personnel information, roles, and organizational hierarchy with enhanced detail.

Focus on:
- Key personnel with titles and companies
- Contact information when available
- Organizational hierarchy and reporting relationships
- Role classification (decision_maker, advisor, contact)
- Department and functional area identification

Provide your response in JSON format with the following structure:
{
  "personnel": [
    {
      "name": "John Smith",
      "title": "Chief Executive Officer",
      "company": "AquaFlow Technologies",
      "role": "decision_maker",
      "department": "executive",
      "confidence": 0.95,
      "context": "CEO of target company",
      "contact": {
        "email": "john.smith@aquaflow.com",
        "phone": "+1-555-0123",
        "address": "123 Business Ave, San Francisco, CA"
      },
      "metadata": {"source_section": "management_team"}
    }
  ],
  "contacts": [
    {
      "email": "deals@investmentbank.com",
      "phone": "+1-555-0456",
      "address": "456 Wall Street, New York, NY",
      "company": "Investment Bank LLC",
      "confidence": 0.90,
      "context": "deal advisor contact",
      "metadata": {"source_section": "advisor_contacts"}
    }
  ],
  "hierarchy": [
    {
      "superior": "John Smith",
      "subordinate": "Jane Doe",
      "confidence": 0.85,
      "context": "CEO to CFO reporting relationship"
    }
  ],
  "confidence": 0.87,
  "metadata": {
    "extraction_method": "ai_analysis",
    "sections_analyzed": ["management_team", "advisor_contacts"]
  },
  "warnings": ["Any extraction warnings"]
}`

	userPrompt := fmt.Sprintf("Extract personnel and role information from this %s document and respond with JSON:\n\n%s", documentType, content)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result PersonnelRoleExtraction
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse personnel and roles: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// ValidateEntitiesAcrossDocuments validates and resolves conflicts between entities from multiple documents
func (cp *ClaudeProvider) ValidateEntitiesAcrossDocuments(ctx context.Context, documentExtractions []DocumentEntityExtraction) (*CrossDocumentValidation, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	// Create a summary of extractions for the AI to analyze
	extractionSummary := make(map[string]interface{})
	for i, extraction := range documentExtractions {
		extractionSummary[fmt.Sprintf("document_%d", i)] = map[string]interface{}{
			"documentId":   extraction.DocumentID,
			"documentType": extraction.DocumentType,
			"companies":    extraction.Companies,
			"financial":    extraction.Financial,
			"personnel":    extraction.Personnel,
			"confidence":   extraction.Confidence,
		}
	}

	systemPrompt := `You are an expert M&A analyst specializing in cross-document validation and conflict resolution.
Analyze entity extractions from multiple documents and identify conflicts, validate consistency, and provide resolutions.

Focus on:
- Identifying conflicting information between documents
- Consolidating consistent entities across documents
- Providing confidence-based conflict resolution
- Generating validation summaries and recommendations

Provide your response in JSON format with the following structure:
{
  "consolidatedEntities": {
    "companies": [
      {
        "name": "AquaFlow Technologies Inc.",
        "role": "target",
        "confidence": 0.95,
        "context": "consolidated from multiple documents",
        "industry": "water treatment",
        "location": "California, USA",
        "metadata": {"consolidation_sources": ["doc1", "doc2"]},
        "validated": true
      }
    ],
    "financial": {
      "revenue": {
        "value": 25000000,
        "confidence": 0.92,
        "source": "consolidated",
        "context": "validated across financial documents",
        "unit": "dollars",
        "validated": true
      }
    },
    "personnel": [],
    "deals": []
  },
  "conflicts": [
    {
      "type": "financial",
      "field": "revenue",
      "values": [
        {
          "value": 25000000,
          "source": "financial_statement",
          "confidence": 0.95,
          "documentId": "doc1"
        },
        {
          "value": 24500000,
          "source": "management_presentation",
          "confidence": 0.85,
          "documentId": "doc2"
        }
      ],
      "severity": "low",
      "description": "Minor revenue discrepancy between financial statement and presentation",
      "metadata": {"variance_percentage": 2.0}
    }
  ],
  "resolutions": [
    {
      "conflictId": "conflict_1",
      "resolution": "precedence",
      "chosenValue": 25000000,
      "reasoning": "Financial statement has higher confidence and is more authoritative",
      "confidence": 0.90,
      "metadata": {"resolution_method": "source_precedence"}
    }
  ],
  "confidence": 0.88,
  "summary": {
    "totalEntities": 15,
    "validatedEntities": 13,
    "conflictsFound": 2,
    "conflictsResolved": 2,
    "overallConfidence": 0.88
  },
  "metadata": {
    "validation_method": "ai_analysis",
    "documents_analyzed": 3
  }
}`

	userPrompt := fmt.Sprintf("Validate and resolve conflicts in these entity extractions and respond with JSON:\n\n%s",
		func() string {
			summaryBytes, _ := json.MarshalIndent(extractionSummary, "", "  ")
			return string(summaryBytes)
		}())

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, err
	}

	var result CrossDocumentValidation
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("failed to parse cross-document validation: %w", err)
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// SEMANTIC FIELD MAPPING ENGINE METHODS FOR TASK 2.1

// AnalyzeFieldSemantics analyzes field meaning and context for semantic understanding
func (cp *ClaudeProvider) AnalyzeFieldSemantics(ctx context.Context, fieldName string, fieldValue interface{}, documentContext string) (*FieldSemanticAnalysis, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	systemPrompt := `You are an expert semantic field analyst for M&A document processing.
Analyze the given field name, value, and document context to understand the semantic meaning and business significance.

Focus on:
- Semantic type classification (currency, date, company_name, percentage, etc.)
- Business category (financial, entity, legal, operational)
- Data type and expected format
- Business rules that apply
- Confidence assessment

Return a JSON response with the analysis results.`

	userPrompt := fmt.Sprintf(`Analyze this field:
Field Name: %s
Field Value: %v
Document Context: %s

Provide semantic analysis including:
1. Semantic type (currency, date, company_name, percentage, text, number, etc.)
2. Business category (financial, entity, legal, operational, etc.)
3. Data type (string, number, date, boolean)
4. Expected format pattern
5. Applicable business rules
6. Confidence score (0.0 to 1.0)
7. Alternative interpretations`, fieldName, fieldValue, documentContext)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("Claude API request failed: %w", err)
	}

	var result FieldSemanticAnalysis
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		// Fallback with structured response
		result = FieldSemanticAnalysis{
			FieldName:        fieldName,
			SemanticType:     inferSemanticType(fieldName, fieldValue),
			BusinessCategory: inferBusinessCategory(fieldName),
			DataType:         inferDataType(fieldValue),
			ExpectedFormat:   inferExpectedFormat(fieldValue),
			ConfidenceScore:  0.7,
			Context:          documentContext,
			Metadata: map[string]interface{}{
				"provider":    "claude",
				"fallback":    true,
				"parse_error": err.Error(),
			},
			Suggestions:   []string{},
			BusinessRules: []string{},
		}
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// CreateSemanticMapping creates intelligent field mappings based on semantic understanding
func (cp *ClaudeProvider) CreateSemanticMapping(ctx context.Context, sourceFields map[string]interface{}, templateFields []string, documentType string) (*SemanticMappingResult, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	systemPrompt := `You are an expert field mapping specialist for M&A document processing.
Create intelligent semantic mappings between source document fields and template fields.

Consider:
- Semantic similarity and business meaning
- Field name variations and synonyms
- Data type compatibility
- Business logic and context
- Confidence scoring for each mapping
- Required transformations

Return a JSON response with detailed mapping results.`

	sourceFieldsJSON, _ := json.Marshal(sourceFields)
	templateFieldsJSON, _ := json.Marshal(templateFields)

	userPrompt := fmt.Sprintf(`Create semantic mappings for:
Document Type: %s
Source Fields: %s
Template Fields: %s

Provide:
1. Individual field mappings with confidence scores
2. Required transformations (format, calculate, lookup, aggregate)
3. Business justification for each mapping
4. Alternative mapping suggestions
5. Unmapped fields and reasons
6. Overall mapping strategy and confidence`, documentType, sourceFieldsJSON, templateFieldsJSON)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("Claude API request failed: %w", err)
	}

	var result SemanticMappingResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		// Fallback with basic mapping
		result = createFallbackMapping(sourceFields, templateFields, documentType)
		result.Metadata["provider"] = "claude_fallback"
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// ResolveFieldConflicts resolves conflicts when multiple sources provide different values for the same field
func (cp *ClaudeProvider) ResolveFieldConflicts(ctx context.Context, conflicts []FieldConflict, resolutionContext *ConflictResolutionContext) (*ConflictResolutionResult, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	systemPrompt := `You are an expert conflict resolution specialist for M&A document processing.
Resolve conflicts between field values from multiple sources using business logic and context.

Consider:
- Source reliability and confidence scores
- Business rules and precedence
- Data quality and consistency
- User preferences and historical patterns
- Resolution justification

Return a JSON response with resolution decisions.`

	conflictsJSON, _ := json.Marshal(conflicts)
	contextJSON, _ := json.Marshal(resolutionContext)

	userPrompt := fmt.Sprintf(`Resolve these field conflicts:
Conflicts: %s
Resolution Context: %s

Provide:
1. Resolved values for each conflict
2. Resolution method used (confidence_based, rule_based, user_preference, manual_review)
3. Detailed justification for each decision
4. Confidence scores for resolutions
5. Flags for conflicts requiring manual review
6. Alternative values to consider`, conflictsJSON, contextJSON)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("Claude API request failed: %w", err)
	}

	var result ConflictResolutionResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		// Fallback with confidence-based resolution
		result = createFallbackResolution(conflicts)
		result.Metadata["provider"] = "claude_fallback"
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// AnalyzeTemplateStructure analyzes template structure and field requirements
func (cp *ClaudeProvider) AnalyzeTemplateStructure(ctx context.Context, templatePath string, templateContent []byte) (*TemplateStructureAnalysis, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	// Truncate content if too long
	contentStr := string(templateContent)
	if len(contentStr) > 20000 {
		contentStr = contentStr[:20000] + "..."
	}

	systemPrompt := `You are an expert template structure analyst for M&A document processing.
Analyze template structure to understand field requirements, relationships, and validation rules.

Focus on:
- Field identification and types
- Required vs optional fields
- Calculated fields and formulas
- Section organization
- Field relationships and dependencies
- Validation rules and constraints
- Complexity assessment

Return a JSON response with comprehensive structure analysis.`

	userPrompt := fmt.Sprintf(`Analyze this template:
Template Path: %s
Template Content: %s

Provide:
1. Identified fields with types and locations
2. Template sections and organization
3. Field relationships and dependencies
4. Required vs optional field classification
5. Calculated fields and formulas
6. Validation rules and constraints
7. Complexity assessment and compatibility score`, templatePath, contentStr)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("Claude API request failed: %w", err)
	}

	var result TemplateStructureAnalysis
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		// Fallback with basic structure analysis
		result = createFallbackStructureAnalysis(templatePath, templateContent)
		result.Metadata["provider"] = "claude_fallback"
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}

// ValidateFieldMapping validates the logical consistency and business rule compliance of field mappings
func (cp *ClaudeProvider) ValidateFieldMapping(ctx context.Context, mapping *FieldMapping, validationRules []ValidationRule) (*MappingValidationResult, error) {
	atomic.AddInt64(&cp.stats.TotalRequests, 1)

	systemPrompt := `You are an expert field mapping validator for M&A document processing.
Validate field mappings against business rules and logical consistency requirements.

Check for:
- Data type compatibility
- Format consistency
- Business rule compliance
- Logical relationships
- Value ranges and constraints
- Completeness and quality

Return a JSON response with detailed validation results.`

	mappingJSON, _ := json.Marshal(mapping)
	rulesJSON, _ := json.Marshal(validationRules)

	userPrompt := fmt.Sprintf(`Validate this field mapping:
Mapping: %s
Validation Rules: %s

Provide:
1. Overall validation status and score
2. Individual field validation results
3. Rule compliance assessment
4. Errors and warnings with severity
5. Recommendations for improvement
6. Audit trail of validation steps`, mappingJSON, rulesJSON)

	response, err := cp.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		return nil, fmt.Errorf("Claude API request failed: %w", err)
	}

	var result MappingValidationResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		atomic.AddInt64(&cp.stats.FailedCalls, 1)
		// Fallback with basic validation
		result = createFallbackValidation(mapping, validationRules)
		result.Metadata["provider"] = "claude_fallback"
	}

	atomic.AddInt64(&cp.stats.SuccessfulCalls, 1)
	return &result, nil
}
