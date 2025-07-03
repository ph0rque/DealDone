package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Template represents a template file
type Template struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"` // xlsx, xls, docx, pptx, csv, txt, md, pdf
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

// TemplateManager handles template file operations
type TemplateManager struct {
	configService *ConfigService
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(configService *ConfigService) *TemplateManager {
	return &TemplateManager{
		configService: configService,
	}
}

// GetSupportedExtensions returns the list of supported template file extensions
func (tm *TemplateManager) GetSupportedExtensions() []string {
	return []string{".xlsx", ".xls", ".docx", ".pptx", ".csv", ".txt", ".md", ".pdf"}
}

// IsTemplateFile checks if a file is a valid template based on extension
func (tm *TemplateManager) IsTemplateFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, supported := range tm.GetSupportedExtensions() {
		if ext == supported {
			return true
		}
	}
	return false
}

// ListTemplates returns all templates in the Templates folder
func (tm *TemplateManager) ListTemplates() ([]Template, error) {
	templatesPath := tm.configService.GetTemplatesPath()

	// Check if templates folder exists
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		return []Template{}, nil
	}

	var templates []Template

	err := filepath.Walk(templatesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a template file
		if !tm.IsTemplateFile(info.Name()) {
			return nil
		}

		// Note: We could get relative path here if needed in the future
		// relPath, err := filepath.Rel(templatesPath, path)

		// Determine type from extension
		ext := strings.ToLower(filepath.Ext(info.Name()))
		fileType := strings.TrimPrefix(ext, ".")

		template := Template{
			Name:     strings.TrimSuffix(info.Name(), ext),
			Path:     path,
			Type:     fileType,
			Size:     info.Size(),
			Modified: info.ModTime().Format("2006-01-02 15:04:05"),
		}

		templates = append(templates, template)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	return templates, nil
}

// ValidateTemplatesFolder checks if the templates folder exists and is accessible
func (tm *TemplateManager) ValidateTemplatesFolder() error {
	templatesPath := tm.configService.GetTemplatesPath()

	// Check if folder exists
	info, err := os.Stat(templatesPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("templates folder does not exist: %s", templatesPath)
	}
	if err != nil {
		return fmt.Errorf("failed to access templates folder: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("templates path is not a directory: %s", templatesPath)
	}

	// Check read permissions
	testFile := filepath.Join(templatesPath, ".test_read")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("templates folder is not writable: %w", err)
	}
	os.Remove(testFile)

	return nil
}

// CopyTemplateToAnalysis copies a template to a deal's analysis folder
func (tm *TemplateManager) CopyTemplateToAnalysis(templatePath, dealName string) (string, error) {
	// Validate template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template does not exist: %s", templatePath)
	}

	// Get destination path
	dealsPath := tm.configService.GetDealsPath()
	analysisPath := filepath.Join(dealsPath, dealName, "analysis")

	// Ensure analysis folder exists
	if err := os.MkdirAll(analysisPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create analysis folder: %w", err)
	}

	// Generate appropriate filename for analysis
	originalName := filepath.Base(templatePath)
	ext := filepath.Ext(originalName)
	baseName := strings.TrimSuffix(originalName, ext)

	// Convert template name to analysis filename
	analysisName := tm.generateAnalysisFilename(baseName, ext)
	destPath := filepath.Join(analysisPath, analysisName)

	// Check if file already exists - if so, skip copying to avoid duplicates
	if _, err := os.Stat(destPath); err == nil {
		// File already exists, return the existing path
		return destPath, nil
	}

	// Copy file
	source, err := os.Open(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to open template: %w", err)
	}
	defer source.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dest.Close()

	if _, err := dest.ReadFrom(source); err != nil {
		return "", fmt.Errorf("failed to copy template: %w", err)
	}

	return destPath, nil
}

// generateAnalysisFilename converts template names to appropriate analysis filenames
func (tm *TemplateManager) generateAnalysisFilename(baseName, ext string) string {
	fmt.Printf("DEBUG generateAnalysisFilename: baseName='%s', ext='%s'\n", baseName, ext)

	// Convert to lowercase
	name := strings.ToLower(baseName)
	fmt.Printf("DEBUG after lowercase: '%s'\n", name)

	// Remove "template" from the name
	name = strings.ReplaceAll(name, "_template", "")
	name = strings.ReplaceAll(name, "-template", "")
	name = strings.ReplaceAll(name, "template_", "")
	name = strings.ReplaceAll(name, "template-", "")
	name = strings.ReplaceAll(name, "template", "")
	fmt.Printf("DEBUG after removing template: '%s'\n", name)

	// Replace spaces and special characters with underscores
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	fmt.Printf("DEBUG after replacing chars: '%s'\n", name)

	// Remove duplicate underscores
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	fmt.Printf("DEBUG after removing duplicate underscores: '%s'\n", name)

	// Trim leading/trailing underscores
	name = strings.Trim(name, "_")
	fmt.Printf("DEBUG after trimming underscores: '%s'\n", name)

	// Handle specific template names
	switch name {
	case "deal_summary":
		fmt.Printf("DEBUG matched deal_summary case\n")
		return "deal_summary" + ext
	case "financial_model":
		fmt.Printf("DEBUG matched financial_model case\n")
		return "financial_model" + ext
	case "due_diligence_checklist":
		fmt.Printf("DEBUG matched due_diligence_checklist case\n")
		return "due_diligence_checklist" + ext
	default:
		fmt.Printf("DEBUG using default case: '%s'\n", name)
		return name + ext
	}
}

// GetTemplateCount returns the number of templates available
func (tm *TemplateManager) GetTemplateCount() (int, error) {
	templates, err := tm.ListTemplates()
	if err != nil {
		return 0, err
	}
	return len(templates), nil
}

// TemplateExists checks if a specific template exists
func (tm *TemplateManager) TemplateExists(templateName string) bool {
	templates, err := tm.ListTemplates()
	if err != nil {
		return false
	}

	for _, template := range templates {
		if template.Name == templateName || filepath.Base(template.Path) == templateName {
			return true
		}
	}

	return false
}

// GenerateDefaultTemplates creates the default template files
func (tm *TemplateManager) GenerateDefaultTemplates() error {
	templatesPath := tm.configService.GetTemplatesPath()

	// Ensure templates directory exists
	if err := os.MkdirAll(templatesPath, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Create default templates generator
	generator := NewDefaultTemplateGenerator(templatesPath)

	// Generate the default templates
	return generator.GenerateDefaultTemplates()
}
