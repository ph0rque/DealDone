package main

import (
	"fmt"
	"math"
	"time"

	"DealDone/internal/domain/analysis"
	"DealDone/internal/infrastructure/ai"
)

// DealValuationCalculator performs various valuation methods for M&A deals
type DealValuationCalculator struct {
	aiService      *ai.AIService
	financialCache map[string]*analysis.FinancialAnalysis
}

// NewDealValuationCalculator creates a new deal valuation calculator
func NewDealValuationCalculator(aiService *ai.AIService) *DealValuationCalculator {
	return &DealValuationCalculator{
		aiService:      aiService,
		financialCache: make(map[string]*analysis.FinancialAnalysis),
	}
}

// ValuationResult contains the results of various valuation methods
type ValuationResult struct {
	DealName      string                 `json:"dealName"`
	ValuationDate time.Time              `json:"valuationDate"`
	DCFValuation  *DCFResult             `json:"dcfValuation,omitempty"`
	Multiples     *MultiplesValuation    `json:"multiples,omitempty"`
	Comps         *ComparableAnalysis    `json:"comparables,omitempty"`
	AssetBased    *AssetBasedValuation   `json:"assetBased,omitempty"`
	SummaryRange  *ValuationRange        `json:"summaryRange"`
	Confidence    float64                `json:"confidence"`
	Assumptions   map[string]interface{} `json:"assumptions"`
	Warnings      []string               `json:"warnings"`
}

// DCFResult contains discounted cash flow analysis results
type DCFResult struct {
	EnterpriseValue float64            `json:"enterpriseValue"`
	EquityValue     float64            `json:"equityValue"`
	TerminalValue   float64            `json:"terminalValue"`
	WACC            float64            `json:"wacc"`
	GrowthRate      float64            `json:"growthRate"`
	ProjectedCF     []float64          `json:"projectedCashFlows"`
	PresentValues   []float64          `json:"presentValues"`
	Assumptions     map[string]float64 `json:"assumptions"`
}

// MultiplesValuation contains valuation based on financial multiples
type MultiplesValuation struct {
	EVToRevenue     *MultipleResult `json:"evToRevenue,omitempty"`
	EVToEBITDA      *MultipleResult `json:"evToEBITDA,omitempty"`
	PEMultiple      *MultipleResult `json:"peMultiple,omitempty"`
	PriceToBook     *MultipleResult `json:"priceToBook,omitempty"`
	IndustryAverage float64         `json:"industryAverage"`
}

// MultipleResult contains a single multiple calculation
type MultipleResult struct {
	Multiple      float64 `json:"multiple"`
	ImpliedValue  float64 `json:"impliedValue"`
	IndustryRange Range   `json:"industryRange"`
}

// ComparableAnalysis contains comparable company analysis
type ComparableAnalysis struct {
	ComparableCompanies []ComparableCompany `json:"comparableCompanies"`
	MedianMultiple      float64             `json:"medianMultiple"`
	ImpliedValue        float64             `json:"impliedValue"`
	Adjustments         map[string]float64  `json:"adjustments"`
}

// ComparableCompany represents a comparable company
type ComparableCompany struct {
	Name       string  `json:"name"`
	Ticker     string  `json:"ticker,omitempty"`
	EVMultiple float64 `json:"evMultiple"`
	Revenue    float64 `json:"revenue"`
	EBITDA     float64 `json:"ebitda"`
	MarketCap  float64 `json:"marketCap"`
	Similarity float64 `json:"similarityScore"`
}

// AssetBasedValuation contains asset-based valuation
type AssetBasedValuation struct {
	TotalAssets      float64            `json:"totalAssets"`
	TotalLiabilities float64            `json:"totalLiabilities"`
	NetAssetValue    float64            `json:"netAssetValue"`
	Adjustments      map[string]float64 `json:"adjustments"`
	IntangibleValue  float64            `json:"intangibleValue"`
}

// ValuationRange represents a range of values
type ValuationRange struct {
	Low      float64 `json:"low"`
	Mid      float64 `json:"mid"`
	High     float64 `json:"high"`
	Currency string  `json:"currency"`
}

// Range represents a simple min/max range
type Range struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// CalculateValuation performs comprehensive deal valuation
func (dvc *DealValuationCalculator) CalculateValuation(dealName string, financialData *analysis.FinancialAnalysis, marketData map[string]interface{}) (*ValuationResult, error) {
	result := &ValuationResult{
		DealName:      dealName,
		ValuationDate: time.Now(),
		Assumptions:   make(map[string]interface{}),
		Warnings:      make([]string, 0),
	}

	// Cache financial data
	dvc.financialCache[dealName] = financialData

	// Perform DCF valuation
	dcf, err := dvc.calculateDCF(financialData, marketData)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("DCF calculation failed: %v", err))
	} else {
		result.DCFValuation = dcf
	}

	// Calculate multiples-based valuation
	multiples := dvc.calculateMultiples(financialData, marketData)
	result.Multiples = multiples

	// Perform comparable analysis
	comps := dvc.performComparableAnalysis(financialData, marketData)
	result.Comps = comps

	// Calculate asset-based valuation
	assetBased := dvc.calculateAssetBased(financialData)
	result.AssetBased = assetBased

	// Calculate summary range
	result.SummaryRange = dvc.calculateSummaryRange(result)

	// Calculate overall confidence
	result.Confidence = dvc.calculateConfidence(result, financialData)

	// Set assumptions used
	result.Assumptions = dvc.extractAssumptions(marketData)

	return result, nil
}

// calculateDCF performs discounted cash flow analysis
func (dvc *DealValuationCalculator) calculateDCF(financial *analysis.FinancialAnalysis, marketData map[string]interface{}) (*DCFResult, error) {
	dcf := &DCFResult{
		Assumptions: make(map[string]float64),
	}

	// Extract or calculate WACC
	wacc := 0.10 // Default 10%
	if w, ok := marketData["wacc"].(float64); ok {
		wacc = w
	}
	dcf.WACC = wacc

	// Terminal growth rate
	growthRate := 0.03 // Default 3%
	if g, ok := marketData["terminalGrowth"].(float64); ok {
		growthRate = g
	}
	dcf.GrowthRate = growthRate

	// Project cash flows (5 years)
	baseCashFlow := financial.CashFlow
	if baseCashFlow == 0 {
		// Estimate from EBITDA if cash flow not available
		baseCashFlow = financial.EBITDA * 0.7 // Rough conversion
	}

	// Growth assumptions
	yearlyGrowth := []float64{0.15, 0.12, 0.10, 0.08, 0.05} // Declining growth

	dcf.ProjectedCF = make([]float64, 5)
	dcf.PresentValues = make([]float64, 5)

	totalPV := 0.0
	for i := 0; i < 5; i++ {
		if i == 0 {
			dcf.ProjectedCF[i] = baseCashFlow * (1 + yearlyGrowth[i])
		} else {
			dcf.ProjectedCF[i] = dcf.ProjectedCF[i-1] * (1 + yearlyGrowth[i])
		}

		// Calculate present value
		discountFactor := math.Pow(1+wacc, float64(i+1))
		dcf.PresentValues[i] = dcf.ProjectedCF[i] / discountFactor
		totalPV += dcf.PresentValues[i]
	}

	// Calculate terminal value
	terminalCF := dcf.ProjectedCF[4] * (1 + growthRate)
	dcf.TerminalValue = terminalCF / (wacc - growthRate)

	// Present value of terminal value
	terminalPV := dcf.TerminalValue / math.Pow(1+wacc, 5)

	// Enterprise value
	dcf.EnterpriseValue = totalPV + terminalPV

	// Calculate equity value (EV - Net Debt)
	netDebt := financial.TotalLiabilities * 0.6 // Assume 60% is debt
	dcf.EquityValue = dcf.EnterpriseValue - netDebt

	// Store assumptions
	dcf.Assumptions["baseCashFlow"] = baseCashFlow
	dcf.Assumptions["avgGrowthRate"] = 0.10
	dcf.Assumptions["netDebt"] = netDebt

	return dcf, nil
}

// calculateMultiples calculates valuation based on financial multiples
func (dvc *DealValuationCalculator) calculateMultiples(financial *analysis.FinancialAnalysis, marketData map[string]interface{}) *MultiplesValuation {
	multiples := &MultiplesValuation{}

	// Get industry multiples
	industryEVRevenue := 2.5 // Default
	industryEVEBITDA := 10.0 // Default
	industryPE := 20.0       // Default
	industryPB := 3.0        // Default

	if indMultiples, ok := marketData["industryMultiples"].(map[string]float64); ok {
		if m, exists := indMultiples["evRevenue"]; exists {
			industryEVRevenue = m
		}
		if m, exists := indMultiples["evEBITDA"]; exists {
			industryEVEBITDA = m
		}
		if m, exists := indMultiples["pe"]; exists {
			industryPE = m
		}
		if m, exists := indMultiples["pb"]; exists {
			industryPB = m
		}
	}

	// EV/Revenue multiple
	if financial.Revenue > 0 {
		multiples.EVToRevenue = &MultipleResult{
			Multiple:     industryEVRevenue,
			ImpliedValue: financial.Revenue * industryEVRevenue,
			IndustryRange: Range{
				Min: industryEVRevenue * 0.8,
				Max: industryEVRevenue * 1.2,
			},
		}
	}

	// EV/EBITDA multiple
	if financial.EBITDA > 0 {
		multiples.EVToEBITDA = &MultipleResult{
			Multiple:     industryEVEBITDA,
			ImpliedValue: financial.EBITDA * industryEVEBITDA,
			IndustryRange: Range{
				Min: industryEVEBITDA * 0.8,
				Max: industryEVEBITDA * 1.2,
			},
		}
	}

	// P/E multiple
	if financial.NetIncome > 0 {
		multiples.PEMultiple = &MultipleResult{
			Multiple:     industryPE,
			ImpliedValue: financial.NetIncome * industryPE,
			IndustryRange: Range{
				Min: industryPE * 0.7,
				Max: industryPE * 1.3,
			},
		}
	}

	// Price/Book multiple
	bookValue := financial.TotalAssets - financial.TotalLiabilities
	if bookValue > 0 {
		multiples.PriceToBook = &MultipleResult{
			Multiple:     industryPB,
			ImpliedValue: bookValue * industryPB,
			IndustryRange: Range{
				Min: industryPB * 0.8,
				Max: industryPB * 1.2,
			},
		}
	}

	// Calculate average industry multiple
	count := 0
	total := 0.0
	if multiples.EVToRevenue != nil {
		total += multiples.EVToRevenue.Multiple
		count++
	}
	if multiples.EVToEBITDA != nil {
		total += multiples.EVToEBITDA.Multiple
		count++
	}
	if count > 0 {
		multiples.IndustryAverage = total / float64(count)
	}

	return multiples
}

// performComparableAnalysis performs comparable company analysis
func (dvc *DealValuationCalculator) performComparableAnalysis(financial *analysis.FinancialAnalysis, marketData map[string]interface{}) *ComparableAnalysis {
	comps := &ComparableAnalysis{
		ComparableCompanies: make([]ComparableCompany, 0),
		Adjustments:         make(map[string]float64),
	}

	// Get comparable companies from market data
	if compData, ok := marketData["comparables"].([]interface{}); ok {
		for _, comp := range compData {
			if compMap, ok := comp.(map[string]interface{}); ok {
				comparable := ComparableCompany{
					Name:      getStringValue(compMap, "name"),
					Ticker:    getStringValue(compMap, "ticker"),
					Revenue:   getFloatValue(compMap, "revenue"),
					EBITDA:    getFloatValue(compMap, "ebitda"),
					MarketCap: getFloatValue(compMap, "marketCap"),
				}

				// Calculate EV multiple
				if comparable.EBITDA > 0 {
					comparable.EVMultiple = comparable.MarketCap / comparable.EBITDA
				}

				// Calculate similarity score
				comparable.Similarity = dvc.calculateSimilarity(financial, &comparable)

				comps.ComparableCompanies = append(comps.ComparableCompanies, comparable)
			}
		}
	}

	// If no comparables provided, use default industry comparables
	if len(comps.ComparableCompanies) == 0 {
		comps.ComparableCompanies = dvc.getDefaultComparables(financial)
	}

	// Calculate median multiple
	if len(comps.ComparableCompanies) > 0 {
		multiples := make([]float64, 0)
		for _, comp := range comps.ComparableCompanies {
			if comp.EVMultiple > 0 {
				multiples = append(multiples, comp.EVMultiple)
			}
		}

		if len(multiples) > 0 {
			comps.MedianMultiple = calculateMedian(multiples)
			comps.ImpliedValue = financial.EBITDA * comps.MedianMultiple
		}
	}

	// Apply adjustments
	comps.Adjustments["sizeDiscount"] = -0.10 // 10% small company discount
	comps.Adjustments["marketPremium"] = 0.05 // 5% market conditions

	return comps
}

// calculateAssetBased performs asset-based valuation
func (dvc *DealValuationCalculator) calculateAssetBased(financial *analysis.FinancialAnalysis) *AssetBasedValuation {
	assetBased := &AssetBasedValuation{
		TotalAssets:      financial.TotalAssets,
		TotalLiabilities: financial.TotalLiabilities,
		Adjustments:      make(map[string]float64),
	}

	// Calculate net asset value
	assetBased.NetAssetValue = assetBased.TotalAssets - assetBased.TotalLiabilities

	// Apply adjustments for market value
	assetBased.Adjustments["marketValueAdjustment"] = assetBased.NetAssetValue * 0.15 // 15% markup
	assetBased.Adjustments["workingCapitalAdj"] = financial.Revenue * 0.1             // 10% of revenue

	// Estimate intangible value
	if financial.Revenue > 0 {
		// Simple excess earnings method
		tangibleReturn := assetBased.NetAssetValue * 0.08 // 8% return on tangible assets
		excessEarnings := financial.NetIncome - tangibleReturn
		if excessEarnings > 0 {
			assetBased.IntangibleValue = excessEarnings * 5 // 5x multiple on excess earnings
		}
	}

	// Add adjustments to NAV
	for _, adj := range assetBased.Adjustments {
		assetBased.NetAssetValue += adj
	}
	assetBased.NetAssetValue += assetBased.IntangibleValue

	return assetBased
}

// calculateSummaryRange calculates the valuation range from all methods
func (dvc *DealValuationCalculator) calculateSummaryRange(result *ValuationResult) *ValuationRange {
	values := make([]float64, 0)

	// Collect all valuation results
	if result.DCFValuation != nil && result.DCFValuation.EquityValue > 0 {
		values = append(values, result.DCFValuation.EquityValue)
	}

	if result.Multiples != nil {
		if result.Multiples.EVToEBITDA != nil {
			values = append(values, result.Multiples.EVToEBITDA.ImpliedValue)
		}
		if result.Multiples.EVToRevenue != nil {
			values = append(values, result.Multiples.EVToRevenue.ImpliedValue)
		}
	}

	if result.Comps != nil && result.Comps.ImpliedValue > 0 {
		values = append(values, result.Comps.ImpliedValue)
	}

	if result.AssetBased != nil && result.AssetBased.NetAssetValue > 0 {
		values = append(values, result.AssetBased.NetAssetValue)
	}

	// Calculate range
	if len(values) == 0 {
		return &ValuationRange{
			Low:      0,
			Mid:      0,
			High:     0,
			Currency: "USD",
		}
	}

	min, max := findMinMax(values)

	return &ValuationRange{
		Low:      min * 0.9, // 10% discount
		Mid:      calculateMean(values),
		High:     max * 1.1, // 10% premium
		Currency: "USD",
	}
}

// calculateConfidence calculates overall confidence in the valuation
func (dvc *DealValuationCalculator) calculateConfidence(result *ValuationResult, financial *analysis.FinancialAnalysis) float64 {
	confidence := 0.0
	factors := 0

	// Check data quality
	if financial.Confidence > 0 {
		confidence += financial.Confidence
		factors++
	}

	// Check number of valuation methods used
	methodsUsed := 0
	if result.DCFValuation != nil {
		methodsUsed++
	}
	if result.Multiples != nil {
		methodsUsed++
	}
	if result.Comps != nil && len(result.Comps.ComparableCompanies) > 0 {
		methodsUsed++
	}
	if result.AssetBased != nil {
		methodsUsed++
	}

	methodConfidence := float64(methodsUsed) / 4.0
	confidence += methodConfidence
	factors++

	// Check consistency of results
	if result.SummaryRange != nil && result.SummaryRange.High > 0 {
		spread := (result.SummaryRange.High - result.SummaryRange.Low) / result.SummaryRange.Mid
		if spread < 0.3 { // Less than 30% spread
			confidence += 0.9
		} else if spread < 0.5 {
			confidence += 0.7
		} else {
			confidence += 0.5
		}
		factors++
	}

	if factors > 0 {
		return confidence / float64(factors)
	}

	return 0.5 // Default medium confidence
}

// extractAssumptions extracts key assumptions from market data
func (dvc *DealValuationCalculator) extractAssumptions(marketData map[string]interface{}) map[string]interface{} {
	assumptions := make(map[string]interface{})

	// Extract all relevant assumptions
	keyAssumptions := []string{
		"wacc", "terminalGrowth", "riskFreeRate", "marketPremium",
		"industryBeta", "taxRate", "inflationRate",
	}

	for _, key := range keyAssumptions {
		if value, exists := marketData[key]; exists {
			assumptions[key] = value
		}
	}

	// Add default assumptions if not provided
	if _, exists := assumptions["wacc"]; !exists {
		assumptions["wacc"] = 0.10
	}
	if _, exists := assumptions["terminalGrowth"]; !exists {
		assumptions["terminalGrowth"] = 0.03
	}
	if _, exists := assumptions["taxRate"]; !exists {
		assumptions["taxRate"] = 0.25
	}

	return assumptions
}

// calculateSimilarity calculates similarity score between target and comparable
func (dvc *DealValuationCalculator) calculateSimilarity(target *analysis.FinancialAnalysis, comp *ComparableCompany) float64 {
	score := 0.0
	factors := 0.0

	// Revenue similarity
	if target.Revenue > 0 && comp.Revenue > 0 {
		revRatio := math.Min(target.Revenue, comp.Revenue) / math.Max(target.Revenue, comp.Revenue)
		score += revRatio
		factors++
	}

	// EBITDA margin similarity
	if target.Revenue > 0 && comp.Revenue > 0 {
		targetMargin := target.EBITDA / target.Revenue
		compMargin := comp.EBITDA / comp.Revenue
		marginDiff := math.Abs(targetMargin - compMargin)
		marginScore := 1.0 - math.Min(marginDiff*2, 1.0) // Max 50% difference
		score += marginScore
		factors++
	}

	// Growth similarity (would need historical data in real implementation)
	// For now, assume similar growth
	score += 0.7
	factors++

	if factors > 0 {
		return score / factors
	}

	return 0.5 // Default medium similarity
}

// getDefaultComparables returns default comparable companies
func (dvc *DealValuationCalculator) getDefaultComparables(financial *analysis.FinancialAnalysis) []ComparableCompany {
	// In a real implementation, this would fetch from a database or API
	// For now, return synthetic comparables based on the target

	baseRevenue := financial.Revenue
	baseEBITDA := financial.EBITDA

	return []ComparableCompany{
		{
			Name:       "Industry Peer A",
			Revenue:    baseRevenue * 1.2,
			EBITDA:     baseEBITDA * 1.1,
			MarketCap:  baseEBITDA * 12,
			EVMultiple: 12,
			Similarity: 0.85,
		},
		{
			Name:       "Industry Peer B",
			Revenue:    baseRevenue * 0.8,
			EBITDA:     baseEBITDA * 0.9,
			MarketCap:  baseEBITDA * 10,
			EVMultiple: 10,
			Similarity: 0.80,
		},
		{
			Name:       "Industry Peer C",
			Revenue:    baseRevenue * 1.5,
			EBITDA:     baseEBITDA * 1.4,
			MarketCap:  baseEBITDA * 14,
			EVMultiple: 14,
			Similarity: 0.75,
		},
	}
}

// Helper functions

func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Sort values
	sorted := make([]float64, len(values))
	copy(sorted, values)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func findMinMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}

	min, max := values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloatValue(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	return 0
}

// CalculateQuickValuation performs a quick valuation based on limited data
func (dvc *DealValuationCalculator) CalculateQuickValuation(revenue, ebitda, netIncome float64) *ValuationRange {
	// Simple multiple-based quick valuation
	values := make([]float64, 0)

	if revenue > 0 {
		values = append(values, revenue*2.5) // 2.5x revenue
	}
	if ebitda > 0 {
		values = append(values, ebitda*10) // 10x EBITDA
	}
	if netIncome > 0 {
		values = append(values, netIncome*20) // 20x earnings
	}

	if len(values) == 0 {
		return &ValuationRange{Currency: "USD"}
	}

	mean := calculateMean(values)
	return &ValuationRange{
		Low:      mean * 0.8,
		Mid:      mean,
		High:     mean * 1.2,
		Currency: "USD",
	}
}

// GenerateValuationReport generates a text summary of the valuation
func (dvc *DealValuationCalculator) GenerateValuationReport(result *ValuationResult) string {
	report := fmt.Sprintf("Valuation Report for %s\n", result.DealName)
	report += fmt.Sprintf("Date: %s\n\n", result.ValuationDate.Format("January 2, 2006"))

	report += "Executive Summary:\n"
	if result.SummaryRange != nil {
		report += fmt.Sprintf("Valuation Range: $%.2fM - $%.2fM (Mid: $%.2fM)\n",
			result.SummaryRange.Low/1000000,
			result.SummaryRange.High/1000000,
			result.SummaryRange.Mid/1000000)
	}
	report += fmt.Sprintf("Confidence Level: %.0f%%\n\n", result.Confidence*100)

	// DCF Results
	if result.DCFValuation != nil {
		report += "Discounted Cash Flow Analysis:\n"
		report += fmt.Sprintf("  Enterprise Value: $%.2fM\n", result.DCFValuation.EnterpriseValue/1000000)
		report += fmt.Sprintf("  Equity Value: $%.2fM\n", result.DCFValuation.EquityValue/1000000)
		report += fmt.Sprintf("  WACC: %.1f%%\n", result.DCFValuation.WACC*100)
		report += fmt.Sprintf("  Terminal Growth: %.1f%%\n\n", result.DCFValuation.GrowthRate*100)
	}

	// Multiples
	if result.Multiples != nil {
		report += "Multiples Analysis:\n"
		if result.Multiples.EVToEBITDA != nil {
			report += fmt.Sprintf("  EV/EBITDA: %.1fx (Implied Value: $%.2fM)\n",
				result.Multiples.EVToEBITDA.Multiple,
				result.Multiples.EVToEBITDA.ImpliedValue/1000000)
		}
		if result.Multiples.EVToRevenue != nil {
			report += fmt.Sprintf("  EV/Revenue: %.1fx (Implied Value: $%.2fM)\n",
				result.Multiples.EVToRevenue.Multiple,
				result.Multiples.EVToRevenue.ImpliedValue/1000000)
		}
		report += "\n"
	}

	// Comparables
	if result.Comps != nil && len(result.Comps.ComparableCompanies) > 0 {
		report += fmt.Sprintf("Comparable Company Analysis:\n")
		report += fmt.Sprintf("  Number of Comparables: %d\n", len(result.Comps.ComparableCompanies))
		report += fmt.Sprintf("  Median Multiple: %.1fx\n", result.Comps.MedianMultiple)
		report += fmt.Sprintf("  Implied Value: $%.2fM\n\n", result.Comps.ImpliedValue/1000000)
	}

	// Warnings
	if len(result.Warnings) > 0 {
		report += "Warnings:\n"
		for _, warning := range result.Warnings {
			report += fmt.Sprintf("  - %s\n", warning)
		}
	}

	return report
}
