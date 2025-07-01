import React, { useState, useEffect } from 'react'
import { FileManagerProvider } from './contexts/FileManagerContext'
import { FileManagerLayout } from './components/FileManagerLayout'
import { ErrorBoundary } from './components/ErrorBoundary'
import { FirstRunSetup } from './components/FirstRunSetup'
import { DealDashboard } from './components/DealDashboard'
import { Settings } from './components/Settings'
import { ProcessingProgress } from './components/ProcessingProgress'
import { ThemeToggle } from './components/ThemeToggle'
import { Toaster } from './components/ui/toaster'
import { IsFirstRun, CompleteFirstRunSetup, IsDealDoneReady } from '../wailsjs/go/main/App'
import './App.css'
import { Button } from './components/ui/button'
import { LayoutDashboard, FolderOpen, Settings as SettingsIcon } from 'lucide-react'

type AppView = 'dashboard' | 'filemanager'

function App() {
  const [isFirstRun, setIsFirstRun] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [isDealDoneReady, setIsDealDoneReady] = useState(false)
  const [currentView, setCurrentView] = useState<AppView>('dashboard')
  const [showSettings, setShowSettings] = useState(false)
  const [showProcessing, setShowProcessing] = useState(false)

  useEffect(() => {
    checkFirstRun()
  }, [])

  const checkFirstRun = async () => {
    try {
      const firstRun = await IsFirstRun()
      setIsFirstRun(firstRun)
      
      if (!firstRun) {
        const ready = await IsDealDoneReady()
        setIsDealDoneReady(ready)
      }
    } catch (error) {
      console.error('Error checking first run:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleSetupComplete = async (path: string) => {
    try {
      await CompleteFirstRunSetup(path)
      setIsFirstRun(false)
      setIsDealDoneReady(true)
    } catch (error) {
      console.error('Error completing setup:', error)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-50 dark:bg-gray-900">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (isFirstRun) {
    return (
      <ErrorBoundary>
        <FirstRunSetup onComplete={handleSetupComplete} />
        <Toaster />
      </ErrorBoundary>
    )
  }

  return (
    <ErrorBoundary>
      <FileManagerProvider>
        <div className="flex flex-col h-screen bg-gray-50 dark:bg-gray-900">
          {/* Header */}
          <header className="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between px-4 py-3">
              <div className="flex items-center space-x-4">
                <h1 className="text-xl font-bold">DealDone</h1>
                
                {isDealDoneReady && (
                  <nav className="flex space-x-2">
                    <Button
                      variant={currentView === 'dashboard' ? 'default' : 'ghost'}
                      size="sm"
                      onClick={() => setCurrentView('dashboard')}
                    >
                      <LayoutDashboard className="h-4 w-4 mr-2" />
                      Dashboard
                    </Button>
                    <Button
                      variant={currentView === 'filemanager' ? 'default' : 'ghost'}
                      size="sm"
                      onClick={() => setCurrentView('filemanager')}
                    >
                      <FolderOpen className="h-4 w-4 mr-2" />
                      File Manager
                    </Button>
                  </nav>
                )}
              </div>
              
              <div className="flex items-center space-x-2">
                <ThemeToggle />
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowSettings(true)}
                >
                  <SettingsIcon className="h-5 w-5" />
                </Button>
              </div>
            </div>
          </header>

          {/* Main Content */}
          <main className="flex-1 overflow-hidden">
            {isDealDoneReady ? (
              currentView === 'dashboard' ? (
                <DealDashboard />
              ) : (
                <FileManagerLayout />
              )
            ) : (
              <div className="flex items-center justify-center h-full">
                <div className="text-center">
                  <p className="text-lg text-gray-600 dark:text-gray-400">
                    DealDone folder is not ready. Please complete the setup.
                  </p>
                  <Button 
                    className="mt-4"
                    onClick={() => setIsFirstRun(true)}
                  >
                    Run Setup
                  </Button>
                </div>
              </div>
            )}
          </main>

          {/* Settings Modal */}
          <Settings 
            isOpen={showSettings} 
            onClose={() => setShowSettings(false)} 
          />

          {/* Processing Progress */}
          <ProcessingProgress 
            isOpen={showProcessing}
            onClose={() => setShowProcessing(false)}
          />

          {/* Toast Notifications */}
          <Toaster />
        </div>
      </FileManagerProvider>
    </ErrorBoundary>
  )
}

export default App
