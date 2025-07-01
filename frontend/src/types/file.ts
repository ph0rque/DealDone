// Base file system item interface
export interface FileSystemItem {
  id: string
  name: string
  path: string
  isDirectory: boolean
  size: number
  modifiedAt: Date
  createdAt: Date
  permissions: FilePermissions
  extension?: string
  mimeType?: string
  children?: FileSystemItem[]
  isExpanded?: boolean
  isLoading?: boolean
}

// File permissions interface
export interface FilePermissions {
  readable: boolean
  writable: boolean
  executable: boolean
}

// File operation result
export interface FileOperationResult {
  success: boolean
  message?: string
  error?: string
}

// Search result interface
export interface SearchResult {
  items: FileSystemItem[]
  query: string
  totalCount: number
}

// Tree node state for UI
export interface TreeNodeState {
  isSelected: boolean
  isExpanded: boolean
  isLoading: boolean
  isDragOver: boolean
}

// Context menu item
export interface ContextMenuItem {
  id: string
  label: string
  icon?: string
  action: () => void
  disabled?: boolean
  separator?: boolean
}

// File manager context state
export interface FileManagerState {
  currentPath: string
  selectedItems: string[]
  clipboardItems: {
    items: FileSystemItem[]
    operation: 'copy' | 'cut'
  } | null
  searchQuery: string
  searchResults: SearchResult | null
  isLoading: boolean
  error: string | null
}

// Theme types
export type Theme = 'light' | 'dark' | 'system'

// File type categories for icons
export type FileType = 
  | 'folder'
  | 'document'
  | 'image'
  | 'video'
  | 'audio'
  | 'archive'
  | 'code'
  | 'unknown'

// File operations
export type FileOperation = 
  | 'create'
  | 'copy'
  | 'move'
  | 'delete'
  | 'rename'
  | 'open' 