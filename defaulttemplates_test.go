package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultTemplateGenerator(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dealdone-default-templates-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templatesPath := filepath.Join(tempDir, "Templates")
	dtg := NewDefaultTemplateGenerator(templatesPath)

	t.Run("GenerateDefaultTemplates creates all templates", func(t *testing.T) {
		err := dtg.GenerateDefaultTemplates()
		if err != nil {
			t.Fatalf("Failed to generate default templates: %v", err)
		}

		// Check that templates directory was created
		if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
			t.Error("Templates directory was not created")
		}

		// Check each template file exists
		expectedTemplates := dtg.GetDefaultTemplateNames()
		for _, template := range expectedTemplates {
			templatePath := filepath.Join(templatesPath, template)
			if _, err := os.Stat(templatePath); os.IsNotExist(err) {
				t.Errorf("Template %s was not created", template)
			}
		}
	})

	t.Run("HasDefaultTemplates detects existing templates", func(t *testing.T) {
		// Initially should not have templates
		dtg2 := NewDefaultTemplateGenerator(filepath.Join(tempDir, "Templates2"))
		if dtg2.HasDefaultTemplates() {
			t.Error("HasDefaultTemplates should return false for empty directory")
		}

		// Generate templates
		os.MkdirAll(filepath.Join(tempDir, "Templates2"), 0755)
		dtg2.GenerateDefaultTemplates()

		// Now should have templates
		if !dtg2.HasDefaultTemplates() {
			t.Error("HasDefaultTemplates should return true after generation")
		}
	})

	t.Run("GenerateDefaultTemplates doesn't overwrite existing", func(t *testing.T) {
		// Create a custom financial model
		customPath := filepath.Join(templatesPath, "Financial_Model_Template.csv")
		customContent := []byte("Custom content")
		os.WriteFile(customPath, customContent, 0644)

		// Generate templates again
		err := dtg.GenerateDefaultTemplates()
		if err != nil {
			t.Fatalf("Failed to generate templates: %v", err)
		}

		// Check that custom content is preserved
		content, err := os.ReadFile(customPath)
		if err != nil {
			t.Fatalf("Failed to read custom template: %v", err)
		}

		if string(content) != "Custom content" {
			t.Error("Default template generation overwrote existing file")
		}
	})

	t.Run("Financial model template has correct structure", func(t *testing.T) {
		// Create a fresh directory for this test
		testTemplatesPath := filepath.Join(tempDir, "TestFinancial")
		testDtg := NewDefaultTemplateGenerator(testTemplatesPath)

		// Generate templates
		err := testDtg.GenerateDefaultTemplates()
		if err != nil {
			t.Fatalf("Failed to generate templates: %v", err)
		}

		financialPath := filepath.Join(testTemplatesPath, "Financial_Model_Template.csv")
		content, err := os.ReadFile(financialPath)
		if err != nil {
			t.Fatalf("Failed to read financial template: %v", err)
		}

		// Check for key elements
		contentStr := string(content)
		expectedElements := []string{
			"Financial Model Template",
			"Income Statement",
			"Revenue",
			"EBITDA",
			"Net Income",
			"Key Metrics",
		}

		for _, element := range expectedElements {
			if !strings.Contains(contentStr, element) {
				t.Errorf("Financial template missing element: %s", element)
			}
		}
	})

	t.Run("Due diligence checklist has correct categories", func(t *testing.T) {
		checklistPath := filepath.Join(templatesPath, "Due_Diligence_Checklist.csv")
		content, err := os.ReadFile(checklistPath)
		if err != nil {
			t.Fatalf("Failed to read checklist template: %v", err)
		}

		// Check for key categories
		contentStr := string(content)
		categories := []string{
			"Financial",
			"Legal",
			"Operations",
			"Market",
		}

		for _, category := range categories {
			if !strings.Contains(contentStr, category) {
				t.Errorf("Checklist missing category: %s", category)
			}
		}
	})

	t.Run("Deal summary template has all sections", func(t *testing.T) {
		summaryPath := filepath.Join(templatesPath, "Deal_Summary_Template.txt")
		content, err := os.ReadFile(summaryPath)
		if err != nil {
			t.Fatalf("Failed to read summary template: %v", err)
		}

		// Check for key sections
		contentStr := string(content)
		sections := []string{
			"EXECUTIVE SUMMARY",
			"COMPANY OVERVIEW",
			"FINANCIAL HIGHLIGHTS",
			"INVESTMENT RATIONALE",
			"TRANSACTION STRUCTURE",
			"KEY RISKS",
			"GROWTH OPPORTUNITIES",
			"NEXT STEPS",
			"CONTACTS",
		}

		for _, section := range sections {
			if !strings.Contains(contentStr, section) {
				t.Errorf("Deal summary missing section: %s", section)
			}
		}
	})
}
