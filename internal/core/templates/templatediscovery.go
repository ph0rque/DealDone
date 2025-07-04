package templates

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TemplateMetadata contains detailed information about a template
type TemplateMetadata struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"` // financial, legal, general
	Type        string                 `json:"type"`     // xlsx, docx, etc.
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Tags        []string               `json:"tags"`
	Fields      []TemplateField        `json:"fields,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// TemplateField represents a field that can be populated in a template
type TemplateField struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // text, number, date, currency
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
	Format      string `json:"format,omitempty"` // for dates, numbers
	Source      string `json:"source,omitempty"` // suggested data source
}

// TemplateInfo combines file info with metadata
type TemplateInfo struct {
	Template
	Metadata    *TemplateMetadata `json:"metadata,omitempty"`
	HasMetadata bool              `json:"has_metadata"`
}

// TemplateDiscovery handles advanced template discovery and management
type TemplateDiscovery struct {
	templateManager *TemplateManager
	metadataCache   map[string]*TemplateMetadata
}

// NewTemplateDiscovery creates a new template discovery service
func NewTemplateDiscovery(templateManager *TemplateManager) *TemplateDiscovery {
	return &TemplateDiscovery{
		templateManager: templateManager,
		metadataCache:   make(map[string]*TemplateMetadata),
	}
}

// DiscoverTemplates performs a comprehensive template discovery
func (td *TemplateDiscovery) DiscoverTemplates() ([]TemplateInfo, error) {
	// Get basic template list
	templates, err := td.templateManager.ListTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	var templateInfos []TemplateInfo

	for _, template := range templates {
		info := TemplateInfo{
			Template:    template,
			HasMetadata: false,
		}

		// Try to load metadata
		metadata, err := td.loadTemplateMetadata(template.Path)
		if err == nil && metadata != nil {
			info.Metadata = metadata
			info.HasMetadata = true
			td.metadataCache[template.Path] = metadata
		} else {
			// Generate basic metadata from file info
			info.Metadata = td.generateBasicMetadata(&template)
		}

		templateInfos = append(templateInfos, info)
	}

	return templateInfos, nil
}

// loadTemplateMetadata loads metadata from a .meta.json file
func (td *TemplateDiscovery) loadTemplateMetadata(templatePath string) (*TemplateMetadata, error) {
	// Check cache first
	if cached, exists := td.metadataCache[templatePath]; exists {
		return cached, nil
	}

	// Look for metadata file
	metadataPath := strings.TrimSuffix(templatePath, filepath.Ext(templatePath)) + ".meta.json"

	file, err := os.Open(metadataPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metadata TemplateMetadata
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &metadata, nil
}

// generateBasicMetadata creates basic metadata from file information
func (td *TemplateDiscovery) generateBasicMetadata(template *Template) *TemplateMetadata {
	// Determine category from path or name
	category := td.detectCategory(template.Name, template.Path)

	// Generate a simple ID
	id := strings.ToLower(strings.ReplaceAll(template.Name, " ", "-"))

	// Parse modified time
	modTime, _ := time.Parse("2006-01-02 15:04:05", template.Modified)

	return &TemplateMetadata{
		ID:          id,
		Name:        template.Name,
		Description: fmt.Sprintf("%s template", strings.Title(template.Type)),
		Category:    category,
		Type:        template.Type,
		Version:     "1.0",
		Author:      "System",
		CreatedAt:   modTime,
		UpdatedAt:   modTime,
		Tags:        td.generateTags(template.Name, category),
		Properties:  make(map[string]interface{}),
	}
}

// detectCategory attempts to determine template category
func (td *TemplateDiscovery) detectCategory(name, path string) string {
	lowerName := strings.ToLower(name)
	lowerPath := strings.ToLower(path)

	// Check for financial indicators
	financialKeywords := []string{"financial", "finance", "model", "valuation",
		"budget", "forecast", "revenue", "profit", "cashflow", "p&l"}
	for _, keyword := range financialKeywords {
		if strings.Contains(lowerName, keyword) || strings.Contains(lowerPath, keyword) {
			return "financial"
		}
	}

	// Check for legal indicators
	legalKeywords := []string{"legal", "contract", "agreement", "nda", "loi",
		"term", "compliance", "regulatory", "diligence"}
	for _, keyword := range legalKeywords {
		if strings.Contains(lowerName, keyword) || strings.Contains(lowerPath, keyword) {
			return "legal"
		}
	}

	return "general"
}

// generateTags creates relevant tags based on template name and category
func (td *TemplateDiscovery) generateTags(name, category string) []string {
	tags := []string{category}

	// Add type-specific tags
	words := strings.Fields(strings.ToLower(name))
	for _, word := range words {
		if len(word) > 3 && word != "template" {
			tags = append(tags, word)
		}
	}

	return tags
}

// SearchTemplates searches templates by various criteria
func (td *TemplateDiscovery) SearchTemplates(query string, filters map[string]string) ([]TemplateInfo, error) {
	allTemplates, err := td.DiscoverTemplates()
	if err != nil {
		return nil, err
	}

	var results []TemplateInfo
	queryLower := strings.ToLower(query)

	for _, template := range allTemplates {
		// Check if template matches query
		if query != "" {
			nameMatch := strings.Contains(strings.ToLower(template.Name), queryLower)
			descMatch := template.Metadata != nil &&
				strings.Contains(strings.ToLower(template.Metadata.Description), queryLower)
			tagMatch := false

			if template.Metadata != nil {
				for _, tag := range template.Metadata.Tags {
					if strings.Contains(strings.ToLower(tag), queryLower) {
						tagMatch = true
						break
					}
				}
			}

			if !nameMatch && !descMatch && !tagMatch {
				continue
			}
		}

		// Apply filters
		if category, ok := filters["category"]; ok && template.Metadata != nil {
			if template.Metadata.Category != category {
				continue
			}
		}

		if fileType, ok := filters["type"]; ok {
			if template.Type != fileType {
				continue
			}
		}

		results = append(results, template)
	}

	return results, nil
}

// GetTemplateByID retrieves a template by its ID
func (td *TemplateDiscovery) GetTemplateByID(id string) (*TemplateInfo, error) {
	templates, err := td.DiscoverTemplates()
	if err != nil {
		return nil, err
	}

	for _, template := range templates {
		if template.Metadata != nil && template.Metadata.ID == id {
			return &template, nil
		}
	}

	return nil, fmt.Errorf("template not found: %s", id)
}

// GetTemplateCategories returns all available template categories
func (td *TemplateDiscovery) GetTemplateCategories() []string {
	return []string{"financial", "legal", "general"}
}

// SaveTemplateMetadata saves metadata for a template
func (td *TemplateDiscovery) SaveTemplateMetadata(templatePath string, metadata *TemplateMetadata) error {
	metadataPath := strings.TrimSuffix(templatePath, filepath.Ext(templatePath)) + ".meta.json"

	file, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	// Update cache
	td.metadataCache[templatePath] = metadata

	return nil
}

// OrganizeTemplatesByCategory organizes templates into category-based structure
func (td *TemplateDiscovery) OrganizeTemplatesByCategory() (map[string][]TemplateInfo, error) {
	templates, err := td.DiscoverTemplates()
	if err != nil {
		return nil, err
	}

	organized := make(map[string][]TemplateInfo)

	for _, template := range templates {
		category := "general"
		if template.Metadata != nil {
			category = template.Metadata.Category
		}

		organized[category] = append(organized[category], template)
	}

	return organized, nil
}

// ImportTemplate imports a new template with metadata
func (td *TemplateDiscovery) ImportTemplate(sourcePath string, metadata *TemplateMetadata) error {
	// Validate source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", sourcePath)
	}

	// Check if it's a supported template type
	if !td.templateManager.IsTemplateFile(sourcePath) {
		return fmt.Errorf("unsupported template type: %s", filepath.Ext(sourcePath))
	}

	// Get destination path
	templatesPath := td.templateManager.configService.GetTemplatesPath()
	destName := filepath.Base(sourcePath)

	// Organize by category if metadata specifies one
	if metadata != nil && metadata.Category != "" {
		categoryPath := filepath.Join(templatesPath, metadata.Category)
		if err := os.MkdirAll(categoryPath, 0755); err != nil {
			return fmt.Errorf("failed to create category directory: %w", err)
		}
		templatesPath = categoryPath
	}

	destPath := filepath.Join(templatesPath, destName)

	// Check if file already exists
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("template already exists: %s", destName)
	}

	// Copy file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Save metadata if provided
	if metadata != nil {
		metadata.CreatedAt = time.Now()
		metadata.UpdatedAt = time.Now()

		if err := td.SaveTemplateMetadata(destPath, metadata); err != nil {
			// Don't fail the import, just log the error
			fmt.Printf("Warning: failed to save metadata: %v\n", err)
		}
	}

	return nil
}
