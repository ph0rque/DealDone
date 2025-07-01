package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFieldMatcher(t *testing.T) {
	aiService := &AIService{}
	matcher := NewFieldMatcher(aiService)

	assert.NotNil(t, matcher)
	assert.Equal(t, aiService, matcher.aiService)
	assert.NotNil(t, matcher.synonyms)
	assert.Greater(t, len(matcher.synonyms), 0)
}

func TestNormalizeString(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"Company Name", "company name"},
		{"Total_Revenue", "total revenue"},
		{"EBITDA-Margin", "ebitda margin"},
		{"Deal.Date", "deal date"},
		{"Revenue ($USD)", "revenue usd"},
		{"  Multiple   Spaces  ", "multiple spaces"},
	}

	for _, tt := range tests {
		result := matcher.normalizeString(tt.input)
		assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
	}
}

func TestTokenize(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	tokens := matcher.tokenize("total revenue amount")
	assert.Equal(t, []string{"total", "revenue", "amount"}, tokens)
}

func TestExactMatches(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	sourceFields := []string{"Revenue", "Company Name", "Date"}
	templateFields := []DataField{
		{Name: "Revenue", DataType: "number"},
		{Name: "Company Name", DataType: "string"},
		{Name: "Transaction Date", DataType: "date"},
	}

	result, err := matcher.MatchFields(sourceFields, templateFields)

	require.NoError(t, err)
	assert.Len(t, result.Matches, 2) // Revenue and Company Name should match exactly

	// Check exact matches
	exactMatches := 0
	for _, match := range result.Matches {
		if match.MatchType == "exact" {
			exactMatches++
			assert.Equal(t, 1.0, match.ConfidenceScore)
		}
	}
	assert.Equal(t, 2, exactMatches)
}

func TestSynonymMatches(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	sourceFields := []string{"Sales", "Corporation", "Closing Date"}
	templateFields := []DataField{
		{Name: "Revenue", DataType: "number"},
		{Name: "Company", DataType: "string"},
		{Name: "Date", DataType: "date"},
	}

	result, err := matcher.MatchFields(sourceFields, templateFields)

	require.NoError(t, err)
	assert.Greater(t, len(result.Matches), 0)

	// Check for synonym matches
	synonymMatches := 0
	for _, match := range result.Matches {
		if match.MatchType == "synonym" {
			synonymMatches++
			assert.Equal(t, 0.9, match.ConfidenceScore)
		}
	}
	assert.Greater(t, synonymMatches, 0)
}

func TestCheckSynonyms(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	tests := []struct {
		field1   string
		field2   string
		expected bool
	}{
		{"revenue", "sales", true},
		{"total revenue", "gross revenue", true},
		{"company", "corporation", true},
		{"date", "closing date", true},
		{"revenue", "expenses", false},
	}

	for _, tt := range tests {
		score, synonym := matcher.checkSynonyms(tt.field1, tt.field2)
		if tt.expected {
			assert.Greater(t, score, 0.0, "Fields: %s and %s", tt.field1, tt.field2)
			assert.NotEmpty(t, synonym)
		} else {
			assert.Equal(t, 0.0, score, "Fields: %s and %s", tt.field1, tt.field2)
			assert.Empty(t, synonym)
		}
	}
}

func TestLevenshteinDistance(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"abc", "abc", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
	}

	for _, tt := range tests {
		result := matcher.levenshteinDistance(tt.s1, tt.s2)
		assert.Equal(t, tt.expected, result, "Strings: %s and %s", tt.s1, tt.s2)
	}
}

func TestLevenshteinSimilarity(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	tests := []struct {
		s1       string
		s2       string
		minScore float64
	}{
		{"revenue", "revenue", 1.0},
		{"revenue", "revenues", 0.8},
		{"company name", "company names", 0.8},
		{"total different", "completely other", 0.0},
	}

	for _, tt := range tests {
		result := matcher.levenshteinSimilarity(tt.s1, tt.s2)
		assert.GreaterOrEqual(t, result, tt.minScore, "Strings: %s and %s", tt.s1, tt.s2)
	}
}

func TestTokenSimilarity(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	tests := []struct {
		tokens1  []string
		tokens2  []string
		expected float64
	}{
		{[]string{"total", "revenue"}, []string{"total", "revenue"}, 1.0},
		{[]string{"total", "revenue"}, []string{"revenue", "total"}, 1.0},
		{[]string{"total"}, []string{"total", "revenue"}, 0.5},
		{[]string{"abc"}, []string{"xyz"}, 0.0},
		{[]string{}, []string{"test"}, 0.0},
	}

	for _, tt := range tests {
		result := matcher.tokenSimilarity(tt.tokens1, tt.tokens2)
		assert.InDelta(t, tt.expected, result, 0.01, "Tokens: %v and %v", tt.tokens1, tt.tokens2)
	}
}

func TestFuzzyMatches(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	sourceFields := []string{"Total Rev", "Cmpny Name", "Trans Date"}
	templateFields := []DataField{
		{Name: "Total Revenue", DataType: "number"},
		{Name: "Company Name", DataType: "string"},
		{Name: "Transaction Date", DataType: "date"},
	}

	result, err := matcher.MatchFields(sourceFields, templateFields)

	require.NoError(t, err)
	assert.Greater(t, len(result.Matches), 0)

	// Check for fuzzy matches
	fuzzyMatches := 0
	for _, match := range result.Matches {
		if match.MatchType == "fuzzy" {
			fuzzyMatches++
			assert.Greater(t, match.ConfidenceScore, 0.7)
		}
	}
	assert.Greater(t, fuzzyMatches, 0)
}

func TestCalculateOverallScore(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	result := &MatchingResult{
		Matches: []FieldMatch{
			{ConfidenceScore: 1.0},
			{ConfidenceScore: 0.9},
			{ConfidenceScore: 0.8},
		},
	}

	score := matcher.calculateOverallScore(result, 5, 5)
	expected := (1.0 + 0.9 + 0.8) / 5.0
	assert.InDelta(t, expected, score, 0.01)

	// Test edge cases
	score = matcher.calculateOverallScore(result, 0, 5)
	assert.Equal(t, 0.0, score)

	score = matcher.calculateOverallScore(result, 5, 0)
	assert.Equal(t, 0.0, score)
}

func TestGetFieldMappingSuggestions(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	unmatchedSource := []string{"Tot Revenue", "Profit Mrgn"}
	templateFields := []DataField{
		{Name: "Total Revenue", DataType: "number"},
		{Name: "Revenue", DataType: "number"},
		{Name: "Gross Margin", DataType: "number"},
		{Name: "Operating Margin", DataType: "number"},
		{Name: "Net Margin", DataType: "number"},
		{Name: "Unrelated Field", DataType: "string"},
	}

	suggestions := matcher.GetFieldMappingSuggestions(unmatchedSource, templateFields)

	assert.Contains(t, suggestions, "Tot Revenue")
	assert.Contains(t, suggestions, "Profit Mrgn")

	// Check that relevant suggestions are provided
	revenueSuggestions := suggestions["Tot Revenue"]
	assert.Contains(t, revenueSuggestions, "Total Revenue")

	marginSuggestions := suggestions["Profit Mrgn"]
	assert.Greater(t, len(marginSuggestions), 0)
}

func TestCompleteMatchingScenario(t *testing.T) {
	matcher := NewFieldMatcher(nil)

	// Simulate a real-world scenario with various match types
	sourceFields := []string{
		"Company Name",     // Exact match
		"Total Sales",      // Synonym match
		"EBITDA Margn",     // Fuzzy match
		"Deal Dt",          // Fuzzy match
		"Random Field XYZ", // No match
	}

	templateFields := []DataField{
		{Name: "Company Name", DataType: "string"},
		{Name: "Revenue", DataType: "number"},
		{Name: "EBITDA Margin", DataType: "number"},
		{Name: "Transaction Date", DataType: "date"},
		{Name: "Purchase Price", DataType: "currency"},
	}

	result, err := matcher.MatchFields(sourceFields, templateFields)

	require.NoError(t, err)

	// Should have at least 3 matches (exact, synonym, and fuzzy)
	assert.GreaterOrEqual(t, len(result.Matches), 3)

	// Check unmatched fields
	assert.Contains(t, result.UnmatchedSource, "Random Field XYZ")
	assert.Contains(t, result.UnmatchedTarget, "Purchase Price")

	// Verify overall score is reasonable
	assert.Greater(t, result.OverallScore, 0.5)
	assert.Less(t, result.OverallScore, 1.0)

	// Verify matches are sorted by confidence
	for i := 1; i < len(result.Matches); i++ {
		assert.GreaterOrEqual(t, result.Matches[i-1].ConfidenceScore, result.Matches[i].ConfidenceScore)
	}
}
