import { useState, useCallback } from 'react'
import { FileManagerAPI } from '../services/fileManagerApi'
import { ErrorService } from '../services/errorService'
import type { FileSystemItem, FileOperationResult, FileOperation } from '../types'

export function useFileOperations() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const executeOperation = useCallback(async (
    operation: FileOperation,
    items: FileSystemItem[],
    targetPath?: string
  ): Promise<FileOperationResult> => {
    setIsLoading(true)
    setError(null)
    
    try {
      console.log(`Executing ${operation} on:`, items, 'Target:', targetPath)
      
      // This is now handled by specific operation methods
      return {
        success: true,
        message: `${operation} operation completed successfully`
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred'
      setError(errorMessage)
      return {
        success: false,
        error: errorMessage
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  const createFile = useCallback(async (path: string, name: string): Promise<FileOperationResult> => {
    setIsLoading(true)
    setError(null)
    
    try {
      const result = await FileManagerAPI.createItem(path, name, false)
      if (!result.success) {
        const appError = ErrorService.handleFileOperationError('create', result.error || 'Unknown error', name)
        setError(appError.message)
        return {
          success: false,
          error: appError.message
        }
      }
      return result
    } catch (err) {
      const appError = ErrorService.handleFileOperationError('create', err instanceof Error ? err : 'Unknown error', name)
      setError(appError.message)
      return {
        success: false,
        error: appError.message
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  const createFolder = useCallback(async (path: string, name: string): Promise<FileOperationResult> => {
    setIsLoading(true)
    setError(null)
    
    try {
      const result = await FileManagerAPI.createItem(path, name, true)
      if (!result.success) {
        setError(result.error || 'Failed to create folder')
      }
      return result
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create folder'
      setError(errorMessage)
      return {
        success: false,
        error: errorMessage
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  const copyItems = useCallback(async (items: FileSystemItem[], targetPath: string): Promise<FileOperationResult> => {
    setIsLoading(true)
    setError(null)
    
    try {
      const sourcePaths = items.map(item => item.path)
      const result = await FileManagerAPI.copyItems(sourcePaths, targetPath)
      if (!result.success) {
        setError(result.error || 'Failed to copy items')
      }
      return result
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to copy items'
      setError(errorMessage)
      return {
        success: false,
        error: errorMessage
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  const moveItems = useCallback(async (items: FileSystemItem[], targetPath: string): Promise<FileOperationResult> => {
    setIsLoading(true)
    setError(null)
    
    try {
      const sourcePaths = items.map(item => item.path)
      const result = await FileManagerAPI.moveItems(sourcePaths, targetPath)
      if (!result.success) {
        setError(result.error || 'Failed to move items')
      }
      return result
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to move items'
      setError(errorMessage)
      return {
        success: false,
        error: errorMessage
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  const deleteItems = useCallback(async (items: FileSystemItem[]): Promise<FileOperationResult> => {
    setIsLoading(true)
    setError(null)
    
    try {
      const paths = items.map(item => item.path)
      const result = await FileManagerAPI.deleteItems(paths)
      if (!result.success) {
        setError(result.error || 'Failed to delete items')
      }
      return result
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to delete items'
      setError(errorMessage)
      return {
        success: false,
        error: errorMessage
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  const renameItem = useCallback(async (item: FileSystemItem, newName: string): Promise<FileOperationResult> => {
    setIsLoading(true)
    setError(null)
    
    try {
      const result = await FileManagerAPI.renameItem(item.path, newName)
      if (!result.success) {
        setError(result.error || 'Failed to rename item')
      }
      return result
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to rename item'
      setError(errorMessage)
      return {
        success: false,
        error: errorMessage
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  const openItem = useCallback(async (item: FileSystemItem): Promise<FileOperationResult> => {
    setIsLoading(true)
    setError(null)
    
    try {
      const result = await FileManagerAPI.openFile(item.path)
      if (!result.success) {
        setError(result.error || 'Failed to open item')
      }
      return result
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to open item'
      setError(errorMessage)
      return {
        success: false,
        error: errorMessage
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  return {
    isLoading,
    error,
    createFile,
    createFolder,
    copyItems,
    moveItems,
    deleteItems,
    renameItem,
    openItem
  }
} 