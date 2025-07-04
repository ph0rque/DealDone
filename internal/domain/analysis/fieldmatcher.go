package analysis

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"
)

// FieldMatcher provides intelligent field matching between documents and templates
type FieldMatcher struct {
	aiService *AIService
	synonyms  map[string][]string
}

// NewFieldMatcher creates a new field matcher
func NewFieldMatcher(aiService *AIService) *FieldMatcher {
	return &FieldMatcher{
		aiService: aiService,
		synonyms:  initializeSynonyms(),
	}
}

// FieldMatch represents a match between a source field and template field
type FieldMatch struct {
	SourceField     string  `json:"sourceField"`
	TemplateField   string  `json:"templateField"`
	ConfidenceScore float64 `json:"confidenceScore"`
	MatchType       string  `json:"matchType"` // "exact", "synonym", "fuzzy", "ai", "semantic"
	Reason          string  `json:"reason"`
}

// MatchingResult contains all field matches and metadata
type MatchingResult struct {
	Matches         []FieldMatch `json:"matches"`
	UnmatchedSource []string     `json:"unmatchedSource"`
	UnmatchedTarget []string     `json:"unmatchedTarget"`
	OverallScore    float64      `json:"overallScore"`
}

// initializeSynonyms creates a map of common business/financial synonyms
func initializeSynonyms() map[string][]string {
	return map[string][]string{
		"revenue":     {"sales", "income", "turnover", "receipts", "earnings", "gross revenue", "total revenue"},
		"ebitda":      {"earnings before interest taxes depreciation amortization", "operating income", "operating profit"},
		"company":     {"corporation", "business", "entity", "organization", "firm", "enterprise", "target"},
		"date":        {"closing date", "transaction date", "deal date", "effective date", "completion date"},
		"price":       {"purchase price", "deal value", "transaction value", "consideration", "amount"},
		"assets":      {"total assets", "asset base", "property", "holdings"},
		"liabilities": {"total liabilities", "debt", "obligations", "payables"},
		"equity":      {"shareholders equity", "net worth", "book value", "owners equity"},
		"margin":      {"profit margin", "gross margin", "operating margin", "net margin"},
		"cash":        {"cash flow", "free cash flow", "operating cash flow", "cash position"},
	}
}

// MatchFields finds the best matches between source fields and template fields
func (fm *FieldMatcher) MatchFields(sourceFields []string, templateFields []DataField) (*MatchingResult, error) {
	result := &MatchingResult{
		Matches:         []FieldMatch{},
		UnmatchedSource: make([]string, 0),
		UnmatchedTarget: make([]string, 0),
	}

	// Create normalized versions of fields for matching
	normalizedSource := fm.normalizeFields(sourceFields)
	normalizedTemplate := fm.normalizeDataFields(templateFields)

	// Track which fields have been matched
	matchedSource := make(map[string]bool)
	matchedTemplate := make(map[string]bool)

	// Phase 1: Exact matches
	fm.findExactMatches(normalizedSource, normalizedTemplate, result, matchedSource, matchedTemplate)

	// Phase 2: Synonym matches
	fm.findSynonymMatches(normalizedSource, normalizedTemplate, result, matchedSource, matchedTemplate)

	// Phase 3: Fuzzy string matches
	fm.findFuzzyMatches(normalizedSource, normalizedTemplate, result, matchedSource, matchedTemplate)

	// Phase 4: AI-powered semantic matches for remaining fields
	if fm.aiService != nil && fm.aiService.IsAvailable() {
		fm.findSemanticMatches(normalizedSource, normalizedTemplate, result, matchedSource, matchedTemplate)
	}

	// Collect unmatched fields
	for i, field := range sourceFields {
		if !matchedSource[normalizedSource[i].normalized] {
			result.UnmatchedSource = append(result.UnmatchedSource, field)
		}
	}

	for i, field := range templateFields {
		if !matchedTemplate[normalizedTemplate[i].normalized] {
			result.UnmatchedTarget = append(result.UnmatchedTarget, field.Name)
		}
	}

	// Calculate overall matching score
	result.OverallScore = fm.calculateOverallScore(result, len(sourceFields), len(templateFields))

	// Sort matches by confidence score
	sort.Slice(result.Matches, func(i, j int) bool {
		return result.Matches[i].ConfidenceScore > result.Matches[j].ConfidenceScore
	})

	return result, nil
}

// normalizedField represents a field with its normalized form
type normalizedField struct {
	original   string
	normalized string
	tokens     []string
}

// normalizeFields normalizes field names for matching
func (fm *FieldMatcher) normalizeFields(fields []string) []normalizedField {
	normalized := make([]normalizedField, len(fields))
	for i, field := range fields {
		norm := fm.normalizeString(field)
		normalized[i] = normalizedField{
			original:   field,
			normalized: norm,
			tokens:     fm.tokenize(norm),
		}
	}
	return normalized
}

// normalizeDataFields normalizes DataField names for matching
func (fm *FieldMatcher) normalizeDataFields(fields []DataField) []normalizedField {
	normalized := make([]normalizedField, len(fields))
	for i, field := range fields {
		norm := fm.normalizeString(field.Name)
		normalized[i] = normalizedField{
			original:   field.Name,
			normalized: norm,
			tokens:     fm.tokenize(norm),
		}
	}
	return normalized
}

// normalizeString converts a string to lowercase and removes special characters
func (fm *FieldMatcher) normalizeString(s string) string {
	var result strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}
	// Collapse multiple spaces
	normalized := strings.Join(strings.Fields(result.String()), " ")
	return normalized
}

// tokenize splits a string into tokens
func (fm *FieldMatcher) tokenize(s string) []string {
	return strings.Fields(s)
}

// findExactMatches finds exact string matches
func (fm *FieldMatcher) findExactMatches(source, template []normalizedField, result *MatchingResult,
	matchedSource, matchedTemplate map[string]bool) {

	for _, src := range source {
		if matchedSource[src.normalized] {
			continue
		}

		for _, tmpl := range template {
			if matchedTemplate[tmpl.normalized] {
				continue
			}

			if src.normalized == tmpl.normalized {
				result.Matches = append(result.Matches, FieldMatch{
					SourceField:     src.original,
					TemplateField:   tmpl.original,
					ConfidenceScore: 1.0,
					MatchType:       "exact",
					Reason:          "Exact name match",
				})
				matchedSource[src.normalized] = true
				matchedTemplate[tmpl.normalized] = true
				break
			}
		}
	}
}

// findSynonymMatches finds matches using synonym dictionary
func (fm *FieldMatcher) findSynonymMatches(source, template []normalizedField, result *MatchingResult,
	matchedSource, matchedTemplate map[string]bool) {

	for _, src := range source {
		if matchedSource[src.normalized] {
			continue
		}

		for _, tmpl := range template {
			if matchedTemplate[tmpl.normalized] {
				continue
			}

			if score, synonym := fm.checkSynonyms(src.normalized, tmpl.normalized); score > 0 {
				result.Matches = append(result.Matches, FieldMatch{
					SourceField:     src.original,
					TemplateField:   tmpl.original,
					ConfidenceScore: score,
					MatchType:       "synonym",
					Reason:          fmt.Sprintf("Synonym match: %s", synonym),
				})
				matchedSource[src.normalized] = true
				matchedTemplate[tmpl.normalized] = true
				break
			}
		}
	}
}

// checkSynonyms checks if two fields are synonyms
func (fm *FieldMatcher) checkSynonyms(field1, field2 string) (float64, string) {
	// Check each synonym group
	for key, synonyms := range fm.synonyms {
		inGroup1 := false
		inGroup2 := false

		// Check if field1 matches key or any synonym
		if strings.Contains(field1, key) {
			inGroup1 = true
		} else {
			for _, syn := range synonyms {
				if strings.Contains(field1, syn) || strings.Contains(syn, field1) {
					inGroup1 = true
					break
				}
			}
		}

		// Check if field2 matches key or any synonym
		if strings.Contains(field2, key) {
			inGroup2 = true
		} else {
			for _, syn := range synonyms {
				if strings.Contains(field2, syn) || strings.Contains(syn, field2) {
					inGroup2 = true
					break
				}
			}
		}

		if inGroup1 && inGroup2 {
			return 0.9, key
		}
	}

	return 0, ""
}

// findFuzzyMatches uses fuzzy string matching algorithms
func (fm *FieldMatcher) findFuzzyMatches(source, template []normalizedField, result *MatchingResult,
	matchedSource, matchedTemplate map[string]bool) {

	threshold := 0.7 // Minimum similarity score

	for _, src := range source {
		if matchedSource[src.normalized] {
			continue
		}

		bestMatch := normalizedField{}
		bestScore := 0.0

		for _, tmpl := range template {
			if matchedTemplate[tmpl.normalized] {
				continue
			}

			// Calculate similarity using multiple methods
			levenshteinScore := fm.levenshteinSimilarity(src.normalized, tmpl.normalized)
			tokenScore := fm.tokenSimilarity(src.tokens, tmpl.tokens)

			// Weighted average of scores
			score := (levenshteinScore*0.6 + tokenScore*0.4)

			if score > bestScore && score >= threshold {
				bestScore = score
				bestMatch = tmpl
			}
		}

		if bestScore > 0 {
			result.Matches = append(result.Matches, FieldMatch{
				SourceField:     src.original,
				TemplateField:   bestMatch.original,
				ConfidenceScore: bestScore,
				MatchType:       "fuzzy",
				Reason:          fmt.Sprintf("Fuzzy match (%.2f similarity)", bestScore),
			})
			matchedSource[src.normalized] = true
			matchedTemplate[bestMatch.normalized] = true
		}
	}
}

// levenshteinSimilarity calculates string similarity using Levenshtein distance
func (fm *FieldMatcher) levenshteinSimilarity(s1, s2 string) float64 {
	distance := fm.levenshteinDistance(s1, s2)
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(distance)/maxLen
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (fm *FieldMatcher) levenshteinDistance(s1, s2 string) int {
	if s1 == s2 {
		return 0
	}

	if len(s1) == 0 {
		return len(s2)
	}

	if len(s2) == 0 {
		return len(s1)
	}

	// Create distance matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first column and row
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = minInt(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// tokenSimilarity calculates similarity based on common tokens
func (fm *FieldMatcher) tokenSimilarity(tokens1, tokens2 []string) float64 {
	if len(tokens1) == 0 || len(tokens2) == 0 {
		return 0
	}

	// Count common tokens
	common := 0
	for _, t1 := range tokens1 {
		for _, t2 := range tokens2 {
			if t1 == t2 {
				common++
				break
			}
		}
	}

	// Jaccard similarity
	union := len(tokens1) + len(tokens2) - common
	if union == 0 {
		return 0
	}

	return float64(common) / float64(union)
}

// findSemanticMatches uses AI to find semantic matches
func (fm *FieldMatcher) findSemanticMatches(source, template []normalizedField, result *MatchingResult,
	matchedSource, matchedTemplate map[string]bool) {

	// Collect unmatched fields
	unmatchedSrc := []string{}
	unmatchedTmpl := []string{}

	for _, src := range source {
		if !matchedSource[src.normalized] {
			unmatchedSrc = append(unmatchedSrc, src.original)
		}
	}

	for _, tmpl := range template {
		if !matchedTemplate[tmpl.normalized] {
			unmatchedTmpl = append(unmatchedTmpl, tmpl.original)
		}
	}

	if len(unmatchedSrc) == 0 || len(unmatchedTmpl) == 0 {
		return
	}

	// Use AI to find semantic matches
	ctx := context.Background()
	aiMatches := fm.getAIFieldMatches(ctx, unmatchedSrc, unmatchedTmpl)

	for _, match := range aiMatches {
		// Find normalized versions
		var srcNorm, tmplNorm string
		for _, src := range source {
			if src.original == match.SourceField {
				srcNorm = src.normalized
				break
			}
		}
		for _, tmpl := range template {
			if tmpl.original == match.TemplateField {
				tmplNorm = tmpl.normalized
				break
			}
		}

		if srcNorm != "" && tmplNorm != "" && !matchedSource[srcNorm] && !matchedTemplate[tmplNorm] {
			result.Matches = append(result.Matches, match)
			matchedSource[srcNorm] = true
			matchedTemplate[tmplNorm] = true
		}
	}
}

// getAIFieldMatches uses AI service to find semantic field matches
func (fm *FieldMatcher) getAIFieldMatches(ctx context.Context, sourceFields, templateFields []string) []FieldMatch {
	// In a real implementation, this would call the AI service
	// For now, return empty matches
	return []FieldMatch{}
}

// calculateOverallScore calculates the overall matching score
func (fm *FieldMatcher) calculateOverallScore(result *MatchingResult, totalSource, totalTemplate int) float64 {
	if totalSource == 0 || totalTemplate == 0 {
		return 0
	}

	// Calculate weighted score based on matches and their confidence
	totalScore := 0.0
	for _, match := range result.Matches {
		totalScore += match.ConfidenceScore
	}

	// Normalize by the maximum possible matches
	maxMatches := math.Min(float64(totalSource), float64(totalTemplate))
	if maxMatches == 0 {
		return 0
	}

	return totalScore / maxMatches
}

// GetFieldMappingSuggestions provides suggestions for unmapped fields
func (fm *FieldMatcher) GetFieldMappingSuggestions(unmatchedSource []string, templateFields []DataField) map[string][]string {
	suggestions := make(map[string][]string)

	for _, srcField := range unmatchedSource {
		srcNorm := fm.normalizeString(srcField)
		fieldSuggestions := []string{}

		// Find top 3 closest template fields
		type scoredField struct {
			field string
			score float64
		}

		scores := []scoredField{}

		for _, tmplField := range templateFields {
			tmplNorm := fm.normalizeString(tmplField.Name)

			// Calculate similarity
			score := fm.levenshteinSimilarity(srcNorm, tmplNorm)
			if score > 0.3 { // Only include if somewhat similar
				scores = append(scores, scoredField{
					field: tmplField.Name,
					score: score,
				})
			}
		}

		// Sort by score
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].score > scores[j].score
		})

		// Take top 3
		for i := 0; i < len(scores) && i < 3; i++ {
			fieldSuggestions = append(fieldSuggestions, scores[i].field)
		}

		if len(fieldSuggestions) > 0 {
			suggestions[srcField] = fieldSuggestions
		}
	}

	return suggestions
}

// minInt returns the minimum of three integers
func minInt(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= a && b <= c {
		return b
	}
	return c
}
