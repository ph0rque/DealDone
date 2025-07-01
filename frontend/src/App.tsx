import React from 'react'
import { FileManagerProvider } from './contexts/FileManagerContext'
import { FileManagerLayout } from './components/FileManagerLayout'
import { ErrorBoundary } from './components/ErrorBoundary'
import { useTheme } from './hooks/useTheme'
import { useKeyboardShortcuts, AccessibilityProvider } from './components/UXEnhancements'
import './App.css'

function ThemeProvider({ children }: { children: React.ReactNode }) {
  useTheme()
  useKeyboardShortcuts()
  return <>{children}</>
}

function App() {
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
