import React, { useState, useEffect } from 'react';
import { Search, Filter, Calendar, FileText, FolderOpen, Clock, CheckCircle, AlertTriangle, X } from 'lucide-react';
import { Input } from './ui/input';
import { Button } from './ui/button';
import { useToast } from '../hooks/use-toast';

export interface DocumentItem {
  id: string;
  name: string;
  type: 'legal' | 'financial' | 'general';
  status: 'pending' | 'processing' | 'completed' | 'error';
  uploadDate: Date;
  analysisDate?: Date;
  size: number;
  confidence?: number;
  path: string;
  tags?: string[];
}

interface DocumentSearchProps {
  documents: DocumentItem[];
  onDocumentSelect: (document: DocumentItem) => void;
  onDocumentAction: (documentId: string, action: 'move' | 'delete' | 'reprocess') => void;
}

interface SearchFilters {
  type: string[];
  status: string[];
  dateRange: {
    start?: Date;
    end?: Date;
  };
  minConfidence?: number;
  tags: string[];
}

export function DocumentSearch({ documents, onDocumentSelect, onDocumentAction }: DocumentSearchProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [showFilters, setShowFilters] = useState(false);
  const [filters, setFilters] = useState<SearchFilters>({
    type: [],
    status: [],
    dateRange: {},
    tags: []
  });
  const [filteredDocuments, setFilteredDocuments] = useState<DocumentItem[]>(documents);
  const [sortBy, setSortBy] = useState<'name' | 'date' | 'confidence' | 'size'>('date');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const { toast } = useToast();

  useEffect(() => {
    filterAndSortDocuments();
  }, [documents, searchTerm, filters, sortBy, sortOrder]);

  const filterAndSortDocuments = () => {
    let filtered = documents.filter(doc => {
      // Text search
      const matchesSearch = searchTerm === '' || 
        doc.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        doc.tags?.some(tag => tag.toLowerCase().includes(searchTerm.toLowerCase()));

      // Type filter
      const matchesType = filters.type.length === 0 || filters.type.includes(doc.type);

      // Status filter
      const matchesStatus = filters.status.length === 0 || filters.status.includes(doc.status);

      // Date range filter
      const matchesDateRange = (!filters.dateRange.start || doc.uploadDate >= filters.dateRange.start) &&
                              (!filters.dateRange.end || doc.uploadDate <= filters.dateRange.end);

      // Confidence filter
      const matchesConfidence = !filters.minConfidence || 
                               (doc.confidence && doc.confidence >= filters.minConfidence);

      // Tags filter
      const matchesTags = filters.tags.length === 0 || 
                         filters.tags.every(tag => doc.tags?.includes(tag));

      return matchesSearch && matchesType && matchesStatus && matchesDateRange && matchesConfidence && matchesTags;
    });

    // Sort documents
    filtered.sort((a, b) => {
      let comparison = 0;
      
      switch (sortBy) {
        case 'name':
          comparison = a.name.localeCompare(b.name);
          break;
        case 'date':
          comparison = a.uploadDate.getTime() - b.uploadDate.getTime();
          break;
        case 'confidence':
          comparison = (a.confidence || 0) - (b.confidence || 0);
          break;
        case 'size':
          comparison = a.size - b.size;
          break;
      }

      return sortOrder === 'asc' ? comparison : -comparison;
    });

    setFilteredDocuments(filtered);
  };

  const toggleFilter = (filterType: keyof SearchFilters, value: string) => {
    setFilters(prev => {
      if (filterType === 'type' || filterType === 'status' || filterType === 'tags') {
        const currentArray = prev[filterType] as string[];
        const newArray = currentArray.includes(value)
          ? currentArray.filter(item => item !== value)
          : [...currentArray, value];
        
        return { ...prev, [filterType]: newArray };
      }
      return prev;
    });
  };

  const clearFilters = () => {
    setFilters({
      type: [],
      status: [],
      dateRange: {},
      tags: []
    });
    setSearchTerm('');
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'processing':
        return <Clock className="h-4 w-4 text-yellow-500" />;
      case 'error':
        return <AlertTriangle className="h-4 w-4 text-red-500" />;
      default:
        return <Clock className="h-4 w-4 text-gray-400" />;
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'legal':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
      case 'financial':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
      case 'general':
        return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200';
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const activeFilterCount = filters.type.length + filters.status.length + filters.tags.length + 
                            (filters.dateRange.start ? 1 : 0) + (filters.dateRange.end ? 1 : 0) +
                            (filters.minConfidence ? 1 : 0);

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm">
      {/* Search Header */}
      <div className="p-4 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center space-x-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              type="text"
              placeholder="Search documents..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
          
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowFilters(!showFilters)}
            className={activeFilterCount > 0 ? 'bg-blue-50 dark:bg-blue-900/20' : ''}
          >
            <Filter className="h-4 w-4 mr-2" />
            Filters
            {activeFilterCount > 0 && (
              <span className="ml-2 bg-blue-600 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center">
                {activeFilterCount}
              </span>
            )}
          </Button>

          {activeFilterCount > 0 && (
            <Button variant="ghost" size="sm" onClick={clearFilters}>
              <X className="h-4 w-4 mr-1" />
              Clear
            </Button>
          )}
        </div>

        {/* Filter Panel */}
        {showFilters && (
          <div className="mt-4 p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg space-y-4">
            {/* Document Type Filter */}
            <div>
              <label className="text-sm font-medium mb-2 block">Document Type</label>
              <div className="flex space-x-2">
                {['legal', 'financial', 'general'].map(type => (
                  <button
                    key={type}
                    onClick={() => toggleFilter('type', type)}
                    className={`px-3 py-1 text-xs rounded-full transition-colors ${
                      filters.type.includes(type)
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-200 dark:bg-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-500'
                    }`}
                  >
                    {type.charAt(0).toUpperCase() + type.slice(1)}
                  </button>
                ))}
              </div>
            </div>

            {/* Status Filter */}
            <div>
              <label className="text-sm font-medium mb-2 block">Status</label>
              <div className="flex space-x-2">
                {['pending', 'processing', 'completed', 'error'].map(status => (
                  <button
                    key={status}
                    onClick={() => toggleFilter('status', status)}
                    className={`px-3 py-1 text-xs rounded-full transition-colors ${
                      filters.status.includes(status)
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-200 dark:bg-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-500'
                    }`}
                  >
                    {status.charAt(0).toUpperCase() + status.slice(1)}
                  </button>
                ))}
              </div>
            </div>

            {/* Sort Options */}
            <div>
              <label className="text-sm font-medium mb-2 block">Sort By</label>
              <div className="flex space-x-2">
                <select
                  value={sortBy}
                  onChange={(e) => setSortBy(e.target.value as any)}
                  className="text-xs bg-gray-200 dark:bg-gray-600 rounded px-2 py-1"
                >
                  <option value="date">Upload Date</option>
                  <option value="name">Name</option>
                  <option value="confidence">Confidence</option>
                  <option value="size">File Size</option>
                </select>
                <button
                  onClick={() => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')}
                  className="text-xs bg-gray-200 dark:bg-gray-600 rounded px-2 py-1 hover:bg-gray-300 dark:hover:bg-gray-500"
                >
                  {sortOrder === 'asc' ? '↑' : '↓'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Results Header */}
      <div className="px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-700/30">
        <div className="flex items-center justify-between text-sm text-gray-600 dark:text-gray-400">
          <span>{filteredDocuments.length} documents found</span>
          <span>
            Showing {Math.min(filteredDocuments.length, 50)} of {filteredDocuments.length}
          </span>
        </div>
      </div>

      {/* Document List */}
      <div className="max-h-96 overflow-y-auto">
        {filteredDocuments.slice(0, 50).map((doc) => (
          <div
            key={doc.id}
            className="p-4 border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer"
            onClick={() => onDocumentSelect(doc)}
          >
            <div className="flex items-start justify-between">
              <div className="flex items-start space-x-3 flex-1">
                <FileText className="h-5 w-5 text-gray-400 mt-0.5" />
                <div className="flex-1 min-w-0">
                  <h4 className="text-sm font-medium truncate">{doc.name}</h4>
                  <div className="flex items-center space-x-2 mt-1">
                    <span className={`text-xs px-2 py-1 rounded ${getTypeColor(doc.type)}`}>
                      {doc.type}
                    </span>
                    <span className="text-xs text-gray-500">
                      {formatFileSize(doc.size)}
                    </span>
                    <span className="text-xs text-gray-500">
                      {doc.uploadDate.toLocaleDateString()}
                    </span>
                    {doc.confidence && (
                      <span className="text-xs text-gray-500">
                        {Math.round(doc.confidence * 100)}% confidence
                      </span>
                    )}
                  </div>
                  {doc.tags && doc.tags.length > 0 && (
                    <div className="flex space-x-1 mt-2">
                      {doc.tags.slice(0, 3).map((tag, idx) => (
                        <span key={idx} className="text-xs bg-gray-100 dark:bg-gray-600 text-gray-600 dark:text-gray-300 px-1 py-0.5 rounded">
                          {tag}
                        </span>
                      ))}
                      {doc.tags.length > 3 && (
                        <span className="text-xs text-gray-500">+{doc.tags.length - 3} more</span>
                      )}
                    </div>
                  )}
                </div>
              </div>
              <div className="flex items-center space-x-2">
                {getStatusIcon(doc.status)}
                <DocumentActionMenu 
                  documentId={doc.id}
                  onAction={(action) => onDocumentAction(doc.id, action)}
                />
              </div>
            </div>
          </div>
        ))}
        
        {filteredDocuments.length === 0 && (
          <div className="p-8 text-center text-gray-500">
            <FolderOpen className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>No documents found</p>
            <p className="text-sm">Try adjusting your search or filters</p>
          </div>
        )}
      </div>
    </div>
  );
}

// Document Action Menu Component (for subtask 6.8)
interface DocumentActionMenuProps {
  documentId: string;
  onAction: (action: 'move' | 'delete' | 'reprocess') => void;
}

function DocumentActionMenu({ documentId, onAction }: DocumentActionMenuProps) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="relative">
      <button
        onClick={(e) => {
          e.stopPropagation();
          setIsOpen(!isOpen);
        }}
        className="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
      >
        <div className="w-1 h-1 bg-gray-400 rounded-full mb-1"></div>
        <div className="w-1 h-1 bg-gray-400 rounded-full mb-1"></div>
        <div className="w-1 h-1 bg-gray-400 rounded-full"></div>
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute right-0 top-8 z-20 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg py-1 min-w-32">
            <button
              onClick={(e) => {
                e.stopPropagation();
                onAction('move');
                setIsOpen(false);
              }}
              className="w-full text-left px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              Move to...
            </button>
            <button
              onClick={(e) => {
                e.stopPropagation();
                onAction('reprocess');
                setIsOpen(false);
              }}
              className="w-full text-left px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              Reprocess
            </button>
            <button
              onClick={(e) => {
                e.stopPropagation();
                onAction('delete');
                setIsOpen(false);
              }}
              className="w-full text-left px-3 py-2 text-sm text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
            >
              Delete
            </button>
          </div>
        </>
      )}
    </div>
  );
} 