import React, { useState, useEffect } from 'react';
import { 
  X, 
  FileText, 
  Brain, 
  TrendingUp, 
  AlertTriangle,
  DollarSign,
  Building,
  Calendar,
  ChevronRight,
  Download,
  Share2,
  Maximize2
} from 'lucide-react';
import { 
  AnalyzeDocument, 
  ExtractFinancialData,
  AnalyzeDocumentRisks,
  GenerateDocumentInsights,
  ExtractDocumentEntities
} from '../../wailsjs/go/main/App';
import { Button } from './ui/button';
import { useToast } from '../hooks/use-toast';

interface DocumentViewerProps {
  documentPath: string;
  documentName: string;
  onClose: () => void;
}

interface AnalysisResults {
  classification?: any;
  financial?: any;
  risks?: any;
  insights?: any;
  entities?: any;
}

export function DocumentViewer({ documentPath, documentName, onClose }: DocumentViewerProps) {
  const [activeTab, setActiveTab] = useState<'preview' | 'analysis'>('preview');
  const [analysisType, setAnalysisType] = useState<'overview' | 'financial' | 'risks' | 'entities'>('overview');
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [analysisResults, setAnalysisResults] = useState<AnalysisResults>({});
  const [documentContent, setDocumentContent] = useState<string>('');
  const { toast } = useToast();

  useEffect(() => {
    // Load document content
    loadDocument();
    // Automatically analyze on open
    analyzeDocument();
  }, [documentPath]);

  const loadDocument = async () => {
    try {
      // In a real implementation, this would load the actual document content
      // For now, we'll use placeholder content
      setDocumentContent(`Document content for ${documentName} would be displayed here.`);
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to load document",
        variant: "destructive",
      });
    }
  };

  const analyzeDocument = async () => {
    setIsAnalyzing(true);
    
    try {
      // Run all analyses in parallel
      const [classification, financial, risks, insights, entities] = await Promise.all([
        AnalyzeDocument(documentPath).catch(err => null),
        ExtractFinancialData(documentPath).catch(err => null),
        AnalyzeDocumentRisks(documentPath).catch(err => null),
        GenerateDocumentInsights(documentPath).catch(err => null),
        ExtractDocumentEntities(documentPath).catch(err => null),
      ]);

      setAnalysisResults({
        classification,
        financial,
        risks,
        insights,
        entities,
      });

      toast({
        title: "Analysis Complete",
        description: "Document has been analyzed successfully",
      });
    } catch (error) {
      console.error('Analysis error:', error);
      toast({
        title: "Analysis Error",
        description: "Some analyses may have failed",
        variant: "destructive",
      });
    } finally {
      setIsAnalyzing(false);
    }
  };

  const renderAnalysisContent = () => {
    switch (analysisType) {
      case 'overview':
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold mb-3">Document Classification</h3>
              {analysisResults.classification ? (
                <div className="bg-gray-50 dark:bg-gray-800 p-4 rounded-lg">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">Type:</span>
                    <span className="text-sm bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 px-2 py-1 rounded">
                      {analysisResults.classification.documentType}
                    </span>
                  </div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">Confidence:</span>
                    <span className="text-sm">{(analysisResults.classification.confidence * 100).toFixed(1)}%</span>
                  </div>
                  {analysisResults.classification.summary && (
                    <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        {analysisResults.classification.summary}
                      </p>
                    </div>
                  )}
                </div>
              ) : (
                <p className="text-sm text-gray-500">No classification data available</p>
              )}
            </div>

            <div>
              <h3 className="text-lg font-semibold mb-3">Key Insights</h3>
              {analysisResults.insights ? (
                <div className="space-y-3">
                  {analysisResults.insights.keyPoints?.map((point: string, idx: number) => (
                    <div key={idx} className="flex items-start space-x-2">
                      <ChevronRight className="h-4 w-4 text-blue-500 mt-0.5" />
                      <p className="text-sm">{point}</p>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-gray-500">No insights available</p>
              )}
            </div>
          </div>
        );

      case 'financial':
        return (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold mb-3">Financial Data</h3>
            {analysisResults.financial ? (
              <div className="grid grid-cols-2 gap-4">
                {Object.entries(analysisResults.financial).map(([key, value]) => {
                  if (typeof value === 'number') {
                    return (
                      <div key={key} className="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg">
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          {key.replace(/([A-Z])/g, ' $1').trim()}
                        </p>
                        <p className="text-lg font-semibold mt-1">
                          {value.toLocaleString('en-US', { 
                            style: 'currency', 
                            currency: analysisResults.financial.currency || 'USD' 
                          })}
                        </p>
                      </div>
                    );
                  }
                  return null;
                })}
              </div>
            ) : (
              <p className="text-sm text-gray-500">No financial data extracted</p>
            )}
          </div>
        );

      case 'risks':
        return (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold mb-3">Risk Analysis</h3>
            {analysisResults.risks ? (
              <>
                <div className="bg-gray-50 dark:bg-gray-800 p-4 rounded-lg">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">Overall Risk Score:</span>
                    <span className={`text-2xl font-bold ${
                      analysisResults.risks.overallRiskScore > 0.7 ? 'text-red-600' :
                      analysisResults.risks.overallRiskScore > 0.4 ? 'text-yellow-600' :
                      'text-green-600'
                    }`}>
                      {(analysisResults.risks.overallRiskScore * 100).toFixed(0)}
                    </span>
                  </div>
                </div>
                
                <div className="space-y-3">
                  {analysisResults.risks.riskCategories?.map((risk: any, idx: number) => (
                    <div key={idx} className="border border-gray-200 dark:border-gray-700 rounded-lg p-3">
                      <div className="flex items-start justify-between mb-2">
                        <h4 className="font-medium text-sm">{risk.category}</h4>
                        <span className={`text-xs px-2 py-1 rounded ${
                          risk.severity === 'critical' ? 'bg-red-100 text-red-800' :
                          risk.severity === 'high' ? 'bg-orange-100 text-orange-800' :
                          risk.severity === 'medium' ? 'bg-yellow-100 text-yellow-800' :
                          'bg-green-100 text-green-800'
                        }`}>
                          {risk.severity}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 dark:text-gray-400">{risk.description}</p>
                      {risk.mitigation && (
                        <p className="text-xs text-blue-600 dark:text-blue-400 mt-2">
                          Mitigation: {risk.mitigation}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              </>
            ) : (
              <p className="text-sm text-gray-500">No risk analysis available</p>
            )}
          </div>
        );

      case 'entities':
        return (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold mb-3">Extracted Entities</h3>
            {analysisResults.entities ? (
              <div className="space-y-4">
                {analysisResults.entities.organizations?.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium mb-2 flex items-center">
                      <Building className="h-4 w-4 mr-1" />
                      Organizations
                    </h4>
                    <div className="flex flex-wrap gap-2">
                      {analysisResults.entities.organizations.map((org: any, idx: number) => (
                        <span key={idx} className="text-xs bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 px-2 py-1 rounded">
                          {org.text}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
                
                {analysisResults.entities.monetaryValues?.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium mb-2 flex items-center">
                      <DollarSign className="h-4 w-4 mr-1" />
                      Monetary Values
                    </h4>
                    <div className="flex flex-wrap gap-2">
                      {analysisResults.entities.monetaryValues.map((value: any, idx: number) => (
                        <span key={idx} className="text-xs bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 px-2 py-1 rounded">
                          {value.text}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
                
                {analysisResults.entities.dates?.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium mb-2 flex items-center">
                      <Calendar className="h-4 w-4 mr-1" />
                      Dates
                    </h4>
                    <div className="flex flex-wrap gap-2">
                      {analysisResults.entities.dates.map((date: any, idx: number) => (
                        <span key={idx} className="text-xs bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 px-2 py-1 rounded">
                          {date.text}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <p className="text-sm text-gray-500">No entities extracted</p>
            )}
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 w-full h-full max-w-7xl max-h-[90vh] rounded-lg shadow-xl flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center space-x-4">
            <FileText className="h-5 w-5 text-gray-400" />
            <h2 className="text-lg font-semibold">{documentName}</h2>
          </div>
          
          <div className="flex items-center space-x-2">
            <Button variant="ghost" size="sm">
              <Download className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="sm">
              <Share2 className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="sm">
              <Maximize2 className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="sm" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex items-center border-b border-gray-200 dark:border-gray-700 px-4">
          <button
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === 'preview' 
                ? 'text-blue-600 border-blue-600' 
                : 'text-gray-500 border-transparent hover:text-gray-700'
            }`}
            onClick={() => setActiveTab('preview')}
          >
            Document Preview
          </button>
          <button
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === 'analysis' 
                ? 'text-blue-600 border-blue-600' 
                : 'text-gray-500 border-transparent hover:text-gray-700'
            }`}
            onClick={() => setActiveTab('analysis')}
          >
            AI Analysis
          </button>
          {isAnalyzing && (
            <div className="ml-4 flex items-center text-sm text-gray-500">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
              Analyzing...
            </div>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 flex overflow-hidden">
          {activeTab === 'preview' ? (
            <div className="flex-1 p-6 overflow-y-auto">
              <div className="prose dark:prose-invert max-w-none">
                {documentContent}
              </div>
            </div>
          ) : (
            <>
              {/* Analysis Sidebar */}
              <div className="w-64 bg-gray-50 dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 p-4">
                <h3 className="text-sm font-semibold mb-3">Analysis Types</h3>
                <div className="space-y-2">
                  <button
                    className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                      analysisType === 'overview' 
                        ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' 
                        : 'hover:bg-gray-100 dark:hover:bg-gray-700'
                    }`}
                    onClick={() => setAnalysisType('overview')}
                  >
                    <Brain className="h-4 w-4 inline mr-2" />
                    Overview
                  </button>
                  
                  <button
                    className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                      analysisType === 'financial' 
                        ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' 
                        : 'hover:bg-gray-100 dark:hover:bg-gray-700'
                    }`}
                    onClick={() => setAnalysisType('financial')}
                  >
                    <TrendingUp className="h-4 w-4 inline mr-2" />
                    Financial Data
                  </button>
                  
                  <button
                    className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                      analysisType === 'risks' 
                        ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' 
                        : 'hover:bg-gray-100 dark:hover:bg-gray-700'
                    }`}
                    onClick={() => setAnalysisType('risks')}
                  >
                    <AlertTriangle className="h-4 w-4 inline mr-2" />
                    Risk Analysis
                  </button>
                  
                  <button
                    className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                      analysisType === 'entities' 
                        ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' 
                        : 'hover:bg-gray-100 dark:hover:bg-gray-700'
                    }`}
                    onClick={() => setAnalysisType('entities')}
                  >
                    <Building className="h-4 w-4 inline mr-2" />
                    Entities
                  </button>
                </div>
                
                <div className="mt-6">
                  <Button 
                    className="w-full" 
                    size="sm"
                    onClick={analyzeDocument}
                    disabled={isAnalyzing}
                  >
                    Re-analyze Document
                  </Button>
                </div>
              </div>

              {/* Analysis Content */}
              <div className="flex-1 p-6 overflow-y-auto">
                {renderAnalysisContent()}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
} 