package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestTemplate(path string, name string) error {
	fullPath := filepath.Join(path, name)
	return os.WriteFile(fullPath, []byte("test template content"), 0644)
}

func TestTemplateManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dealdone-template-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config service with test directory
	cs := &ConfigService{
		config: &Config{
			DealDoneRoot: filepath.Join(tempDir, "DealDone"),
		},
	}

	// Create folder structure
	fm := NewFolderManager(cs)
	if err := fm.InitializeFolderStructure(); err != nil {
		t.Fatalf("Failed to initialize folder structure: %v", err)
	}

	tm := NewTemplateManager(cs)

	t.Run("GetSupportedExtensions returns correct extensions", func(t *testing.T) {
		extensions := tm.GetSupportedExtensions()
		expected := []string{".xlsx", ".xls", ".docx", ".pptx"}

		if len(extensions) != len(expected) {
			t.Errorf("Expected %d extensions, got %d", len(expected), len(extensions))
		}

		for i, ext := range expected {
			if i < len(extensions) && extensions[i] != ext {
				t.Errorf("Expected extension %s, got %s", ext, extensions[i])
			}
		}
	})

	t.Run("IsTemplateFile identifies valid templates", func(t *testing.T) {
		validFiles := []string{
			"template.xlsx",
			"TEMPLATE.XLSX",
			"my-template.xls",
			"document.docx",
			"presentation.pptx",
		}

		for _, file := range validFiles {
			if !tm.IsTemplateFile(file) {
				t.Errorf("Expected %s to be identified as template", file)
			}
		}

		invalidFiles := []string{
			"file.txt",
			"image.png",
			"data.csv",
			"script.py",
			"noextension",
		}

		for _, file := range invalidFiles {
			if tm.IsTemplateFile(file) {
				t.Errorf("Expected %s NOT to be identified as template", file)
			}
		}
	})

	t.Run("ListTemplates returns templates in folder", func(t *testing.T) {
		templatesPath := cs.GetTemplatesPath()

		// Create test templates
		createTestTemplate(templatesPath, "financial-model.xlsx")
		createTestTemplate(templatesPath, "deal-memo.docx")
		createTestTemplate(templatesPath, "pitch-deck.pptx")
		createTestTemplate(templatesPath, "not-a-template.txt") // Should be ignored

		templates, err := tm.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		if len(templates) != 3 {
			t.Errorf("Expected 3 templates, got %d", len(templates))
		}

		// Check template properties
		templateNames := make(map[string]bool)
		for _, tmpl := range templates {
			templateNames[tmpl.Name] = true

			// Verify type is set correctly
			ext := filepath.Ext(tmpl.Path)
			expectedType := ext[1:] // Remove leading dot
			if tmpl.Type != expectedType {
				t.Errorf("Expected type %s for %s, got %s", expectedType, tmpl.Name, tmpl.Type)
			}
		}

		expectedNames := []string{"financial-model", "deal-memo", "pitch-deck"}
		for _, name := range expectedNames {
			if !templateNames[name] {
				t.Errorf("Expected template %s not found", name)
			}
		}
	})

	t.Run("ValidateTemplatesFolder checks folder properly", func(t *testing.T) {
		// Should pass with existing folder
		err := tm.ValidateTemplatesFolder()
		if err != nil {
			t.Errorf("ValidateTemplatesFolder failed on valid folder: %v", err)
		}

		// Remove folder and check again
		os.RemoveAll(cs.GetTemplatesPath())
		err = tm.ValidateTemplatesFolder()
		if err == nil {
			t.Error("ValidateTemplatesFolder should fail when folder doesn't exist")
		}
	})

	t.Run("CopyTemplateToAnalysis copies template correctly", func(t *testing.T) {
		// Recreate folder structure and templates
		fm.InitializeFolderStructure()
		templatesPath := cs.GetTemplatesPath()
		templateFile := "analysis-template.xlsx"
		createTestTemplate(templatesPath, templateFile)

		// Create a deal
		dealName := "TestDeal"
		fm.CreateDealFolder(dealName)

		// Copy template
		templatePath := filepath.Join(templatesPath, templateFile)
		destPath, err := tm.CopyTemplateToAnalysis(templatePath, dealName)
		if err != nil {
			t.Fatalf("Failed to copy template: %v", err)
		}

		// Check if file exists at destination
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			t.Error("Template was not copied to analysis folder")
		}

		// Verify it's in the analysis folder
		expectedDir := filepath.Join(cs.GetDealsPath(), dealName, "analysis")
		if filepath.Dir(destPath) != expectedDir {
			t.Errorf("Template copied to wrong location: %s", destPath)
		}
	})

	t.Run("TemplateExists checks correctly", func(t *testing.T) {
		fm.InitializeFolderStructure()
		templatesPath := cs.GetTemplatesPath()
		createTestTemplate(templatesPath, "existing-template.xlsx")

		if !tm.TemplateExists("existing-template") {
			t.Error("TemplateExists should return true for existing template")
		}

		if !tm.TemplateExists("existing-template.xlsx") {
			t.Error("TemplateExists should return true for full filename")
		}

		if tm.TemplateExists("non-existent-template") {
			t.Error("TemplateExists should return false for non-existent template")
		}
	})
}
