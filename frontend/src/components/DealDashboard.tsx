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
import { GetDealsList, ProcessFolder } from '../../wailsjs/go/main/App';
import { DocumentUpload } from './DocumentUpload';
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
  const { toast } = useToast();

  useEffect(() => {
    loadDeals();
  }, []);

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
      
      if (transformedDeals.length > 0 && !selectedDeal) {
        setSelectedDeal(transformedDeals[0].name);
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

  const createNewDeal = () => {
    // This would open a dialog to create a new deal
    toast({
      title: "Create New Deal",
      description: "Deal creation dialog would open here",
    });
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
                    onUploadComplete={() => {
                      setShowUpload(false);
                      loadDeals();
                    }}
                  />
                </div>
              )}

              {/* Document Categories */}
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
                <h3 className="text-lg font-semibold mb-4">Document Categories</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-sm font-medium">Legal Documents</span>
                      <span className="text-sm text-gray-500">15</span>
                    </div>
                    <div className="text-xs text-gray-500">
                      Contracts, agreements, legal opinions
                    </div>
                  </div>
                  
                  <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-sm font-medium">Financial Documents</span>
                      <span className="text-sm text-gray-500">22</span>
                    </div>
                    <div className="text-xs text-gray-500">
                      Financial statements, models, valuations
                    </div>
                  </div>
                  
                  <div className="p-4 bg-purple-50 dark:bg-purple-900/20 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-sm font-medium">General Documents</span>
                      <span className="text-sm text-gray-500">13</span>
                    </div>
                    <div className="text-xs text-gray-500">
                      Presentations, reports, correspondence
                    </div>
                  </div>
                </div>
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
    </div>
  );
} 