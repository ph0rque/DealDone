import React, { createContext, useContext, useState, ReactNode } from 'react'
import type { FileManagerState, FileSystemItem, SearchResult } from '../types'

interface FileManagerContextType {
  state: FileManagerState
  setCurrentPath: (path: string) => void
  setSelectedItems: (items: string[]) => void
  setSearchQuery: (query: string) => void
  setSearchResults: (results: SearchResult | null) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
}

const FileManagerContext = createContext<FileManagerContextType | null>(null)

interface FileManagerProviderProps {
  children: ReactNode
}

export function FileManagerProvider({ children }: FileManagerProviderProps) {
  const [state, setState] = useState<FileManagerState>({
    currentPath: '/',
    selectedItems: [],
    clipboardItems: null,
    searchQuery: '',
    searchResults: null,
    isLoading: false,
    error: null
  })

  const setCurrentPath = (path: string) => {
    setState(prev => ({ ...prev, currentPath: path, selectedItems: [] }))
  }

  const setSelectedItems = (items: string[]) => {
    setState(prev => ({ ...prev, selectedItems: items }))
  }

  const setSearchQuery = (query: string) => {
    setState(prev => ({ ...prev, searchQuery: query }))
  }

  const setSearchResults = (results: SearchResult | null) => {
    setState(prev => ({ ...prev, searchResults: results }))
  }

  const setLoading = (loading: boolean) => {
    setState(prev => ({ ...prev, isLoading: loading }))
  }

  const setError = (error: string | null) => {
    setState(prev => ({ ...prev, error }))
  }

  return (
    <FileManagerContext.Provider value={{
      state,
      setCurrentPath,
      setSelectedItems,
      setSearchQuery,
      setSearchResults,
      setLoading,
      setError
    }}>
      {children}
    </FileManagerContext.Provider>
  )
}

export function useFileManager() {
  const context = useContext(FileManagerContext)
  if (!context) {
    throw new Error('useFileManager must be used within a FileManagerProvider')
  }
  return context
} 