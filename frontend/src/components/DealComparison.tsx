import React, { useState, useEffect } from 'react';
import { TrendingUp, TrendingDown, Minus, ArrowRight, BarChart3, FileText, X, Plus, Eye, Filter } from 'lucide-react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { useToast } from '../hooks/use-toast';

interface Deal {
  id: string;
  name: string;
  status: 'active' | 'completed' | 'on-hold' | 'cancelled';
  industry: string;
  dealValue: number;
  dealType: 'acquisition' | 'merger' | 'investment' | 'partnership';
  analysisDate: Date;
  documents: number;
  riskScore: number;
  completionPercentage: number;
  keyMetrics: Record<string, number | string>;
  financials: {
    revenue: number;
    ebitda: number;
    netIncome: number;
    assets: number;
    valuation: number;
  };
  riskFactors: string[];
  opportunities: string[];
}

interface DealComparisonProps {
  onClose: () => void;
  preSelectedDeals?: string[];
}

export function DealComparison({ onClose, preSelectedDeals = [] }: DealComparisonProps) {
  const [availableDeals, setAvailableDeals] = useState<Deal[]>([]);
  const [selectedDeals, setSelectedDeals] = useState<string[]>(preSelectedDeals);
  const [searchTerm, setSearchTerm] = useState('');
  const [comparisonView, setComparisonView] = useState<'overview' | 'financial' | 'risk' | 'detailed'>('overview');
  const [isLoading, setIsLoading] = useState(true);
  const { toast } = useToast();

  useEffect(() => {
    loadDeals();
  }, []);

  const loadDeals = async () => {
    try {
      setIsLoading(true);
      
      // Sample deal data
      const sampleDeals: Deal[] = [
        {
          id: '1',
          name: 'TechCorp Acquisition',
          status: 'active',
          industry: 'Technology',
          dealValue: 250000000,
          dealType: 'acquisition',
          analysisDate: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000),
          documents: 45,
          riskScore: 0.35,
          completionPercentage: 78,
          keyMetrics: {
            'Revenue Growth': '15%',
            'Market Share': '8%',
            'Employee Count': 1200,
            'Customer Retention': '94%'
          },
          financials: {
            revenue: 120000000,
            ebitda: 35000000,
            netIncome: 18000000,
            assets: 180000000,
            valuation: 250000000
          },
          riskFactors: ['Market volatility', 'Integration complexity', 'Regulatory changes'],
          opportunities: ['Market expansion', 'Synergy potential', 'Technology IP']
        },
        {
          id: '2',
          name: 'HealthFirst Merger',
          status: 'completed',
          industry: 'Healthcare',
          dealValue: 180000000,
          dealType: 'merger',
          analysisDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
          documents: 67,
          riskScore: 0.42,
          completionPercentage: 100,
          keyMetrics: {
            'Revenue Growth': '8%',
            'Market Share': '12%',
            'Employee Count': 2100,
            'Customer Retention': '89%'
          },
          financials: {
            revenue: 95000000,
            ebitda: 22000000,
            netIncome: 12000000,
            assets: 140000000,
            valuation: 180000000
          },
          riskFactors: ['Regulatory compliance', 'Cultural integration', 'Competitive pressure'],
          opportunities: ['Geographic expansion', 'Service diversification', 'Cost optimization']
        },
        {
          id: '3',
          name: 'GreenEnergy Investment',
          status: 'active',
          industry: 'Energy',
          dealValue: 75000000,
          dealType: 'investment',
          analysisDate: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000),
          documents: 28,
          riskScore: 0.58,
          completionPercentage: 45,
          keyMetrics: {
            'Revenue Growth': '25%',
            'Market Share': '3%',
            'Employee Count': 450,
            'Customer Retention': '91%'
          },
          financials: {
            revenue: 35000000,
            ebitda: 8500000,
            netIncome: 4200000,
            assets: 65000000,
            valuation: 75000000
          },
          riskFactors: ['Technology risk', 'Market adoption', 'Regulatory uncertainty'],
          opportunities: ['ESG focus', 'Government incentives', 'Growing market']
        }
      ];

      setAvailableDeals(sampleDeals);
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

  const selectedDealData = availableDeals.filter(deal => selectedDeals.includes(deal.id));
  const filteredAvailableDeals = availableDeals.filter(deal => 
    !selectedDeals.includes(deal.id) &&
    (searchTerm === '' || deal.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
     deal.industry.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', { 
      style: 'currency', 
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(amount);
  };

  const formatPercentage = (value: number) => {
    return `${(value * 100).toFixed(1)}%`;
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
      case 'completed':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
      case 'on-hold':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200';
      case 'cancelled':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200';
    }
  };

  const getRiskColor = (risk: number) => {
    if (risk <= 0.3) return 'text-green-600 bg-green-50 dark:bg-green-900/20';
    if (risk <= 0.6) return 'text-yellow-600 bg-yellow-50 dark:bg-yellow-900/20';
    return 'text-red-600 bg-red-50 dark:bg-red-900/20';
  };

  const renderTrendIcon = (current: number, comparison: number) => {
    if (current > comparison) return <TrendingUp className="h-4 w-4 text-green-600" />;
    if (current < comparison) return <TrendingDown className="h-4 w-4 text-red-600" />;
    return <Minus className="h-4 w-4 text-gray-600" />;
  };

  const renderOverviewComparison = () => (
    <div className="space-y-6">
      {/* Basic Info */}
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        <div className="p-4 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
          <h3 className="font-semibold">Deal Overview</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 dark:bg-gray-800/50">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium">Metric</th>
                {selectedDealData.map(deal => (
                  <th key={deal.id} className="px-4 py-3 text-left text-sm font-medium">{deal.name}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
              <tr>
                <td className="px-4 py-3 text-sm font-medium">Status</td>
                {selectedDealData.map(deal => (
                  <td key={deal.id} className="px-4 py-3">
                    <span className={`inline-block px-2 py-1 text-xs rounded ${getStatusColor(deal.status)}`}>
                      {deal.status}
                    </span>
                  </td>
                ))}
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm font-medium">Industry</td>
                {selectedDealData.map(deal => (
                  <td key={deal.id} className="px-4 py-3 text-sm">{deal.industry}</td>
                ))}
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm font-medium">Deal Value</td>
                {selectedDealData.map(deal => (
                  <td key={deal.id} className="px-4 py-3 text-sm font-medium">{formatCurrency(deal.dealValue)}</td>
                ))}
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm font-medium">Type</td>
                {selectedDealData.map(deal => (
                  <td key={deal.id} className="px-4 py-3 text-sm capitalize">{deal.dealType}</td>
                ))}
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm font-medium">Risk Score</td>
                {selectedDealData.map(deal => (
                  <td key={deal.id} className="px-4 py-3">
                    <span className={`inline-block px-2 py-1 text-xs rounded ${getRiskColor(deal.riskScore)}`}>
                      {formatPercentage(deal.riskScore)}
                    </span>
                  </td>
                ))}
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm font-medium">Documents</td>
                {selectedDealData.map(deal => (
                  <td key={deal.id} className="px-4 py-3 text-sm">{deal.documents}</td>
                ))}
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        <div className="p-4 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
          <h3 className="font-semibold">Key Metrics</h3>
        </div>
        <div className="p-4 space-y-4">
          {Object.keys(selectedDealData[0]?.keyMetrics || {}).map((metric) => (
            <div key={metric} className="flex items-center justify-between">
              <span className="text-sm font-medium">{metric}:</span>
              <div className="flex items-center space-x-4">
                {selectedDealData.map((deal, idx) => (
                  <div key={deal.id} className="flex items-center space-x-2">
                    <span className="text-sm">{deal.keyMetrics[metric]}</span>
                    {idx < selectedDealData.length - 1 && (
                      <ArrowRight className="h-3 w-3 text-gray-400" />
                    )}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );

  const renderFinancialComparison = () => (
    <div className="space-y-6">
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        <div className="p-4 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
          <h3 className="font-semibold">Financial Comparison</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 dark:bg-gray-800/50">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium">Financial Metric</th>
                {selectedDealData.map(deal => (
                  <th key={deal.id} className="px-4 py-3 text-left text-sm font-medium">{deal.name}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
              {[
                { key: 'revenue', label: 'Revenue' },
                { key: 'ebitda', label: 'EBITDA' },
                { key: 'netIncome', label: 'Net Income' },
                { key: 'assets', label: 'Total Assets' },
                { key: 'valuation', label: 'Valuation' }
              ].map(({ key, label }) => (
                <tr key={key}>
                  <td className="px-4 py-3 text-sm font-medium">{label}</td>
                  {selectedDealData.map((deal, idx) => (
                    <td key={deal.id} className="px-4 py-3">
                      <div className="flex items-center space-x-2">
                        <span className="text-sm font-medium">
                          {formatCurrency(deal.financials[key as keyof typeof deal.financials])}
                        </span>
                        {idx > 0 && renderTrendIcon(
                          deal.financials[key as keyof typeof deal.financials],
                          selectedDealData[0].financials[key as keyof typeof deal.financials]
                        )}
                      </div>
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );

  const renderRiskComparison = () => (
    <div className="space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Risk Factors */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="p-4 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
            <h3 className="font-semibold">Risk Factors</h3>
          </div>
          <div className="p-4 space-y-4">
            {selectedDealData.map(deal => (
              <div key={deal.id}>
                <h4 className="font-medium mb-2">{deal.name}</h4>
                <ul className="space-y-1">
                  {deal.riskFactors.map((risk, idx) => (
                    <li key={idx} className="text-sm text-red-600 flex items-center space-x-2">
                      <div className="w-1 h-1 bg-red-600 rounded-full" />
                      <span>{risk}</span>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>

        {/* Opportunities */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="p-4 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
            <h3 className="font-semibold">Opportunities</h3>
          </div>
          <div className="p-4 space-y-4">
            {selectedDealData.map(deal => (
              <div key={deal.id}>
                <h4 className="font-medium mb-2">{deal.name}</h4>
                <ul className="space-y-1">
                  {deal.opportunities.map((opportunity, idx) => (
                    <li key={idx} className="text-sm text-green-600 flex items-center space-x-2">
                      <div className="w-1 h-1 bg-green-600 rounded-full" />
                      <span>{opportunity}</span>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );

  if (isLoading) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-white dark:bg-gray-900 rounded-lg p-8">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="text-center mt-4">Loading deals...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 w-full h-full max-w-7xl max-h-[95vh] rounded-lg shadow-xl flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div>
            <h2 className="text-2xl font-bold">Deal Comparison</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              Compare multiple deals side by side
            </p>
          </div>
          <Button variant="ghost" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        <div className="flex flex-1 overflow-hidden">
          {/* Sidebar - Deal Selection */}
          <div className="w-80 bg-gray-50 dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 overflow-y-auto">
            <div className="p-4 space-y-4">
              <div>
                <h3 className="font-semibold mb-2">Selected Deals ({selectedDeals.length})</h3>
                <div className="space-y-2">
                  {selectedDealData.map(deal => (
                    <div key={deal.id} className="flex items-center justify-between p-2 bg-blue-50 dark:bg-blue-900/20 rounded">
                      <span className="text-sm font-medium">{deal.name}</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setSelectedDeals(prev => prev.filter(id => id !== deal.id))}
                      >
                        <X className="h-3 w-3" />
                      </Button>
                    </div>
                  ))}
                </div>
              </div>

              <div>
                <h3 className="font-semibold mb-2">Available Deals</h3>
                <Input
                  type="text"
                  placeholder="Search deals..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="mb-2"
                />
                <div className="space-y-2">
                  {filteredAvailableDeals.map(deal => (
                    <div key={deal.id} className="flex items-center justify-between p-2 border border-gray-200 dark:border-gray-700 rounded">
                      <div>
                        <div className="text-sm font-medium">{deal.name}</div>
                        <div className="text-xs text-gray-500">{deal.industry} • {formatCurrency(deal.dealValue)}</div>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          if (selectedDeals.length < 4) {
                            setSelectedDeals(prev => [...prev, deal.id]);
                          } else {
                            toast({
                              title: "Maximum reached",
                              description: "You can compare up to 4 deals at once",
                              variant: "destructive",
                            });
                          }
                        }}
                        disabled={selectedDeals.length >= 4}
                      >
                        <Plus className="h-3 w-3" />
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Main Content */}
          <div className="flex-1 flex flex-col">
            {/* View Tabs */}
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <div className="flex space-x-1">
                {[
                  { id: 'overview', label: 'Overview', icon: <Eye className="h-4 w-4" /> },
                  { id: 'financial', label: 'Financial', icon: <BarChart3 className="h-4 w-4" /> },
                  { id: 'risk', label: 'Risk Analysis', icon: <Filter className="h-4 w-4" /> }
                ].map(({ id, label, icon }) => (
                  <button
                    key={id}
                    onClick={() => setComparisonView(id as any)}
                    className={`flex items-center space-x-2 px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                      comparisonView === id
                        ? 'text-blue-600 border-blue-600'
                        : 'text-gray-500 border-transparent hover:text-gray-700'
                    }`}
                  >
                    {icon}
                    <span>{label}</span>
                  </button>
                ))}
              </div>
            </div>

            {/* Comparison Content */}
            <div className="flex-1 overflow-y-auto p-6">
              {selectedDeals.length < 2 ? (
                <div className="text-center py-12">
                  <BarChart3 className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
                    Select deals to compare
                  </h3>
                  <p className="text-gray-500 dark:text-gray-400">
                    Choose at least 2 deals from the sidebar to start comparing
                  </p>
                </div>
              ) : (
                <>
                  {comparisonView === 'overview' && renderOverviewComparison()}
                  {comparisonView === 'financial' && renderFinancialComparison()}
                  {comparisonView === 'risk' && renderRiskComparison()}
                </>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 