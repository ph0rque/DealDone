import React from 'react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import {
  Copy,
  Cut,
  Paste,
  Trash2,
  Edit3,
  FolderPlus,
  FilePlus,
  Eye,
  Info
} from 'lucide-react'
import type { FileSystemItem, ContextMenuItem } from '../types'

interface ContextMenuProps {
  item: FileSystemItem
  children: React.ReactNode
  onAction: (action: string, item: FileSystemItem) => void
  canPaste?: boolean
}

export function ContextMenu({ item, children, onAction, canPaste = false }: ContextMenuProps) {
  const menuItems: ContextMenuItem[] = [
    {
      id: 'open',
      label: 'Open',
      icon: 'Eye',
      action: () => onAction('open', item)
    },
    {
      id: 'separator1',
      label: '',
      separator: true,
      action: () => {}
    },
    {
      id: 'copy',
      label: 'Copy',
      icon: 'Copy',
      action: () => onAction('copy', item)
    },
    {
      id: 'cut',
      label: 'Cut',
      icon: 'Cut',
      action: () => onAction('cut', item)
    },
    {
      id: 'paste',
      label: 'Paste',
      icon: 'Paste',
      action: () => onAction('paste', item),
      disabled: !canPaste
    },
    {
      id: 'separator2',
      label: '',
      separator: true,
      action: () => {}
    },
    {
      id: 'rename',
      label: 'Rename',
      icon: 'Edit3',
      action: () => onAction('rename', item)
    },
    {
      id: 'delete',
      label: 'Delete',
      icon: 'Trash2',
      action: () => onAction('delete', item)
    }
  ]

  // Add folder-specific items
  if (item.isDirectory) {
    menuItems.splice(3, 0, {
      id: 'separator_new',
      label: '',
      separator: true,
      action: () => {}
    }, {
      id: 'new_folder',
      label: 'New Folder',
      icon: 'FolderPlus',
      action: () => onAction('new_folder', item)
    }, {
      id: 'new_file',
      label: 'New File',
      icon: 'FilePlus',
      action: () => onAction('new_file', item)
    })
  }

  const getIcon = (iconName: string) => {
    const iconProps = { className: "w-4 h-4 mr-2" }
    switch (iconName) {
      case 'Eye': return <Eye {...iconProps} />
      case 'Copy': return <Copy {...iconProps} />
      case 'Cut': return <Cut {...iconProps} />
      case 'Paste': return <Paste {...iconProps} />
      case 'Edit3': return <Edit3 {...iconProps} />
      case 'Trash2': return <Trash2 {...iconProps} />
      case 'FolderPlus': return <FolderPlus {...iconProps} />
      case 'FilePlus': return <FilePlus {...iconProps} />
      case 'Info': return <Info {...iconProps} />
      default: return null
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        {children}
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-48">
        {menuItems.map((menuItem) => {
          if (menuItem.separator) {
            return <DropdownMenuSeparator key={menuItem.id} />
          }
          
          return (
            <DropdownMenuItem
              key={menuItem.id}
              onClick={menuItem.action}
              disabled={menuItem.disabled}
              className="cursor-pointer"
            >
              {menuItem.icon && getIcon(menuItem.icon)}
              {menuItem.label}
            </DropdownMenuItem>
          )
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  )
} 