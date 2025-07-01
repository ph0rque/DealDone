import React, { useEffect } from 'react'
import { useToast } from '../hooks/use-toast'

// Keyboard shortcuts handler
export function useKeyboardShortcuts() {
  const { toast } = useToast()

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // Cmd/Ctrl + / to show shortcuts help
      if ((event.metaKey || event.ctrlKey) && event.key === '/') {
        event.preventDefault()
        toast({
          title: "Keyboard Shortcuts",
          description: "F5: Refresh • Cmd/Ctrl+F: Search • Escape: Clear/Cancel • Enter: Confirm • Delete: Delete selected",
        })
      }

      // F5 to refresh
      if (event.key === 'F5') {
        event.preventDefault()
        window.location.reload()
      }

      // Escape to close dialogs/modals (handled by individual components)
      if (event.key === 'Escape') {
        // This is handled by individual dialog components
        console.log('Escape key pressed')
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [toast])

  return null
}

// Performance monitoring for file operations
export function usePerformanceMonitoring() {
  const { toast } = useToast()

  const monitorOperation = (operationName: string, fn: () => Promise<any>) => {
    return async () => {
      const startTime = performance.now()
      
      try {
        const result = await fn()
        const endTime = performance.now()
        const duration = endTime - startTime

        // Show performance warning for slow operations (>2 seconds)
        if (duration > 2000) {
          console.warn(`Slow operation detected: ${operationName} took ${duration.toFixed(2)}ms`)
          
          // Optionally show user feedback for very slow operations
          if (duration > 5000) {
            toast({
              title: "Slow Operation",
              description: `${operationName} took longer than expected. Consider optimizing your file system.`,
            })
          }
        }

        return result
      } catch (error) {
        const endTime = performance.now()
        console.error(`Operation failed: ${operationName} failed after ${(endTime - startTime).toFixed(2)}ms`, error)
        throw error
      }
    }
  }

  return { monitorOperation }
}

// Accessibility enhancements
export function AccessibilityProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    // Add focus visible styles for keyboard navigation
    const style = document.createElement('style')
    style.textContent = `
      .focus-visible {
        outline: 2px solid hsl(var(--ring));
        outline-offset: 2px;
      }
      
      /* Ensure proper contrast for focus states */
      button:focus-visible,
      input:focus-visible,
      select:focus-visible,
      textarea:focus-visible {
        outline: 2px solid hsl(var(--ring));
        outline-offset: 2px;
      }
    `
    document.head.appendChild(style)

    return () => {
      document.head.removeChild(style)
    }
  }, [])

  return <>{children}</>
}

// Loading state improvements
export function useOptimisticUpdates() {
  const [optimisticOperations, setOptimisticOperations] = React.useState<Set<string>>(new Set())

  const startOptimisticOperation = (operationId: string) => {
    setOptimisticOperations(prev => new Set(prev).add(operationId))
  }

  const endOptimisticOperation = (operationId: string) => {
    setOptimisticOperations(prev => {
      const newSet = new Set(prev)
      newSet.delete(operationId)
      return newSet
    })
  }

  const isOperationPending = (operationId: string) => {
    return optimisticOperations.has(operationId)
  }

  return {
    startOptimisticOperation,
    endOptimisticOperation,
    isOperationPending
  }
}

// User feedback helpers
export function useFeedbackHelpers() {
  const { toast } = useToast()

  const showSuccess = (message: string, title?: string) => {
    toast({
      title: title || "Success",
      description: message,
    })
  }

  const showWarning = (message: string, title?: string) => {
    toast({
      title: title || "Warning",
      description: message,
    })
  }

  const showInfo = (message: string, title?: string) => {
    toast({
      title: title || "Info",
      description: message,
    })
  }

  return {
    showSuccess,
    showWarning,
    showInfo
  }
} 