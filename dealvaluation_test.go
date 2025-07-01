package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDealValuationCalculator(t *testing.T) {
	aiService := &AIService{}
	calc := NewDealValuationCalculator(aiService)

	assert.NotNil(t, calc)
	assert.Equal(t, aiService, calc.aiService)
	assert.NotNil(t, calc.financialCache)
}

func TestCalculateDCF(t *testing.T) {
	calc := NewDealValuationCalculator(nil)

	financial := &FinancialAnalysis{
		Revenue:          10000000, // $10M
		EBITDA:           2000000,  // $2M
		NetIncome:        1500000,  // $1.5M
		CashFlow:         1800000,  // $1.8M
		TotalAssets:      15000000, // $15M
		TotalLiabilities: 5000000,  // $5M
	}

	marketData := map[string]interface{}{
		"wacc":           0.12,  // 12%
		"terminalGrowth": 0.025, // 2.5%
	}

	dcf, err := calc.calculateDCF(financial, marketData)

	require.NoError(t, err)
	assert.NotNil(t, dcf)
	assert.Equal(t, 0.12, dcf.WACC)
	assert.Equal(t, 0.025, dcf.GrowthRate)
	assert.Len(t, dcf.ProjectedCF, 5)
	assert.Len(t, dcf.PresentValues, 5)
	assert.Greater(t, dcf.EnterpriseValue, 0.0)
	assert.Greater(t, dcf.EquityValue, 0.0)
	assert.Greater(t, dcf.TerminalValue, 0.0)

	// Check that projected cash flows are increasing (initially)
	assert.Greater(t, dcf.ProjectedCF[1], dcf.ProjectedCF[0])

	// Check that present values are discounted
	assert.Less(t, dcf.PresentValues[0], dcf.ProjectedCF[0])
}

func TestCalculateMultiples(t *testing.T) {
	calc := NewDealValuationCalculator(nil)

	financial := &FinancialAnalysis{
		Revenue:          10000000, // $10M
		EBITDA:           2000000,  // $2M
		NetIncome:        1500000,  // $1.5M
		TotalAssets:      15000000, // $15M
		TotalLiabilities: 5000000,  // $5M
	}

	marketData := map[string]interface{}{
		"industryMultiples": map[string]float64{
			"evRevenue": 3.0,
			"evEBITDA":  12.0,
			"pe":        25.0,
			"pb":        4.0,
		},
	}

	multiples := calc.calculateMultiples(financial, marketData)

	assert.NotNil(t, multiples)
	assert.NotNil(t, multiples.EVToRevenue)
	assert.Equal(t, 3.0, multiples.EVToRevenue.Multiple)
	assert.Equal(t, 30000000.0, multiples.EVToRevenue.ImpliedValue) // $30M

	assert.NotNil(t, multiples.EVToEBITDA)
	assert.Equal(t, 12.0, multiples.EVToEBITDA.Multiple)
	assert.Equal(t, 24000000.0, multiples.EVToEBITDA.ImpliedValue) // $24M

	assert.NotNil(t, multiples.PEMultiple)
	assert.Equal(t, 25.0, multiples.PEMultiple.Multiple)

	assert.NotNil(t, multiples.PriceToBook)
	assert.Greater(t, multiples.IndustryAverage, 0.0)
}

func TestPerformComparableAnalysis(t *testing.T) {
	calc := NewDealValuationCalculator(nil)

	financial := &FinancialAnalysis{
		Revenue: 10000000, // $10M
		EBITDA:  2000000,  // $2M
	}

	marketData := map[string]interface{}{
		"comparables": []interface{}{
			map[string]interface{}{
				"name":      "Comp A",
				"ticker":    "CMPA",
				"revenue":   12000000.0,
				"ebitda":    2400000.0,
				"marketCap": 36000000.0,
			},
			map[string]interface{}{
				"name":      "Comp B",
				"ticker":    "CMPB",
				"revenue":   8000000.0,
				"ebitda":    1800000.0,
				"marketCap": 22000000.0,
			},
		},
	}

	comps := calc.performComparableAnalysis(financial, marketData)

	assert.NotNil(t, comps)
	assert.Len(t, comps.ComparableCompanies, 2)
	assert.Greater(t, comps.MedianMultiple, 0.0)
	assert.Greater(t, comps.ImpliedValue, 0.0)
	assert.Contains(t, comps.Adjustments, "sizeDiscount")
	assert.Contains(t, comps.Adjustments, "marketPremium")
}

func TestCalculateAssetBased(t *testing.T) {
	calc := NewDealValuationCalculator(nil)

	financial := &FinancialAnalysis{
		Revenue:          10000000, // $10M
		NetIncome:        1500000,  // $1.5M
		TotalAssets:      15000000, // $15M
		TotalLiabilities: 5000000,  // $5M
	}

	assetBased := calc.calculateAssetBased(financial)

	assert.NotNil(t, assetBased)
	assert.Equal(t, 15000000.0, assetBased.TotalAssets)
	assert.Equal(t, 5000000.0, assetBased.TotalLiabilities)
	assert.Greater(t, assetBased.NetAssetValue, 10000000.0) // Should be > book value due to adjustments
	assert.Greater(t, assetBased.IntangibleValue, 0.0)
	assert.Contains(t, assetBased.Adjustments, "marketValueAdjustment")
	assert.Contains(t, assetBased.Adjustments, "workingCapitalAdj")
}

func TestCalculateSummaryRange(t *testing.T) {
	calc := NewDealValuationCalculator(nil)

	result := &ValuationResult{
		DCFValuation: &DCFResult{
			EquityValue: 25000000, // $25M
		},
		Multiples: &MultiplesValuation{
			EVToEBITDA: &MultipleResult{
				ImpliedValue: 24000000, // $24M
			},
			EVToRevenue: &MultipleResult{
				ImpliedValue: 30000000, // $30M
			},
		},
		Comps: &ComparableAnalysis{
			ImpliedValue: 26000000, // $26M
		},
		AssetBased: &AssetBasedValuation{
			NetAssetValue: 12000000, // $12M
		},
	}

	summaryRange := calc.calculateSummaryRange(result)

	assert.NotNil(t, summaryRange)
	assert.Greater(t, summaryRange.Low, 0.0)
	assert.Greater(t, summaryRange.Mid, summaryRange.Low)
	assert.Greater(t, summaryRange.High, summaryRange.Mid)
	assert.Equal(t, "USD", summaryRange.Currency)

	// Check that low is ~10% below minimum
	assert.Less(t, summaryRange.Low, 12000000.0)

	// Check that high is ~10% above maximum
	assert.Greater(t, summaryRange.High, 30000000.0)
}

func TestCalculateConfidence(t *testing.T) {
	calc := NewDealValuationCalculator(nil)

	financial := &FinancialAnalysis{
		Confidence: 0.8,
	}

	// High confidence scenario - all methods used, tight range
	result1 := &ValuationResult{
		DCFValuation: &DCFResult{},
		Multiples:    &MultiplesValuation{},
		Comps: &ComparableAnalysis{
			ComparableCompanies: []ComparableCompany{{}, {}},
		},
		AssetBased: &AssetBasedValuation{},
		SummaryRange: &ValuationRange{
			Low:  20000000,
			Mid:  22000000,
			High: 24000000,
		},
	}

	confidence1 := calc.calculateConfidence(result1, financial)
	assert.Greater(t, confidence1, 0.8) // Should be high

	// Low confidence scenario - few methods, wide range
	result2 := &ValuationResult{
		DCFValuation: &DCFResult{},
		SummaryRange: &ValuationRange{
			Low:  10000000,
			Mid:  20000000,
			High: 30000000,
		},
	}

	confidence2 := calc.calculateConfidence(result2, financial)
	assert.Less(t, confidence2, confidence1) // Should be lower
}

func TestCalculateSimilarity(t *testing.T) {
	calc := NewDealValuationCalculator(nil)

	target := &FinancialAnalysis{
		Revenue: 10000000, // $10M
		EBITDA:  2000000,  // $2M (20% margin)
	}

	// Very similar company
	comp1 := &ComparableCompany{
		Revenue: 11000000, // $11M
		EBITDA:  2200000,  // $2.2M (20% margin)
	}

	similarity1 := calc.calculateSimilarity(target, comp1)
	assert.Greater(t, similarity1, 0.8) // Should be high

	// Less similar company
	comp2 := &ComparableCompany{
		Revenue: 50000000, // $50M
		EBITDA:  5000000,  // $5M (10% margin)
	}

	similarity2 := calc.calculateSimilarity(target, comp2)
	assert.Less(t, similarity2, similarity1) // Should be lower
}

func TestCalculateQuickValuation(t *testing.T) {
	calc := NewDealValuationCalculator(nil)

	valRange := calc.CalculateQuickValuation(10000000, 2000000, 1500000)

	assert.NotNil(t, valRange)
	assert.Greater(t, valRange.Low, 0.0)
	assert.Greater(t, valRange.Mid, valRange.Low)
	assert.Greater(t, valRange.High, valRange.Mid)
	assert.Equal(t, "USD", valRange.Currency)
}

func TestGenerateValuationReport(t *testing.T) {
	calc := NewDealValuationCalculator(nil)

	result := &ValuationResult{
		DealName:      "Test Deal",
		ValuationDate: time.Now(),
		DCFValuation: &DCFResult{
			EnterpriseValue: 30000000,
			EquityValue:     25000000,
			WACC:            0.12,
			GrowthRate:      0.03,
		},
		Multiples: &MultiplesValuation{
			EVToEBITDA: &MultipleResult{
				Multiple:     12.0,
				ImpliedValue: 24000000,
			},
		},
		SummaryRange: &ValuationRange{
			Low:  20000000,
			Mid:  25000000,
			High: 30000000,
		},
		Confidence: 0.85,
		Warnings:   []string{"Limited historical data"},
	}

	report := calc.GenerateValuationReport(result)

	assert.Contains(t, report, "Test Deal")
	assert.Contains(t, report, "Valuation Range:")
	assert.Contains(t, report, "Discounted Cash Flow Analysis:")
	assert.Contains(t, report, "Multiples Analysis:")
	assert.Contains(t, report, "85%")                     // Confidence
	assert.Contains(t, report, "Limited historical data") // Warning
}

func TestHelperFunctions(t *testing.T) {
	// Test calculateMedian
	t.Run("calculateMedian", func(t *testing.T) {
		assert.Equal(t, 3.0, calculateMedian([]float64{1, 2, 3, 4, 5}))
		assert.Equal(t, 3.5, calculateMedian([]float64{1, 2, 4, 5}))
		assert.Equal(t, 0.0, calculateMedian([]float64{}))
	})

	// Test calculateMean
	t.Run("calculateMean", func(t *testing.T) {
		assert.Equal(t, 3.0, calculateMean([]float64{1, 2, 3, 4, 5}))
		assert.Equal(t, 0.0, calculateMean([]float64{}))
	})

	// Test findMinMax
	t.Run("findMinMax", func(t *testing.T) {
		min, max := findMinMax([]float64{5, 2, 8, 1, 9})
		assert.Equal(t, 1.0, min)
		assert.Equal(t, 9.0, max)

		min2, max2 := findMinMax([]float64{})
		assert.Equal(t, 0.0, min2)
		assert.Equal(t, 0.0, max2)
	})

	// Test getStringValue
	t.Run("getStringValue", func(t *testing.T) {
		m := map[string]interface{}{
			"name": "Test",
			"age":  30,
		}
		assert.Equal(t, "Test", getStringValue(m, "name"))
		assert.Equal(t, "", getStringValue(m, "missing"))
		assert.Equal(t, "", getStringValue(m, "age"))
	})

	// Test getFloatValue
	t.Run("getFloatValue", func(t *testing.T) {
		m := map[string]interface{}{
			"price":    100.5,
			"quantity": 10,
			"name":     "Test",
		}
		assert.Equal(t, 100.5, getFloatValue(m, "price"))
		assert.Equal(t, 10.0, getFloatValue(m, "quantity"))
		assert.Equal(t, 0.0, getFloatValue(m, "missing"))
		assert.Equal(t, 0.0, getFloatValue(m, "name"))
	})
}

func TestCalculateValuationIntegration(t *testing.T) {
	calc := NewDealValuationCalculator(nil)

	financial := &FinancialAnalysis{
		Revenue:          20000000, // $20M
		EBITDA:           4000000,  // $4M
		NetIncome:        3000000,  // $3M
		CashFlow:         3500000,  // $3.5M
		TotalAssets:      30000000, // $30M
		TotalLiabilities: 10000000, // $10M
		Confidence:       0.85,
	}

	marketData := map[string]interface{}{
		"wacc":           0.11,
		"terminalGrowth": 0.03,
		"industryMultiples": map[string]float64{
			"evRevenue": 2.5,
			"evEBITDA":  11.0,
			"pe":        22.0,
			"pb":        3.5,
		},
		"comparables": []interface{}{
			map[string]interface{}{
				"name":      "Peer Co A",
				"revenue":   25000000.0,
				"ebitda":    5000000.0,
				"marketCap": 60000000.0,
			},
		},
	}

	result, err := calc.CalculateValuation("Integration Test Deal", financial, marketData)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Integration Test Deal", result.DealName)
	assert.NotNil(t, result.DCFValuation)
	assert.NotNil(t, result.Multiples)
	assert.NotNil(t, result.Comps)
	assert.NotNil(t, result.AssetBased)
	assert.NotNil(t, result.SummaryRange)
	assert.Greater(t, result.Confidence, 0.5)
	assert.NotEmpty(t, result.Assumptions)

	// Verify financial data was cached
	cached, exists := calc.financialCache["Integration Test Deal"]
	assert.True(t, exists)
	assert.Equal(t, financial, cached)
}
