import React, { useState, useEffect } from 'react'
import { FileManagerProvider } from './contexts/FileManagerContext'
import { FileManagerLayout } from './components/FileManagerLayout'
import { ErrorBoundary } from './components/ErrorBoundary'
import { FirstRunSetup } from './components/FirstRunSetup'
import { useTheme } from './hooks/useTheme'
import { useKeyboardShortcuts, AccessibilityProvider } from './components/UXEnhancements'
import { IsFirstRun } from '../wailsjs/go/main/App'
import './App.css'

function ThemeProvider({ children }: { children: React.ReactNode }) {
  useTheme()
  useKeyboardShortcuts()
  return <>{children}</>
}

function App() {
  const [isFirstRun, setIsFirstRun] = useState<boolean | null>(null)

  useEffect(() => {
    // Check if this is the first run
    IsFirstRun().then(setIsFirstRun)
  }, [])

  // Show loading state while checking
  if (isFirstRun === null) {
    return (
      <div className="h-screen flex items-center justify-center">
        <div className="text-gray-500 dark:text-gray-400">Loading...</div>
      </div>
    )
  }

  // Show first run setup if needed
  if (isFirstRun) {
    return (
      <ErrorBoundary>
        <ThemeProvider>
          <FirstRunSetup onComplete={() => setIsFirstRun(false)} />
        </ThemeProvider>
      </ErrorBoundary>
    )
  }

  // Show main app
  return (
    <ErrorBoundary>
      <AccessibilityProvider>
        <ThemeProvider>
          <FileManagerProvider>
            <FileManagerLayout />
          </FileManagerProvider>
        </ThemeProvider>
      </AccessibilityProvider>
    </ErrorBoundary>
  )
}

export default App
