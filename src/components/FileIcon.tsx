import React from 'react'
import {
  Folder,
  FolderOpen,
  File,
  FileText,
  FileImage,
  FileVideo,
  FileAudio,
  Archive,
  FileCode,
  Loader2
} from 'lucide-react'
import type { FileType } from '../types'
import { getFileType } from '../utils/fileUtils'

interface FileIconProps {
  filename: string
  isDirectory: boolean
  isExpanded?: boolean
  isLoading?: boolean
  className?: string
}

export function FileIcon({ 
  filename, 
  isDirectory, 
  isExpanded = false, 
  isLoading = false,
  className = "w-4 h-4"
}: FileIconProps) {
  if (isLoading) {
    return <Loader2 className={`${className} animate-spin`} />
  }

  const fileType = getFileType(filename, isDirectory)
  
  const getIcon = (type: FileType) => {
    switch (type) {
      case 'folder':
        return isExpanded ? 
          <FolderOpen className={className} /> : 
          <Folder className={className} />
      case 'document':
        return <FileText className={className} />
      case 'image':
        return <FileImage className={className} />
      case 'video':
        return <FileVideo className={className} />
      case 'audio':
        return <FileAudio className={className} />
      case 'archive':
        return <Archive className={className} />
      case 'code':
        return <FileCode className={className} />
      default:
        return <File className={className} />
    }
  }

  return getIcon(fileType)
} 