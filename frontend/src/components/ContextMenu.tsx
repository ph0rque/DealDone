import React, { useState } from 'react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './ui/dialog'
import { Button } from './ui/button'
import { Input } from './ui/input'
import {
  Copy,
  Scissors,
  Clipboard,
  Trash2,
  Edit3,
  FolderPlus,
  FilePlus,
  Eye,
  Info
} from 'lucide-react'
import { useFileOperations } from '../hooks/useFileOperations'
import { useToast } from '../hooks/use-toast'
import type { FileSystemItem, ContextMenuItem } from '../types'

interface ContextMenuProps {
  item: FileSystemItem
  children: React.ReactNode
  onRefresh?: () => void
  clipboardItems?: FileSystemItem[] | null
  onCopy?: (items: FileSystemItem[]) => void
  onCut?: (items: FileSystemItem[]) => void
}

export function ContextMenu({ 
  item, 
  children, 
  onRefresh,
  clipboardItems,
  onCopy,
  onCut
}: ContextMenuProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [renameDialogOpen, setRenameDialogOpen] = useState(false)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [createType, setCreateType] = useState<'file' | 'folder'>('file')
  const [renameValue, setRenameValue] = useState(item.name)
  const [createValue, setCreateValue] = useState('')
  
  const fileOps = useFileOperations()
  const { toast } = useToast()

  const handleAction = async (action: string) => {
    try {
      let result
      
      switch (action) {
        case 'open':
          result = await fileOps.openItem(item)
          if (result.success) {
            toast({
              title: "Success",
              description: result.message || `Opened ${item.name}`
            })
          } else {
            toast({
              title: "Error",
              description: result.error || "Failed to open item",
              variant: "destructive"
            })
          }
          break
          
        case 'copy':
          onCopy?.([item])
          toast({
            title: "Copied",
            description: `${item.name} copied to clipboard`
          })
          break
          
        case 'cut':
          onCut?.([item])
          toast({
            title: "Cut",
            description: `${item.name} cut to clipboard`
          })
          break
          
        case 'paste':
          if (clipboardItems && clipboardItems.length > 0) {
            // Determine if this was a cut or copy operation
            // For now, we'll assume copy - this could be enhanced
            result = await fileOps.copyItems(clipboardItems, item.path)
            if (result.success) {
              toast({
                title: "Success",
                description: `Pasted ${clipboardItems.length} item(s)`
              })
              onRefresh?.()
            } else {
              toast({
                title: "Error",
                description: result.error || "Failed to paste items",
                variant: "destructive"
              })
            }
          }
          break
          
        case 'delete':
          setDeleteDialogOpen(true)
          break
          
        case 'rename':
          setRenameValue(item.name)
          setRenameDialogOpen(true)
          break
          
        case 'new_file':
          setCreateType('file')
          setCreateValue('')
          setCreateDialogOpen(true)
          break
          
        case 'new_folder':
          setCreateType('folder')
          setCreateValue('')
          setCreateDialogOpen(true)
          break
      }
    } catch (error) {
      console.error('Context menu action error:', error)
      toast({
        title: "Error",
        description: "An unexpected error occurred",
        variant: "destructive"
      })
    }
  }

  const handleDelete = async () => {
    const result = await fileOps.deleteItems([item])
    if (result.success) {
      toast({
        title: "Success",
        description: `${item.name} deleted successfully`
      })
      onRefresh?.()
    } else {
      toast({
        title: "Error",
        description: result.error || "Failed to delete item",
        variant: "destructive"
      })
    }
    setDeleteDialogOpen(false)
  }

  const handleRename = async () => {
    if (renameValue.trim() && renameValue !== item.name) {
      const result = await fileOps.renameItem(item, renameValue.trim())
      if (result.success) {
        toast({
          title: "Success",
          description: `Renamed to ${renameValue}`
        })
        onRefresh?.()
      } else {
        toast({
          title: "Error",
          description: result.error || "Failed to rename item",
          variant: "destructive"
        })
      }
    }
    setRenameDialogOpen(false)
  }

  const handleCreate = async () => {
    if (createValue.trim()) {
      const result = createType === 'file' 
        ? await fileOps.createFile(item.path, createValue.trim())
        : await fileOps.createFolder(item.path, createValue.trim())
      
      if (result.success) {
        toast({
          title: "Success",
          description: `${createType === 'file' ? 'File' : 'Folder'} created successfully`
        })
        onRefresh?.()
      } else {
        toast({
          title: "Error",
          description: result.error || `Failed to create ${createType}`,
          variant: "destructive"
        })
      }
    }
    setCreateDialogOpen(false)
  }
  const menuItems: ContextMenuItem[] = [
    {
      id: 'open',
      label: 'Open',
      icon: 'Eye',
      action: () => handleAction('open')
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
      action: () => handleAction('copy')
    },
    {
      id: 'cut',
      label: 'Cut',
      icon: 'Scissors',
      action: () => handleAction('cut')
    },
    {
      id: 'paste',
      label: 'Paste',
      icon: 'Clipboard',
      action: () => handleAction('paste'),
      disabled: !clipboardItems || clipboardItems.length === 0
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
      action: () => handleAction('rename')
    },
    {
      id: 'delete',
      label: 'Delete',
      icon: 'Trash2',
      action: () => handleAction('delete')
    }
  ]

  // Add folder-specific items
  if (item.isDirectory) {
    menuItems.splice(5, 0, {
      id: 'separator_new',
      label: '',
      separator: true,
      action: () => {}
    }, {
      id: 'new_folder',
      label: 'New Folder',
      icon: 'FolderPlus',
      action: () => handleAction('new_folder')
    }, {
      id: 'new_file',
      label: 'New File',
      icon: 'FilePlus',
      action: () => handleAction('new_file')
    })
  }

  const getIcon = (iconName: string) => {
    const iconProps = { className: "w-4 h-4 mr-2" }
    switch (iconName) {
      case 'Eye': return <Eye {...iconProps} />
      case 'Copy': return <Copy {...iconProps} />
      case 'Scissors': return <Scissors {...iconProps} />
      case 'Clipboard': return <Clipboard {...iconProps} />
      case 'Edit3': return <Edit3 {...iconProps} />
      case 'Trash2': return <Trash2 {...iconProps} />
      case 'FolderPlus': return <FolderPlus {...iconProps} />
      case 'FilePlus': return <FilePlus {...iconProps} />
      case 'Info': return <Info {...iconProps} />
      default: return null
    }
  }

  return (
    <>
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

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete {item.name}?</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete "{item.name}"? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Rename Dialog */}
      <Dialog open={renameDialogOpen} onOpenChange={setRenameDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Rename {item.name}</DialogTitle>
            <DialogDescription>
              Enter a new name for this {item.isDirectory ? 'folder' : 'file'}:
            </DialogDescription>
          </DialogHeader>
          <Input
            value={renameValue}
            onChange={(e) => setRenameValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                handleRename()
              } else if (e.key === 'Escape') {
                setRenameDialogOpen(false)
              }
            }}
            placeholder="Enter new name"
            autoFocus
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setRenameDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleRename}>
              Rename
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create File/Folder Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New {createType === 'file' ? 'File' : 'Folder'}</DialogTitle>
            <DialogDescription>
              Enter a name for the new {createType}:
            </DialogDescription>
          </DialogHeader>
          <Input
            value={createValue}
            onChange={(e) => setCreateValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                handleCreate()
              } else if (e.key === 'Escape') {
                setCreateDialogOpen(false)
              }
            }}
            placeholder={`Enter ${createType} name`}
            autoFocus
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate}>
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
} 