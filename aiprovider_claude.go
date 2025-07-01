package main

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
func (cp *ClaudeProvider) GetProvider() AIProvider {
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
