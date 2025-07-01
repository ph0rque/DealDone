package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTemplateDiscovery(t *testing.T) {
	t.Run("NewTemplateDiscovery creates discovery service", func(t *testing.T) {
		config, err := NewConfigService()
		if err != nil {
			t.Fatalf("Failed to create config service: %v", err)
		}
		tm := NewTemplateManager(config)
		td := NewTemplateDiscovery(tm)

		if td == nil {
			t.Fatal("Expected TemplateDiscovery to be created")
		}
		if td.templateManager != tm {
			t.Error("Expected templateManager to be set")
		}
		if td.metadataCache == nil {
			t.Error("Expected metadataCache to be initialized")
		}
	})

	t.Run("detectCategory identifies categories correctly", func(t *testing.T) {
		config, err := NewConfigService()
		if err != nil {
			t.Fatalf("Failed to create config service: %v", err)
		}
		tm := NewTemplateManager(config)
		td := NewTemplateDiscovery(tm)

		tests := []struct {
			name     string
			path     string
			expected string
		}{
			{"Financial Model", "templates/financial/model.xlsx", "financial"},
			{"Revenue Forecast", "templates/revenue_forecast.xlsx", "financial"},
			{"Contract Template", "templates/legal/contract.docx", "legal"},
			{"NDA Agreement", "templates/nda_agreement.docx", "legal"},
			{"General Report", "templates/report.docx", "general"},
			{"Meeting Notes", "templates/notes.docx", "general"},
		}

		for _, test := range tests {
			result := td.detectCategory(test.name, test.path)
			if result != test.expected {
				t.Errorf("detectCategory(%s, %s) = %s, want %s",
					test.name, test.path, result, test.expected)
			}
		}
	})

	t.Run("generateTags creates relevant tags", func(t *testing.T) {
		config, err := NewConfigService()
		if err != nil {
			t.Fatalf("Failed to create config service: %v", err)
		}
		tm := NewTemplateManager(config)
		td := NewTemplateDiscovery(tm)

		tags := td.generateTags("Financial Model Template", "financial")

		// Should include category
		found := false
		for _, tag := range tags {
			if tag == "financial" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'financial' tag to be included")
		}

		// Should include relevant words
		expectedWords := []string{"model"}
		for _, word := range expectedWords {
			found := false
			for _, tag := range tags {
				if tag == word {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected '%s' tag to be included", word)
			}
		}
	})

	t.Run("DiscoverTemplates with metadata files", func(t *testing.T) {
		// Create test directory
		testDir := "test_templates_discovery"
		os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		config := &ConfigService{
			config: &Config{
				DealDoneRoot: testDir,
			},
		}

		// Create template and metadata files
		templatePath := filepath.Join(testDir, "Templates", "test_template.xlsx")
		os.MkdirAll(filepath.Dir(templatePath), 0755)
		os.WriteFile(templatePath, []byte("test"), 0644)

		// Create metadata file
		metadata := &TemplateMetadata{
			ID:          "test-template",
			Name:        "Test Template",
			Description: "A test template",
			Category:    "financial",
			Type:        "xlsx",
			Version:     "1.0",
			Author:      "Test",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Tags:        []string{"test", "financial"},
		}

		metadataPath := filepath.Join(testDir, "Templates", "test_template.meta.json")
		metadataFile, _ := os.Create(metadataPath)
		json.NewEncoder(metadataFile).Encode(metadata)
		metadataFile.Close()

		tm := NewTemplateManager(config)
		td := NewTemplateDiscovery(tm)

		templates, err := td.DiscoverTemplates()
		if err != nil {
			t.Fatalf("DiscoverTemplates failed: %v", err)
		}

		if len(templates) != 1 {
			t.Fatalf("Expected 1 template, got %d", len(templates))
		}

		template := templates[0]
		if !template.HasMetadata {
			t.Error("Expected template to have metadata")
		}
		if template.Metadata.ID != "test-template" {
			t.Errorf("Expected metadata ID 'test-template', got %s", template.Metadata.ID)
		}
	})

	t.Run("SearchTemplates filters correctly", func(t *testing.T) {
		// Create test directory
		testDir := "test_templates_search"
		os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		config := &ConfigService{
			config: &Config{
				DealDoneRoot: testDir,
			},
		}

		// Create multiple templates
		templatesDir := filepath.Join(testDir, "Templates")
		os.MkdirAll(templatesDir, 0755)

		// Financial template
		os.WriteFile(filepath.Join(templatesDir, "financial_model.xlsx"), []byte("test"), 0644)
		// Legal template
		os.WriteFile(filepath.Join(templatesDir, "contract_template.docx"), []byte("test"), 0644)
		// General template
		os.WriteFile(filepath.Join(templatesDir, "report.docx"), []byte("test"), 0644)

		tm := NewTemplateManager(config)
		td := NewTemplateDiscovery(tm)

		// Search by query
		results, err := td.SearchTemplates("financial", nil)
		if err != nil {
			t.Fatalf("SearchTemplates failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result for 'financial' query, got %d", len(results))
		}

		// Search by category filter
		filters := map[string]string{"category": "legal"}
		results, err = td.SearchTemplates("", filters)
		if err != nil {
			t.Fatalf("SearchTemplates with filter failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result for legal category filter, got %d", len(results))
		}

		// Search by type filter
		filters = map[string]string{"type": "docx"}
		results, err = td.SearchTemplates("", filters)
		if err != nil {
			t.Fatalf("SearchTemplates with type filter failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results for docx type filter, got %d", len(results))
		}
	})

	t.Run("SaveTemplateMetadata creates metadata file", func(t *testing.T) {
		// Create test directory
		testDir := "test_templates_metadata"
		os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		config := &ConfigService{
			config: &Config{
				DealDoneRoot: testDir,
			},
		}

		// Create template file
		templatePath := filepath.Join(testDir, "Templates", "test.xlsx")
		os.MkdirAll(filepath.Dir(templatePath), 0755)
		os.WriteFile(templatePath, []byte("test"), 0644)

		tm := NewTemplateManager(config)
		td := NewTemplateDiscovery(tm)

		metadata := &TemplateMetadata{
			ID:          "test",
			Name:        "Test",
			Description: "Test template",
			Category:    "general",
			Type:        "xlsx",
			Version:     "1.0",
			Author:      "Test",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := td.SaveTemplateMetadata(templatePath, metadata)
		if err != nil {
			t.Fatalf("SaveTemplateMetadata failed: %v", err)
		}

		// Check metadata file exists
		metadataPath := filepath.Join(testDir, "Templates", "test.meta.json")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			t.Error("Expected metadata file to be created")
		}

		// Check cache is updated
		if cached, exists := td.metadataCache[templatePath]; !exists || cached.ID != "test" {
			t.Error("Expected metadata cache to be updated")
		}
	})

	t.Run("OrganizeTemplatesByCategory groups templates", func(t *testing.T) {
		// Create test directory
		testDir := "test_templates_organize"
		os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		config := &ConfigService{
			config: &Config{
				DealDoneRoot: testDir,
			},
		}

		// Create templates in different categories
		templatesDir := filepath.Join(testDir, "Templates")
		os.MkdirAll(templatesDir, 0755)

		os.WriteFile(filepath.Join(templatesDir, "financial_model.xlsx"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(templatesDir, "revenue_forecast.xlsx"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(templatesDir, "contract.docx"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(templatesDir, "general_report.docx"), []byte("test"), 0644)

		tm := NewTemplateManager(config)
		td := NewTemplateDiscovery(tm)

		organized, err := td.OrganizeTemplatesByCategory()
		if err != nil {
			t.Fatalf("OrganizeTemplatesByCategory failed: %v", err)
		}

		// Check financial category
		if len(organized["financial"]) != 2 {
			t.Errorf("Expected 2 financial templates, got %d", len(organized["financial"]))
		}

		// Check legal category
		if len(organized["legal"]) != 1 {
			t.Errorf("Expected 1 legal template, got %d", len(organized["legal"]))
		}

		// Check general category
		if len(organized["general"]) != 1 {
			t.Errorf("Expected 1 general template, got %d", len(organized["general"]))
		}
	})

	t.Run("ImportTemplate copies and categorizes", func(t *testing.T) {
		// Create test directories
		testDir := "test_templates_import"
		sourceDir := "test_source_templates"
		os.MkdirAll(testDir, 0755)
		os.MkdirAll(sourceDir, 0755)
		defer os.RemoveAll(testDir)
		defer os.RemoveAll(sourceDir)

		config := &ConfigService{
			config: &Config{
				DealDoneRoot: testDir,
			},
		}

		// Create source template
		sourcePath := filepath.Join(sourceDir, "import_test.xlsx")
		os.WriteFile(sourcePath, []byte("test content"), 0644)

		tm := NewTemplateManager(config)
		td := NewTemplateDiscovery(tm)

		metadata := &TemplateMetadata{
			ID:          "import-test",
			Name:        "Import Test",
			Description: "Imported template",
			Category:    "financial",
			Type:        "xlsx",
			Version:     "1.0",
			Author:      "Test",
		}

		err := td.ImportTemplate(sourcePath, metadata)
		if err != nil {
			t.Fatalf("ImportTemplate failed: %v", err)
		}

		// Check file was copied to correct location
		expectedPath := filepath.Join(testDir, "Templates", "financial", "import_test.xlsx")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Error("Expected template to be copied to financial category folder")
		}

		// Check metadata file was created
		metadataPath := filepath.Join(testDir, "Templates", "financial", "import_test.meta.json")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			t.Error("Expected metadata file to be created")
		}
	})
}
