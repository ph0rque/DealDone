import React from 'react'
import { FileTree } from './FileTree'
import { SearchBar } from './SearchBar'
import { Toaster } from './ui/toaster'

export function FileManagerLayout() {
  return (
    <div className="flex flex-col h-screen bg-background text-foreground">
      {/* Header with search */}
      <header className="border-b border-border p-4">
        <div className="max-w-md">
          <SearchBar />
        </div>
      </header>

      {/* Main content area */}
      <main className="flex-1 flex overflow-hidden">
        {/* File tree panel */}
        <div className="w-80 border-r border-border bg-card">
          <div className="h-full overflow-auto">
            <FileTree />
          </div>
        </div>

        {/* Content area (placeholder for future features) */}
        <div className="flex-1 flex items-center justify-center bg-muted/30">
          <div className="text-center space-y-2">
            <h2 className="text-xl font-medium text-muted-foreground">
              File Manager
            </h2>
            <p className="text-sm text-muted-foreground">
              Use the file tree on the left to navigate your files and folders
            </p>
          </div>
        </div>
      </main>

      {/* Toast notifications */}
      <Toaster />
    </div>
  )
} 