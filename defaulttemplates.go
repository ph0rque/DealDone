package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultTemplateGenerator creates example templates for new installations
type DefaultTemplateGenerator struct {
	templatesPath string
}

// NewDefaultTemplateGenerator creates a new template generator
func NewDefaultTemplateGenerator(templatesPath string) *DefaultTemplateGenerator {
	return &DefaultTemplateGenerator{
		templatesPath: templatesPath,
	}
}

// GenerateDefaultTemplates creates a set of basic example templates
func (dtg *DefaultTemplateGenerator) GenerateDefaultTemplates() error {
	// Ensure templates directory exists
	if err := os.MkdirAll(dtg.templatesPath, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Generate each template
	if err := dtg.createFinancialModelTemplate(); err != nil {
		return fmt.Errorf("failed to create financial model template: %w", err)
	}

	if err := dtg.createDueDiligenceChecklistTemplate(); err != nil {
		return fmt.Errorf("failed to create due diligence checklist: %w", err)
	}

	if err := dtg.createDealSummaryTemplate(); err != nil {
		return fmt.Errorf("failed to create deal summary template: %w", err)
	}

	return nil
}

// createFinancialModelTemplate creates a basic financial model CSV template
func (dtg *DefaultTemplateGenerator) createFinancialModelTemplate() error {
	filename := filepath.Join(dtg.templatesPath, "Financial_Model_Template.csv")

	// Check if already exists
	if _, err := os.Stat(filename); err == nil {
		return nil // Already exists, skip
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	headers := [][]string{
		{"Financial Model Template - DealDone"},
		{"Company Name:", "[To be filled]"},
		{"Deal Date:", "[To be filled]"},
		{""},
		{"Income Statement", "Year 1", "Year 2", "Year 3", "Year 4", "Year 5"},
	}

	for _, row := range headers {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Revenue section
	revenue := [][]string{
		{"Revenue", "", "", "", "", ""},
		{"  Product Revenue", "0", "0", "0", "0", "0"},
		{"  Service Revenue", "0", "0", "0", "0", "0"},
		{"Total Revenue", "=B7+B8", "=C7+C8", "=D7+D8", "=E7+E8", "=F7+F8"},
		{""},
		{"Operating Expenses", "", "", "", "", ""},
		{"  Cost of Goods Sold", "0", "0", "0", "0", "0"},
		{"  Sales & Marketing", "0", "0", "0", "0", "0"},
		{"  General & Administrative", "0", "0", "0", "0", "0"},
		{"  Research & Development", "0", "0", "0", "0", "0"},
		{"Total Operating Expenses", "=B12+B13+B14+B15", "=C12+C13+C14+C15", "=D12+D13+D14+D15", "=E12+E13+E14+E15", "=F12+F13+F14+F15"},
		{""},
		{"EBITDA", "=B9-B16", "=C9-C16", "=D9-D16", "=E9-E16", "=F9-F16"},
		{"  Depreciation", "0", "0", "0", "0", "0"},
		{"  Amortization", "0", "0", "0", "0", "0"},
		{"EBIT", "=B18-B19-B20", "=C18-C19-C20", "=D18-D19-D20", "=E18-E19-E20", "=F18-F19-F20"},
		{"  Interest Expense", "0", "0", "0", "0", "0"},
		{"Pre-tax Income", "=B21-B22", "=C21-C22", "=D21-D22", "=E21-E22", "=F21-F22"},
		{"  Tax", "0", "0", "0", "0", "0"},
		{"Net Income", "=B23-B24", "=C23-C24", "=D23-D24", "=E23-E24", "=F23-F24"},
		{""},
		{"Key Metrics", "", "", "", "", ""},
		{"Revenue Growth %", "", "=(C9-B9)/B9*100", "=(D9-C9)/C9*100", "=(E9-D9)/D9*100", "=(F9-E9)/E9*100"},
		{"EBITDA Margin %", "=B18/B9*100", "=C18/C9*100", "=D18/D9*100", "=E18/E9*100", "=F18/F9*100"},
		{"Net Margin %", "=B25/B9*100", "=C25/C9*100", "=D25/D9*100", "=E25/E9*100", "=F25/F9*100"},
	}

	for _, row := range revenue {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// createDueDiligenceChecklistTemplate creates a due diligence checklist
func (dtg *DefaultTemplateGenerator) createDueDiligenceChecklistTemplate() error {
	filename := filepath.Join(dtg.templatesPath, "Due_Diligence_Checklist.csv")

	// Check if already exists
	if _, err := os.Stat(filename); err == nil {
		return nil // Already exists, skip
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Checklist items
	checklist := [][]string{
		{"Due Diligence Checklist - DealDone"},
		{"Company:", "[To be filled]"},
		{"Date:", "[To be filled]"},
		{""},
		{"Category", "Item", "Status", "Notes", "Reviewer"},
		{"Financial", "3 Years Financial Statements", "Pending", "", ""},
		{"Financial", "Tax Returns (3 years)", "Pending", "", ""},
		{"Financial", "Accounts Receivable Aging", "Pending", "", ""},
		{"Financial", "Accounts Payable Aging", "Pending", "", ""},
		{"Financial", "Bank Statements", "Pending", "", ""},
		{"Financial", "Debt Schedules", "Pending", "", ""},
		{""},
		{"Legal", "Corporate Structure", "Pending", "", ""},
		{"Legal", "Material Contracts", "Pending", "", ""},
		{"Legal", "Litigation History", "Pending", "", ""},
		{"Legal", "Intellectual Property", "Pending", "", ""},
		{"Legal", "Employment Agreements", "Pending", "", ""},
		{"Legal", "Regulatory Compliance", "Pending", "", ""},
		{""},
		{"Operations", "Customer List", "Pending", "", ""},
		{"Operations", "Supplier Agreements", "Pending", "", ""},
		{"Operations", "Product/Service Details", "Pending", "", ""},
		{"Operations", "Operational Metrics", "Pending", "", ""},
		{"Operations", "IT Systems Overview", "Pending", "", ""},
		{""},
		{"Market", "Market Analysis", "Pending", "", ""},
		{"Market", "Competitive Analysis", "Pending", "", ""},
		{"Market", "Growth Strategy", "Pending", "", ""},
		{"Market", "Customer Concentration", "Pending", "", ""},
	}

	for _, row := range checklist {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// createDealSummaryTemplate creates a deal summary template
func (dtg *DefaultTemplateGenerator) createDealSummaryTemplate() error {
	filename := filepath.Join(dtg.templatesPath, "Deal_Summary_Template.txt")

	// Check if already exists
	if _, err := os.Stat(filename); err == nil {
		return nil // Already exists, skip
	}

	content := `DEAL SUMMARY TEMPLATE - DealDone
================================

EXECUTIVE SUMMARY
-----------------
Deal Name: [To be filled]
Target Company: [To be filled]
Deal Type: [Acquisition/Merger/Investment]
Deal Value: $[Amount]
Date: [Date]

COMPANY OVERVIEW
----------------
Company Name: [Name]
Industry: [Industry]
Founded: [Year]
Headquarters: [Location]
Employees: [Number]
Website: [URL]

Business Description:
[Provide a brief description of the company's business model, products/services, and market position]

FINANCIAL HIGHLIGHTS
--------------------
Revenue (Last Year): $[Amount]
EBITDA (Last Year): $[Amount]
EBITDA Margin: [%]
Revenue Growth (3-yr CAGR): [%]

Key Financial Metrics:
- [Metric 1]: [Value]
- [Metric 2]: [Value]
- [Metric 3]: [Value]

INVESTMENT RATIONALE
--------------------
1. [Key reason 1]
2. [Key reason 2]
3. [Key reason 3]

TRANSACTION STRUCTURE
---------------------
Purchase Price: $[Amount]
Enterprise Value: $[Amount]
EV/Revenue Multiple: [X]
EV/EBITDA Multiple: [X]
Equity Investment: $[Amount]
Debt Financing: $[Amount]

KEY RISKS
----------
1. [Risk 1]
2. [Risk 2]
3. [Risk 3]

GROWTH OPPORTUNITIES
--------------------
1. [Opportunity 1]
2. [Opportunity 2]
3. [Opportunity 3]

NEXT STEPS
-----------
1. [Action item 1]
2. [Action item 2]
3. [Action item 3]

CONTACTS
---------
Deal Lead: [Name, Title, Email, Phone]
Financial Advisor: [Name, Company, Email]
Legal Counsel: [Name, Firm, Email]

APPENDICES
-----------
- Financial Statements
- Due Diligence Report
- Valuation Analysis
- Legal Documentation
`

	return os.WriteFile(filename, []byte(content), 0644)
}

// HasDefaultTemplates checks if default templates already exist
func (dtg *DefaultTemplateGenerator) HasDefaultTemplates() bool {
	templates := []string{
		"Financial_Model_Template.csv",
		"Due_Diligence_Checklist.csv",
		"Deal_Summary_Template.txt",
	}

	for _, template := range templates {
		path := filepath.Join(dtg.templatesPath, template)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// GetDefaultTemplateNames returns the names of default templates
func (dtg *DefaultTemplateGenerator) GetDefaultTemplateNames() []string {
	return []string{
		"Financial_Model_Template.csv",
		"Due_Diligence_Checklist.csv",
		"Deal_Summary_Template.txt",
	}
}
