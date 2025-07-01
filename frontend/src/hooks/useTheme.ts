import { useState, useEffect, useCallback } from 'react'
import type { Theme } from '@/types'

export function useTheme() {
  const [theme, setTheme] = useState<Theme>('system')
  const [resolvedTheme, setResolvedTheme] = useState<'light' | 'dark'>('light')

  // Check system theme preference
  const getSystemTheme = useCallback((): 'light' | 'dark' => {
    if (typeof window !== 'undefined' && window.matchMedia) {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    }
    return 'light'
  }, [])

  // Apply theme to document
  const applyTheme = useCallback((appliedTheme: 'light' | 'dark' | 'dracula') => {
    const root = window.document.documentElement
    root.classList.remove('light', 'dark', 'dracula')
    root.classList.add(appliedTheme)
    setResolvedTheme(appliedTheme === 'dracula' ? 'dark' : appliedTheme)
  }, [])

  // Set theme
  const changeTheme = useCallback((newTheme: Theme) => {
    setTheme(newTheme)
    localStorage.setItem('theme', newTheme)

    let themeToApply: 'light' | 'dark' | 'dracula'
    if (newTheme === 'system') {
      themeToApply = getSystemTheme()
    } else if (newTheme === 'dracula') {
      themeToApply = 'dracula'
    } else {
      themeToApply = newTheme
    }
    applyTheme(themeToApply)
  }, [getSystemTheme, applyTheme])

  // Initialize theme on mount
  useEffect(() => {
    const savedTheme = localStorage.getItem('theme') as Theme | null
    const initialTheme = savedTheme || 'system'
    setTheme(initialTheme)

    let themeToApply: 'light' | 'dark' | 'dracula'
    if (initialTheme === 'system') {
      themeToApply = getSystemTheme()
    } else if (initialTheme === 'dracula') {
      themeToApply = 'dracula'
    } else {
      themeToApply = initialTheme
    }
    applyTheme(themeToApply)
  }, [getSystemTheme, applyTheme])

  // Listen for system theme changes
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    
    const handleChange = () => {
      if (theme === 'system') {
        applyTheme(getSystemTheme())
      }
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [theme, getSystemTheme, applyTheme])

  return {
    theme,
    resolvedTheme,
    setTheme: changeTheme
  }
} 