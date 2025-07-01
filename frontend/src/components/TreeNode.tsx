import React, { useState } from 'react'
import { ChevronRight, ChevronDown } from 'lucide-react'
import { FileIcon } from './FileIcon'
import { ContextMenu } from './ContextMenu'
import type { FileSystemItem } from '../types'
import { cn } from '../lib/utils'

interface TreeNodeProps {
  item: FileSystemItem
  level: number
  isSelected: boolean
  selectedItemId?: string | null
  onSelect: (item: FileSystemItem) => void
  onExpand: (item: FileSystemItem) => void
  onContextMenu?: (item: FileSystemItem, event: React.MouseEvent) => void
  onRefresh?: () => void
  clipboardItems?: FileSystemItem[] | null
  onCopy?: (items: FileSystemItem[]) => void
  onCut?: (items: FileSystemItem[]) => void
  isLast?: boolean
  parentPath?: string[]
}

export function TreeNode({
  item,
  level,
  isSelected,
  selectedItemId,
  onSelect,
  onExpand,
  onContextMenu,
  onRefresh,
  clipboardItems,
  onCopy,
  onCut,
  isLast = false,
  parentPath = []
}: TreeNodeProps) {
  const [isHovered, setIsHovered] = useState(false)

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    onSelect(item)
    
    if (item.isDirectory) {
      onExpand(item)
    }
  }

  const handleExpandClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (item.isDirectory) {
      onExpand(item)
    }
  }

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault()
    onContextMenu?.(item, e)
  }

  return (
    <div className="relative">

      <ContextMenu
        item={item}
        onRefresh={onRefresh}
        clipboardItems={clipboardItems}
        onCopy={onCopy}
        onCut={onCut}
      >
        <div
          className={cn(
            "flex items-center gap-1 px-2 py-1 cursor-pointer text-sm relative",
            "hover:bg-accent hover:text-accent-foreground",
            "transition-colors duration-150",
            isSelected && "bg-accent text-accent-foreground",
            isHovered && !isSelected && "bg-muted",
            level > 0 && "border-l-2 border-transparent"
          )}
          style={{ 
            paddingLeft: `${level * 24 + 8}px`,
            marginLeft: level > 0 ? '2px' : 0 
          }}
          onClick={handleClick}
          onContextMenu={handleContextMenu}
          onMouseEnter={() => setIsHovered(true)}
          onMouseLeave={() => setIsHovered(false)}
        >
          {/* Expand/collapse button for directories */}
          {item.isDirectory ? (
            <button
              className="p-0.5 hover:bg-muted rounded-sm flex-shrink-0"
              onClick={handleExpandClick}
            >
              {item.isExpanded ? (
                <ChevronDown className="w-3 h-3 transition-transform duration-200" />
              ) : (
                <ChevronRight className="w-3 h-3 transition-transform duration-200" />
              )}
            </button>
          ) : (
            <div className="w-4 h-4" /> // Placeholder for alignment
          )}

          {/* File icon */}
          <FileIcon
            filename={item.name}
            isDirectory={item.isDirectory}
            isExpanded={item.isExpanded}
            isLoading={item.isLoading}
            className="w-4 h-4 flex-shrink-0"
          />

          {/* File name */}
          <span 
            className="truncate flex-1 select-none"
            title={item.name}
          >
            {item.name}
          </span>

          {/* Children count indicator for collapsed folders */}
          {item.isDirectory && !item.isExpanded && item.children && item.children.length > 0 && (
            <span className="text-xs text-muted-foreground ml-2">
              {item.children.length}
            </span>
          )}
        </div>
      </ContextMenu>

      {/* Child items */}
      {item.isDirectory && item.isExpanded && item.children && item.children.length > 0 && (
        <div className="relative">
          {item.children.map((child, index) => {
            const childIsLast = index === item.children!.length - 1
            const newParentPath = isLast ? parentPath : [...parentPath, level.toString()]
            
            return (
              <TreeNode
                key={child.id}
                item={child}
                level={level + 1}
                isSelected={selectedItemId === child.id}
                selectedItemId={selectedItemId}
                onSelect={onSelect}
                onExpand={onExpand}
                onContextMenu={onContextMenu}
                onRefresh={onRefresh}
                clipboardItems={clipboardItems}
                onCopy={onCopy}
                onCut={onCut}
                isLast={childIsLast}
                parentPath={newParentPath}
              />
            )
          })}
        </div>
      )}
    </div>
  )
} 