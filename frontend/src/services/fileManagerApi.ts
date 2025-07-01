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
      
      // Convert backend items to frontend format with Date objects
      return items.map(item => ({
        ...item,
        modifiedAt: new Date(item.modifiedAt),
        createdAt: new Date(item.createdAt),
        isExpanded: false,
        isLoading: false
      }))
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
      
      // Convert backend items to frontend format
      const items = result.items.map(item => ({
        ...item,
        modifiedAt: new Date(item.modifiedAt),
        createdAt: new Date(item.createdAt),
        isExpanded: false,
        isLoading: false
      }))
      
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