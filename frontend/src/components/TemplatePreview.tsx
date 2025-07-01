import React, { useState, useEffect } from 'react';
import { FileSpreadsheet, Download, Eye, Edit, CheckCircle, AlertTriangle, Info, X, ZoomIn, ZoomOut } from 'lucide-react';
import { Button } from './ui/button';
import { useToast } from '../hooks/use-toast';

interface TemplateField {
  name: string;
  value: any;
  confidence: number;
  source: 'ai' | 'extracted' | 'calculated' | 'default';
  originalValue?: string;
  isFormula?: boolean;
  isRequired?: boolean;
  dataType: 'string' | 'number' | 'currency' | 'date' | 'percentage';
}

interface PopulatedTemplate {
  id: string;
  templateName: string;
  dealName: string;
  populatedDate: Date;
  totalFields: number;
  populatedFields: number;
  averageConfidence: number;
  fields: Record<string, TemplateField>;
  sheets?: Record<string, Record<string, TemplateField>>;
  validationErrors: string[];
  warnings: string[];
}

interface TemplatePreviewProps {
  template: PopulatedTemplate;
  onClose: () => void;
  onDownload?: () => void;
  onEdit?: () => void;
  onApprove?: () => void;
}

export function TemplatePreview({ template, onClose, onDownload, onEdit, onApprove }: TemplatePreviewProps) {
  const [activeSheet, setActiveSheet] = useState<string>('main');
  const [zoomLevel, setZoomLevel] = useState(100);
  const [showValidation, setShowValidation] = useState(true);
  const [selectedField, setSelectedField] = useState<string | null>(null);
  const { toast } = useToast();

  const sheets = template.sheets ? Object.keys(template.sheets) : ['main'];
  const currentFields = template.sheets ? template.sheets[activeSheet] : template.fields;

  const getConfidenceColor = (confidence: number) => {
    if (confidence >= 0.9) return 'text-green-600 bg-green-50 dark:bg-green-900/20';
    if (confidence >= 0.7) return 'text-yellow-600 bg-yellow-50 dark:bg-yellow-900/20';
    return 'text-red-600 bg-red-50 dark:bg-red-900/20';
  };

  const getSourceIcon = (source: string) => {
    switch (source) {
      case 'ai':
        return <div className="w-2 h-2 bg-blue-500 rounded-full" title="AI Analysis" />;
      case 'extracted':
        return <div className="w-2 h-2 bg-green-500 rounded-full" title="Document Extraction" />;
      case 'calculated':
        return <div className="w-2 h-2 bg-purple-500 rounded-full" title="Calculated Value" />;
      case 'default':
        return <div className="w-2 h-2 bg-gray-400 rounded-full" title="Default Value" />;
      default:
        return <div className="w-2 h-2 bg-gray-400 rounded-full" />;
    }
  };

  const formatValue = (field: TemplateField) => {
    if (field.value === null || field.value === undefined) return 'N/A';
    
    switch (field.dataType) {
      case 'currency':
        return new Intl.NumberFormat('en-US', { 
          style: 'currency', 
          currency: 'USD' 
        }).format(Number(field.value));
      case 'percentage':
        return `${(Number(field.value) * 100).toFixed(2)}%`;
      case 'number':
        return Number(field.value).toLocaleString();
      case 'date':
        return new Date(field.value).toLocaleDateString();
      default:
        return String(field.value);
    }
  };

  const completionPercentage = Math.round((template.populatedFields / template.totalFields) * 100);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 w-full h-full max-w-7xl max-h-[95vh] rounded-lg shadow-xl flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center space-x-4">
            <FileSpreadsheet className="h-6 w-6 text-blue-600" />
            <div>
              <h2 className="text-xl font-bold">{template.templateName}</h2>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                Deal: {template.dealName} • Populated {template.populatedDate.toLocaleDateString()}
              </p>
            </div>
          </div>
          
          <div className="flex items-center space-x-2">
            <Button variant="outline" size="sm" onClick={() => setZoomLevel(Math.max(50, zoomLevel - 10))}>
              <ZoomOut className="h-4 w-4" />
            </Button>
            <span className="text-sm text-gray-500">{zoomLevel}%</span>
            <Button variant="outline" size="sm" onClick={() => setZoomLevel(Math.min(150, zoomLevel + 10))}>
              <ZoomIn className="h-4 w-4" />
            </Button>
            
            <div className="w-px h-6 bg-gray-300 dark:bg-gray-600 mx-2" />
            
            {onDownload && (
              <Button variant="outline" size="sm" onClick={onDownload}>
                <Download className="h-4 w-4 mr-2" />
                Download
              </Button>
            )}
            {onEdit && (
              <Button variant="outline" size="sm" onClick={onEdit}>
                <Edit className="h-4 w-4 mr-2" />
                Edit
              </Button>
            )}
            {onApprove && (
              <Button size="sm" onClick={onApprove}>
                <CheckCircle className="h-4 w-4 mr-2" />
                Approve
              </Button>
            )}
            <Button variant="ghost" size="sm" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Stats Bar */}
        <div className="px-6 py-4 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-6">
              <div className="flex items-center space-x-2">
                <span className="text-sm font-medium">Completion:</span>
                <div className="flex items-center space-x-2">
                  <div className="w-24 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                    <div 
                      className="bg-blue-600 h-2 rounded-full" 
                      style={{ width: `${completionPercentage}%` }}
                    />
                  </div>
                  <span className="text-sm font-medium">{completionPercentage}%</span>
                </div>
              </div>
              
              <div className="flex items-center space-x-2">
                <span className="text-sm font-medium">Fields:</span>
                <span className="text-sm">{template.populatedFields} / {template.totalFields}</span>
              </div>
              
              <div className="flex items-center space-x-2">
                <span className="text-sm font-medium">Avg. Confidence:</span>
                <span className={`text-sm px-2 py-1 rounded ${getConfidenceColor(template.averageConfidence)}`}>
                  {Math.round(template.averageConfidence * 100)}%
                </span>
              </div>
            </div>
            
            <div className="flex items-center space-x-4">
              {template.validationErrors.length > 0 && (
                <div className="flex items-center space-x-1 text-red-600">
                  <AlertTriangle className="h-4 w-4" />
                  <span className="text-sm">{template.validationErrors.length} errors</span>
                </div>
              )}
              {template.warnings.length > 0 && (
                <div className="flex items-center space-x-1 text-yellow-600">
                  <Info className="h-4 w-4" />
                  <span className="text-sm">{template.warnings.length} warnings</span>
                </div>
              )}
              
              <Button 
                variant="ghost" 
                size="sm" 
                onClick={() => setShowValidation(!showValidation)}
              >
                {showValidation ? 'Hide' : 'Show'} Validation
              </Button>
            </div>
          </div>
        </div>

        {/* Sheet Tabs */}
        {sheets.length > 1 && (
          <div className="px-6 border-b border-gray-200 dark:border-gray-700">
            <div className="flex space-x-1">
              {sheets.map((sheet) => (
                <button
                  key={sheet}
                  onClick={() => setActiveSheet(sheet)}
                  className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                    activeSheet === sheet
                      ? 'text-blue-600 border-blue-600'
                      : 'text-gray-500 border-transparent hover:text-gray-700'
                  }`}
                >
                  {sheet.charAt(0).toUpperCase() + sheet.slice(1)}
                </button>
              ))}
            </div>
          </div>
        )}

        <div className="flex flex-1 overflow-hidden">
          {/* Main Content */}
          <div className="flex-1 overflow-y-auto" style={{ fontSize: `${zoomLevel}%` }}>
            <div className="p-6">
              <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
                <div className="grid grid-cols-12 gap-px bg-gray-200 dark:bg-gray-700">
                  {/* Header Row */}
                  <div className="col-span-5 bg-gray-100 dark:bg-gray-800 p-3 font-medium text-sm">
                    Field Name
                  </div>
                  <div className="col-span-3 bg-gray-100 dark:bg-gray-800 p-3 font-medium text-sm">
                    Value
                  </div>
                  <div className="col-span-2 bg-gray-100 dark:bg-gray-800 p-3 font-medium text-sm">
                    Confidence
                  </div>
                  <div className="col-span-2 bg-gray-100 dark:bg-gray-800 p-3 font-medium text-sm">
                    Source
                  </div>

                  {/* Data Rows */}
                  {Object.entries(currentFields).map(([fieldName, field]) => (
                    <React.Fragment key={fieldName}>
                      <div 
                        className={`col-span-5 bg-white dark:bg-gray-900 p-3 text-sm cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 ${
                          selectedField === fieldName ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                        }`}
                        onClick={() => setSelectedField(fieldName)}
                      >
                        <div className="flex items-center space-x-2">
                          <span className={field.isRequired ? 'font-medium' : ''}>{field.name}</span>
                          {field.isRequired && <span className="text-red-500">*</span>}
                          {field.isFormula && <span className="text-blue-500 text-xs">ƒ</span>}
                        </div>
                      </div>
                      
                      <div className={`col-span-3 bg-white dark:bg-gray-900 p-3 text-sm ${
                        selectedField === fieldName ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                      }`}>
                        {field.value !== null && field.value !== undefined ? (
                          <span className={field.confidence < 0.7 ? 'text-yellow-600' : ''}>
                            {formatValue(field)}
                          </span>
                        ) : (
                          <span className="text-gray-400 italic">Not populated</span>
                        )}
                      </div>
                      
                      <div className={`col-span-2 bg-white dark:bg-gray-900 p-3 text-sm ${
                        selectedField === fieldName ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                      }`}>
                        <span className={`px-2 py-1 rounded text-xs ${getConfidenceColor(field.confidence)}`}>
                          {Math.round(field.confidence * 100)}%
                        </span>
                      </div>
                      
                      <div className={`col-span-2 bg-white dark:bg-gray-900 p-3 text-sm ${
                        selectedField === fieldName ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                      }`}>
                        <div className="flex items-center space-x-2">
                          {getSourceIcon(field.source)}
                          <span className="text-xs text-gray-500">{field.source}</span>
                        </div>
                      </div>
                    </React.Fragment>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Validation Panel */}
          {showValidation && (template.validationErrors.length > 0 || template.warnings.length > 0) && (
            <div className="w-80 bg-gray-50 dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 overflow-y-auto">
              <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                <h3 className="font-semibold">Validation Results</h3>
              </div>
              
              <div className="p-4 space-y-4">
                {template.validationErrors.length > 0 && (
                  <div>
                    <h4 className="font-medium text-red-600 mb-2 flex items-center">
                      <AlertTriangle className="h-4 w-4 mr-1" />
                      Errors ({template.validationErrors.length})
                    </h4>
                    <div className="space-y-2">
                      {template.validationErrors.map((error, idx) => (
                        <div key={idx} className="text-sm text-red-600 bg-red-50 dark:bg-red-900/20 p-2 rounded">
                          {error}
                        </div>
                      ))}
                    </div>
                  </div>
                )}
                
                {template.warnings.length > 0 && (
                  <div>
                    <h4 className="font-medium text-yellow-600 mb-2 flex items-center">
                      <Info className="h-4 w-4 mr-1" />
                      Warnings ({template.warnings.length})
                    </h4>
                    <div className="space-y-2">
                      {template.warnings.map((warning, idx) => (
                        <div key={idx} className="text-sm text-yellow-600 bg-yellow-50 dark:bg-yellow-900/20 p-2 rounded">
                          {warning}
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
} 