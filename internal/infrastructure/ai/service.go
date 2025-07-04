package ai

import (
	"context"
	"time"
)

// Service provides a unified interface for AI operations
type Service interface {
	// Provider Management
	SetProvider(provider Provider) error
	GetCurrentProvider() Provider
	GetAvailableProviders() []Provider
	GetProviderHealth(provider Provider) (*ProviderHealth, error)

	// Document Classification
	ClassifyDocument(ctx context.Context, content string, filePath string) (*AIClassificationResult, error)

	// Financial Analysis
	AnalyzeFinancialData(ctx context.Context, content string) (*FinancialAnalysis, error)
	ExtractFinancialMetrics(ctx context.Context, content string) (*FinancialMetricsExtraction, error)

	// Risk Analysis
	AnalyzeRisks(ctx context.Context, content string) (*RiskAnalysis, error)

	// Document Insights
	GenerateInsights(ctx context.Context, content string) (*DocumentInsights, error)

	// Entity Extraction
	ExtractEntities(ctx context.Context, content string) (*EntityExtraction, error)
	ExtractCompanyAndDealInfo(ctx context.Context, content string) (*CompanyDealExtraction, error)
	ExtractPersonnelAndRoles(ctx context.Context, content string) (*PersonnelRoleExtraction, error)
	ExtractDocumentEntities(ctx context.Context, content string, docType string) (*DocumentEntityExtraction, error)

	// Field Extraction and Mapping
	ExtractDocumentFields(ctx context.Context, content string, templateFields []string) (*DocumentFieldExtraction, error)
	MapFieldsToTemplate(ctx context.Context, sourceFields []string, templateFields []string) (*FieldMappingResult, error)
	SuggestFieldMappings(ctx context.Context, unmatchedFields []string, templateFields []string) (map[string][]MappingSuggestion, error)
	AnalyzeFieldSemantics(ctx context.Context, sourceFields []string, content string) (*FieldSemanticAnalysis, error)
	MapFieldsWithSemantics(ctx context.Context, sourceAnalysis *FieldSemanticAnalysis, templateStructure *TemplateStructureAnalysis) (*SemanticMappingResult, error)

	// Value Formatting
	FormatFieldValue(ctx context.Context, fieldName string, value string, targetFormat string) (*FormattedFieldValue, error)

	// Cross-Document Validation
	ValidateAcrossDocuments(ctx context.Context, entities map[string]*DocumentEntityExtraction) (*CrossDocumentValidation, error)

	// Template Analysis
	AnalyzeTemplateStructure(ctx context.Context, templatePath string, templateContent string) (*TemplateStructureAnalysis, error)

	// Conflict Resolution
	ResolveFieldConflicts(ctx context.Context, conflicts []FieldConflict, context ConflictResolutionContext) (*ConflictResolutionResult, error)

	// Validation
	ValidateFieldMapping(ctx context.Context, mapping *SemanticMappingResult, businessRules []BusinessRule) (*MappingValidationResult, error)

	// Usage and Statistics
	GetUsageStats() (*AIUsageStats, error)
	GetUsageMetrics(provider Provider, period time.Duration) (*UsageMetrics, error)
	ResetUsageStats() error

	// Cache Management
	ClearCache() error
	GetCacheStats() (map[string]interface{}, error)

	// Configuration
	UpdateConfig(config *AIConfig) error
	GetConfig() *AIConfig
}
