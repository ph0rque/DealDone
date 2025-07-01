package main

import (
	"context"
	"os"
	"path/filepath"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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
