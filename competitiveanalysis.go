package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"DealDone/internal/core/documents"
	"DealDone/internal/infrastructure/ai"
)

// CompetitiveAnalyzer performs competitive analysis for M&A deals
type CompetitiveAnalyzer struct {
	aiService         *ai.AIService
	documentProcessor *documents.DocumentProcessor
}

// NewCompetitiveAnalyzer creates a new competitive analyzer
func NewCompetitiveAnalyzer(aiService *ai.AIService, documentProcessor *documents.DocumentProcessor) *CompetitiveAnalyzer {
	return &CompetitiveAnalyzer{
		aiService:         aiService,
		documentProcessor: documentProcessor,
	}
}

// CompetitiveAnalysis contains comprehensive competitive analysis results
type CompetitiveAnalysis struct {
	DealName       string                    `json:"dealName"`
	AnalysisDate   time.Time                 `json:"analysisDate"`
	MarketPosition *MarketPositionAnalysis   `json:"marketPosition"`
	Competitors    []CompetitorProfile       `json:"competitors"`
	SWOT           *SWOTAnalysis             `json:"swot"`
	MarketTrends   []MarketTrend             `json:"marketTrends"`
	StrategicValue *StrategicValueAssessment `json:"strategicValue"`
	Synergies      *SynergyAnalysis          `json:"synergies"`
	Risks          []CompetitiveRisk         `json:"risks"`
	Opportunities  []GrowthOpportunity       `json:"opportunities"`
	Summary        string                    `json:"summary"`
}

// MarketPositionAnalysis describes the target's position in the market
type MarketPositionAnalysis struct {
	MarketShare     float64              `json:"marketShare"`
	MarketRank      int                  `json:"marketRank"`
	GrowthRate      float64              `json:"growthRate"`
	MarketSize      float64              `json:"marketSize"`
	Segments        []MarketSegment      `json:"segments"`
	GeographicReach []string             `json:"geographicReach"`
	CustomerBase    CustomerBaseAnalysis `json:"customerBase"`
	BrandStrength   float64              `json:"brandStrength"` // 0-1 score
}

// MarketSegment represents a specific market segment
type MarketSegment struct {
	Name          string  `json:"name"`
	Size          float64 `json:"size"`
	Growth        float64 `json:"growth"`
	MarketShare   float64 `json:"marketShare"`
	Profitability float64 `json:"profitability"`
}

// CustomerBaseAnalysis describes the customer base
type CustomerBaseAnalysis struct {
	TotalCustomers    int               `json:"totalCustomers"`
	CustomerSegments  []CustomerSegment `json:"segments"`
	ChurnRate         float64           `json:"churnRate"`
	CustomerLifetime  float64           `json:"customerLifetimeValue"`
	ConcentrationRisk float64           `json:"concentrationRisk"` // 0-1, higher = more concentrated
}

// CustomerSegment represents a customer segment
type CustomerSegment struct {
	Name          string  `json:"name"`
	Percentage    float64 `json:"percentage"`
	Revenue       float64 `json:"revenue"`
	Profitability float64 `json:"profitability"`
	GrowthRate    float64 `json:"growthRate"`
}

// CompetitorProfile contains detailed information about a competitor
type CompetitorProfile struct {
	Name              string            `json:"name"`
	MarketShare       float64           `json:"marketShare"`
	Revenue           float64           `json:"revenue"`
	GrowthRate        float64           `json:"growthRate"`
	Strengths         []string          `json:"strengths"`
	Weaknesses        []string          `json:"weaknesses"`
	Products          []string          `json:"products"`
	GeographicFocus   []string          `json:"geographicFocus"`
	FinancialStrength float64           `json:"financialStrength"` // 0-1 score
	ThreatLevel       string            `json:"threatLevel"`       // "high", "medium", "low"
	RecentMoves       []CompetitiveMove `json:"recentMoves"`
}

// CompetitiveMove represents a recent competitive action
type CompetitiveMove struct {
	Date        time.Time `json:"date"`
	Type        string    `json:"type"` // "acquisition", "product_launch", "expansion", etc.
	Description string    `json:"description"`
	Impact      string    `json:"impact"` // "high", "medium", "low"
}

// SWOTAnalysis contains strengths, weaknesses, opportunities, and threats
type SWOTAnalysis struct {
	Strengths     []SWOTItem `json:"strengths"`
	Weaknesses    []SWOTItem `json:"weaknesses"`
	Opportunities []SWOTItem `json:"opportunities"`
	Threats       []SWOTItem `json:"threats"`
}

// SWOTItem represents a single SWOT element
type SWOTItem struct {
	Description string  `json:"description"`
	Impact      string  `json:"impact"` // "high", "medium", "low"
	Score       float64 `json:"score"`  // 0-1, importance/severity
}

// MarketTrend represents a market trend
type MarketTrend struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Timeline    string  `json:"timeline"`  // "short-term", "medium-term", "long-term"
	Impact      string  `json:"impact"`    // "positive", "negative", "neutral"
	Relevance   float64 `json:"relevance"` // 0-1, how relevant to the deal
}

// StrategicValueAssessment evaluates strategic value
type StrategicValueAssessment struct {
	OverallScore      float64 `json:"overallScore"` // 0-1
	MarketAccessValue float64 `json:"marketAccessValue"`
	TechnologyValue   float64 `json:"technologyValue"`
	CustomerBaseValue float64 `json:"customerBaseValue"`
	BrandValue        float64 `json:"brandValue"`
	OperationalValue  float64 `json:"operationalValue"`
	StrategicFit      float64 `json:"strategicFit"`
	Rationale         string  `json:"rationale"`
}

// SynergyAnalysis identifies potential synergies
type SynergyAnalysis struct {
	RevenueSynergies   []Synergy `json:"revenueSynergies"`
	CostSynergies      []Synergy `json:"costSynergies"`
	TotalValue         float64   `json:"totalValue"`
	TimeToRealize      string    `json:"timeToRealize"`      // e.g., "12-18 months"
	ImplementationRisk string    `json:"implementationRisk"` // "high", "medium", "low"
}

// Synergy represents a specific synergy opportunity
type Synergy struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Timeline    string  `json:"timeline"`
	Probability float64 `json:"probability"` // 0-1
}

// CompetitiveRisk represents a competitive risk
type CompetitiveRisk struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Likelihood  float64 `json:"likelihood"` // 0-1
	Impact      float64 `json:"impact"`     // 0-1
	Mitigation  string  `json:"mitigation"`
}

// GrowthOpportunity represents a growth opportunity
type GrowthOpportunity struct {
	Type           string  `json:"type"`
	Description    string  `json:"description"`
	PotentialValue float64 `json:"potentialValue"`
	TimeHorizon    string  `json:"timeHorizon"`
	Requirements   string  `json:"requirements"`
	Probability    float64 `json:"probability"` // 0-1
}

// AnalyzeCompetitiveLandscape performs comprehensive competitive analysis
func (ca *CompetitiveAnalyzer) AnalyzeCompetitiveLandscape(ctx context.Context, dealName string, targetCompany string, documents []documents.DocumentInfo, marketData map[string]interface{}) (*CompetitiveAnalysis, error) {
	analysis := &CompetitiveAnalysis{
		DealName:      dealName,
		AnalysisDate:  time.Now(),
		Competitors:   make([]CompetitorProfile, 0),
		MarketTrends:  make([]MarketTrend, 0),
		Risks:         make([]CompetitiveRisk, 0),
		Opportunities: make([]GrowthOpportunity, 0),
	}

	// Extract competitive intelligence from documents
	competitiveData, err := ca.extractCompetitiveData(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("failed to extract competitive data: %w", err)
	}

	// Analyze market position
	analysis.MarketPosition = ca.analyzeMarketPosition(targetCompany, competitiveData, marketData)

	// Identify and profile competitors
	analysis.Competitors = ca.profileCompetitors(competitiveData, marketData)

	// Perform SWOT analysis
	analysis.SWOT = ca.performSWOT(targetCompany, competitiveData, analysis.Competitors)

	// Identify market trends
	analysis.MarketTrends = ca.identifyMarketTrends(competitiveData, marketData)

	// Assess strategic value
	analysis.StrategicValue = ca.assessStrategicValue(analysis.MarketPosition, analysis.SWOT, competitiveData)

	// Analyze potential synergies
	analysis.Synergies = ca.analyzeSynergies(targetCompany, competitiveData, marketData)

	// Identify risks and opportunities
	analysis.Risks = ca.identifyCompetitiveRisks(analysis.Competitors, analysis.MarketTrends)
	analysis.Opportunities = ca.identifyGrowthOpportunities(analysis.MarketPosition, analysis.MarketTrends, competitiveData)

	// Generate summary
	analysis.Summary = ca.generateCompetitiveSummary(analysis)

	return analysis, nil
}

// extractCompetitiveData extracts competitive intelligence from documents
func (ca *CompetitiveAnalyzer) extractCompetitiveData(ctx context.Context, documents []documents.DocumentInfo) (map[string]interface{}, error) {
	competitiveData := make(map[string]interface{})

	// Process each document for competitive insights
	for _, doc := range documents {
		// In a real implementation, this would use AI to extract competitive data
		// For now, we'll simulate with document type-based extraction

		switch doc.Type {
		case string(documents.DocTypeFinancial):
			// Extract financial competitive data
			competitiveData["marketSize"] = 1000000000.0 // $1B market
			competitiveData["marketGrowth"] = 0.15       // 15% growth

		case string(documents.DocTypeLegal):
			// Extract regulatory and competitive landscape
			competitiveData["regulatoryBarriers"] = []string{
				"Licensing requirements",
				"Data protection compliance",
			}
		}
	}

	// Use AI for deeper analysis if available
	if ca.aiService != nil && ca.aiService.IsAvailable() {
		// Would call AI service here for competitive intelligence
		competitiveData["aiInsights"] = true
	}

	return competitiveData, nil
}

// analyzeMarketPosition analyzes the target's market position
func (ca *CompetitiveAnalyzer) analyzeMarketPosition(targetCompany string, competitiveData map[string]interface{}, marketData map[string]interface{}) *MarketPositionAnalysis {
	position := &MarketPositionAnalysis{
		MarketShare:     0.15, // Default 15%
		MarketRank:      3,
		GrowthRate:      0.20,         // 20% growth
		MarketSize:      1000000000.0, // $1B
		Segments:        make([]MarketSegment, 0),
		GeographicReach: []string{"North America", "Europe"},
	}

	// Extract from market data if available
	if share, ok := marketData["marketShare"].(float64); ok {
		position.MarketShare = share
	}
	if size, ok := competitiveData["marketSize"].(float64); ok {
		position.MarketSize = size
	}
	if growth, ok := competitiveData["marketGrowth"].(float64); ok {
		position.GrowthRate = growth
	}

	// Define market segments
	position.Segments = []MarketSegment{
		{
			Name:          "Enterprise",
			Size:          position.MarketSize * 0.6,
			Growth:        0.25,
			MarketShare:   0.20,
			Profitability: 0.35,
		},
		{
			Name:          "Mid-Market",
			Size:          position.MarketSize * 0.3,
			Growth:        0.15,
			MarketShare:   0.10,
			Profitability: 0.25,
		},
		{
			Name:          "SMB",
			Size:          position.MarketSize * 0.1,
			Growth:        0.10,
			MarketShare:   0.05,
			Profitability: 0.15,
		},
	}

	// Customer base analysis
	position.CustomerBase = CustomerBaseAnalysis{
		TotalCustomers:    5000,
		ChurnRate:         0.10,  // 10% annual churn
		CustomerLifetime:  50000, // $50k LTV
		ConcentrationRisk: 0.3,   // 30% revenue from top customers
		CustomerSegments: []CustomerSegment{
			{
				Name:          "Enterprise",
				Percentage:    0.20,
				Revenue:       0.60,
				Profitability: 0.40,
				GrowthRate:    0.30,
			},
			{
				Name:          "Mid-Market",
				Percentage:    0.50,
				Revenue:       0.30,
				Profitability: 0.25,
				GrowthRate:    0.20,
			},
			{
				Name:          "SMB",
				Percentage:    0.30,
				Revenue:       0.10,
				Profitability: 0.15,
				GrowthRate:    0.10,
			},
		},
	}

	// Calculate brand strength (0-1 score)
	position.BrandStrength = ca.calculateBrandStrength(position.MarketShare, position.GrowthRate)

	return position
}

// profileCompetitors creates detailed competitor profiles
func (ca *CompetitiveAnalyzer) profileCompetitors(competitiveData map[string]interface{}, marketData map[string]interface{}) []CompetitorProfile {
	// In a real implementation, this would extract from documents and market data
	// For now, create sample competitor profiles

	competitors := []CompetitorProfile{
		{
			Name:              "Market Leader Corp",
			MarketShare:       0.35,
			Revenue:           500000000, // $500M
			GrowthRate:        0.15,
			Strengths:         []string{"Brand recognition", "Scale economies", "Technology platform"},
			Weaknesses:        []string{"Legacy systems", "Slow innovation", "High prices"},
			Products:          []string{"Enterprise Solution", "Cloud Platform", "Analytics Suite"},
			GeographicFocus:   []string{"Global"},
			FinancialStrength: 0.9,
			ThreatLevel:       "high",
			RecentMoves: []CompetitiveMove{
				{
					Date:        time.Now().AddDate(0, -3, 0),
					Type:        "acquisition",
					Description: "Acquired AI startup for $100M",
					Impact:      "high",
				},
			},
		},
		{
			Name:              "Innovative Challenger Inc",
			MarketShare:       0.20,
			Revenue:           250000000, // $250M
			GrowthRate:        0.30,
			Strengths:         []string{"Innovation", "Agility", "Customer focus"},
			Weaknesses:        []string{"Limited scale", "Funding constraints", "Geographic coverage"},
			Products:          []string{"Next-Gen Platform", "Mobile Solution", "API Services"},
			GeographicFocus:   []string{"North America", "Europe"},
			FinancialStrength: 0.7,
			ThreatLevel:       "medium",
			RecentMoves: []CompetitiveMove{
				{
					Date:        time.Now().AddDate(0, -6, 0),
					Type:        "product_launch",
					Description: "Launched AI-powered analytics",
					Impact:      "medium",
				},
			},
		},
		{
			Name:              "Regional Player Ltd",
			MarketShare:       0.10,
			Revenue:           100000000, // $100M
			GrowthRate:        0.10,
			Strengths:         []string{"Local presence", "Customer relationships", "Domain expertise"},
			Weaknesses:        []string{"Limited resources", "Technology gaps", "Scale"},
			Products:          []string{"Regional Solution", "Custom Services"},
			GeographicFocus:   []string{"Asia Pacific"},
			FinancialStrength: 0.5,
			ThreatLevel:       "low",
			RecentMoves:       []CompetitiveMove{},
		},
	}

	// Sort by threat level
	sort.Slice(competitors, func(i, j int) bool {
		threatOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		return threatOrder[competitors[i].ThreatLevel] > threatOrder[competitors[j].ThreatLevel]
	})

	return competitors
}

// performSWOT performs SWOT analysis
func (ca *CompetitiveAnalyzer) performSWOT(targetCompany string, competitiveData map[string]interface{}, competitors []CompetitorProfile) *SWOTAnalysis {
	swot := &SWOTAnalysis{
		Strengths:     make([]SWOTItem, 0),
		Weaknesses:    make([]SWOTItem, 0),
		Opportunities: make([]SWOTItem, 0),
		Threats:       make([]SWOTItem, 0),
	}

	// Strengths
	swot.Strengths = []SWOTItem{
		{
			Description: "Strong technology platform with AI capabilities",
			Impact:      "high",
			Score:       0.9,
		},
		{
			Description: "Established customer base with low churn",
			Impact:      "high",
			Score:       0.85,
		},
		{
			Description: "Experienced management team",
			Impact:      "medium",
			Score:       0.75,
		},
		{
			Description: "Profitable operations with strong margins",
			Impact:      "high",
			Score:       0.8,
		},
	}

	// Weaknesses
	swot.Weaknesses = []SWOTItem{
		{
			Description: "Limited geographic presence",
			Impact:      "medium",
			Score:       0.7,
		},
		{
			Description: "Dependency on key customers",
			Impact:      "high",
			Score:       0.8,
		},
		{
			Description: "Lack of mobile capabilities",
			Impact:      "medium",
			Score:       0.65,
		},
	}

	// Opportunities
	swot.Opportunities = []SWOTItem{
		{
			Description: "Growing market demand for AI solutions",
			Impact:      "high",
			Score:       0.9,
		},
		{
			Description: "Potential for international expansion",
			Impact:      "high",
			Score:       0.85,
		},
		{
			Description: "Cross-selling opportunities to existing customers",
			Impact:      "medium",
			Score:       0.75,
		},
		{
			Description: "Strategic partnerships with tech giants",
			Impact:      "medium",
			Score:       0.7,
		},
	}

	// Threats
	swot.Threats = []SWOTItem{
		{
			Description: fmt.Sprintf("Aggressive competition from %s", competitors[0].Name),
			Impact:      "high",
			Score:       0.85,
		},
		{
			Description: "Rapid technology changes requiring constant innovation",
			Impact:      "medium",
			Score:       0.75,
		},
		{
			Description: "Potential economic downturn affecting IT spending",
			Impact:      "medium",
			Score:       0.7,
		},
		{
			Description: "Increasing cybersecurity threats",
			Impact:      "high",
			Score:       0.8,
		},
	}

	return swot
}

// identifyMarketTrends identifies key market trends
func (ca *CompetitiveAnalyzer) identifyMarketTrends(competitiveData map[string]interface{}, marketData map[string]interface{}) []MarketTrend {
	trends := []MarketTrend{
		{
			Name:        "AI and Machine Learning Adoption",
			Description: "Rapid adoption of AI/ML across all industry segments",
			Timeline:    "short-term",
			Impact:      "positive",
			Relevance:   0.95,
		},
		{
			Name:        "Cloud Migration Acceleration",
			Description: "Enterprises moving to cloud-native solutions",
			Timeline:    "medium-term",
			Impact:      "positive",
			Relevance:   0.9,
		},
		{
			Name:        "Cybersecurity Focus",
			Description: "Increased investment in security solutions",
			Timeline:    "long-term",
			Impact:      "positive",
			Relevance:   0.85,
		},
		{
			Name:        "Market Consolidation",
			Description: "Larger players acquiring smaller competitors",
			Timeline:    "medium-term",
			Impact:      "neutral",
			Relevance:   0.8,
		},
		{
			Name:        "Remote Work Transformation",
			Description: "Permanent shift to hybrid work models",
			Timeline:    "long-term",
			Impact:      "positive",
			Relevance:   0.75,
		},
	}

	// Sort by relevance
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Relevance > trends[j].Relevance
	})

	return trends
}

// assessStrategicValue assesses the strategic value of the acquisition
func (ca *CompetitiveAnalyzer) assessStrategicValue(position *MarketPositionAnalysis, swot *SWOTAnalysis, competitiveData map[string]interface{}) *StrategicValueAssessment {
	assessment := &StrategicValueAssessment{}

	// Calculate individual value components (0-1 scale)
	assessment.MarketAccessValue = ca.calculateMarketAccessValue(position)
	assessment.TechnologyValue = ca.calculateTechnologyValue(swot)
	assessment.CustomerBaseValue = ca.calculateCustomerValue(position)
	assessment.BrandValue = position.BrandStrength
	assessment.OperationalValue = ca.calculateOperationalValue(swot)
	assessment.StrategicFit = ca.calculateStrategicFit(swot)

	// Calculate overall score (weighted average)
	weights := map[string]float64{
		"market":      0.20,
		"technology":  0.25,
		"customer":    0.20,
		"brand":       0.10,
		"operational": 0.15,
		"strategic":   0.10,
	}

	assessment.OverallScore = (assessment.MarketAccessValue * weights["market"]) +
		(assessment.TechnologyValue * weights["technology"]) +
		(assessment.CustomerBaseValue * weights["customer"]) +
		(assessment.BrandValue * weights["brand"]) +
		(assessment.OperationalValue * weights["operational"]) +
		(assessment.StrategicFit * weights["strategic"])

	// Generate rationale
	assessment.Rationale = ca.generateStrategicRationale(assessment)

	return assessment
}

// analyzeSynergies identifies and quantifies potential synergies
func (ca *CompetitiveAnalyzer) analyzeSynergies(targetCompany string, competitiveData map[string]interface{}, marketData map[string]interface{}) *SynergyAnalysis {
	synergies := &SynergyAnalysis{
		RevenueSynergies:   make([]Synergy, 0),
		CostSynergies:      make([]Synergy, 0),
		TimeToRealize:      "12-24 months",
		ImplementationRisk: "medium",
	}

	// Revenue synergies
	synergies.RevenueSynergies = []Synergy{
		{
			Type:        "Cross-selling",
			Description: "Sell acquirer products to target's customer base",
			Value:       50000000, // $50M
			Timeline:    "6-12 months",
			Probability: 0.8,
		},
		{
			Type:        "Geographic expansion",
			Description: "Leverage target's presence in new markets",
			Value:       30000000, // $30M
			Timeline:    "12-18 months",
			Probability: 0.7,
		},
		{
			Type:        "Product bundling",
			Description: "Create integrated solution offerings",
			Value:       20000000, // $20M
			Timeline:    "9-15 months",
			Probability: 0.75,
		},
	}

	// Cost synergies
	synergies.CostSynergies = []Synergy{
		{
			Type:        "Operational efficiency",
			Description: "Consolidate operations and eliminate redundancies",
			Value:       40000000, // $40M
			Timeline:    "12-24 months",
			Probability: 0.85,
		},
		{
			Type:        "Technology consolidation",
			Description: "Merge technology platforms and reduce licensing",
			Value:       25000000, // $25M
			Timeline:    "18-24 months",
			Probability: 0.7,
		},
		{
			Type:        "Procurement savings",
			Description: "Leverage combined purchasing power",
			Value:       15000000, // $15M
			Timeline:    "6-12 months",
			Probability: 0.9,
		},
	}

	// Calculate total value
	totalValue := 0.0
	for _, syn := range synergies.RevenueSynergies {
		totalValue += syn.Value * syn.Probability
	}
	for _, syn := range synergies.CostSynergies {
		totalValue += syn.Value * syn.Probability
	}
	synergies.TotalValue = totalValue

	return synergies
}

// identifyCompetitiveRisks identifies competitive risks
func (ca *CompetitiveAnalyzer) identifyCompetitiveRisks(competitors []CompetitorProfile, trends []MarketTrend) []CompetitiveRisk {
	risks := []CompetitiveRisk{
		{
			Type:        "Competitive response",
			Description: fmt.Sprintf("%s may respond aggressively to acquisition", competitors[0].Name),
			Likelihood:  0.8,
			Impact:      0.7,
			Mitigation:  "Prepare defensive strategies and customer retention programs",
		},
		{
			Type:        "Market share erosion",
			Description: "Competitors may target customers during integration",
			Likelihood:  0.7,
			Impact:      0.6,
			Mitigation:  "Fast-track integration and maintain service quality",
		},
		{
			Type:        "Talent poaching",
			Description: "Key employees may be recruited by competitors",
			Likelihood:  0.6,
			Impact:      0.8,
			Mitigation:  "Implement retention bonuses and career development programs",
		},
		{
			Type:        "Technology disruption",
			Description: "New entrants with disruptive technology",
			Likelihood:  0.5,
			Impact:      0.9,
			Mitigation:  "Invest in R&D and monitor emerging technologies",
		},
		{
			Type:        "Regulatory changes",
			Description: "New regulations may favor competitors",
			Likelihood:  0.4,
			Impact:      0.7,
			Mitigation:  "Engage with regulators and maintain compliance flexibility",
		},
	}

	// Sort by risk score (likelihood * impact)
	sort.Slice(risks, func(i, j int) bool {
		return (risks[i].Likelihood * risks[i].Impact) > (risks[j].Likelihood * risks[j].Impact)
	})

	return risks
}

// identifyGrowthOpportunities identifies growth opportunities
func (ca *CompetitiveAnalyzer) identifyGrowthOpportunities(position *MarketPositionAnalysis, trends []MarketTrend, competitiveData map[string]interface{}) []GrowthOpportunity {
	opportunities := []GrowthOpportunity{
		{
			Type:           "Market expansion",
			Description:    "Enter Asia Pacific market leveraging combined capabilities",
			PotentialValue: 100000000, // $100M
			TimeHorizon:    "2-3 years",
			Requirements:   "Local partnerships and regulatory compliance",
			Probability:    0.7,
		},
		{
			Type:           "Product innovation",
			Description:    "Develop AI-powered next-generation platform",
			PotentialValue: 150000000, // $150M
			TimeHorizon:    "3-5 years",
			Requirements:   "R&D investment and talent acquisition",
			Probability:    0.65,
		},
		{
			Type:           "Vertical integration",
			Description:    "Acquire complementary service providers",
			PotentialValue: 80000000, // $80M
			TimeHorizon:    "1-2 years",
			Requirements:   "M&A capability and integration expertise",
			Probability:    0.75,
		},
		{
			Type:           "Platform ecosystem",
			Description:    "Build partner ecosystem around combined platform",
			PotentialValue: 60000000, // $60M
			TimeHorizon:    "2-3 years",
			Requirements:   "API development and partner programs",
			Probability:    0.8,
		},
		{
			Type:           "Enterprise upsell",
			Description:    "Move mid-market customers to enterprise tier",
			PotentialValue: 40000000, // $40M
			TimeHorizon:    "1-2 years",
			Requirements:   "Sales enablement and product enhancements",
			Probability:    0.85,
		},
	}

	// Sort by expected value (potential * probability)
	sort.Slice(opportunities, func(i, j int) bool {
		return (opportunities[i].PotentialValue * opportunities[i].Probability) >
			(opportunities[j].PotentialValue * opportunities[j].Probability)
	})

	return opportunities
}

// Helper methods

func (ca *CompetitiveAnalyzer) calculateBrandStrength(marketShare, growthRate float64) float64 {
	// Simple brand strength calculation
	return math.Min(1.0, (marketShare*2+growthRate)/2)
}

func (ca *CompetitiveAnalyzer) calculateMarketAccessValue(position *MarketPositionAnalysis) float64 {
	// Based on market share, growth, and geographic reach
	geoScore := float64(len(position.GeographicReach)) / 10.0 // Normalize to 0-1
	return math.Min(1.0, (position.MarketShare*3+position.GrowthRate*2+geoScore)/3)
}

func (ca *CompetitiveAnalyzer) calculateTechnologyValue(swot *SWOTAnalysis) float64 {
	// Based on technology-related strengths
	techScore := 0.0
	techCount := 0

	for _, strength := range swot.Strengths {
		if strings.Contains(strings.ToLower(strength.Description), "technology") ||
			strings.Contains(strings.ToLower(strength.Description), "platform") ||
			strings.Contains(strings.ToLower(strength.Description), "ai") {
			techScore += strength.Score
			techCount++
		}
	}

	if techCount > 0 {
		return techScore / float64(techCount)
	}
	return 0.5 // Default medium value
}

func (ca *CompetitiveAnalyzer) calculateCustomerValue(position *MarketPositionAnalysis) float64 {
	// Based on customer metrics
	churnScore := 1.0 - position.CustomerBase.ChurnRate                 // Lower churn = higher score
	concentrationScore := 1.0 - position.CustomerBase.ConcentrationRisk // Lower concentration = higher score

	return (churnScore + concentrationScore) / 2
}

func (ca *CompetitiveAnalyzer) calculateOperationalValue(swot *SWOTAnalysis) float64 {
	// Based on operational strengths minus weaknesses
	strengthScore := 0.0
	for _, strength := range swot.Strengths {
		if strings.Contains(strings.ToLower(strength.Description), "operation") ||
			strings.Contains(strings.ToLower(strength.Description), "margin") ||
			strings.Contains(strings.ToLower(strength.Description), "efficien") {
			strengthScore += strength.Score
		}
	}

	weaknessScore := 0.0
	for _, weakness := range swot.Weaknesses {
		if strings.Contains(strings.ToLower(weakness.Description), "operation") ||
			strings.Contains(strings.ToLower(weakness.Description), "cost") {
			weaknessScore += weakness.Score
		}
	}

	return math.Max(0, math.Min(1.0, strengthScore-weaknessScore/2))
}

func (ca *CompetitiveAnalyzer) calculateStrategicFit(swot *SWOTAnalysis) float64 {
	// Based on opportunities and threat mitigation potential
	oppScore := 0.0
	for _, opp := range swot.Opportunities {
		oppScore += opp.Score
	}

	if len(swot.Opportunities) > 0 {
		return math.Min(1.0, oppScore/float64(len(swot.Opportunities)))
	}
	return 0.5
}

func (ca *CompetitiveAnalyzer) generateStrategicRationale(assessment *StrategicValueAssessment) string {
	rationale := "Strategic acquisition rationale: "

	highValueAreas := []string{}
	if assessment.MarketAccessValue > 0.7 {
		highValueAreas = append(highValueAreas, "strong market position")
	}
	if assessment.TechnologyValue > 0.7 {
		highValueAreas = append(highValueAreas, "valuable technology assets")
	}
	if assessment.CustomerBaseValue > 0.7 {
		highValueAreas = append(highValueAreas, "loyal customer base")
	}

	if len(highValueAreas) > 0 {
		rationale += fmt.Sprintf("The target offers %s", strings.Join(highValueAreas, ", "))
		rationale += fmt.Sprintf(". Overall strategic value score of %.0f%% indicates a %s strategic fit.",
			assessment.OverallScore*100,
			ca.getValueDescription(assessment.OverallScore))
	}

	return rationale
}

func (ca *CompetitiveAnalyzer) getValueDescription(score float64) string {
	switch {
	case score >= 0.8:
		return "highly attractive"
	case score >= 0.6:
		return "good"
	case score >= 0.4:
		return "moderate"
	default:
		return "limited"
	}
}

// generateCompetitiveSummary generates an executive summary
func (ca *CompetitiveAnalyzer) generateCompetitiveSummary(analysis *CompetitiveAnalysis) string {
	summary := fmt.Sprintf("Competitive Analysis Summary for %s:\n\n", analysis.DealName)

	// Market position
	summary += fmt.Sprintf("Market Position: The target holds a %.1f%% market share (rank #%d) in a $%.0fB market growing at %.0f%% annually.\n\n",
		analysis.MarketPosition.MarketShare*100,
		analysis.MarketPosition.MarketRank,
		analysis.MarketPosition.MarketSize/1000000000,
		analysis.MarketPosition.GrowthRate*100)

	// Key competitors
	summary += "Key Competitors:\n"
	for i, comp := range analysis.Competitors {
		if i >= 3 {
			break
		}
		summary += fmt.Sprintf("- %s (%.1f%% share, %s threat)\n",
			comp.Name, comp.MarketShare*100, comp.ThreatLevel)
	}
	summary += "\n"

	// Strategic value
	if analysis.StrategicValue != nil {
		summary += fmt.Sprintf("Strategic Value: %.0f%% overall score. %s\n\n",
			analysis.StrategicValue.OverallScore*100,
			analysis.StrategicValue.Rationale)
	}

	// Synergies
	if analysis.Synergies != nil {
		summary += fmt.Sprintf("Synergy Potential: $%.0fM total value achievable in %s\n",
			analysis.Synergies.TotalValue/1000000,
			analysis.Synergies.TimeToRealize)

		revSynergies := 0.0
		for _, syn := range analysis.Synergies.RevenueSynergies {
			revSynergies += syn.Value * syn.Probability
		}
		costSynergies := 0.0
		for _, syn := range analysis.Synergies.CostSynergies {
			costSynergies += syn.Value * syn.Probability
		}

		summary += fmt.Sprintf("- Revenue synergies: $%.0fM\n", revSynergies/1000000)
		summary += fmt.Sprintf("- Cost synergies: $%.0fM\n\n", costSynergies/1000000)
	}

	// Top risks and opportunities
	summary += "Top Risks:\n"
	for i, risk := range analysis.Risks {
		if i >= 3 {
			break
		}
		riskScore := risk.Likelihood * risk.Impact
		summary += fmt.Sprintf("- %s (risk score: %.0f%%)\n", risk.Description, riskScore*100)
	}

	summary += "\nTop Opportunities:\n"
	for i, opp := range analysis.Opportunities {
		if i >= 3 {
			break
		}
		expValue := opp.PotentialValue * opp.Probability
		summary += fmt.Sprintf("- %s ($%.0fM expected value)\n", opp.Description, expValue/1000000)
	}

	return summary
}

// QuickCompetitiveAssessment performs a quick competitive assessment
func (ca *CompetitiveAnalyzer) QuickCompetitiveAssessment(targetCompany string, revenue float64, marketShare float64) map[string]interface{} {
	assessment := make(map[string]interface{})

	// Quick competitive position
	if marketShare > 0.3 {
		assessment["position"] = "Market leader"
	} else if marketShare > 0.15 {
		assessment["position"] = "Strong challenger"
	} else if marketShare > 0.05 {
		assessment["position"] = "Niche player"
	} else {
		assessment["position"] = "Small player"
	}

	// Estimated market size
	if marketShare > 0 {
		assessment["estimatedMarketSize"] = revenue / marketShare
	}

	// Competitive intensity
	assessment["competitiveIntensity"] = ca.estimateCompetitiveIntensity(marketShare)

	// Strategic recommendations
	recommendations := []string{}
	if marketShare < 0.1 {
		recommendations = append(recommendations, "Consider market consolidation opportunities")
	}
	if marketShare > 0.2 {
		recommendations = append(recommendations, "Leverage market position for premium pricing")
	}
	recommendations = append(recommendations, "Focus on differentiation and innovation")

	assessment["recommendations"] = recommendations

	return assessment
}

func (ca *CompetitiveAnalyzer) estimateCompetitiveIntensity(marketShare float64) string {
	// Simple heuristic based on market share distribution
	if marketShare > 0.4 {
		return "Low - dominant position"
	} else if marketShare > 0.2 {
		return "Medium - oligopolistic market"
	} else {
		return "High - fragmented market"
	}
}
