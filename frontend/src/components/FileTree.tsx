import React, { useState, useEffect, useCallback } from 'react'
import { TreeNode } from './TreeNode'
import { LoadingSpinner } from './LoadingSpinner'
import { useFileManager } from '../contexts/FileManagerContext'
import { FileManagerAPI } from '../services/fileManagerApi'
import { useToast } from '../hooks/use-toast'
import type { FileSystemItem } from '../types'

export function FileTree() {
  const { state, setCurrentPath, setSelectedItems, setLoading, setError } = useFileManager()
  const [rootItems, setRootItems] = useState<FileSystemItem[]>([])
  const [selectedItemId, setSelectedItemId] = useState<string | null>(null)
  const [showHidden, setShowHidden] = useState(false)
  const [clipboardItems, setClipboardItems] = useState<FileSystemItem[] | null>(null)
  const [clipboardOperation, setClipboardOperation] = useState<'copy' | 'cut' | null>(null)
  const { toast } = useToast()

  // Load initial directory
  useEffect(() => {
    // Start with user's home directory
    const initializeFileTree = async () => {
      try {
        const directories = await FileManagerAPI.getCommonDirectories()
        setCurrentPath(directories.home)
      } catch (error) {
        console.error('Failed to get home directory:', error)
        setCurrentPath('/')
      }
    }
    
    initializeFileTree()
  }, [setCurrentPath])

  // Load directory when current path changes
  useEffect(() => {
    if (state.currentPath) {
      loadDirectory(state.currentPath)
    }
  }, [state.currentPath, showHidden])

  const loadDirectory = async (path: string) => {
    setLoading(true)
    setError(null)
    
    try {
      console.log('Loading directory:', path)
      const items = await FileManagerAPI.listDirectory(path, showHidden)
      setRootItems(items)
    } catch (error) {
      console.error('Failed to load directory:', error)
      const errorMessage = error instanceof Error ? error.message : 'Failed to load directory'
      setError(errorMessage)
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive"
      })
    } finally {
      setLoading(false)
    }
  }

  const handleItemSelect = useCallback((item: FileSystemItem) => {
    setSelectedItemId(item.id)
    setSelectedItems([item.id])
    
    if (!item.isDirectory) {
      // TODO: Handle file selection/preview
      console.log('File selected:', item.name)
    }
  }, [setSelectedItems])

  const handleItemExpand = useCallback(async (item: FileSystemItem) => {
    if (!item.isDirectory) return

    const isCurrentlyExpanded = item.isExpanded
    const targetId = item.id

    // Simple toggle for collapse
    if (isCurrentlyExpanded) {
      setRootItems(prev => updateItemExpanded(prev, targetId, !isCurrentlyExpanded))
      return
    }

    // For expansion, check if we need to load children
    if (!isCurrentlyExpanded && (!item.children || item.children.length === 0)) {
      // Set loading state
      setRootItems(prev => updateItemExpanded(prev, targetId, !isCurrentlyExpanded, true))

      try {
        console.log('Loading children for:', item.path)
        const children = await FileManagerAPI.listDirectory(item.path, showHidden)
        console.log('Loaded children:', children.length, 'items')
        
        // Update with children and remove loading state
        setRootItems(prev => updateItemChildren(prev, targetId, children, true))
      } catch (error) {
        console.error('Failed to load children:', error)
        toast({
          title: "Error",
          description: `Failed to load folder contents: ${error instanceof Error ? error.message : 'Unknown error'}`,
          variant: "destructive"
        })
        
        // Collapse the item on error and remove loading state
        setRootItems(prev => updateItemExpanded(prev, targetId, false, false))
      }
    } else {
      // Just toggle if children already exist
      setRootItems(prev => updateItemExpanded(prev, targetId, !isCurrentlyExpanded))
    }
  }, [showHidden, toast])

  // Helper function to update item expanded state
  const updateItemExpanded = (items: FileSystemItem[], targetId: string, isExpanded: boolean, isLoading = false): FileSystemItem[] => {
    return items.map(currentItem => {
      if (currentItem.id === targetId) {
        return {
          ...currentItem,
          isExpanded,
          isLoading
        }
      }
      
      if (currentItem.children) {
        return {
          ...currentItem,
          children: updateItemExpanded(currentItem.children, targetId, isExpanded, isLoading)
        }
      }
      
      return currentItem
    })
  }

  // Helper function to update item children
  const updateItemChildren = (items: FileSystemItem[], targetId: string, children: FileSystemItem[], isExpanded = true): FileSystemItem[] => {
    return items.map(currentItem => {
      if (currentItem.id === targetId) {
        return {
          ...currentItem,
          children,
          isLoading: false,
          isExpanded
        }
      }
      
      if (currentItem.children) {
        return {
          ...currentItem,
          children: updateItemChildren(currentItem.children, targetId, children, isExpanded)
        }
      }
      
      return currentItem
    })
  }

  const handleCopy = useCallback((items: FileSystemItem[]) => {
    setClipboardItems(items)
    setClipboardOperation('copy')
  }, [])

  const handleCut = useCallback((items: FileSystemItem[]) => {
    setClipboardItems(items)
    setClipboardOperation('cut')
  }, [])

  const handleRefresh = useCallback(() => {
    if (state.currentPath) {
      loadDirectory(state.currentPath)
    }
  }, [state.currentPath, showHidden])

  const handleContextMenu = useCallback((item: FileSystemItem, event: React.MouseEvent) => {
    console.log('Context menu for:', item.name, 'at', event.clientX, event.clientY)
  }, [])

  if (state.isLoading && rootItems.length === 0) {
    return (
      <div className="flex items-center justify-center h-32">
        <LoadingSpinner />
      </div>
    )
  }

  if (state.error) {
    return (
      <div className="p-4 text-center">
        <p className="text-sm text-destructive">{state.error}</p>
        <button
          onClick={() => loadDirectory(state.currentPath)}
          className="mt-2 text-xs text-muted-foreground hover:text-foreground underline"
        >
          Retry
        </button>
      </div>
    )
  }

  return (
    <div className="p-2">
      {/* Current path breadcrumb and controls */}
      <div className="mb-2 space-y-2">
        <div className="p-2 bg-muted rounded text-xs text-muted-foreground font-mono">
          {state.currentPath}
        </div>
        
        {/* Show hidden files toggle */}
        <div className="flex items-center gap-2 text-xs">
          <input
            type="checkbox"
            id="show-hidden"
            checked={showHidden}
            onChange={(e) => setShowHidden(e.target.checked)}
            className="w-3 h-3"
          />
          <label htmlFor="show-hidden" className="text-muted-foreground cursor-pointer">
            Show hidden files
          </label>
        </div>
      </div>

      {/* Quick navigation */}
      <div className="mb-3">
        <QuickNavigation onNavigate={setCurrentPath} />
      </div>

      {/* Tree items */}
      <div className="space-y-0">
        {rootItems.map((item, index) => (
          <TreeNode
            key={item.id}
            item={item}
            level={0}
            isSelected={selectedItemId === item.id}
            selectedItemId={selectedItemId}
            onSelect={handleItemSelect}
            onExpand={handleItemExpand}
            onContextMenu={handleContextMenu}
            onRefresh={handleRefresh}
            clipboardItems={clipboardItems}
            onCopy={handleCopy}
            onCut={handleCut}
            isLast={index === rootItems.length - 1}
          />
        ))}
      </div>

      {rootItems.length === 0 && !state.isLoading && (
        <div className="p-4 text-center text-sm text-muted-foreground">
          This folder is empty
        </div>
      )}
    </div>
  )
}

// Quick navigation component
function QuickNavigation({ onNavigate }: { onNavigate: (path: string) => void }) {
  const [directories, setDirectories] = useState<Record<string, string> | null>(null)

  useEffect(() => {
    FileManagerAPI.getCommonDirectories()
      .then(setDirectories)
      .catch(console.error)
  }, [])

  if (!directories) return null

  const navItems = [
    { label: 'Home', path: directories.home, icon: '🏠' },
    { label: 'Documents', path: directories.documents, icon: '📄' },
    { label: 'Desktop', path: directories.desktop, icon: '🖥️' },
    { label: 'Downloads', path: directories.downloads, icon: '📥' }
  ]

  return (
    <div className="space-y-1">
      {navItems.map((item) => (
        <button
          key={item.path}
          onClick={() => onNavigate(item.path)}
          className="w-full text-left px-2 py-1 text-xs hover:bg-muted rounded transition-colors flex items-center gap-2"
        >
          <span>{item.icon}</span>
          <span>{item.label}</span>
        </button>
      ))}
    </div>
  )
} 