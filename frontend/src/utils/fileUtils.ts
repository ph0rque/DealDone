import type { FileType } from '../types'

// File type detection based on extension
export function getFileType(filename: string, isDirectory: boolean): FileType {
  if (isDirectory) return 'folder'
  
  const extension = filename.split('.').pop()?.toLowerCase()
  if (!extension) return 'unknown'
  
  // Document types
  const documentExtensions = ['txt', 'doc', 'docx', 'pdf', 'rtf', 'odt', 'md', 'html', 'htm']
  if (documentExtensions.includes(extension)) return 'document'
  
  // Image types
  const imageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'webp', 'ico', 'tiff', 'tif']
  if (imageExtensions.includes(extension)) return 'image'
  
  // Video types
  const videoExtensions = ['mp4', 'avi', 'mkv', 'mov', 'wmv', 'flv', 'webm', 'm4v', '3gp']
  if (videoExtensions.includes(extension)) return 'video'
  
  // Audio types
  const audioExtensions = ['mp3', 'wav', 'flac', 'aac', 'ogg', 'wma', 'm4a', 'opus']
  if (audioExtensions.includes(extension)) return 'audio'
  
  // Archive types
  const archiveExtensions = ['zip', 'rar', '7z', 'tar', 'gz', 'bz2', 'xz', 'dmg', 'iso']
  if (archiveExtensions.includes(extension)) return 'archive'
  
  // Code types
  const codeExtensions = [
    'js', 'ts', 'jsx', 'tsx', 'py', 'java', 'cpp', 'c', 'h', 'cs', 'php', 'rb', 'go', 
    'rs', 'swift', 'kt', 'scala', 'r', 'sql', 'json', 'xml', 'yaml', 'yml', 'toml',
    'css', 'scss', 'sass', 'less', 'vue', 'svelte', 'dart', 'sh', 'bash', 'ps1'
  ]
  if (codeExtensions.includes(extension)) return 'code'
  
  return 'unknown'
}

// Format file size
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`
}

// Format date
export function formatDate(date: Date): string {
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

// Check if item is hidden (starts with dot)
export function isHiddenFile(name: string): boolean {
  return name.startsWith('.')
}

// Get parent directory path
export function getParentPath(path: string): string {
  const parts = path.split('/').filter(p => p.length > 0)
  if (parts.length <= 1) return '/'
  return '/' + parts.slice(0, -1).join('/')
}

// Get filename without extension
export function getNameWithoutExtension(filename: string): string {
  const lastDotIndex = filename.lastIndexOf('.')
  return lastDotIndex > 0 ? filename.substring(0, lastDotIndex) : filename
} 