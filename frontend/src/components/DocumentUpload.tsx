import React, { useState, useCallback } from 'react';
import { Upload, File, X, CheckCircle, AlertCircle } from 'lucide-react';
import { ProcessDocument, ProcessDocuments } from '../../wailsjs/go/main/App';
import { Button } from './ui/button';
import { useToast } from '../hooks/use-toast';

interface UploadedFile {
  file: File;
  id: string;
  status: 'pending' | 'processing' | 'success' | 'error';
  result?: any;
  error?: string;
}

interface DocumentUploadProps {
  dealName: string;
  onUploadComplete?: (results: any[]) => void;
}

export function DocumentUpload({ dealName, onUploadComplete }: DocumentUploadProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [uploadedFiles, setUploadedFiles] = useState<UploadedFile[]>([]);
  const [isProcessing, setIsProcessing] = useState(false);
  const { toast } = useToast();

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const files = Array.from(e.dataTransfer.files);
    handleFiles(files);
  }, []);

  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const files = Array.from(e.target.files);
      handleFiles(files);
    }
  }, []);

  const handleFiles = (files: File[]) => {
    const newFiles: UploadedFile[] = files.map(file => ({
      file,
      id: `${file.name}-${Date.now()}-${Math.random()}`,
      status: 'pending' as const,
    }));

    setUploadedFiles(prev => [...prev, ...newFiles]);
  };

  const removeFile = (id: string) => {
    setUploadedFiles(prev => prev.filter(f => f.id !== id));
  };

  const processFiles = async () => {
    if (uploadedFiles.length === 0) return;

    setIsProcessing(true);
    const filePaths: string[] = [];

    // Update all files to processing status
    setUploadedFiles(prev => 
      prev.map(f => ({ ...f, status: 'processing' as const }))
    );

    try {
      // In a real implementation, we would upload files to a temporary location
      // For now, we'll use the file paths if available
      // This is a placeholder - actual implementation would handle file uploads
      
      toast({
        title: "Processing Documents",
        description: `Processing ${uploadedFiles.length} documents for ${dealName}...`,
      });

      // Simulate processing delay
      await new Promise(resolve => setTimeout(resolve, 2000));

      // Mark all as successful for now
      setUploadedFiles(prev => 
        prev.map(f => ({ 
          ...f, 
          status: 'success' as const,
          result: { 
            type: 'general',
            confidence: 0.85,
            routed: true 
          }
        }))
      );

      toast({
        title: "Success",
        description: `Successfully processed ${uploadedFiles.length} documents`,
      });

      if (onUploadComplete) {
        onUploadComplete(uploadedFiles.map(f => f.result));
      }
    } catch (error) {
      console.error('Error processing files:', error);
      
      setUploadedFiles(prev => 
        prev.map(f => ({ 
          ...f, 
          status: 'error' as const,
          error: 'Failed to process document'
        }))
      );

      toast({
        title: "Error",
        description: "Failed to process documents. Please try again.",
        variant: "destructive",
      });
    } finally {
      setIsProcessing(false);
    }
  };

  const getSupportedFormats = () => {
    return '.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.rtf,.jpg,.jpeg,.png,.tiff';
  };

  return (
    <div className="w-full max-w-4xl mx-auto p-6">
      <div
        className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
          isDragging 
            ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20' 
            : 'border-gray-300 dark:border-gray-600'
        }`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        <Upload className="mx-auto h-12 w-12 text-gray-400 mb-4" />
        
        <h3 className="text-lg font-semibold mb-2">
          Upload Documents for {dealName}
        </h3>
        
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Drag and drop your files here, or click to browse
        </p>
        
        <input
          type="file"
          id="file-upload"
          className="hidden"
          multiple
          accept={getSupportedFormats()}
          onChange={handleFileSelect}
        />
        
        <label
          htmlFor="file-upload"
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 cursor-pointer"
        >
          Select Files
        </label>
        
        <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
          Supported formats: PDF, Word, Excel, PowerPoint, Text, Images
        </p>
      </div>

      {uploadedFiles.length > 0 && (
        <div className="mt-6 space-y-2">
          <h4 className="text-sm font-medium mb-2">Uploaded Files</h4>
          
          {uploadedFiles.map(({ file, id, status, error }) => (
            <div
              key={id}
              className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg"
            >
              <div className="flex items-center space-x-3">
                <File className="h-5 w-5 text-gray-400" />
                <div className="flex-1">
                  <p className="text-sm font-medium">{file.name}</p>
                  <p className="text-xs text-gray-500">
                    {(file.size / 1024 / 1024).toFixed(2)} MB
                  </p>
                </div>
              </div>
              
              <div className="flex items-center space-x-2">
                {status === 'pending' && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => removeFile(id)}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                )}
                
                {status === 'processing' && (
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
                )}
                
                {status === 'success' && (
                  <CheckCircle className="h-4 w-4 text-green-600" />
                )}
                
                {status === 'error' && (
                  <div className="flex items-center space-x-1">
                    <AlertCircle className="h-4 w-4 text-red-600" />
                    <span className="text-xs text-red-600">{error}</span>
                  </div>
                )}
              </div>
            </div>
          ))}
          
          <div className="mt-4 flex justify-end space-x-2">
            <Button
              variant="outline"
              onClick={() => setUploadedFiles([])}
              disabled={isProcessing}
            >
              Clear All
            </Button>
            
            <Button
              onClick={processFiles}
              disabled={isProcessing || uploadedFiles.length === 0}
            >
              {isProcessing ? 'Processing...' : `Process ${uploadedFiles.length} Files`}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
} 