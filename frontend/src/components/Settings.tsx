import React, { useState, useEffect } from 'react';
import {
  Settings as SettingsIcon,
  Brain,
  FolderOpen,
  Eye,
  Shield,
  Download,
  Upload,
  Save,
  X,
  Check,
  AlertCircle,
  Key,
  Globe,
  FileText,
  Database
} from 'lucide-react';
import { 
  GetAIConfig,
  SaveAIConfig,
  GetAppConfig,
  SaveAppConfig,
  GetAvailableAIProviders,
  TestAIProvider,
  ExportAIConfig,
  ImportAIConfig
} from '../../wailsjs/go/main/App';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { useToast } from '../hooks/use-toast';

interface SettingsProps {
  isOpen: boolean;
  onClose: () => void;
}

type SettingsTab = 'ai' | 'folders' | 'analysis' | 'security';

export function Settings({ isOpen, onClose }: SettingsProps) {
  const [activeTab, setActiveTab] = useState<SettingsTab>('ai');
  const [isSaving, setIsSaving] = useState(false);
  const [isDirty, setIsDirty] = useState(false);
  const { toast } = useToast();

  // AI Settings
  const [aiProvider, setAIProvider] = useState('openai');
  const [apiKey, setApiKey] = useState('');
  const [modelName, setModelName] = useState('gpt-4');
  const [maxTokens, setMaxTokens] = useState(4000);
  const [temperature, setTemperature] = useState(0.7);
  const [enableCache, setEnableCache] = useState(true);
  const [cacheExpiry, setCacheExpiry] = useState(3600);

  // Folder Settings
  const [dealDonePath, setDealDonePath] = useState('');
  const [autoCreateFolders, setAutoCreateFolders] = useState(true);
  const [folderStructure, setFolderStructure] = useState('standard');

  // Analysis Settings
  const [autoAnalyze, setAutoAnalyze] = useState(true);
  const [extractFinancial, setExtractFinancial] = useState(true);
  const [extractRisks, setExtractRisks] = useState(true);
  const [extractEntities, setExtractEntities] = useState(true);
  const [confidenceThreshold, setConfidenceThreshold] = useState(0.7);

  // Security Settings
  const [encryptAPIKeys, setEncryptAPIKeys] = useState(true);
  const [allowedFileTypes, setAllowedFileTypes] = useState('pdf,doc,docx,xls,xlsx');
  const [maxFileSize, setMaxFileSize] = useState(50); // MB

  useEffect(() => {
    if (isOpen) {
      loadSettings();
    }
  }, [isOpen]);

  const loadSettings = async () => {
    try {
      // Load AI config
      const aiConfig = await GetAIConfig();
      if (aiConfig) {
        setAIProvider(aiConfig.provider || 'openai');
        setApiKey(aiConfig.apiKey || '');
        setModelName(aiConfig.modelName || 'gpt-4');
        setMaxTokens(aiConfig.maxTokens || 4000);
        setTemperature(aiConfig.temperature || 0.7);
        setEnableCache(aiConfig.enableCache !== false);
        setCacheExpiry(aiConfig.cacheExpiry || 3600);
      }

      // Load app config
      const appConfig = await GetAppConfig();
      if (appConfig) {
        setDealDonePath(appConfig.dealDonePath || '');
        setAutoCreateFolders(appConfig.autoCreateFolders !== false);
        setAutoAnalyze(appConfig.autoAnalyze !== false);
        setExtractFinancial(appConfig.extractFinancial !== false);
        setExtractRisks(appConfig.extractRisks !== false);
        setExtractEntities(appConfig.extractEntities !== false);
        setConfidenceThreshold(appConfig.confidenceThreshold || 0.7);
      }
    } catch (error) {
      console.error('Error loading settings:', error);
      toast({
        title: "Error",
        description: "Failed to load settings",
        variant: "destructive",
      });
    }
  };

  const saveSettings = async () => {
    setIsSaving(true);
    
    try {
      // Save AI config
      await SaveAIConfig({
        provider: aiProvider,
        apiKey: apiKey,
        modelName: modelName,
        maxTokens: maxTokens,
        temperature: temperature,
        enableCache: enableCache,
        cacheExpiry: cacheExpiry,
      });

      // Save app config
      await SaveAppConfig({
        dealDonePath: dealDonePath,
        autoCreateFolders: autoCreateFolders,
        autoAnalyze: autoAnalyze,
        extractFinancial: extractFinancial,
        extractRisks: extractRisks,
        extractEntities: extractEntities,
        confidenceThreshold: confidenceThreshold,
        encryptAPIKeys: encryptAPIKeys,
        allowedFileTypes: allowedFileTypes,
        maxFileSize: maxFileSize,
      });

      setIsDirty(false);
      toast({
        title: "Success",
        description: "Settings saved successfully",
      });
    } catch (error) {
      console.error('Error saving settings:', error);
      toast({
        title: "Error",
        description: "Failed to save settings",
        variant: "destructive",
      });
    } finally {
      setIsSaving(false);
    }
  };

  const testAIConnection = async () => {
    try {
      const result = await TestAIProvider(aiProvider, apiKey);
      if (result.success) {
        toast({
          title: "Success",
          description: "AI provider connection successful",
        });
      } else {
        toast({
          title: "Error",
          description: result.error || "Connection failed",
          variant: "destructive",
        });
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to test AI connection",
        variant: "destructive",
      });
    }
  };

  const exportSettings = async () => {
    try {
      const exported = await ExportAIConfig();
      // In a real app, this would trigger a file download
      toast({
        title: "Success",
        description: "Settings exported successfully",
      });
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to export settings",
        variant: "destructive",
      });
    }
  };

  const importSettings = async () => {
    // In a real app, this would open a file picker
    toast({
      title: "Import Settings",
      description: "File picker would open here",
    });
  };

  const renderTabContent = () => {
    switch (activeTab) {
      case 'ai':
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold mb-4">AI Provider Settings</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-2">AI Provider</label>
                  <select
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800"
                    value={aiProvider}
                    onChange={(e) => {
                      setAIProvider(e.target.value);
                      setIsDirty(true);
                    }}
                  >
                    <option value="openai">OpenAI</option>
                    <option value="claude">Claude (Anthropic)</option>
                    <option value="default">Rule-based (No API)</option>
                  </select>
                </div>

                {aiProvider !== 'default' && (
                  <>
                    <div>
                      <label className="block text-sm font-medium mb-2">API Key</label>
                      <div className="flex space-x-2">
                        <Input
                          type="password"
                          value={apiKey}
                          onChange={(e) => {
                            setApiKey(e.target.value);
                            setIsDirty(true);
                          }}
                          placeholder="Enter your API key"
                        />
                        <Button
                          variant="outline"
                          onClick={testAIConnection}
                        >
                          Test
                        </Button>
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium mb-2">Model</label>
                      <Input
                        value={modelName}
                        onChange={(e) => {
                          setModelName(e.target.value);
                          setIsDirty(true);
                        }}
                        placeholder="e.g., gpt-4, claude-3-opus"
                      />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium mb-2">Max Tokens</label>
                        <Input
                          type="number"
                          value={maxTokens}
                          onChange={(e) => {
                            setMaxTokens(parseInt(e.target.value));
                            setIsDirty(true);
                          }}
                          min="100"
                          max="8000"
                        />
                      </div>

                      <div>
                        <label className="block text-sm font-medium mb-2">Temperature</label>
                        <Input
                          type="number"
                          value={temperature}
                          onChange={(e) => {
                            setTemperature(parseFloat(e.target.value));
                            setIsDirty(true);
                          }}
                          min="0"
                          max="1"
                          step="0.1"
                        />
                      </div>
                    </div>
                  </>
                )}

                <div className="border-t pt-4">
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={enableCache}
                      onChange={(e) => {
                        setEnableCache(e.target.checked);
                        setIsDirty(true);
                      }}
                      className="rounded"
                    />
                    <span className="text-sm">Enable response caching</span>
                  </label>
                  
                  {enableCache && (
                    <div className="mt-2 ml-6">
                      <label className="block text-sm text-gray-600 dark:text-gray-400 mb-1">
                        Cache expiry (seconds)
                      </label>
                      <Input
                        type="number"
                        value={cacheExpiry}
                        onChange={(e) => {
                          setCacheExpiry(parseInt(e.target.value));
                          setIsDirty(true);
                        }}
                        min="60"
                        max="86400"
                        className="w-32"
                      />
                    </div>
                  )}
                </div>
              </div>
            </div>

            <div className="flex justify-end space-x-2 pt-4 border-t">
              <Button variant="outline" onClick={exportSettings}>
                <Download className="h-4 w-4 mr-2" />
                Export
              </Button>
              <Button variant="outline" onClick={importSettings}>
                <Upload className="h-4 w-4 mr-2" />
                Import
              </Button>
            </div>
          </div>
        );

      case 'folders':
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold mb-4">Folder Settings</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-2">DealDone Location</label>
                  <div className="flex space-x-2">
                    <Input
                      value={dealDonePath}
                      onChange={(e) => {
                        setDealDonePath(e.target.value);
                        setIsDirty(true);
                      }}
                      placeholder="/path/to/DealDone"
                      readOnly
                    />
                    <Button variant="outline">
                      Browse
                    </Button>
                  </div>
                  <p className="text-xs text-gray-500 mt-1">
                    This is where your deals and templates are stored
                  </p>
                </div>

                <div>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={autoCreateFolders}
                      onChange={(e) => {
                        setAutoCreateFolders(e.target.checked);
                        setIsDirty(true);
                      }}
                      className="rounded"
                    />
                    <span className="text-sm">Automatically create deal folders</span>
                  </label>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">Folder Structure</label>
                  <select
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800"
                    value={folderStructure}
                    onChange={(e) => {
                      setFolderStructure(e.target.value);
                      setIsDirty(true);
                    }}
                  >
                    <option value="standard">Standard (legal, financial, general, analysis)</option>
                    <option value="custom">Custom</option>
                  </select>
                </div>
              </div>
            </div>
          </div>
        );

      case 'analysis':
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold mb-4">Analysis Settings</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={autoAnalyze}
                      onChange={(e) => {
                        setAutoAnalyze(e.target.checked);
                        setIsDirty(true);
                      }}
                      className="rounded"
                    />
                    <span className="text-sm">Automatically analyze uploaded documents</span>
                  </label>
                </div>

                <div className="pl-6 space-y-2">
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={extractFinancial}
                      onChange={(e) => {
                        setExtractFinancial(e.target.checked);
                        setIsDirty(true);
                      }}
                      className="rounded"
                      disabled={!autoAnalyze}
                    />
                    <span className="text-sm">Extract financial data</span>
                  </label>

                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={extractRisks}
                      onChange={(e) => {
                        setExtractRisks(e.target.checked);
                        setIsDirty(true);
                      }}
                      className="rounded"
                      disabled={!autoAnalyze}
                    />
                    <span className="text-sm">Analyze risks</span>
                  </label>

                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={extractEntities}
                      onChange={(e) => {
                        setExtractEntities(e.target.checked);
                        setIsDirty(true);
                      }}
                      className="rounded"
                      disabled={!autoAnalyze}
                    />
                    <span className="text-sm">Extract entities</span>
                  </label>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">
                    Confidence Threshold
                  </label>
                  <div className="flex items-center space-x-4">
                    <input
                      type="range"
                      min="0"
                      max="1"
                      step="0.05"
                      value={confidenceThreshold}
                      onChange={(e) => {
                        setConfidenceThreshold(parseFloat(e.target.value));
                        setIsDirty(true);
                      }}
                      className="flex-1"
                    />
                    <span className="text-sm font-medium w-12">
                      {(confidenceThreshold * 100).toFixed(0)}%
                    </span>
                  </div>
                  <p className="text-xs text-gray-500 mt-1">
                    Documents below this confidence will be flagged for review
                  </p>
                </div>
              </div>
            </div>
          </div>
        );

      case 'security':
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold mb-4">Security Settings</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={encryptAPIKeys}
                      onChange={(e) => {
                        setEncryptAPIKeys(e.target.checked);
                        setIsDirty(true);
                      }}
                      className="rounded"
                    />
                    <span className="text-sm">Encrypt API keys in storage</span>
                  </label>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">
                    Allowed File Types
                  </label>
                  <Input
                    value={allowedFileTypes}
                    onChange={(e) => {
                      setAllowedFileTypes(e.target.value);
                      setIsDirty(true);
                    }}
                    placeholder="pdf,doc,docx,xls,xlsx"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Comma-separated list of allowed file extensions
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">
                    Maximum File Size (MB)
                  </label>
                  <Input
                    type="number"
                    value={maxFileSize}
                    onChange={(e) => {
                      setMaxFileSize(parseInt(e.target.value));
                      setIsDirty(true);
                    }}
                    min="1"
                    max="500"
                  />
                </div>

                <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
                  <div className="flex items-start space-x-2">
                    <AlertCircle className="h-5 w-5 text-yellow-600 mt-0.5" />
                    <div className="text-sm">
                      <p className="font-medium text-yellow-800 dark:text-yellow-200">
                        Security Notice
                      </p>
                      <p className="text-yellow-600 dark:text-yellow-400 mt-1">
                        API keys are stored locally on your device. Enable encryption for additional security.
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 w-full max-w-4xl max-h-[90vh] rounded-lg shadow-xl flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-2xl font-bold">Settings</h2>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="h-5 w-5" />
          </Button>
        </div>

        {/* Content */}
        <div className="flex flex-1 overflow-hidden">
          {/* Sidebar */}
          <div className="w-64 bg-gray-50 dark:bg-gray-800 p-4">
            <nav className="space-y-2">
              <button
                className={`w-full text-left px-4 py-2 rounded-lg flex items-center space-x-2 transition-colors ${
                  activeTab === 'ai' 
                    ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' 
                    : 'hover:bg-gray-100 dark:hover:bg-gray-700'
                }`}
                onClick={() => setActiveTab('ai')}
              >
                <Brain className="h-5 w-5" />
                <span>AI Configuration</span>
              </button>

              <button
                className={`w-full text-left px-4 py-2 rounded-lg flex items-center space-x-2 transition-colors ${
                  activeTab === 'folders' 
                    ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' 
                    : 'hover:bg-gray-100 dark:hover:bg-gray-700'
                }`}
                onClick={() => setActiveTab('folders')}
              >
                <FolderOpen className="h-5 w-5" />
                <span>Folders</span>
              </button>

              <button
                className={`w-full text-left px-4 py-2 rounded-lg flex items-center space-x-2 transition-colors ${
                  activeTab === 'analysis' 
                    ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' 
                    : 'hover:bg-gray-100 dark:hover:bg-gray-700'
                }`}
                onClick={() => setActiveTab('analysis')}
              >
                <Eye className="h-5 w-5" />
                <span>Analysis</span>
              </button>

              <button
                className={`w-full text-left px-4 py-2 rounded-lg flex items-center space-x-2 transition-colors ${
                  activeTab === 'security' 
                    ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' 
                    : 'hover:bg-gray-100 dark:hover:bg-gray-700'
                }`}
                onClick={() => setActiveTab('security')}
              >
                <Shield className="h-5 w-5" />
                <span>Security</span>
              </button>
            </nav>
          </div>

          {/* Tab Content */}
          <div className="flex-1 p-6 overflow-y-auto">
            {renderTabContent()}
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-6 border-t border-gray-200 dark:border-gray-700">
          <div>
            {isDirty && (
              <p className="text-sm text-yellow-600 dark:text-yellow-400">
                You have unsaved changes
              </p>
            )}
          </div>
          
          <div className="flex space-x-2">
            <Button variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button 
              onClick={saveSettings}
              disabled={!isDirty || isSaving}
            >
              {isSaving ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Saving...
                </>
              ) : (
                <>
                  <Save className="h-4 w-4 mr-2" />
                  Save Changes
                </>
              )}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
} 