package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FolderManager handles the creation and management of DealDone folder structure
type FolderManager struct {
	configService *ConfigService
}

// NewFolderManager creates a new folder manager
func NewFolderManager(configService *ConfigService) *FolderManager {
	return &FolderManager{
		configService: configService,
	}
}

// InitializeFolderStructure creates the DealDone folder structure
func (fm *FolderManager) InitializeFolderStructure() error {
	root := fm.configService.GetDealDoneRoot()

	// Create main DealDone folder
	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("failed to create DealDone root folder: %w", err)
	}

	// Create Templates folder
	templatesPath := fm.configService.GetTemplatesPath()
	if err := os.MkdirAll(templatesPath, 0755); err != nil {
		return fmt.Errorf("failed to create Templates folder: %w", err)
	}

	// Create Deals folder
	dealsPath := fm.configService.GetDealsPath()
	if err := os.MkdirAll(dealsPath, 0755); err != nil {
		return fmt.Errorf("failed to create Deals folder: %w", err)
	}

	return nil
}

// CreateDealFolder creates a new deal folder with subfolders
func (fm *FolderManager) CreateDealFolder(dealName string) (string, error) {
	if dealName == "" {
		return "", fmt.Errorf("deal name cannot be empty")
	}

	dealsPath := fm.configService.GetDealsPath()
	dealPath := filepath.Join(dealsPath, dealName)

	// Create deal folder
	if err := os.MkdirAll(dealPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create deal folder: %w", err)
	}

	// Create subfolders
	subfolders := []string{"legal", "financial", "general", "analysis"}
	for _, folder := range subfolders {
		subfolder := filepath.Join(dealPath, folder)
		if err := os.MkdirAll(subfolder, 0755); err != nil {
			return "", fmt.Errorf("failed to create %s folder: %w", folder, err)
		}
	}

	return dealPath, nil
}

// ValidateFolderStructure checks if the folder structure exists and is valid
func (fm *FolderManager) ValidateFolderStructure() error {
	root := fm.configService.GetDealDoneRoot()

	// Check if root exists
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return fmt.Errorf("DealDone root folder does not exist: %s", root)
	}

	// Check if Templates folder exists
	templatesPath := fm.configService.GetTemplatesPath()
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		return fmt.Errorf("Templates folder does not exist: %s", templatesPath)
	}

	// Check if Deals folder exists
	dealsPath := fm.configService.GetDealsPath()
	if _, err := os.Stat(dealsPath); os.IsNotExist(err) {
		return fmt.Errorf("Deals folder does not exist: %s", dealsPath)
	}

	return nil
}

// ListDeals returns a list of all deal folders
func (fm *FolderManager) ListDeals() ([]string, error) {
	dealsPath := fm.configService.GetDealsPath()

	entries, err := os.ReadDir(dealsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read deals directory: %w", err)
	}

	var deals []string
	for _, entry := range entries {
		if entry.IsDir() {
			deals = append(deals, entry.Name())
		}
	}

	return deals, nil
}

// DealExists checks if a deal folder exists
func (fm *FolderManager) DealExists(dealName string) bool {
	dealsPath := fm.configService.GetDealsPath()
	dealPath := filepath.Join(dealsPath, dealName)

	_, err := os.Stat(dealPath)
	return err == nil
}

// GetDealPath returns the full path to a deal folder
func (fm *FolderManager) GetDealPath(dealName string) string {
	return filepath.Join(fm.configService.GetDealsPath(), dealName)
}

// GetDealSubfolderPath returns the path to a specific subfolder in a deal
func (fm *FolderManager) GetDealSubfolderPath(dealName, subfolder string) string {
	return filepath.Join(fm.GetDealPath(dealName), subfolder)
}

// EnsureFolderExists creates a folder if it doesn't exist
func (fm *FolderManager) EnsureFolderExists(path string) error {
	return os.MkdirAll(path, 0755)
}

// DealInfo represents information about a deal
type DealInfo struct {
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	CreatedAt     time.Time `json:"createdAt"`
	DocumentCount int       `json:"documentCount"`
}

// IsDealDoneReady checks if the DealDone folder structure is ready
func (fm *FolderManager) IsDealDoneReady() bool {
	root := fm.configService.GetDealDoneRoot()

	// Check if root exists
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return false
	}

	// Check if Templates and Deals folders exist
	templatesPath := fm.configService.GetTemplatesPath()
	dealsPath := fm.configService.GetDealsPath()

	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		return false
	}

	if _, err := os.Stat(dealsPath); os.IsNotExist(err) {
		return false
	}

	return true
}

// GetAllDeals returns information about all deals
func (fm *FolderManager) GetAllDeals() ([]DealInfo, error) {
	dealsPath := fm.configService.GetDealsPath()

	entries, err := os.ReadDir(dealsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []DealInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read deals directory: %w", err)
	}

	deals := make([]DealInfo, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dealPath := filepath.Join(dealsPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Count documents in deal
		docCount := fm.countDocumentsInDeal(dealPath)

		deals = append(deals, DealInfo{
			Name:          entry.Name(),
			Path:          dealPath,
			CreatedAt:     info.ModTime(),
			DocumentCount: docCount,
		})
	}

	return deals, nil
}

// countDocumentsInDeal counts documents in all subfolders of a deal
func (fm *FolderManager) countDocumentsInDeal(dealPath string) int {
	count := 0

	subfolders := []string{"legal", "financial", "general", "analysis"}
	for _, subfolder := range subfolders {
		subPath := filepath.Join(dealPath, subfolder)
		entries, err := os.ReadDir(subPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				count++
			}
		}
	}

	return count
}
