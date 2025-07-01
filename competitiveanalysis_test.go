package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompetitiveAnalyzer(t *testing.T) {
	aiService := &AIService{}
	docProcessor := &DocumentProcessor{}

	analyzer := NewCompetitiveAnalyzer(aiService, docProcessor)

	assert.NotNil(t, analyzer)
	assert.Equal(t, aiService, analyzer.aiService)
	assert.Equal(t, docProcessor, analyzer.documentProcessor)
}

func TestAnalyzeMarketPosition(t *testing.T) {
	analyzer := NewCompetitiveAnalyzer(nil, nil)

	competitiveData := map[string]interface{}{
		"marketSize":   2000000000.0, // $2B
		"marketGrowth": 0.25,         // 25%
	}

	marketData := map[string]interface{}{
		"marketShare": 0.20, // 20%
	}

	position := analyzer.analyzeMarketPosition("TargetCo", competitiveData, marketData)

	assert.NotNil(t, position)
	assert.Equal(t, 0.20, position.MarketShare)
	assert.Equal(t, 2000000000.0, position.MarketSize)
	assert.Equal(t, 0.25, position.GrowthRate)
	assert.NotEmpty(t, position.Segments)
	assert.NotEmpty(t, position.GeographicReach)
	assert.NotNil(t, position.CustomerBase)
	assert.Greater(t, position.BrandStrength, 0.0)

	// Check segments
	assert.Len(t, position.Segments, 3)
	totalSegmentSize := 0.0
	for _, segment := range position.Segments {
		totalSegmentSize += segment.Size
	}
	assert.InDelta(t, position.MarketSize, totalSegmentSize, 1.0)

	// Check customer base
	assert.Greater(t, position.CustomerBase.TotalCustomers, 0)
	assert.NotEmpty(t, position.CustomerBase.CustomerSegments)
	assert.Greater(t, position.CustomerBase.CustomerLifetime, 0.0)
}

func TestProfileCompetitors(t *testing.T) {
	analyzer := NewCompetitiveAnalyzer(nil, nil)

	competitors := analyzer.profileCompetitors(map[string]interface{}{}, map[string]interface{}{})

	assert.NotEmpty(t, competitors)
	assert.GreaterOrEqual(t, len(competitors), 3)

	// Check first competitor (should be highest threat)
	topCompetitor := competitors[0]
	assert.Equal(t, "high", topCompetitor.ThreatLevel)
	assert.Greater(t, topCompetitor.MarketShare, 0.0)
	assert.Greater(t, topCompetitor.Revenue, 0.0)
	assert.NotEmpty(t, topCompetitor.Strengths)
	assert.NotEmpty(t, topCompetitor.Weaknesses)
	assert.NotEmpty(t, topCompetitor.Products)
	assert.Greater(t, topCompetitor.FinancialStrength, 0.0)

	// Check threat level ordering
	for i := 1; i < len(competitors); i++ {
		prevThreat := competitors[i-1].ThreatLevel
		currThreat := competitors[i].ThreatLevel
		threatOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		assert.GreaterOrEqual(t, threatOrder[prevThreat], threatOrder[currThreat])
	}
}

func TestPerformSWOT(t *testing.T) {
	analyzer := NewCompetitiveAnalyzer(nil, nil)

	competitors := []CompetitorProfile{
		{Name: "BigCorp", ThreatLevel: "high"},
	}

	swot := analyzer.performSWOT("TargetCo", map[string]interface{}{}, competitors)

	assert.NotNil(t, swot)
	assert.NotEmpty(t, swot.Strengths)
	assert.NotEmpty(t, swot.Weaknesses)
	assert.NotEmpty(t, swot.Opportunities)
	assert.NotEmpty(t, swot.Threats)

	// Check all items have required fields
	for _, item := range swot.Strengths {
		assert.NotEmpty(t, item.Description)
		assert.Contains(t, []string{"high", "medium", "low"}, item.Impact)
		assert.Greater(t, item.Score, 0.0)
		assert.LessOrEqual(t, item.Score, 1.0)
	}

	// Check that competitor is mentioned in threats
	foundCompetitorThreat := false
	for _, threat := range swot.Threats {
		if contains(threat.Description, "BigCorp") {
			foundCompetitorThreat = true
			break
		}
	}
	assert.True(t, foundCompetitorThreat)
}

func TestIdentifyMarketTrends(t *testing.T) {
	analyzer := NewCompetitiveAnalyzer(nil, nil)

	trends := analyzer.identifyMarketTrends(map[string]interface{}{}, map[string]interface{}{})

	assert.NotEmpty(t, trends)

	// Check trends are properly structured
	for _, trend := range trends {
		assert.NotEmpty(t, trend.Name)
		assert.NotEmpty(t, trend.Description)
		assert.Contains(t, []string{"short-term", "medium-term", "long-term"}, trend.Timeline)
		assert.Contains(t, []string{"positive", "negative", "neutral"}, trend.Impact)
		assert.Greater(t, trend.Relevance, 0.0)
		assert.LessOrEqual(t, trend.Relevance, 1.0)
	}

	// Check trends are sorted by relevance
	for i := 1; i < len(trends); i++ {
		assert.GreaterOrEqual(t, trends[i-1].Relevance, trends[i].Relevance)
	}
}

func TestAssessStrategicValue(t *testing.T) {
	analyzer := NewCompetitiveAnalyzer(nil, nil)

	position := &MarketPositionAnalysis{
		MarketShare:   0.25,
		GrowthRate:    0.30,
		BrandStrength: 0.8,
		CustomerBase: CustomerBaseAnalysis{
			ChurnRate:         0.05,
			ConcentrationRisk: 0.2,
		},
	}

	swot := &SWOTAnalysis{
		Strengths: []SWOTItem{
			{Description: "Strong technology platform", Score: 0.9},
			{Description: "Operational efficiency", Score: 0.8},
		},
		Opportunities: []SWOTItem{
			{Description: "Market expansion", Score: 0.85},
		},
	}

	assessment := analyzer.assessStrategicValue(position, swot, map[string]interface{}{})

	assert.NotNil(t, assessment)
	assert.Greater(t, assessment.OverallScore, 0.0)
	assert.LessOrEqual(t, assessment.OverallScore, 1.0)
	assert.Greater(t, assessment.MarketAccessValue, 0.0)
	assert.Greater(t, assessment.TechnologyValue, 0.0)
	assert.Greater(t, assessment.CustomerBaseValue, 0.0)
	assert.Equal(t, position.BrandStrength, assessment.BrandValue)
	assert.NotEmpty(t, assessment.Rationale)
}

func TestAnalyzeSynergies(t *testing.T) {
	analyzer := NewCompetitiveAnalyzer(nil, nil)

	synergies := analyzer.analyzeSynergies("TargetCo", map[string]interface{}{}, map[string]interface{}{})

	assert.NotNil(t, synergies)
	assert.NotEmpty(t, synergies.RevenueSynergies)
	assert.NotEmpty(t, synergies.CostSynergies)
	assert.Greater(t, synergies.TotalValue, 0.0)
	assert.NotEmpty(t, synergies.TimeToRealize)
	assert.Contains(t, []string{"high", "medium", "low"}, synergies.ImplementationRisk)

	// Check synergy details
	for _, syn := range synergies.RevenueSynergies {
		assert.NotEmpty(t, syn.Type)
		assert.NotEmpty(t, syn.Description)
		assert.Greater(t, syn.Value, 0.0)
		assert.NotEmpty(t, syn.Timeline)
		assert.Greater(t, syn.Probability, 0.0)
		assert.LessOrEqual(t, syn.Probability, 1.0)
	}

	// Verify total value calculation
	calculatedTotal := 0.0
	for _, syn := range synergies.RevenueSynergies {
		calculatedTotal += syn.Value * syn.Probability
	}
	for _, syn := range synergies.CostSynergies {
		calculatedTotal += syn.Value * syn.Probability
	}
	assert.InDelta(t, calculatedTotal, synergies.TotalValue, 1.0)
}

func TestIdentifyCompetitiveRisks(t *testing.T) {
	analyzer := NewCompetitiveAnalyzer(nil, nil)

	competitors := []CompetitorProfile{
		{Name: "AggressiveCorp", ThreatLevel: "high"},
	}

	trends := []MarketTrend{
		{Name: "Disruption", Impact: "negative"},
	}

	risks := analyzer.identifyCompetitiveRisks(competitors, trends)

	assert.NotEmpty(t, risks)

	// Check risk structure
	for _, risk := range risks {
		assert.NotEmpty(t, risk.Type)
		assert.NotEmpty(t, risk.Description)
		assert.Greater(t, risk.Likelihood, 0.0)
		assert.LessOrEqual(t, risk.Likelihood, 1.0)
		assert.Greater(t, risk.Impact, 0.0)
		assert.LessOrEqual(t, risk.Impact, 1.0)
		assert.NotEmpty(t, risk.Mitigation)
	}

	// Check risks are sorted by score (likelihood * impact)
	for i := 1; i < len(risks); i++ {
		prevScore := risks[i-1].Likelihood * risks[i-1].Impact
		currScore := risks[i].Likelihood * risks[i].Impact
		assert.GreaterOrEqual(t, prevScore, currScore)
	}
}

func TestIdentifyGrowthOpportunities(t *testing.T) {
	analyzer := NewCompetitiveAnalyzer(nil, nil)

	position := &MarketPositionAnalysis{
		MarketShare: 0.15,
		GrowthRate:  0.25,
	}

	trends := []MarketTrend{
		{Name: "AI Adoption", Impact: "positive", Relevance: 0.9},
	}

	opportunities := analyzer.identifyGrowthOpportunities(position, trends, map[string]interface{}{})

	assert.NotEmpty(t, opportunities)

	// Check opportunity structure
	for _, opp := range opportunities {
		assert.NotEmpty(t, opp.Type)
		assert.NotEmpty(t, opp.Description)
		assert.Greater(t, opp.PotentialValue, 0.0)
		assert.NotEmpty(t, opp.TimeHorizon)
		assert.NotEmpty(t, opp.Requirements)
		assert.Greater(t, opp.Probability, 0.0)
		assert.LessOrEqual(t, opp.Probability, 1.0)
	}

	// Check opportunities are sorted by expected value
	for i := 1; i < len(opportunities); i++ {
		prevExpected := opportunities[i-1].PotentialValue * opportunities[i-1].Probability
		currExpected := opportunities[i].PotentialValue * opportunities[i].Probability
		assert.GreaterOrEqual(t, prevExpected, currExpected)
	}
}

func TestCalculateBrandStrength(t *testing.T) {
	analyzer := NewCompetitiveAnalyzer(nil, nil)

	tests := []struct {
		name        string
		marketShare float64
		growthRate  float64
		minExpected float64
	}{
		{"Strong position", 0.3, 0.25, 0.5},
		{"Weak position", 0.05, 0.1, 0.0},
		{"High growth", 0.1, 0.5, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := analyzer.calculateBrandStrength(tt.marketShare, tt.growthRate)
			assert.GreaterOrEqual(t, strength, tt.minExpected)
			assert.LessOrEqual(t, strength, 1.0)
		})
	}
}

func TestQuickCompetitiveAssessment(t *testing.T) {
	analyzer := NewCompetitiveAnalyzer(nil, nil)

	tests := []struct {
		name        string
		revenue     float64
		marketShare float64
		expectedPos string
	}{
		{"Market leader", 500000000, 0.35, "Market leader"},
		{"Strong challenger", 200000000, 0.20, "Strong challenger"},
		{"Niche player", 50000000, 0.08, "Niche player"},
		{"Small player", 10000000, 0.02, "Small player"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessment := analyzer.QuickCompetitiveAssessment("TestCo", tt.revenue, tt.marketShare)

			assert.Equal(t, tt.expectedPos, assessment["position"])
			assert.NotEmpty(t, assessment["competitiveIntensity"])
			assert.NotEmpty(t, assessment["recommendations"])

			if tt.marketShare > 0 {
				marketSize, exists := assessment["estimatedMarketSize"]
				assert.True(t, exists)
				assert.Greater(t, marketSize.(float64), 0.0)
			}
		})
	}
}

func TestGenerateCompetitiveSummary(t *testing.T) {
	analyzer := NewCompetitiveAnalyzer(nil, nil)

	analysis := &CompetitiveAnalysis{
		DealName: "Test Deal",
		MarketPosition: &MarketPositionAnalysis{
			MarketShare: 0.15,
			MarketRank:  3,
			MarketSize:  2000000000,
			GrowthRate:  0.20,
		},
		Competitors: []CompetitorProfile{
			{Name: "Leader Corp", MarketShare: 0.35, ThreatLevel: "high"},
			{Name: "Challenger Inc", MarketShare: 0.20, ThreatLevel: "medium"},
		},
		StrategicValue: &StrategicValueAssessment{
			OverallScore: 0.75,
			Rationale:    "Strong strategic fit",
		},
		Synergies: &SynergyAnalysis{
			TotalValue:    150000000,
			TimeToRealize: "12-24 months",
			RevenueSynergies: []Synergy{
				{Value: 50000000, Probability: 0.8},
			},
			CostSynergies: []Synergy{
				{Value: 100000000, Probability: 0.7},
			},
		},
		Risks: []CompetitiveRisk{
			{Description: "Competitive response", Likelihood: 0.8, Impact: 0.7},
		},
		Opportunities: []GrowthOpportunity{
			{Description: "Market expansion", PotentialValue: 100000000, Probability: 0.7},
		},
	}

	summary := analyzer.generateCompetitiveSummary(analysis)

	assert.Contains(t, summary, "Test Deal")
	assert.Contains(t, summary, "15.0%") // Market share
	assert.Contains(t, summary, "$2.0B") // Market size
	assert.Contains(t, summary, "Leader Corp")
	assert.Contains(t, summary, "75%")   // Strategic value
	assert.Contains(t, summary, "$150M") // Total synergies
	assert.Contains(t, summary, "Competitive response")
	assert.Contains(t, summary, "Market expansion")
}

func TestAnalyzeCompetitiveLandscapeIntegration(t *testing.T) {
	aiService := &AIService{}
	docProcessor := &DocumentProcessor{}
	analyzer := NewCompetitiveAnalyzer(aiService, docProcessor)

	documents := []DocumentInfo{
		{
			Name: "financials.xlsx",
			Type: DocTypeFinancial,
			Path: "/test/financials.xlsx",
		},
		{
			Name: "legal.pdf",
			Type: DocTypeLegal,
			Path: "/test/legal.pdf",
		},
	}

	marketData := map[string]interface{}{
		"marketShare": 0.18,
		"industryMultiples": map[string]float64{
			"evRevenue": 3.5,
			"evEBITDA":  12.0,
		},
	}

	ctx := context.Background()
	analysis, err := analyzer.AnalyzeCompetitiveLandscape(ctx, "Integration Deal", "TargetCo", documents, marketData)

	require.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, "Integration Deal", analysis.DealName)
	assert.NotNil(t, analysis.MarketPosition)
	assert.NotEmpty(t, analysis.Competitors)
	assert.NotNil(t, analysis.SWOT)
	assert.NotEmpty(t, analysis.MarketTrends)
	assert.NotNil(t, analysis.StrategicValue)
	assert.NotNil(t, analysis.Synergies)
	assert.NotEmpty(t, analysis.Risks)
	assert.NotEmpty(t, analysis.Opportunities)
	assert.NotEmpty(t, analysis.Summary)
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}
