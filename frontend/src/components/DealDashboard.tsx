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
  Filter
} from 'lucide-react';
import { GetDealsList, ProcessFolder, CreateDeal } from '../../wailsjs/go/main/App';
import { DocumentUpload } from './DocumentUpload';
import { DocumentSearch, DocumentItem } from './DocumentSearch';
import { DocumentViewer } from './DocumentViewer';
import { DealCreationDialog } from './DealCreationDialog';
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
      const transformedDeals: Deal[] = dealsList.map((deal: any) => ({
        name: deal.name,
        createdAt: new Date(deal.modTime),
        documentCount: Math.floor(Math.random() * 50) + 10, // Placeholder
        analysisComplete: Math.random() > 0.3,
        riskScore: Math.random() * 100,
        completeness: Math.random() * 100,
      }));
      
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

  const processDealFolder = async (dealName: string) => {
    try {
      const results = await ProcessFolder(`Deals/${dealName}`, dealName);
      toast({
        title: "Folder Processed",
        description: `Processed ${results.length} documents`,
      });
      await loadDeals();
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to process deal folder",
        variant: "destructive",
      });
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
                      onClick={() => processDealFolder(selectedDealData.name)}
                    >
                      <BarChart3 className="h-4 w-4 mr-2" />
                      Analyze All
                    </Button>
                    <Button onClick={() => setShowUpload(!showUpload)}>
                      <Plus className="h-4 w-4 mr-2" />
                      Add Documents
                    </Button>
                  </div>
                </div>
              </div>
            </div>

            {/* Stats Grid */}
            <div className="p-6">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
                <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-gray-500 dark:text-gray-400">Total Documents</p>
                      <p className="text-2xl font-bold mt-1">{selectedDealData.documentCount}</p>
                    </div>
                    <FileText className="h-8 w-8 text-blue-500 opacity-20" />
                  </div>
                </div>
                
                <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-gray-500 dark:text-gray-400">Completeness</p>
                      <p className="text-2xl font-bold mt-1">
                        {Math.round(selectedDealData.completeness || 0)}%
                      </p>
                    </div>
                    <CheckCircle className="h-8 w-8 text-green-500 opacity-20" />
                  </div>
                </div>
                
                <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-gray-500 dark:text-gray-400">Risk Score</p>
                      <p className="text-2xl font-bold mt-1">
                        {Math.round(selectedDealData.riskScore || 0)}
                      </p>
                    </div>
                    <AlertTriangle className="h-8 w-8 text-yellow-500 opacity-20" />
                  </div>
                </div>
                
                <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-gray-500 dark:text-gray-400">Analysis Status</p>
                      <p className="text-sm font-medium mt-1">
                        {selectedDealData.analysisComplete ? 'Complete' : 'In Progress'}
                      </p>
                    </div>
                    <TrendingUp className="h-8 w-8 text-purple-500 opacity-20" />
                  </div>
                </div>
              </div>

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