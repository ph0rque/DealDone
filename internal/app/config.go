package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config represents the application configuration
type Config struct {
	DealDoneRoot    string `json:"dealdone_root"`
	FirstRun        bool   `json:"first_run"`
	DefaultTemplate string `json:"default_template"`
	LastOpenedDeal  string `json:"last_opened_deal"`
}

// ConfigService handles configuration management
type ConfigService struct {
	configPath string
	config     *Config
}

// NewConfigService creates a new configuration service
func NewConfigService() (*ConfigService, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")

	cs := &ConfigService{
		configPath: configPath,
	}

	// Load existing config or create default
	if err := cs.Load(); err != nil {
		cs.config = cs.getDefaultConfig()
		if err := cs.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	return cs, nil
}

// getConfigDir is a variable for testability
var getConfigDir = getDefaultConfigDir

// getDefaultConfigDir returns the appropriate config directory for the OS
func getDefaultConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, "Library", "Application Support", "DealDone")
	case "windows":
		configDir = filepath.Join(os.Getenv("APPDATA"), "DealDone")
	default: // Linux and others
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config", "dealdone")
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

// getDefaultConfig returns default configuration
func (cs *ConfigService) getDefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	defaultRoot := filepath.Join(home, "Desktop", "DealDone")

	return &Config{
		DealDoneRoot:    defaultRoot,
		FirstRun:        true,
		DefaultTemplate: "",
		LastOpenedDeal:  "",
	}
}

// Load reads configuration from disk
func (cs *ConfigService) Load() error {
	data, err := os.ReadFile(cs.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cs.config = cs.getDefaultConfig()
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	cs.config = &config
	return nil
}

// Save writes configuration to disk
func (cs *ConfigService) Save() error {
	data, err := json.MarshalIndent(cs.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cs.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration
func (cs *ConfigService) GetConfig() *Config {
	return cs.config
}

// SetDealDoneRoot updates the DealDone root folder path
func (cs *ConfigService) SetDealDoneRoot(path string) error {
	cs.config.DealDoneRoot = path
	return cs.Save()
}

// SetFirstRun updates the first run flag
func (cs *ConfigService) SetFirstRun(firstRun bool) error {
	cs.config.FirstRun = firstRun
	return cs.Save()
}

// GetDealDoneRoot returns the DealDone root folder path
func (cs *ConfigService) GetDealDoneRoot() string {
	return cs.config.DealDoneRoot
}

// IsFirstRun returns whether this is the first run
func (cs *ConfigService) IsFirstRun() bool {
	return cs.config.FirstRun
}

// GetTemplatesPath returns the path to the Templates folder
func (cs *ConfigService) GetTemplatesPath() string {
	return filepath.Join(cs.config.DealDoneRoot, "Templates")
}

// GetDealsPath returns the path to the Deals folder
func (cs *ConfigService) GetDealsPath() string {
	return filepath.Join(cs.config.DealDoneRoot, "Deals")
}
