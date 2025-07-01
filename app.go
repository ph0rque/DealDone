package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// App struct
type App struct {
	ctx             context.Context
	configService   *ConfigService
	folderManager   *FolderManager
	templateManager *TemplateManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize configuration service
	configService, err := NewConfigService()
	if err != nil {
		// Log error but continue - will handle in UI
		println("Error initializing config service:", err.Error())
		return
	}
	a.configService = configService

	// Initialize folder manager
	a.folderManager = NewFolderManager(configService)

	// Initialize template manager
	a.templateManager = NewTemplateManager(configService)
}

// GetHomeDirectory returns the user's home directory
func (a *App) GetHomeDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return homeDir
}

// GetCurrentDirectory returns the current working directory
func (a *App) GetCurrentDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return a.GetHomeDirectory()
	}
	return wd
}

// GetDesktopDirectory returns the desktop directory path
func (a *App) GetDesktopDirectory() string {
	homeDir := a.GetHomeDirectory()
	return filepath.Join(homeDir, "Desktop")
}

// GetDocumentsDirectory returns the documents directory path
func (a *App) GetDocumentsDirectory() string {
	homeDir := a.GetHomeDirectory()
	return filepath.Join(homeDir, "Documents")
}

// GetDownloadsDirectory returns the downloads directory path
func (a *App) GetDownloadsDirectory() string {
	homeDir := a.GetHomeDirectory()
	return filepath.Join(homeDir, "Downloads")
}

// IsFirstRun checks if this is the first time the app is running
func (a *App) IsFirstRun() bool {
	if a.configService == nil {
		return true
	}
	return a.configService.IsFirstRun()
}

// GetDealDoneRoot returns the configured DealDone root folder path
func (a *App) GetDealDoneRoot() string {
	if a.configService == nil {
		return ""
	}
	return a.configService.GetDealDoneRoot()
}

// SetDealDoneRoot sets the DealDone root folder path and initializes the folder structure
func (a *App) SetDealDoneRoot(path string) error {
	if a.configService == nil {
		return fmt.Errorf("configuration service not initialized")
	}

	// Set the path in config
	if err := a.configService.SetDealDoneRoot(path); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Initialize folder structure
	if err := a.folderManager.InitializeFolderStructure(); err != nil {
		return fmt.Errorf("failed to create folder structure: %w", err)
	}

	// Generate default templates if none exist
	templatesPath := a.configService.GetTemplatesPath()
	dtg := NewDefaultTemplateGenerator(templatesPath)
	if !dtg.HasDefaultTemplates() {
		if err := dtg.GenerateDefaultTemplates(); err != nil {
			// Log error but don't fail - templates are optional
			println("Warning: Failed to generate default templates:", err.Error())
		}
	}

	// Mark first run as complete
	if err := a.configService.SetFirstRun(false); err != nil {
		return fmt.Errorf("failed to update first run status: %w", err)
	}

	return nil
}

// ValidateFolderStructure checks if the DealDone folder structure is valid
func (a *App) ValidateFolderStructure() error {
	if a.folderManager == nil {
		return fmt.Errorf("folder manager not initialized")
	}
	return a.folderManager.ValidateFolderStructure()
}

// GetDefaultDealDonePath returns the suggested default path for DealDone
func (a *App) GetDefaultDealDonePath() string {
	return filepath.Join(a.GetDesktopDirectory(), "DealDone")
}

// CheckFolderWritePermission checks if the app can write to the specified path
func (a *App) CheckFolderWritePermission(path string) bool {
	// Try to create a test directory
	testPath := filepath.Join(path, ".dealdone_test")
	err := os.MkdirAll(testPath, 0755)
	if err != nil {
		return false
	}

	// Clean up test directory
	os.RemoveAll(testPath)
	return true
}

// GetConfiguredTemplatesPath returns the path to the Templates folder
func (a *App) GetConfiguredTemplatesPath() string {
	if a.configService == nil {
		return ""
	}
	return a.configService.GetTemplatesPath()
}

// GetConfiguredDealsPath returns the path to the Deals folder
func (a *App) GetConfiguredDealsPath() string {
	if a.configService == nil {
		return ""
	}
	return a.configService.GetDealsPath()
}
