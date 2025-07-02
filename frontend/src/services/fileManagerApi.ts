// API Integration Service for Wails File Manager Backend
import {
  ListDirectory,
  CreateFile,
  CopyItems,
  MoveItems,
  DeleteItems,
  RenameItem,
  OpenFile,
  SearchFiles,
  GetHomeDirectory,
  GetDocumentsDirectory,
  GetDesktopDirectory,
  GetDownloadsDirectory
} from '../../wailsjs/go/main/App'

import { main } from '../../wailsjs/go/models'

// Type imports from our frontend types
import type { 
  FileSystemItem, 
  FileOperationResult, 
  SearchResult as LocalSearchResult 
} from '../types'

// Helper function to convert Go Time to JavaScript Date
function convertTimeToDate(time: any): Date {
  if (!time) return new Date()
  
  // If it's already a Date, return it
  if (time instanceof Date) return time
  
  // If it's a string, parse it
  if (typeof time === 'string') return new Date(time)
  
  // If it's a Go Time object with a time string property
  if (time.time) return new Date(time.time)
  
  // If it has nanoseconds since epoch
  if (typeof time === 'number') return new Date(time / 1000000) // Convert nanoseconds to milliseconds
  
  // Fallback to current time
  return new Date()
}

// Helper function to convert backend item to frontend format
function convertToFrontendItem(item: any): FileSystemItem {
  return {
    id: item.id || '',
    name: item.name || '',
    path: item.path || '',
    isDirectory: item.isDirectory || false,
    size: item.size || 0,
    modifiedAt: convertTimeToDate(item.modifiedAt),
    createdAt: convertTimeToDate(item.createdAt),
    permissions: item.permissions || { readable: false, writable: false, executable: false },
    extension: item.extension,
    mimeType: item.mimeType,
    children: item.children ? item.children.map(convertToFrontendItem) : undefined,
    isExpanded: false,
    isLoading: false
  }
}

export class FileManagerAPI {
  /**
   * List directory contents
   */
  static async listDirectory(path: string, showHidden = false): Promise<FileSystemItem[]> {
    try {
      const request = new main.DirectoryListRequest({
        path,
        showHidden
      })
      
      const items = await ListDirectory(request)
      
      // Ensure items is an array - handle null/undefined case
      if (!items || !Array.isArray(items)) {
        console.warn('ListDirectory returned null/undefined, returning empty array')
        return []
      }
      
      // Convert backend items to frontend format with proper Date conversion
      return items.map(convertToFrontendItem)
    } catch (error) {
      console.error('Failed to list directory:', error)
      throw new Error(`Failed to list directory: ${error}`)
    }
  }

  /**
   * Create a new file or folder
   */
  static async createItem(path: string, name: string, isDirectory: boolean): Promise<FileOperationResult> {
    try {
      const request = new main.CreateFileRequest({
        path,
        name,
        isDirectory
      })
      
      const result = await CreateFile(request)
      return {
        success: result.success,
        message: result.message,
        error: result.error
      }
    } catch (error) {
      console.error('Failed to create item:', error)
      return {
        success: false,
        error: `Failed to create item: ${error}`
      }
    }
  }

  /**
   * Copy items to target location
   */
  static async copyItems(sourcePaths: string[], targetPath: string): Promise<FileOperationResult> {
    try {
      const request = new main.CopyMoveRequest({
        sourcePaths,
        targetPath,
        operation: 'copy'
      })
      
      const result = await CopyItems(request)
      return {
        success: result.success,
        message: result.message,
        error: result.error
      }
    } catch (error) {
      console.error('Failed to copy items:', error)
      return {
        success: false,
        error: `Failed to copy items: ${error}`
      }
    }
  }

  /**
   * Move items to target location
   */
  static async moveItems(sourcePaths: string[], targetPath: string): Promise<FileOperationResult> {
    try {
      const request = new main.CopyMoveRequest({
        sourcePaths,
        targetPath,
        operation: 'move'
      })
      
      const result = await MoveItems(request)
      return {
        success: result.success,
        message: result.message,
        error: result.error
      }
    } catch (error) {
      console.error('Failed to move items:', error)
      return {
        success: false,
        error: `Failed to move items: ${error}`
      }
    }
  }

  /**
   * Delete items
   */
  static async deleteItems(paths: string[]): Promise<FileOperationResult> {
    try {
      const request = new main.DeleteRequest({
        paths
      })
      
      const result = await DeleteItems(request)
      return {
        success: result.success,
        message: result.message,
        error: result.error
      }
    } catch (error) {
      console.error('Failed to delete items:', error)
      return {
        success: false,
        error: `Failed to delete items: ${error}`
      }
    }
  }

  /**
   * Rename an item
   */
  static async renameItem(path: string, newName: string): Promise<FileOperationResult> {
    try {
      const request = new main.RenameRequest({
        path,
        newName
      })
      
      const result = await RenameItem(request)
      return {
        success: result.success,
        message: result.message,
        error: result.error
      }
    } catch (error) {
      console.error('Failed to rename item:', error)
      return {
        success: false,
        error: `Failed to rename item: ${error}`
      }
    }
  }

  /**
   * Open a file with the system default application
   */
  static async openFile(path: string): Promise<FileOperationResult> {
    try {
      const request = new main.OpenFileRequest({
        path
      })
      
      const result = await OpenFile(request)
      return {
        success: result.success,
        message: result.message,
        error: result.error
      }
    } catch (error) {
      console.error('Failed to open file:', error)
      return {
        success: false,
        error: `Failed to open file: ${error}`
      }
    }
  }

  /**
   * Search for files and folders
   */
  static async searchFiles(path: string, query: string): Promise<LocalSearchResult> {
    try {
      const request = new main.SearchRequest({
        path,
        query
      })
      
      const result = await SearchFiles(request)
      
      // Ensure result.items is an array - handle null/undefined case
      const itemsArray = result.items || []
      
      // Convert backend items to frontend format with proper Date conversion
      const items = itemsArray.map(convertToFrontendItem)
      
      return {
        items,
        query: result.query,
        totalCount: result.totalCount
      }
    } catch (error) {
      console.error('Failed to search files:', error)
      return {
        items: [],
        query,
        totalCount: 0
      }
    }
  }

  /**
   * Get common directory paths
   */
  static async getCommonDirectories() {
    try {
      const [home, documents, desktop, downloads] = await Promise.all([
        GetHomeDirectory(),
        GetDocumentsDirectory(),
        GetDesktopDirectory(),
        GetDownloadsDirectory()
      ])
      
      return {
        home,
        documents,
        desktop,
        downloads
      }
    } catch (error) {
      console.error('Failed to get common directories:', error)
      throw new Error(`Failed to get common directories: ${error}`)
    }
  }
}

// Export convenience functions
export const {
  listDirectory,
  createItem,
  copyItems,
  moveItems,
  deleteItems,
  renameItem,
  openFile,
  searchFiles,
  getCommonDirectories
} = FileManagerAPI 