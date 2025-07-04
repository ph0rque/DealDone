package deals

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Service provides a unified interface for deal and folder operations
type Service interface {
	// Deal Management
	CreateDeal(ctx context.Context, name string, metadata *DealMetadata) (*Deal, error)
	GetDeal(ctx context.Context, dealID string) (*Deal, error)
	GetDealByName(ctx context.Context, name string) (*Deal, error)
	ListDeals(ctx context.Context, criteria *DealSearchCriteria) ([]*Deal, error)
	UpdateDeal(ctx context.Context, dealID string, updates *DealMetadata) (*Deal, error)
	DeleteDeal(ctx context.Context, dealID string) error

	// Folder Operations
	CreateDealFolder(ctx context.Context, dealName string) (string, error)
	CreateSubfolder(ctx context.Context, dealName string, folderType FolderType) (string, error)
	GetFolderStructure(ctx context.Context, dealName string) (*FolderNode, error)
	DeleteFolder(ctx context.Context, path string) error
	RenameFolder(ctx context.Context, oldPath string, newName string) error

	// Deal Statistics
	GetDealStatistics(ctx context.Context, dealName string) (*DealStatistics, error)
	GetDealActivity(ctx context.Context, dealID string, limit int) ([]*DealActivity, error)
	RecordActivity(ctx context.Context, activity *DealActivity) error

	// Folder Templates
	ApplyFolderTemplate(ctx context.Context, dealName string, templateName string) error
	GetAvailableTemplates(ctx context.Context) ([]*FolderTemplate, error)
	CreateFolderTemplate(ctx context.Context, template *FolderTemplate) error

	// Utility Functions
	ValidateDealName(name string) error
	GetDealPath(dealName string) string
	GetSubfolderPath(dealName string, folderType FolderType) string
	IsValidFolderType(folderType string) bool
}

// Manager implements the Service interface
type Manager struct {
	folderManager *FolderManager
	dealsPath     string
}

// NewManager creates a new deals manager
func NewManager(folderManager *FolderManager, dealsPath string) *Manager {
	return &Manager{
		folderManager: folderManager,
		dealsPath:     dealsPath,
	}
}

// Deal Management
func (m *Manager) CreateDeal(ctx context.Context, name string, metadata *DealMetadata) (*Deal, error) {
	// Create the deal folder structure
	path, err := m.folderManager.CreateDealFolder(name)
	if err != nil {
		return nil, err
	}

	// Create and return the deal object
	deal := &Deal{
		ID:        generateDealID(name),
		Name:      name,
		RootPath:  path,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    DealStatusActive,
		Metadata:  *metadata,
	}

	return deal, nil
}

func (m *Manager) GetDeal(ctx context.Context, dealID string) (*Deal, error) {
	// Implementation would retrieve deal from storage
	return nil, nil
}

func (m *Manager) GetDealByName(ctx context.Context, name string) (*Deal, error) {
	// Implementation would retrieve deal by name
	return nil, nil
}

func (m *Manager) ListDeals(ctx context.Context, criteria *DealSearchCriteria) ([]*Deal, error) {
	// Implementation would list deals based on criteria
	return nil, nil
}

func (m *Manager) UpdateDeal(ctx context.Context, dealID string, updates *DealMetadata) (*Deal, error) {
	// Implementation would update deal metadata
	return nil, nil
}

func (m *Manager) DeleteDeal(ctx context.Context, dealID string) error {
	// Implementation would delete deal and its folder structure
	return nil
}

// Folder Operations
func (m *Manager) CreateDealFolder(ctx context.Context, dealName string) (string, error) {
	return m.folderManager.CreateDealFolder(dealName)
}

func (m *Manager) CreateSubfolder(ctx context.Context, dealName string, folderType FolderType) (string, error) {
	// Map FolderType to string for the existing API
	var folderName string
	switch folderType {
	case FolderTypeLegal:
		folderName = "Legal"
	case FolderTypeFinancial:
		folderName = "Financial"
	case FolderTypeGeneral:
		folderName = "General"
	case FolderTypeAnalysis:
		folderName = "Analysis"
	default:
		folderName = string(folderType)
	}

	// Get the subfolder path and ensure it exists
	path := m.folderManager.GetDealSubfolderPath(dealName, folderName)
	err := m.folderManager.EnsureFolderExists(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (m *Manager) GetFolderStructure(ctx context.Context, dealName string) (*FolderNode, error) {
	// Implementation would build the folder tree
	return nil, nil
}

func (m *Manager) DeleteFolder(ctx context.Context, path string) error {
	// Implementation would delete folder
	return nil
}

func (m *Manager) RenameFolder(ctx context.Context, oldPath string, newName string) error {
	// Implementation would rename folder
	return nil
}

// Deal Statistics
func (m *Manager) GetDealStatistics(ctx context.Context, dealName string) (*DealStatistics, error) {
	// Implementation would calculate deal statistics
	return nil, nil
}

func (m *Manager) GetDealActivity(ctx context.Context, dealID string, limit int) ([]*DealActivity, error) {
	// Implementation would retrieve activity log
	return nil, nil
}

func (m *Manager) RecordActivity(ctx context.Context, activity *DealActivity) error {
	// Implementation would record activity
	return nil
}

// Folder Templates
func (m *Manager) ApplyFolderTemplate(ctx context.Context, dealName string, templateName string) error {
	// Implementation would apply template
	return nil
}

func (m *Manager) GetAvailableTemplates(ctx context.Context) ([]*FolderTemplate, error) {
	// Return default templates
	templates := []*FolderTemplate{
		{
			Name:        "Standard Deal",
			Description: "Standard deal folder structure",
			Structure: map[string]FolderTemplate{
				"Legal":     {},
				"Financial": {},
				"General":   {},
				"Analysis":  {},
			},
		},
	}
	return templates, nil
}

func (m *Manager) CreateFolderTemplate(ctx context.Context, template *FolderTemplate) error {
	// Implementation would save template
	return nil
}

// Utility Functions
func (m *Manager) ValidateDealName(name string) error {
	// Simple validation - in production would be more comprehensive
	if name == "" {
		return fmt.Errorf("deal name cannot be empty")
	}
	if strings.ContainsAny(name, "/\\:*?\"<>|") {
		return fmt.Errorf("deal name contains invalid characters")
	}
	return nil
}

func (m *Manager) GetDealPath(dealName string) string {
	return m.folderManager.GetDealPath(dealName)
}

func (m *Manager) GetSubfolderPath(dealName string, folderType FolderType) string {
	// Map FolderType to string
	var folderName string
	switch folderType {
	case FolderTypeLegal:
		folderName = "Legal"
	case FolderTypeFinancial:
		folderName = "Financial"
	case FolderTypeGeneral:
		folderName = "General"
	case FolderTypeAnalysis:
		folderName = "Analysis"
	default:
		folderName = string(folderType)
	}

	return m.folderManager.GetDealSubfolderPath(dealName, folderName)
}

func (m *Manager) IsValidFolderType(folderType string) bool {
	validTypes := []string{"Legal", "Financial", "General", "Analysis"}
	for _, valid := range validTypes {
		if folderType == valid {
			return true
		}
	}
	return false
}

// Helper function to generate deal ID
func generateDealID(name string) string {
	// Simple implementation - in production would use UUID or similar
	return fmt.Sprintf("deal_%s_%d", strings.ToLower(strings.ReplaceAll(name, " ", "_")), time.Now().Unix())
}
