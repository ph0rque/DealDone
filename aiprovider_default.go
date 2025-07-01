package main

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// DefaultProvider provides rule-based fallback when AI services are unavailable
type DefaultProvider struct {
	stats *AIUsageStats
}

// NewDefaultProvider creates a new default provider
func NewDefaultProvider() *DefaultProvider {
	return &DefaultProvider{
		stats: &AIUsageStats{
			LastReset: time.Now(),
		},
	}
}

// ClassifyDocument performs rule-based document classification
func (dp *DefaultProvider) ClassifyDocument(ctx context.Context, content string, metadata map[string]interface{}) (*AIClassificationResult, error) {
	atomic.AddInt64(&dp.stats.TotalRequests, 1)

	// Basic rule-based classification
	lowerContent := strings.ToLower(content)

	// Extract keywords
	keywords := dp.extractKeywords(lowerContent)

	// Determine document type
	docType := "general"
	confidence := 0.6

	legalScore := dp.calculateLegalScore(lowerContent)
	financialScore := dp.calculateFinancialScore(lowerContent)

	if legalScore > financialScore && legalScore > 0.3 {
		docType = "legal"
		confidence = min(0.9, 0.6+legalScore)
	} else if financialScore > 0.3 {
		docType = "financial"
		confidence = min(0.9, 0.6+financialScore)
	}

	// Extract summary (first 200 chars)
	summary := content
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}

	atomic.AddInt64(&dp.stats.SuccessfulCalls, 1)

	return &AIClassificationResult{
		DocumentType: docType,
		Confidence:   confidence,
		Keywords:     keywords,
		Categories:   []string{docType},
		Language:     "en", // Default to English
		Summary:      summary,
		Metadata:     metadata,
	}, nil
}

// ExtractFinancialData extracts financial data using patterns
func (dp *DefaultProvider) ExtractFinancialData(ctx context.Context, content string) (*FinancialAnalysis, error) {
	atomic.AddInt64(&dp.stats.TotalRequests, 1)

	// Basic pattern matching for financial data
	result := &FinancialAnalysis{
		DataPoints: make(map[string]float64),
		Currency:   "USD", // Default
		Confidence: 0.5,   // Low confidence for rule-based
		Warnings:   []string{"Rule-based extraction - results may be incomplete"},
	}

	// Look for common financial patterns
	lowerContent := strings.ToLower(content)

	// Revenue patterns
	if strings.Contains(lowerContent, "revenue") {
		// Placeholder - would parse numbers near "revenue"
		result.Revenue = 0
	}

	// EBITDA patterns
	if strings.Contains(lowerContent, "ebitda") {
		result.EBITDA = 0
	}

	atomic.AddInt64(&dp.stats.SuccessfulCalls, 1)

	return result, nil
}

// AnalyzeRisks performs basic risk analysis
func (dp *DefaultProvider) AnalyzeRisks(ctx context.Context, content string, docType string) (*RiskAnalysis, error) {
	atomic.AddInt64(&dp.stats.TotalRequests, 1)

	risks := []RiskItem{}
	lowerContent := strings.ToLower(content)

	// Check for risk keywords
	riskKeywords := map[string]string{
		"litigation":    "legal",
		"lawsuit":       "legal",
		"debt":          "financial",
		"loss":          "financial",
		"compliance":    "regulatory",
		"violation":     "regulatory",
		"breach":        "security",
		"default":       "financial",
		"bankruptcy":    "financial",
		"investigation": "regulatory",
	}

	for keyword, category := range riskKeywords {
		if strings.Contains(lowerContent, keyword) {
			risks = append(risks, RiskItem{
				Category:    category,
				Description: fmt.Sprintf("Document contains reference to %s", keyword),
				Severity:    "medium",
				Score:       0.5,
				Mitigation:  "Further review recommended",
			})
		}
	}

	overallScore := float64(len(risks)) * 0.1
	if overallScore > 1.0 {
		overallScore = 1.0
	}

	atomic.AddInt64(&dp.stats.SuccessfulCalls, 1)

	return &RiskAnalysis{
		OverallRiskScore: overallScore,
		RiskCategories:   risks,
		Recommendations:  []string{"Manual review recommended for comprehensive risk assessment"},
		CriticalIssues:   []string{},
		Confidence:       0.5,
	}, nil
}

// GenerateInsights creates basic insights
func (dp *DefaultProvider) GenerateInsights(ctx context.Context, content string, docType string) (*DocumentInsights, error) {
	atomic.AddInt64(&dp.stats.TotalRequests, 1)

	insights := &DocumentInsights{
		KeyPoints:     []string{"Document requires AI analysis for detailed insights"},
		Opportunities: []string{},
		Concerns:      []string{},
		ActionItems:   []string{"Enable AI service for comprehensive analysis"},
		Confidence:    0.3,
	}

	atomic.AddInt64(&dp.stats.SuccessfulCalls, 1)

	return insights, nil
}

// ExtractEntities performs basic entity extraction
func (dp *DefaultProvider) ExtractEntities(ctx context.Context, content string) (*EntityExtraction, error) {
	atomic.AddInt64(&dp.stats.TotalRequests, 1)

	result := &EntityExtraction{
		People:         []Entity{},
		Organizations:  []Entity{},
		Locations:      []Entity{},
		Dates:          []Entity{},
		MonetaryValues: []Entity{},
		Percentages:    []Entity{},
		Products:       []Entity{},
	}

	// Basic pattern matching could be added here
	// For now, return empty results

	atomic.AddInt64(&dp.stats.SuccessfulCalls, 1)

	return result, nil
}

// GetProvider returns the provider name
func (dp *DefaultProvider) GetProvider() AIProvider {
	return ProviderDefault
}

// IsAvailable always returns true for default provider
func (dp *DefaultProvider) IsAvailable() bool {
	return true
}

// GetUsage returns usage statistics
func (dp *DefaultProvider) GetUsage() *AIUsageStats {
	return dp.stats
}

// Helper methods

func (dp *DefaultProvider) extractKeywords(content string) []string {
	// Simple keyword extraction - find frequently occurring words
	commonWords := map[string]bool{
		"the": true, "and": true, "or": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "of": true, "a": true,
		"an": true, "is": true, "are": true, "was": true, "were": true,
		"been": true, "be": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true,
	}

	words := strings.Fields(content)
	wordCount := make(map[string]int)

	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 3 && !commonWords[word] {
			wordCount[word]++
		}
	}

	// Get top 10 keywords
	keywords := []string{}
	for word, count := range wordCount {
		if count > 2 {
			keywords = append(keywords, word)
		}
		if len(keywords) >= 10 {
			break
		}
	}

	return keywords
}

func (dp *DefaultProvider) calculateLegalScore(content string) float64 {
	legalTerms := []string{
		"agreement", "contract", "legal", "law", "clause", "party", "parties",
		"liability", "indemnity", "warranty", "representation", "covenant",
		"breach", "termination", "dispute", "arbitration", "jurisdiction",
		"governing law", "confidential", "nda", "non-disclosure",
	}

	score := 0.0
	for _, term := range legalTerms {
		if strings.Contains(content, term) {
			score += 0.05
		}
	}

	return min(score, 0.8)
}

func (dp *DefaultProvider) calculateFinancialScore(content string) float64 {
	financialTerms := []string{
		"revenue", "profit", "loss", "income", "expense", "cash flow",
		"balance sheet", "assets", "liabilities", "equity", "ebitda",
		"margin", "growth", "forecast", "budget", "financial", "fiscal",
		"quarter", "annual", "ytd", "roi", "irr", "npv",
	}

	score := 0.0
	for _, term := range financialTerms {
		if strings.Contains(content, term) {
			score += 0.05
		}
	}

	return min(score, 0.8)
}
