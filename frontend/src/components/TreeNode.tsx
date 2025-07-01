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
  onSelect: (item: FileSystemItem) => void
  onExpand: (item: FileSystemItem) => void
  onContextMenu?: (item: FileSystemItem, event: React.MouseEvent) => void
  onRefresh?: () => void
  clipboardItems?: FileSystemItem[] | null
  onCopy?: (items: FileSystemItem[]) => void
  onCut?: (items: FileSystemItem[]) => void
}

export function TreeNode({
  item,
  level,
  isSelected,
  onSelect,
  onExpand,
  onContextMenu,
  onRefresh,
  clipboardItems,
  onCopy,
  onCut
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
    <div>
      <ContextMenu
        item={item}
        onRefresh={onRefresh}
        clipboardItems={clipboardItems}
        onCopy={onCopy}
        onCut={onCut}
      >
        <div
          className={cn(
            "flex items-center gap-1 px-2 py-1 cursor-pointer text-sm",
            "hover:bg-accent hover:text-accent-foreground",
            "transition-colors duration-150",
            isSelected && "bg-accent text-accent-foreground",
            isHovered && !isSelected && "bg-muted"
          )}
          style={{ paddingLeft: `${level * 16 + 8}px` }}
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
                <ChevronDown className="w-3 h-3" />
              ) : (
                <ChevronRight className="w-3 h-3" />
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
        </div>
      </ContextMenu>

      {/* Child items */}
      {item.isDirectory && item.isExpanded && item.children && (
        <div>
          {item.children.map((child) => (
            <TreeNode
              key={child.id}
              item={child}
              level={level + 1}
              isSelected={false} // Child selection logic would be handled by parent
              onSelect={onSelect}
              onExpand={onExpand}
              onContextMenu={onContextMenu}
              onRefresh={onRefresh}
              clipboardItems={clipboardItems}
              onCopy={onCopy}
              onCut={onCut}
            />
          ))}
        </div>
      )}
    </div>
  )
} 