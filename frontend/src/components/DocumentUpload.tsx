import React, { useState, useCallback } from 'react';
import { Upload, File, X, CheckCircle, AlertCircle, Info } from 'lucide-react';
import { UploadDocument, UploadDocuments } from '../../wailsjs/go/main/App';
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

  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const files = Array.from(e.dataTransfer.files);
    handleFiles(files);
  }, []);

  const handleFileInput = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const files = Array.from(e.target.files);
      handleFiles(files);
    }
  }, []);

  const handleFiles = useCallback((files: File[]) => {
    const newFiles: UploadedFile[] = files.map(file => ({
      file,
      id: `${Date.now()}-${Math.random()}`,
      status: 'pending'
    }));

    setUploadedFiles(prev => [...prev, ...newFiles]);
  }, []);

  const removeFile = useCallback((id: string) => {
    setUploadedFiles(prev => prev.filter(f => f.id !== id));
  }, []);

  const processFiles = useCallback(async () => {
    const pendingFiles = uploadedFiles.filter(f => f.status === 'pending');
    if (pendingFiles.length === 0) return;

    setIsProcessing(true);

    try {
      if (pendingFiles.length === 1) {
        // Single file upload
        const uploadFile = pendingFiles[0];
        
        // Update status to processing
        setUploadedFiles(prev => prev.map(f => 
          f.id === uploadFile.id ? { ...f, status: 'processing' } : f
        ));

        // Read file as array buffer
        const arrayBuffer = await uploadFile.file.arrayBuffer();
        const uint8Array = new Uint8Array(arrayBuffer);

        // Call backend upload method
        const result = await UploadDocument(dealName, uploadFile.file.name, Array.from(uint8Array));

        // Update status to success
        setUploadedFiles(prev => prev.map(f => 
          f.id === uploadFile.id ? { ...f, status: 'success', result } : f
        ));

        toast({
          title: "Upload Successful",
          description: `${uploadFile.file.name} has been uploaded and processed.`,
        });

      } else {
        // Multiple file upload
        const fileData: { [key: string]: number[] } = {};
        
        // Update all to processing
        setUploadedFiles(prev => prev.map(f => 
          pendingFiles.some(pf => pf.id === f.id) ? { ...f, status: 'processing' } : f
        ));

        // Read all files
        for (const uploadFile of pendingFiles) {
          const arrayBuffer = await uploadFile.file.arrayBuffer();
          const uint8Array = new Uint8Array(arrayBuffer);
          fileData[uploadFile.file.name] = Array.from(uint8Array);
        }

        // Call backend batch upload method
        const results = await UploadDocuments(dealName, fileData);

        // Update statuses based on results
        setUploadedFiles(prev => prev.map(f => {
          const pendingFile = pendingFiles.find(pf => pf.id === f.id);
          if (pendingFile) {
            const fileResult = results.find((r: any) => r.fileName === pendingFile.file.name);
            return {
              ...f,
              status: fileResult ? 'success' : 'error',
              result: fileResult,
              error: fileResult ? undefined : 'Upload failed'
            };
          }
          return f;
        }));

        toast({
          title: "Batch Upload Complete",
          description: `${pendingFiles.length} files have been uploaded and processed.`,
        });
      }

      // Call completion callback
      if (onUploadComplete) {
        const results = uploadedFiles
          .filter(f => f.status === 'success')
          .map(f => f.result);
        onUploadComplete(results);
      }

    } catch (error) {
      console.error('Upload error:', error);
      
      // Update failed files to error status
      setUploadedFiles(prev => prev.map(f => 
        pendingFiles.some(pf => pf.id === f.id) 
          ? { ...f, status: 'error', error: error instanceof Error ? error.message : 'Upload failed' }
          : f
      ));

      toast({
        title: "Upload Failed",
        description: error instanceof Error ? error.message : 'An error occurred during upload.',
        variant: "destructive",
      });
    } finally {
      setIsProcessing(false);
    }
  }, [uploadedFiles, dealName, onUploadComplete, toast]);

  const clearAll = useCallback(() => {
    setUploadedFiles([]);
  }, []);

  const pendingCount = uploadedFiles.filter(f => f.status === 'pending').length;
  const processingCount = uploadedFiles.filter(f => f.status === 'processing').length;
  const successCount = uploadedFiles.filter(f => f.status === 'success').length;
  const errorCount = uploadedFiles.filter(f => f.status === 'error').length;

  return (
    <div className="space-y-4">
      {/* Info Banner */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div className="flex items-start gap-3">
          <Info className="h-5 w-5 text-blue-600 mt-0.5 flex-shrink-0" />
          <div className="text-sm">
            <p className="text-blue-800 font-medium">Real File Upload & n8n Processing</p>
            <p className="text-blue-700 mt-1">
              Files will be saved to the deal folder and automatically processed by n8n workflows for document analysis, routing, and classification.
            </p>
          </div>
        </div>
      </div>

      {/* Upload Area */}
      <div
        className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
          isDragging
            ? 'border-blue-400 bg-blue-50'
            : 'border-gray-300 hover:border-gray-400'
        }`}
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
      >
        <Upload className="mx-auto h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          Upload Documents
        </h3>
        <p className="text-gray-600 mb-4">
          Drag and drop files here, or click to select files
        </p>
        <input
          type="file"
          multiple
          onChange={handleFileInput}
          className="hidden"
          id="file-upload"
          accept=".pdf,.doc,.docx,.txt,.rtf"
        />
        <label
          htmlFor="file-upload"
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 cursor-pointer"
        >
          Select Files
        </label>
      </div>

      {/* File List */}
      {uploadedFiles.length > 0 && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <h4 className="font-medium text-gray-900">
              Files ({uploadedFiles.length})
            </h4>
            <div className="flex gap-2">
              {pendingCount > 0 && (
                <Button
                  onClick={processFiles}
                  disabled={isProcessing}
                  size="sm"
                  className="bg-green-600 hover:bg-green-700"
                >
                  {isProcessing ? 'Processing...' : `Upload ${pendingCount} File${pendingCount !== 1 ? 's' : ''}`}
                </Button>
              )}
              <Button
                onClick={clearAll}
                variant="outline"
                size="sm"
                disabled={isProcessing}
              >
                Clear All
              </Button>
            </div>
          </div>

          {/* Status Summary */}
          {(processingCount > 0 || successCount > 0 || errorCount > 0) && (
            <div className="flex gap-4 text-sm">
              {processingCount > 0 && (
                <span className="text-blue-600">
                  {processingCount} processing...
                </span>
              )}
              {successCount > 0 && (
                <span className="text-green-600">
                  {successCount} completed
                </span>
              )}
              {errorCount > 0 && (
                <span className="text-red-600">
                  {errorCount} failed
                </span>
              )}
            </div>
          )}

          {/* File Items */}
          <div className="space-y-2 max-h-60 overflow-y-auto">
            {uploadedFiles.map((uploadedFile) => (
              <div
                key={uploadedFile.id}
                className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
              >
                <div className="flex items-center gap-3">
                  <File className="h-5 w-5 text-gray-400" />
                  <div>
                    <p className="text-sm font-medium text-gray-900">
                      {uploadedFile.file.name}
                    </p>
                    <p className="text-xs text-gray-500">
                      {(uploadedFile.file.size / 1024 / 1024).toFixed(2)} MB
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {uploadedFile.status === 'pending' && (
                    <span className="text-xs text-gray-500">Pending</span>
                  )}
                  {uploadedFile.status === 'processing' && (
                    <span className="text-xs text-blue-600">Processing...</span>
                  )}
                  {uploadedFile.status === 'success' && (
                    <CheckCircle className="h-5 w-5 text-green-600" />
                  )}
                  {uploadedFile.status === 'error' && (
                    <AlertCircle className="h-5 w-5 text-red-600" />
                  )}
                  <button
                    onClick={() => removeFile(uploadedFile.id)}
                    className="p-1 hover:bg-gray-200 rounded"
                    disabled={uploadedFile.status === 'processing'}
                  >
                    <X className="h-4 w-4 text-gray-400" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
} 