import React, { useState, useEffect } from 'react';
import { FileSpreadsheet, Search, Grid, List, Filter, Download, Upload, Check, Star, Clock } from 'lucide-react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { useToast } from '../hooks/use-toast';

interface Template {
  id: string;
  name: string;
  description: string;
  category: 'financial' | 'legal' | 'operational' | 'custom';
  type: 'excel' | 'csv' | 'word' | 'powerpoint';
  size: number;
  fields: number;
  lastModified: Date;
  isDefault: boolean;
  previewUrl?: string;
  tags: string[];
  compatibility: number; // 0-1 score for how well it matches current deal
}

interface TemplateSelectorProps {
  onTemplateSelect: (template: Template) => void;
  onClose: () => void;
  dealName?: string;
  documentsAvailable?: number;
}

export function TemplateSelector({ onTemplateSelect, onClose, dealName, documentsAvailable = 0 }: TemplateSelectorProps) {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [filteredTemplates, setFilteredTemplates] = useState<Template[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [sortBy, setSortBy] = useState<'name' | 'compatibility' | 'modified'>('compatibility');
  const [isLoading, setIsLoading] = useState(true);
  const { toast } = useToast();

  useEffect(() => {
    loadTemplates();
  }, []);

  useEffect(() => {
    filterAndSortTemplates();
  }, [templates, searchTerm, selectedCategory, sortBy]);

  const loadTemplates = async () => {
    try {
      setIsLoading(true);
      
      // Sample template data
      const sampleTemplates: Template[] = [
        {
          id: '1',
          name: 'Financial Model Template',
          description: 'Comprehensive financial model with DCF, comparables, and scenario analysis',
          category: 'financial',
          type: 'excel',
          size: 2456789,
          fields: 45,
          lastModified: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000),
          isDefault: true,
          tags: ['dcf', 'valuation', 'financial-model'],
          compatibility: 0.95
        },
        {
          id: '2',
          name: 'Due Diligence Checklist',
          description: 'Complete due diligence checklist for M&A transactions',
          category: 'legal',
          type: 'excel',
          size: 987654,
          fields: 78,
          lastModified: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000),
          isDefault: true,
          tags: ['due-diligence', 'legal', 'checklist'],
          compatibility: 0.88
        },
        {
          id: '3',
          name: 'Deal Summary Template',
          description: 'Executive summary template for deal presentations',
          category: 'operational',
          type: 'word',
          size: 234567,
          fields: 25,
          lastModified: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000),
          isDefault: true,
          tags: ['summary', 'presentation', 'executive'],
          compatibility: 0.92
        }
      ];

      setTemplates(sampleTemplates);
    } catch (error) {
      console.error('Error loading templates:', error);
      toast({
        title: "Error",
        description: "Failed to load templates",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const filterAndSortTemplates = () => {
    let filtered = templates.filter(template => {
      const matchesSearch = searchTerm === '' || 
        template.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        template.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        template.tags.some(tag => tag.toLowerCase().includes(searchTerm.toLowerCase()));

      const matchesCategory = selectedCategory === 'all' || template.category === selectedCategory;

      return matchesSearch && matchesCategory;
    });

    // Sort templates
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'compatibility':
          return b.compatibility - a.compatibility;
        case 'modified':
          return b.lastModified.getTime() - a.lastModified.getTime();
        default:
          return 0;
      }
    });

    setFilteredTemplates(filtered);
  };

  const handleTemplateSelect = (template: Template) => {
    onTemplateSelect(template);
    toast({
      title: "Template Selected",
      description: `"${template.name}" will be used for data population`,
    });
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getCategoryColor = (category: string) => {
    switch (category) {
      case 'financial':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
      case 'legal':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
      case 'operational':
        return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200';
      case 'custom':
        return 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200';
    }
  };

  if (isLoading) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-white dark:bg-gray-900 rounded-lg p-8">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="text-center mt-4">Loading templates...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 w-full h-full max-w-6xl max-h-[90vh] rounded-lg shadow-xl flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div>
            <h2 className="text-2xl font-bold">Select Template</h2>
            {dealName && (
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                For deal: {dealName} • {documentsAvailable} documents available
              </p>
            )}
          </div>
          <Button variant="ghost" onClick={onClose}>
            ✕
          </Button>
        </div>

        {/* Controls */}
        <div className="p-6 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
          <div className="flex items-center space-x-4 mb-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                type="text"
                placeholder="Search templates..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
            
            <select
              value={selectedCategory}
              onChange={(e) => setSelectedCategory(e.target.value)}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800"
            >
              <option value="all">All Categories</option>
              <option value="financial">Financial</option>
              <option value="legal">Legal</option>
              <option value="operational">Operational</option>
              <option value="custom">Custom</option>
            </select>

            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as any)}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800"
            >
              <option value="compatibility">Best Match</option>
              <option value="name">Name</option>
              <option value="modified">Recently Modified</option>
            </select>
          </div>
        </div>

        {/* Template Grid */}
        <div className="flex-1 overflow-y-auto p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredTemplates.map((template) => (
              <div
                key={template.id}
                className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:shadow-lg transition-shadow cursor-pointer"
                onClick={() => handleTemplateSelect(template)}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center space-x-2">
                    <FileSpreadsheet className="h-5 w-5 text-green-600" />
                    {template.isDefault && (
                      <Star className="h-4 w-4 text-yellow-500 fill-current" />
                    )}
                  </div>
                  <div className="flex items-center space-x-1">
                    <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                    <span className="text-xs text-gray-500">
                      {Math.round(template.compatibility * 100)}% match
                    </span>
                  </div>
                </div>
                
                <h3 className="font-semibold text-lg mb-2">{template.name}</h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
                  {template.description}
                </p>
                
                <div className="space-y-2 mb-4">
                  <div className="flex items-center justify-between text-xs text-gray-500">
                    <span>{template.fields} fields</span>
                    <span>{formatFileSize(template.size)}</span>
                  </div>
                  <div className="flex items-center text-xs text-gray-500">
                    <Clock className="h-3 w-3 mr-1" />
                    Modified {template.lastModified.toLocaleDateString()}
                  </div>
                </div>
                
                <div className="flex items-center justify-between">
                  <span className={`text-xs px-2 py-1 rounded ${getCategoryColor(template.category)}`}>
                    {template.category}
                  </span>
                  <Button size="sm">
                    Select
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
} 