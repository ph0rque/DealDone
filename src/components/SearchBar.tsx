import React, { useState, useCallback, useEffect } from 'react'
import { Search, X, Loader2 } from 'lucide-react'
import { Input } from './ui/input'
import { Button } from './ui/button'
import { useFileManager } from '../contexts/FileManagerContext'
import { cn } from '../lib/utils'

interface SearchBarProps {
  placeholder?: string
  className?: string
}

export function SearchBar({ 
  placeholder = "Search files and folders...",
  className 
}: SearchBarProps) {
  const { state, setSearchQuery, setSearchResults } = useFileManager()
  const [localQuery, setLocalQuery] = useState(state.searchQuery)
  const [isSearching, setIsSearching] = useState(false)

  // Debounced search function
  const debouncedSearch = useCallback((query: string) => {
    const timeoutId = setTimeout(async () => {
      if (query.trim()) {
        setIsSearching(true)
        setSearchQuery(query)
        
        try {
          // TODO: Integrate with Wails backend search API
          // For now, this is a placeholder
          console.log('Searching for:', query)
          
          // Simulate API call delay
          await new Promise(resolve => setTimeout(resolve, 500))
          
          // Mock search results (replace with actual API call)
          const mockResults = {
            items: [],
            query,
            totalCount: 0
          }
          
          setSearchResults(mockResults)
        } catch (error) {
          console.error('Search error:', error)
          setSearchResults(null)
        } finally {
          setIsSearching(false)
        }
      } else {
        setSearchQuery('')
        setSearchResults(null)
      }
    }, 300) // 300ms debounce

    return () => clearTimeout(timeoutId)
  }, [setSearchQuery, setSearchResults])

  // Handle input changes
  useEffect(() => {
    const cleanup = debouncedSearch(localQuery)
    return cleanup
  }, [localQuery, debouncedSearch])

  const handleClear = () => {
    setLocalQuery('')
    setSearchQuery('')
    setSearchResults(null)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      handleClear()
    }
  }

  return (
    <div className={cn("relative", className)}>
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
        
        <Input
          type="text"
          value={localQuery}
          onChange={(e) => setLocalQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          className="pl-10 pr-10"
        />

        <div className="absolute right-2 top-1/2 transform -translate-y-1/2 flex items-center gap-1">
          {isSearching && (
            <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />
          )}
          
          {localQuery && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleClear}
              className="h-auto p-1 hover:bg-transparent"
            >
              <X className="w-4 h-4 text-muted-foreground hover:text-foreground" />
            </Button>
          )}
        </div>
      </div>

      {/* Search results count */}
      {state.searchResults && (
        <div className="mt-2 text-xs text-muted-foreground">
          {state.searchResults.totalCount > 0 ? (
            `Found ${state.searchResults.totalCount} item${state.searchResults.totalCount !== 1 ? 's' : ''}`
          ) : (
            'No results found'
          )}
        </div>
      )}
    </div>
  )
} 