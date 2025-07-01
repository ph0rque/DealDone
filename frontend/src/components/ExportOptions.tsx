import React, { useState } from 'react';
import { Download, FileSpreadsheet, FileText, BarChart, Image, Settings, CheckCircle, Clock, X, File } from 'lucide-react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { useToast } from '../hooks/use-toast';

interface ExportFormat {
  id: string;
  name: string;
  description: string;
  icon: React.ReactNode;
  fileExtension: string;
  supportsCustomization: boolean;
  estimatedSize: string;
}

interface ExportOptions {
  format: string;
  includeRawData: boolean;
  includeAnalysis: boolean;
  includeVisualization: boolean;
  includeMetadata: boolean;
  dateRange: {
    start?: Date;
    end?: Date;
  };
  customFields: string[];
  fileName: string;
  compressionLevel: 'none' | 'standard' | 'maximum';
}

interface ExportOptionsProps {
  dealName: string;
  availableData: {
    documents: number;
    analyses: number;
    templates: number;
    financials: number;
  };
  onExport: (options: ExportOptions) => void;
  onClose: () => void;
}

export function ExportOptions({ dealName, availableData, onExport, onClose }: ExportOptionsProps) {
  const [exportOptions, setExportOptions] = useState<ExportOptions>({
    format: 'excel',
    includeRawData: true,
    includeAnalysis: true,
    includeVisualization: false,
    includeMetadata: true,
    dateRange: {},
    customFields: [],
    fileName: `${dealName}_analysis_${new Date().toISOString().split('T')[0]}`,
    compressionLevel: 'standard'
  });

  const [isExporting, setIsExporting] = useState(false);
  const [exportProgress, setExportProgress] = useState(0);
  const { toast } = useToast();

  const exportFormats: ExportFormat[] = [
    {
      id: 'excel',
      name: 'Excel Workbook',
      description: 'Complete analysis with multiple sheets, formulas, and charts',
      icon: <FileSpreadsheet className="h-5 w-5 text-green-600" />,
      fileExtension: '.xlsx',
      supportsCustomization: true,
      estimatedSize: '2-5 MB'
    },
    {
      id: 'csv',
      name: 'CSV Files',
      description: 'Raw data in comma-separated format, suitable for analysis tools',
      icon: <File className="h-5 w-5 text-blue-600" />,
      fileExtension: '.csv',
      supportsCustomization: false,
      estimatedSize: '500 KB - 2 MB'
    },
    {
      id: 'pdf',
      name: 'PDF Report',
      description: 'Professional report with analysis, charts, and insights',
      icon: <FileText className="h-5 w-5 text-red-600" />,
      fileExtension: '.pdf',
      supportsCustomization: true,
      estimatedSize: '5-15 MB'
    },
    {
      id: 'powerpoint',
      name: 'PowerPoint Presentation',
      description: 'Executive summary slides with key findings and visualizations',
      icon: <BarChart className="h-5 w-5 text-orange-600" />,
      fileExtension: '.pptx',
      supportsCustomization: true,
      estimatedSize: '10-25 MB'
    },
    {
      id: 'json',
      name: 'JSON Data',
      description: 'Structured data format for API integration and custom applications',
      icon: <Settings className="h-5 w-5 text-purple-600" />,
      fileExtension: '.json',
      supportsCustomization: false,
      estimatedSize: '100-500 KB'
    }
  ];

  const handleExport = async () => {
    setIsExporting(true);
    setExportProgress(0);

    try {
      // Simulate export progress
      const progressSteps = [10, 25, 50, 75, 90, 100];
      for (const step of progressSteps) {
        await new Promise(resolve => setTimeout(resolve, 300));
        setExportProgress(step);
      }

      onExport(exportOptions);
      
      toast({
        title: "Export Complete",
        description: `${exportOptions.fileName}${getSelectedFormat()?.fileExtension} has been generated`,
      });
      
    } catch (error) {
      toast({
        title: "Export Failed",
        description: "An error occurred during export",
        variant: "destructive",
      });
    } finally {
      setIsExporting(false);
      setExportProgress(0);
    }
  };

  const getSelectedFormat = () => exportFormats.find(f => f.id === exportOptions.format);

  const updateOption = <K extends keyof ExportOptions>(key: K, value: ExportOptions[K]) => {
    setExportOptions(prev => ({ ...prev, [key]: value }));
  };

  const toggleCustomField = (field: string) => {
    setExportOptions(prev => ({
      ...prev,
      customFields: prev.customFields.includes(field)
        ? prev.customFields.filter(f => f !== field)
        : [...prev.customFields, field]
    }));
  };

  const getEstimatedSize = () => {
    const baseSize = getSelectedFormat()?.estimatedSize || '1-5 MB';
    const multiplier = 
      (exportOptions.includeRawData ? 1 : 0) +
      (exportOptions.includeAnalysis ? 1 : 0) +
      (exportOptions.includeVisualization ? 2 : 0) +
      (exportOptions.includeMetadata ? 0.5 : 0);
    
    return multiplier > 2 ? 'Large (10+ MB)' : multiplier > 1 ? 'Medium (2-10 MB)' : 'Small (<2 MB)';
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 w-full h-full max-w-4xl max-h-[90vh] rounded-lg shadow-xl flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div>
            <h2 className="text-2xl font-bold">Export Analysis</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              Deal: {dealName} • {availableData.documents} docs, {availableData.analyses} analyses
            </p>
          </div>
          <Button variant="ghost" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {isExporting ? (
          /* Export Progress */
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center max-w-md">
              <div className="relative w-32 h-32 mx-auto mb-6">
                <div className="absolute inset-0 rounded-full border-8 border-gray-200 dark:border-gray-700"></div>
                <div 
                  className="absolute inset-0 rounded-full border-8 border-blue-600 border-t-transparent transition-all duration-300"
                  style={{ 
                    transform: `rotate(${(exportProgress / 100) * 360}deg)`,
                    borderTopColor: exportProgress === 100 ? '#10b981' : 'transparent'
                  }}
                ></div>
                <div className="absolute inset-0 flex items-center justify-center">
                  {exportProgress === 100 ? (
                    <CheckCircle className="h-8 w-8 text-green-600" />
                  ) : (
                    <span className="text-xl font-bold">{exportProgress}%</span>
                  )}
                </div>
              </div>
              
              <h3 className="text-lg font-semibold mb-2">
                {exportProgress === 100 ? 'Export Complete!' : 'Generating Export...'}
              </h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                {exportProgress < 30 && 'Preparing data...'}
                {exportProgress >= 30 && exportProgress < 60 && 'Processing analysis...'}
                {exportProgress >= 60 && exportProgress < 90 && 'Generating format...'}
                {exportProgress >= 90 && exportProgress < 100 && 'Finalizing export...'}
                {exportProgress === 100 && 'Your file is ready for download'}
              </p>
            </div>
          </div>
        ) : (
          <div className="flex-1 overflow-y-auto p-6">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Format Selection */}
              <div className="lg:col-span-2">
                <h3 className="text-lg font-semibold mb-4">Export Format</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                  {exportFormats.map((format) => (
                    <div
                      key={format.id}
                      className={`border rounded-lg p-4 cursor-pointer transition-all ${
                        exportOptions.format === format.id
                          ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                          : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
                      }`}
                      onClick={() => updateOption('format', format.id)}
                    >
                      <div className="flex items-start space-x-3">
                        <div className="flex-shrink-0">{format.icon}</div>
                        <div className="flex-1">
                          <h4 className="font-medium">{format.name}</h4>
                          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                            {format.description}
                          </p>
                          <div className="flex items-center justify-between mt-2">
                            <span className="text-xs text-gray-400">{format.estimatedSize}</span>
                            {exportOptions.format === format.id && (
                              <CheckCircle className="h-4 w-4 text-blue-500" />
                            )}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>

                {/* Content Options */}
                <h3 className="text-lg font-semibold mb-4">What to Include</h3>
                <div className="space-y-3 mb-6">
                  <label className="flex items-center space-x-3">
                    <input
                      type="checkbox"
                      checked={exportOptions.includeRawData}
                      onChange={(e) => updateOption('includeRawData', e.target.checked)}
                      className="w-4 h-4 text-blue-600 rounded"
                    />
                    <div>
                      <span className="font-medium">Raw Document Data</span>
                      <p className="text-sm text-gray-500">Original extracted data from documents</p>
                    </div>
                  </label>

                  <label className="flex items-center space-x-3">
                    <input
                      type="checkbox"
                      checked={exportOptions.includeAnalysis}
                      onChange={(e) => updateOption('includeAnalysis', e.target.checked)}
                      className="w-4 h-4 text-blue-600 rounded"
                    />
                    <div>
                      <span className="font-medium">AI Analysis Results</span>
                      <p className="text-sm text-gray-500">Financial analysis, risk assessment, insights</p>
                    </div>
                  </label>

                  <label className="flex items-center space-x-3">
                    <input
                      type="checkbox"
                      checked={exportOptions.includeVisualization}
                      onChange={(e) => updateOption('includeVisualization', e.target.checked)}
                      className="w-4 h-4 text-blue-600 rounded"
                    />
                    <div>
                      <span className="font-medium">Charts & Visualizations</span>
                      <p className="text-sm text-gray-500">Graphs, charts, and visual analysis</p>
                    </div>
                  </label>

                  <label className="flex items-center space-x-3">
                    <input
                      type="checkbox"
                      checked={exportOptions.includeMetadata}
                      onChange={(e) => updateOption('includeMetadata', e.target.checked)}
                      className="w-4 h-4 text-blue-600 rounded"
                    />
                    <div>
                      <span className="font-medium">Metadata & Sources</span>
                      <p className="text-sm text-gray-500">Confidence scores, data sources, timestamps</p>
                    </div>
                  </label>
                </div>

                {/* Custom Fields */}
                {getSelectedFormat()?.supportsCustomization && (
                  <>
                    <h3 className="text-lg font-semibold mb-4">Additional Options</h3>
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium mb-2">Custom Fields</label>
                        <div className="grid grid-cols-2 gap-2">
                          {['Valuation Metrics', 'Risk Scores', 'Competitive Analysis', 'Market Data', 'Legal Issues', 'Integration Notes'].map((field) => (
                            <label key={field} className="flex items-center space-x-2">
                              <input
                                type="checkbox"
                                checked={exportOptions.customFields.includes(field)}
                                onChange={() => toggleCustomField(field)}
                                className="w-4 h-4 text-blue-600 rounded"
                              />
                              <span className="text-sm">{field}</span>
                            </label>
                          ))}
                        </div>
                      </div>

                      <div>
                        <label className="block text-sm font-medium mb-2">Compression Level</label>
                        <select
                          value={exportOptions.compressionLevel}
                          onChange={(e) => updateOption('compressionLevel', e.target.value as any)}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800"
                        >
                          <option value="none">None (Fastest)</option>
                          <option value="standard">Standard (Balanced)</option>
                          <option value="maximum">Maximum (Smallest File)</option>
                        </select>
                      </div>
                    </div>
                  </>
                )}
              </div>

              {/* Summary Panel */}
              <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                <h3 className="font-semibold mb-4">Export Summary</h3>
                
                <div className="space-y-3 text-sm">
                  <div className="flex justify-between">
                    <span>Format:</span>
                    <span className="font-medium">{getSelectedFormat()?.name}</span>
                  </div>
                  
                  <div className="flex justify-between">
                    <span>Estimated Size:</span>
                    <span className="font-medium">{getEstimatedSize()}</span>
                  </div>
                  
                  <div className="flex justify-between">
                    <span>Items Included:</span>
                    <span className="font-medium">
                      {[
                        exportOptions.includeRawData && 'Data',
                        exportOptions.includeAnalysis && 'Analysis',
                        exportOptions.includeVisualization && 'Charts',
                        exportOptions.includeMetadata && 'Metadata'
                      ].filter(Boolean).join(', ')}
                    </span>
                  </div>
                  
                  <div className="pt-3 border-t border-gray-200 dark:border-gray-700">
                    <label className="block text-sm font-medium mb-2">File Name</label>
                    <Input
                      value={exportOptions.fileName}
                      onChange={(e) => updateOption('fileName', e.target.value)}
                      placeholder="Enter file name"
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      {exportOptions.fileName}{getSelectedFormat()?.fileExtension}
                    </p>
                  </div>
                </div>

                <div className="mt-6 space-y-2">
                  <Button 
                    className="w-full" 
                    onClick={handleExport}
                    disabled={!exportOptions.fileName.trim()}
                  >
                    <Download className="h-4 w-4 mr-2" />
                    Generate Export
                  </Button>
                  
                  <Button variant="outline" className="w-full" onClick={onClose}>
                    Cancel
                  </Button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
} 