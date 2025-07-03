import React, { useState, useEffect } from 'react';
import { 
  FileText, 
  TrendingUp, 
  AlertTriangle, 
  CheckCircle, 
  Clock,
  BarChart3,
  FolderOpen,
  Plus,
  Search,
  Filter,
  Loader2
} from 'lucide-react';
import { GetDealsList, ProcessFolder, CreateDeal, GetDealFolderPath, GetPopulatedTemplateData } from '../../wailsjs/go/main/App';
import { DocumentUpload } from './DocumentUpload';
import { DocumentSearch, DocumentItem } from './DocumentSearch';
import { DocumentViewer } from './DocumentViewer';
import { DealCreationDialog } from './DealCreationDialog';
import { AnalysisProgress } from './AnalysisProgress';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { useToast } from '../hooks/use-toast';

interface Deal {
  name: string;
  createdAt: Date;
  documentCount: number;
  analysisComplete: boolean;
  riskScore?: number;
  completeness?: number;
}

interface DealStats {
  totalDocuments: number;
  analyzedDocuments: number;
  legalDocuments: number;
  financialDocuments: number;
  generalDocuments: number;
  averageConfidence: number;
}

export function DealDashboard() {
  const [deals, setDeals] = useState<Deal[]>([]);
  const [selectedDeal, setSelectedDeal] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [showUpload, setShowUpload] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [dealStats, setDealStats] = useState<DealStats | null>(null);
  const [documents, setDocuments] = useState<DocumentItem[]>([]);
  const [selectedDocument, setSelectedDocument] = useState<DocumentItem | null>(null);
  const [showDocumentViewer, setShowDocumentViewer] = useState(false);
  const [showDealCreation, setShowDealCreation] = useState(false);
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [analysisProgress, setAnalysisProgress] = useState(0);
  const [populatedTemplates, setPopulatedTemplates] = useState<any[]>([]);
  const [showTemplateResults, setShowTemplateResults] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    loadDeals();
  }, []);

  useEffect(() => {
    if (selectedDeal) {
      loadDealDocuments(selectedDeal);
    }
  }, [selectedDeal]);

  const loadDeals = async () => {
    try {
      setIsLoading(true);
      const dealsList = await GetDealsList();
      
      // Transform the deals data
      const transformedDeals: Deal[] = dealsList.map((deal: any) => {
        // Handle date parsing with fallback
        let createdAt: Date;
        if (deal.createdAt) {
          createdAt = new Date(deal.createdAt);
          // Check if date is valid
          if (isNaN(createdAt.getTime())) {
            createdAt = new Date(); // Fallback to current date
          }
        } else {
          createdAt = new Date(); // Fallback to current date
        }

        return {
          name: deal.name,
          createdAt: createdAt,
          documentCount: deal.documentCount || 0, // Use actual document count from backend
          analysisComplete: Math.random() > 0.3,
          riskScore: Math.random() * 100,
          completeness: Math.random() * 100,
        };
      });
      
      setDeals(transformedDeals);
      
      // Only auto-select first deal if no deal is currently selected and deals exist
      if (transformedDeals.length > 0 && !selectedDeal) {
        setSelectedDeal(transformedDeals[0].name);
      }
      
      // If a deal was selected but no longer exists, clear selection
      if (selectedDeal && !transformedDeals.find(d => d.name === selectedDeal)) {
        setSelectedDeal(null);
      }
    } catch (error) {
      console.error('Error loading deals:', error);
      toast({
        title: "Error",
        description: "Failed to load deals",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const loadDealDocuments = async (dealName: string) => {
    try {
      // In a real implementation, this would fetch actual documents from the deal folder
      // For now, we'll generate sample documents
      const sampleDocuments: DocumentItem[] = [
        {
          id: '1',
          name: 'Purchase Agreement.pdf',
          type: 'legal',
          status: 'completed',
          uploadDate: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000),
          analysisDate: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000),
          size: 2456789,
          confidence: 0.92,
          path: `Deals/${dealName}/legal/Purchase Agreement.pdf`,
          tags: ['contract', 'acquisition', 'terms']
        },
        {
          id: '2',
          name: 'Financial Statements Q3.xlsx',
          type: 'financial',
          status: 'completed',
          uploadDate: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000),
          analysisDate: new Date(Date.now() - 4 * 24 * 60 * 60 * 1000),
          size: 1234567,
          confidence: 0.88,
          path: `Deals/${dealName}/financial/Financial Statements Q3.xlsx`,
          tags: ['quarterly', 'financials', 'revenue']
        },
        {
          id: '3',
          name: 'Due Diligence Report.docx',
          type: 'legal',
          status: 'processing',
          uploadDate: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000),
          size: 987654,
          path: `Deals/${dealName}/legal/Due Diligence Report.docx`,
          tags: ['due-diligence', 'review']
        },
        {
          id: '4',
          name: 'Market Analysis.pptx',
          type: 'general',
          status: 'completed',
          uploadDate: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000),
          analysisDate: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000),
          size: 3456789,
          confidence: 0.76,
          path: `Deals/${dealName}/general/Market Analysis.pptx`,
          tags: ['market', 'analysis', 'competition']
        },
        {
          id: '5',
          name: 'Valuation Model.xlsx',
          type: 'financial',
          status: 'error',
          uploadDate: new Date(Date.now() - 6 * 24 * 60 * 60 * 1000),
          size: 5678901,
          path: `Deals/${dealName}/financial/Valuation Model.xlsx`,
          tags: ['valuation', 'dcf', 'model']
        }
      ];
      
      setDocuments(sampleDocuments);
    } catch (error) {
      console.error('Error loading deal documents:', error);
      toast({
        title: "Error",
        description: "Failed to load deal documents",
        variant: "destructive",
      });
    }
  };

  const handleDocumentSelect = (document: DocumentItem) => {
    setSelectedDocument(document);
    setShowDocumentViewer(true);
  };

  const handleDocumentAction = async (documentId: string, action: 'move' | 'delete' | 'reprocess') => {
    const document = documents.find(d => d.id === documentId);
    if (!document) return;

    try {
      switch (action) {
        case 'move':
          toast({
            title: "Move Document",
            description: `Move dialog for "${document.name}" would open here`,
          });
          break;
        case 'delete':
          // In a real implementation, this would delete the document
          setDocuments(prev => prev.filter(d => d.id !== documentId));
          toast({
            title: "Document Deleted",
            description: `"${document.name}" has been deleted`,
          });
          break;
        case 'reprocess':
          // In a real implementation, this would trigger reprocessing
          setDocuments(prev => prev.map(d => 
            d.id === documentId 
              ? { ...d, status: 'processing' as const }
              : d
          ));
          toast({
            title: "Reprocessing Started",
            description: `"${document.name}" is being reprocessed`,
          });
          // Simulate processing completion
          setTimeout(() => {
            setDocuments(prev => prev.map(d => 
              d.id === documentId 
                ? { ...d, status: 'completed' as const, analysisDate: new Date() }
                : d
            ));
            toast({
              title: "Reprocessing Complete",
              description: `"${document.name}" has been reprocessed`,
            });
          }, 3000);
          break;
      }
    } catch (error) {
      toast({
        title: "Error",
        description: `Failed to ${action} document`,
        variant: "destructive",
      });
    }
  };

  const createNewDeal = () => {
    setShowDealCreation(true);
  };

  const handleDealCreated = async (dealData: any) => {
    try {
      // Actually create the deal folder using the backend API
      await CreateDeal(dealData.name);
      
      // Refresh the deals list from the backend to get the newly created deal
      await loadDeals();
      
      // Select the newly created deal
      setSelectedDeal(dealData.name);
      
      toast({
        title: "Deal Created Successfully",
        description: `"${dealData.name}" has been created and is ready for document upload`,
      });
      
    } catch (error) {
      console.error('Error creating deal:', error);
      toast({
        title: "Error",
        description: "Failed to create deal",
        variant: "destructive",
      });
    }
  };

  const analyzeAll = async (dealName: string) => {
    setIsAnalyzing(true);
    setAnalysisProgress(0);
    
    try {
      // Simulate progress steps
      const progressSteps = [
        { progress: 20, message: "Preparing analysis..." },
        { progress: 40, message: "Discovering templates..." },
        { progress: 60, message: "Extracting document data..." },
        { progress: 80, message: "Populating templates..." },
        { progress: 95, message: "Finalizing analysis..." }
      ];
      
      // Start progress animation
      progressSteps.forEach((step, index) => {
        setTimeout(() => {
          setAnalysisProgress(step.progress);
        }, index * 500);
      });
      
      // Get the full path to the deal folder
      const dealFolderPath = await GetDealFolderPath(dealName);
      
      if (!dealFolderPath) {
        toast({
          title: "Error",
          description: "Could not determine deal folder path",
          variant: "destructive",
        });
        return;
      }
      
      const results = await ProcessFolder(dealFolderPath, dealName);
      
      // Complete progress
      setAnalysisProgress(100);
      
      if (results && results.length > 0) {
        // Count already processed vs newly processed files
        const alreadyProcessed = results.filter((r: any) => r.alreadyProcessed).length;
        const newlyProcessed = results.filter((r: any) => r.success && !r.alreadyProcessed).length;
        const failed = results.filter((r: any) => !r.success).length;
        
        if (alreadyProcessed > 0 && newlyProcessed === 0) {
          toast({
            title: "Files Already Processed",
            description: `All ${alreadyProcessed} documents in this deal have already been analyzed and routed to the correct folders.`,
          });
        } else if (newlyProcessed > 0 && alreadyProcessed === 0) {
          toast({
            title: "Folder Processed",
            description: `Successfully processed ${newlyProcessed} new documents`,
          });
        } else if (newlyProcessed > 0 && alreadyProcessed > 0) {
          toast({
            title: "Folder Processed",
            description: `Processed ${newlyProcessed} new documents. ${alreadyProcessed} documents were already processed.`,
          });
        } else if (failed > 0) {
          toast({
            title: "Processing Complete with Errors",
            description: `${failed} documents failed to process. Check the logs for details.`,
            variant: "destructive",
          });
        }
      } else {
        toast({
          title: "Analysis Complete",
          description: "No documents found to process in this deal folder",
        });
      }
      
      // Fetch populated template data after analysis
      try {
        const templateData = await GetPopulatedTemplateData(dealName);
        if (templateData && templateData.templates) {
          setPopulatedTemplates(templateData.templates);
          if (templateData.templates.length > 0) {
            setShowTemplateResults(true);
            toast({
              title: "Template Analysis Complete",
              description: `Successfully populated ${templateData.templates.length} templates with document data`,
            });
          }
        }
      } catch (templateError) {
        console.error('Error fetching populated template data:', templateError);
      }
      
      await loadDeals();
    } catch (error) {
      console.error('Error processing deal folder:', error);
      
      // Check if the error is due to an empty folder or missing folder
      const errorMessage = error instanceof Error ? error.message : String(error);
      console.log('Full error message:', errorMessage);
      
      if (errorMessage.includes('no such file or directory') || 
          errorMessage.includes('does not exist') ||
          errorMessage.includes('failed to list files in folder')) {
        toast({
          title: "Analysis Complete",
          description: "No documents found in this deal folder. Upload documents first to analyze them.",
        });
      } else if (errorMessage.includes('document router not initialized') ||
                 errorMessage.includes('document processor not initialized')) {
        toast({
          title: "Service Error",
          description: "Document analysis service is not available. Please check your configuration.",
          variant: "destructive",
        });
      } else {
        toast({
          title: "Error",
          description: `Failed to process deal folder: ${errorMessage}`,
          variant: "destructive",
        });
      }
    } finally {
      // Reset animation state after a brief delay
      setTimeout(() => {
        setIsAnalyzing(false);
        setAnalysisProgress(0);
      }, 1000);
    }
  };

  const filteredDeals = deals.filter(deal =>
    deal.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const selectedDealData = deals.find(d => d.name === selectedDeal);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-gray-50 dark:bg-gray-900">
      {/* Sidebar */}
      <div className="w-80 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold">Deals</h2>
            <Button size="sm" onClick={createNewDeal}>
              <Plus className="h-4 w-4 mr-1" />
              New Deal
            </Button>
          </div>
          
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              type="text"
              placeholder="Search deals..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
        </div>
        
        <div className="overflow-y-auto h-full pb-20">
          {filteredDeals.map((deal) => (
            <div
              key={deal.name}
              className={`p-4 border-b border-gray-200 dark:border-gray-700 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50 ${
                selectedDeal === deal.name ? 'bg-blue-50 dark:bg-blue-900/20' : ''
              }`}
              onClick={() => setSelectedDeal(deal.name)}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <h3 className="font-medium text-sm">{deal.name}</h3>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    {deal.documentCount} documents
                  </p>
                </div>
                {deal.analysisComplete ? (
                  <CheckCircle className="h-5 w-5 text-green-500" />
                ) : (
                  <Clock className="h-5 w-5 text-yellow-500" />
                )}
              </div>
              
              <div className="mt-2 space-y-1">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-gray-500">Completeness</span>
                  <span className="font-medium">{Math.round(deal.completeness || 0)}%</span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-600 rounded-full h-1.5">
                  <div 
                    className="bg-blue-600 h-1.5 rounded-full"
                    style={{ width: `${deal.completeness || 0}%` }}
                  />
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-y-auto">
        {selectedDealData ? (
          <>
            {/* Header */}
            <div className="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
              <div className="px-6 py-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h1 className="text-2xl font-bold">{selectedDealData.name}</h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                      Created {selectedDealData.createdAt.toLocaleDateString()}
                    </p>
                  </div>
                  <div className="flex space-x-2">
                    <Button
                      variant="outline"
                      onClick={() => analyzeAll(selectedDealData.name)}
                      disabled={isAnalyzing}
                      className={isAnalyzing ? "opacity-75" : ""}
                    >
                      {isAnalyzing ? (
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      ) : (
                        <BarChart3 className="h-4 w-4 mr-2" />
                      )}
                      {isAnalyzing ? "Analyzing..." : "Analyze All"}
                    </Button>
                    <Button onClick={() => setShowUpload(!showUpload)}>
                      <Plus className="h-4 w-4 mr-2" />
                      Add Documents
                    </Button>
                  </div>
                </div>
              </div>
            </div>

            {/* Analysis Progress Indicator */}
            <AnalysisProgress 
              isVisible={isAnalyzing} 
              progress={analysisProgress} 
            />

            {/* Stats Grid */}
            <div className="p-6">
              <div className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6 transition-all duration-300 ${
                isAnalyzing ? 'opacity-90' : 'opacity-100'
              }`}>
                <div className={`bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm transition-all duration-300 ${
                  isAnalyzing ? 'ring-2 ring-blue-200 dark:ring-blue-800' : ''
                }`}>
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-gray-500 dark:text-gray-400">Total Documents</p>
                      <p className="text-2xl font-bold mt-1">{selectedDealData.documentCount}</p>
                    </div>
                    <FileText className={`h-8 w-8 text-blue-500 transition-all duration-300 ${
                      isAnalyzing ? 'opacity-40 animate-pulse' : 'opacity-20'
                    }`} />
                  </div>
                </div>
                
                <div className={`bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm transition-all duration-300 ${
                  isAnalyzing ? 'ring-2 ring-green-200 dark:ring-green-800' : ''
                }`}>
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-gray-500 dark:text-gray-400">Completeness</p>
                      <p className="text-2xl font-bold mt-1">
                        {Math.round(selectedDealData.completeness || 0)}%
                      </p>
                    </div>
                    <CheckCircle className={`h-8 w-8 text-green-500 transition-all duration-300 ${
                      isAnalyzing ? 'opacity-40 animate-pulse' : 'opacity-20'
                    }`} />
                  </div>
                </div>
                
                <div className={`bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm transition-all duration-300 ${
                  isAnalyzing ? 'ring-2 ring-yellow-200 dark:ring-yellow-800' : ''
                }`}>
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-gray-500 dark:text-gray-400">Risk Score</p>
                      <p className="text-2xl font-bold mt-1">
                        {Math.round(selectedDealData.riskScore || 0)}
                      </p>
                    </div>
                    <AlertTriangle className={`h-8 w-8 text-yellow-500 transition-all duration-300 ${
                      isAnalyzing ? 'opacity-40 animate-pulse' : 'opacity-20'
                    }`} />
                  </div>
                </div>
                
                <div className={`bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm transition-all duration-300 ${
                  isAnalyzing ? 'ring-2 ring-purple-200 dark:ring-purple-800' : ''
                }`}>
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-gray-500 dark:text-gray-400">Analysis Status</p>
                      <p className="text-sm font-medium mt-1">
                        {isAnalyzing ? 'Analyzing...' : selectedDealData.analysisComplete ? 'Complete' : 'In Progress'}
                      </p>
                    </div>
                    <TrendingUp className={`h-8 w-8 text-purple-500 transition-all duration-300 ${
                      isAnalyzing ? 'opacity-40 animate-pulse' : 'opacity-20'
                    }`} />
                  </div>
                </div>
              </div>

              {/* Template Analysis Results */}
              {showTemplateResults && populatedTemplates.length > 0 && (
                <div className="mb-6">
                  <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
                    <div className="flex items-center justify-between mb-4">
                      <h3 className="text-lg font-semibold text-green-900 dark:text-green-100">
                        📊 Template Analysis Results
                      </h3>
                      <div className="flex items-center space-x-2">
                        <button
                          onClick={async () => {
                            try {
                              const templateData = await GetPopulatedTemplateData(selectedDealData.name);
                              if (templateData && templateData.templates) {
                                setPopulatedTemplates(templateData.templates);
                              }
                            } catch (error) {
                              console.error('Error refreshing template data:', error);
                            }
                          }}
                          className="text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
                        >
                          🔄 Refresh
                        </button>
                        <button
                          onClick={() => setShowTemplateResults(false)}
                          className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                        >
                          ✕
                        </button>
                      </div>
                    </div>
                    
                    <div className="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-4 mb-4">
                      <p className="text-sm text-green-800 dark:text-green-200">
                        ✅ Analysis complete! Your templates have been populated with real data extracted from the uploaded documents.
                      </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      {populatedTemplates.map((template, index) => (
                        <div
                          key={index}
                          className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:shadow-md transition-shadow"
                        >
                          <div className="flex items-start justify-between mb-3">
                            <div>
                              <h4 className="font-medium text-gray-900 dark:text-gray-100">
                                {template.name}
                              </h4>
                              <p className="text-sm text-gray-500 dark:text-gray-400">
                                {template.type.toUpperCase()} • {template.fieldCount} fields
                              </p>
                            </div>
                            {template.hasFormulas && (
                              <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300">
                                📋 Formulas
                              </span>
                            )}
                          </div>

                          {/* Sample Data Preview */}
                          {template.sampleData && template.sampleData.length > 0 && (
                            <div className="mt-3">
                              <p className="text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Sample Data:
                              </p>
                              <div className="bg-gray-50 dark:bg-gray-800 rounded p-2 text-xs">
                                {Object.entries(template.sampleData[0]).slice(0, 3).map(([key, value]) => (
                                  <div key={key} className="flex justify-between mb-1">
                                    <span className="text-gray-600 dark:text-gray-400 truncate mr-2">
                                      {key}:
                                    </span>
                                    <span className="text-gray-900 dark:text-gray-100 font-medium truncate">
                                      {String(value) || 'N/A'}
                                    </span>
                                  </div>
                                ))}
                                {Object.keys(template.sampleData[0]).length > 3 && (
                                  <div className="text-gray-500 dark:text-gray-400 text-center">
                                    +{Object.keys(template.sampleData[0]).length - 3} more fields
                                  </div>
                                )}
                              </div>
                            </div>
                          )}

                          {/* Template Stats */}
                          <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
                            <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400">
                              <span>Updated: {new Date(template.lastModified).toLocaleDateString()}</span>
                              <span>{Math.round(template.size / 1024)} KB</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>

                    <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        💡 <strong>Next steps:</strong> Review the populated templates in your deal's analysis folder. 
                        The AI has extracted and mapped data from your documents while preserving formulas and formatting.
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* Document Upload Section */}
              {showUpload && (
                <div className="mb-6">
                  <DocumentUpload 
                    dealName={selectedDealData.name}
                    onUploadComplete={async () => {
                      setShowUpload(false);
                      // Refresh deals list but keep current selection
                      const currentSelection = selectedDeal;
                      await loadDeals();
                      if (currentSelection) {
                        setSelectedDeal(currentSelection);
                      }
                    }}
                  />
                </div>
              )}

              {/* Document Search and Management */}
              <div className="mb-6">
                <h3 className="text-lg font-semibold mb-4">Documents</h3>
                <DocumentSearch
                  documents={documents}
                  onDocumentSelect={handleDocumentSelect}
                  onDocumentAction={handleDocumentAction}
                />
              </div>

              {/* Recent Activity */}
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6 mt-6">
                <h3 className="text-lg font-semibold mb-4">Recent Activity</h3>
                <div className="space-y-3">
                  <div className="flex items-start space-x-3">
                    <div className="p-2 bg-blue-100 dark:bg-blue-900/20 rounded">
                      <FileText className="h-4 w-4 text-blue-600" />
                    </div>
                    <div className="flex-1">
                      <p className="text-sm">
                        <span className="font-medium">Purchase Agreement.pdf</span> was uploaded
                      </p>
                      <p className="text-xs text-gray-500">2 hours ago</p>
                    </div>
                  </div>
                  
                  <div className="flex items-start space-x-3">
                    <div className="p-2 bg-green-100 dark:bg-green-900/20 rounded">
                      <CheckCircle className="h-4 w-4 text-green-600" />
                    </div>
                    <div className="flex-1">
                      <p className="text-sm">
                        <span className="font-medium">Financial analysis</span> completed
                      </p>
                      <p className="text-xs text-gray-500">5 hours ago</p>
                    </div>
                  </div>
                  
                  <div className="flex items-start space-x-3">
                    <div className="p-2 bg-yellow-100 dark:bg-yellow-900/20 rounded">
                      <AlertTriangle className="h-4 w-4 text-yellow-600" />
                    </div>
                    <div className="flex-1">
                      <p className="text-sm">
                        <span className="font-medium">Risk assessment</span> identified 3 issues
                      </p>
                      <p className="text-xs text-gray-500">1 day ago</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </>
        ) : (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <FolderOpen className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                No deal selected
              </h3>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">
                Select a deal from the sidebar or create a new one
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Document Viewer Modal */}
      {showDocumentViewer && selectedDocument && (
        <DocumentViewer
          documentPath={selectedDocument.path}
          documentName={selectedDocument.name}
          onClose={() => {
            setShowDocumentViewer(false);
            setSelectedDocument(null);
          }}
        />
      )}

      {/* Deal Creation Dialog */}
      {showDealCreation && (
        <DealCreationDialog
          onClose={() => setShowDealCreation(false)}
          onDealCreated={handleDealCreated}
        />
      )}
    </div>
  );
} 