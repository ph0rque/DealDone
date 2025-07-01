import React, { useState, useEffect, useCallback } from 'react'
import { TreeNode } from './TreeNode'
import { LoadingSpinner } from './LoadingSpinner'
import { useFileManager } from '../contexts/FileManagerContext'
import type { FileSystemItem } from '../types'

export function FileTree() {
  const { state, setCurrentPath, setSelectedItems, setLoading, setError } = useFileManager()
  const [rootItems, setRootItems] = useState<FileSystemItem[]>([])
  const [selectedItemId, setSelectedItemId] = useState<string | null>(null)

  // Load initial directory
  useEffect(() => {
    loadDirectory(state.currentPath)
  }, [state.currentPath])

  const loadDirectory = async (path: string) => {
    setLoading(true)
    setError(null)
    
    try {
      // TODO: Replace with actual Wails API call
      // This is a mock implementation
      console.log('Loading directory:', path)
      
      // Simulate API call delay
      await new Promise(resolve => setTimeout(resolve, 500))
      
      // Mock data - replace with actual API call to backend
      const mockItems: FileSystemItem[] = [
        {
          id: '1',
          name: 'Documents',
          path: '/Documents',
          isDirectory: true,
          size: 0,
          modifiedAt: new Date(),
          createdAt: new Date(),
          permissions: { readable: true, writable: true, executable: true },
          isExpanded: false,
          children: []
        },
        {
          id: '2',
          name: 'Downloads',
          path: '/Downloads',
          isDirectory: true,
          size: 0,
          modifiedAt: new Date(),
          createdAt: new Date(),
          permissions: { readable: true, writable: true, executable: true },
          isExpanded: false,
          children: []
        },
        {
          id: '3',
          name: 'example.txt',
          path: '/example.txt',
          isDirectory: false,
          size: 1024,
          modifiedAt: new Date(),
          createdAt: new Date(),
          permissions: { readable: true, writable: true, executable: false },
          extension: 'txt',
          mimeType: 'text/plain'
        }
      ]
      
      setRootItems(mockItems)
    } catch (error) {
      console.error('Failed to load directory:', error)
      setError('Failed to load directory')
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

    // Update the item's expanded state
    const updateItemExpanded = (items: FileSystemItem[], targetId: string): FileSystemItem[] => {
      return items.map(currentItem => {
        if (currentItem.id === targetId) {
          const updatedItem = { ...currentItem, isExpanded: !currentItem.isExpanded }
          
          // Load children if expanding and not already loaded
          if (updatedItem.isExpanded && !updatedItem.children?.length) {
            // TODO: Load children from backend
            updatedItem.isLoading = true
            // Mock loading children
            setTimeout(() => {
              // This would be replaced with actual API call
              updatedItem.children = []
              updatedItem.isLoading = false
              setRootItems(prev => [...prev])
            }, 500)
          }
          
          return updatedItem
        }
        
        if (currentItem.children) {
          return {
            ...currentItem,
            children: updateItemExpanded(currentItem.children, targetId)
          }
        }
        
        return currentItem
      })
    }

    setRootItems(prev => updateItemExpanded(prev, item.id))
  }, [])

  const handleContextMenu = useCallback((item: FileSystemItem, event: React.MouseEvent) => {
    // TODO: Implement context menu actions
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
      {/* Current path breadcrumb */}
      <div className="mb-2 p-2 bg-muted rounded text-xs text-muted-foreground">
        {state.currentPath}
      </div>

      {/* Tree items */}
      <div className="space-y-0.5">
        {rootItems.map((item) => (
          <TreeNode
            key={item.id}
            item={item}
            level={0}
            isSelected={selectedItemId === item.id}
            onSelect={handleItemSelect}
            onExpand={handleItemExpand}
            onContextMenu={handleContextMenu}
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