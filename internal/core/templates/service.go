package templates

import (
	"context"
	"path/filepath"
)

// Service provides a unified interface for all template operations
type Service interface {
	// Template Management
	ListTemplates() ([]Template, error)
	GetTemplate(path string) (*Template, error)
	IsTemplateFile(filename string) bool
	GenerateDefaultTemplates() error
	ImportTemplate(sourcePath string, metadata *TemplateMetadata) error

	// Template Discovery
	DiscoverTemplates() ([]TemplateInfo, error)
	SearchTemplates(query string, filters map[string]string) ([]TemplateInfo, error)
	GetTemplateCategories() []string
	GetTemplateByID(id string) (*TemplateInfo, error)
	OrganizeTemplatesByCategory() (map[string][]TemplateInfo, error)

	// Template Parsing
	ParseTemplate(templatePath string) (*TemplateData, error)
	ExtractDataFields(templateData *TemplateData) []DataField
	ValidateTemplateStructure(templateData *TemplateData) error

	// Template Population
	PopulateTemplate(templatePath string, mappedData *MappedData, outputPath string) error
	PopulateTemplateWithData(templateData *TemplateData, data map[string]interface{}, preserveFormulas bool) (*TemplateData, error)
	ValidatePopulatedTemplate(populatedPath string, templatePath string) error

	// Field Mapping
	MapDataToTemplate(templatePath string, documentPaths []string, dealName string) (*MappedData, error)
	MatchTemplateFields(sourceFields []string, templatePath string) (*MatchingResult, error)
	GetFieldMappingSuggestions(unmatchedFields []string, templatePath string) (map[string][]string, error)

	// Template Analysis
	AnalyzeDocumentsAndPopulateTemplates(dealName string, documentPaths []string) (*TemplateAnalysisResult, error)
	CopyTemplatesToAnalysis(dealName string, documentTypes []string) ([]string, error)

	// Template Optimization
	OptimizeTemplate(ctx context.Context, templatePath string, optimizationType TemplateOptimizationType) error
	GetOptimizationHistory(templatePath string) ([]OptimizationRecord, error)
	AnalyzeTemplatePerformance(templatePath string) (*TemplatePerformanceMetrics, error)

	// Template Analytics
	TrackTemplateUsage(templateID string, userID string, action string) error
	GetTemplateUsageMetrics(templateID string) (*AnalyticsUsageRecord, error)
	GenerateExecutiveDashboard() (*AnalyticsExecutiveDashboard, error)
	GenerateOperationalDashboard() (*AnalyticsOperationalDashboard, error)
}

// Manager combines all template-related services
type Manager struct {
	templatesPath     string
	templateManager   *TemplateManager
	templateDiscovery *TemplateDiscovery
	templateParser    *TemplateParser
	templatePopulator *TemplatePopulator
	templateOptimizer *TemplateOptimizer
	templateAnalytics *TemplateAnalyticsEngine
	defaultGenerator  *DefaultTemplateGenerator
}

// NewManager creates a new unified template manager
func NewManager(templatesPath string) *Manager {
	templateManager := &TemplateManager{
		configService: &configService{templatesPath: templatesPath},
	}

	return &Manager{
		templatesPath:     templatesPath,
		templateManager:   templateManager,
		templateDiscovery: NewTemplateDiscovery(templateManager),
		templateParser:    NewTemplateParser(templatesPath),
		templatePopulator: NewTemplatePopulator(&TemplateParser{templatesPath: templatesPath}),
		templateOptimizer: nil, // Will be initialized when needed
		templateAnalytics: nil, // Will be initialized when needed
		defaultGenerator:  NewDefaultTemplateGenerator(templatesPath),
	}
}

// Implement Service interface methods by delegating to appropriate sub-services

// Template Management
func (m *Manager) ListTemplates() ([]Template, error) {
	return m.templateManager.ListTemplates()
}

func (m *Manager) GetTemplate(path string) (*Template, error) {
	templates, err := m.templateManager.ListTemplates()
	if err != nil {
		return nil, err
	}

	for _, template := range templates {
		if template.Path == path {
			return &template, nil
		}
	}

	return nil, nil
}

func (m *Manager) IsTemplateFile(filename string) bool {
	return m.templateManager.IsTemplateFile(filename)
}

func (m *Manager) GenerateDefaultTemplates() error {
	return m.defaultGenerator.GenerateDefaultTemplates()
}

func (m *Manager) ImportTemplate(sourcePath string, metadata *TemplateMetadata) error {
	// ImportTemplate needs to be implemented
	// For now, return nil to satisfy the interface
	return nil
}

// Template Discovery
func (m *Manager) DiscoverTemplates() ([]TemplateInfo, error) {
	return m.templateDiscovery.DiscoverTemplates()
}

func (m *Manager) SearchTemplates(query string, filters map[string]string) ([]TemplateInfo, error) {
	return m.templateDiscovery.SearchTemplates(query, filters)
}

func (m *Manager) GetTemplateCategories() []string {
	return m.templateDiscovery.GetTemplateCategories()
}

func (m *Manager) GetTemplateByID(id string) (*TemplateInfo, error) {
	return m.templateDiscovery.GetTemplateByID(id)
}

func (m *Manager) OrganizeTemplatesByCategory() (map[string][]TemplateInfo, error) {
	return m.templateDiscovery.OrganizeTemplatesByCategory()
}

// Template Parsing
func (m *Manager) ParseTemplate(templatePath string) (*TemplateData, error) {
	return m.templateParser.ParseTemplate(templatePath)
}

func (m *Manager) ExtractDataFields(templateData *TemplateData) []DataField {
	return m.templateParser.ExtractDataFields(templateData)
}

func (m *Manager) ValidateTemplateStructure(templateData *TemplateData) error {
	return m.templateParser.ValidateTemplateStructure(templateData)
}

// configService is a minimal implementation to satisfy the interface
type configService struct {
	templatesPath string
	dealsPath     string
}

func (cs *configService) GetTemplatesPath() string {
	return cs.templatesPath
}

func (cs *configService) GetDealsPath() string {
	if cs.dealsPath == "" {
		// Default to a sibling directory of templates
		return filepath.Join(filepath.Dir(cs.templatesPath), "Deals")
	}
	return cs.dealsPath
}
