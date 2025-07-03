import { useEffect, useState } from "react";
import { ThemeToggle } from "./ThemeToggle";
import { Button } from "./ui/button";
import { toast } from "../hooks/use-toast";
import { Input } from "./ui/input";

// Type for AI provider status
type ProviderStatus = {
  configured: boolean;
  model: string;
  enabled: boolean;
};

type AIProviderStatus = {
  [key: string]: ProviderStatus;
};

// Type for n8n status
type N8nStatus = {
  success: boolean;
  message: string;
  config?: any;
};

const Settings = () => {
  const [aiConfig, setAiConfig] = useState<any>({});
  const [providerStatus, setProviderStatus] = useState<AIProviderStatus | null>(
    null
  );
  const [apiKey, setApiKey] = useState("");
  const [selectedProvider, setSelectedProvider] = useState("openai");
  const [n8nStatus, setN8nStatus] = useState<N8nStatus | null>(null);
  const [isTestingN8n, setIsTestingN8n] = useState(false);
  const [isN8nInitializing, setIsN8nInitializing] = useState(true);

  useEffect(() => {
    fetchAIConfig();
    fetchProviderStatus();
    checkN8nInitialization();
  }, []);

  const fetchAIConfig = async () => {
    try {
      const config = await window.go.main.App.GetAIConfig();
      setAiConfig(config);
    } catch (error) {
      console.error("Failed to fetch AI config:", error);
    }
  };

  const fetchProviderStatus = async () => {
    try {
      const status = await window.go.main.App.GetAIProviderStatus();
      setProviderStatus(status);
    } catch (error) {
      console.error("Failed to fetch provider status:", error);
    }
  };

  const checkN8nInitialization = async () => {
    let attempts = 0;
    const maxAttempts = 30; // Check for up to 15 seconds (30 * 500ms)
    
    const checkStatus = async () => {
      try {
        // Try to get n8n integration status to see if it's initialized
        const status = await window.go.main.App.GetN8nIntegrationStatus();
        if (status && status.isRunning) {
          setIsN8nInitializing(false);
          return;
        }
      } catch (error) {
        // Service might not be initialized yet, continue checking
      }
      
      attempts++;
      if (attempts < maxAttempts) {
        setTimeout(checkStatus, 500); // Check every 500ms
      } else {
        // Give up after max attempts and allow testing anyway
        setIsN8nInitializing(false);
      }
    };
    
    checkStatus();
  };

  const handleSaveAIConfig = async () => {
    try {
      await window.go.main.App.SaveAIConfig(aiConfig);
      toast({
        title: "Success",
        description: "AI configuration saved successfully.",
      });
    } catch (error) {
      console.error("Failed to save AI config:", error);
      toast({
        title: "Error",
        description: "Failed to save AI configuration.",
        variant: "destructive",
      });
    }
  };

  const testN8nConnection = async () => {
    setIsTestingN8n(true);
    setN8nStatus(null);
    try {
      const result = await window.go.main.App.TestN8nConnection();
      // Backend returns { status: "success"/"failed", baseURL, responseTime, timestamp }
      const isSuccess = result.status === "success";
      const statusForUI = { 
        success: isSuccess, 
        message: isSuccess 
          ? `Connected to ${result.baseURL} (${result.responseTime}ms)`
          : result.error || "Failed to connect to n8n."
      };
      
      setN8nStatus(statusForUI);
      
      if (isSuccess) {
        toast({
          title: "Success",
          description: `Successfully connected to n8n instance at ${result.baseURL}`,
        });
      } else {
        toast({
          title: "Error", 
          description: result.error || "Failed to connect to n8n.",
          variant: "destructive",
        });
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "An unknown error occurred.";
      setN8nStatus({ success: false, message: errorMessage });
      toast({
        title: "Error",
        description: `Failed to test n8n connection: ${errorMessage}`,
        variant: "destructive",
      });
    } finally {
      setIsTestingN8n(false);
    }
  };

  return (
    <div className="p-6 bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 h-full overflow-y-auto">
      <h1 className="text-3xl font-bold mb-6">Settings</h1>

      <div className="space-y-8">
        <div>
          <h2 className="text-2xl font-semibold mb-4">Appearance</h2>
          <div className="flex items-center justify-between p-4 border rounded-lg dark:border-gray-600">
            <span>Toggle light or dark mode</span>
            <ThemeToggle />
          </div>
        </div>

        <div>
          <h2 className="text-2xl font-semibold mb-4">AI Configuration</h2>
          <div className="space-y-4 p-4 border rounded-lg dark:border-gray-600">
            {/* AI Provider Status and API Key Input */}
            {/* This section can be expanded with more AI settings */}
            <Button onClick={handleSaveAIConfig}>Save AI Configuration</Button>
          </div>
        </div>
      </div>

      <div className="mt-8 pt-8 border-t border-gray-200 dark:border-gray-700">
        <h2 className="text-2xl font-semibold mb-4">Integrations</h2>
        <div className="p-4 border rounded-lg dark:border-gray-600">
          <h3 className="text-xl font-semibold mb-2">n8n Workflow Automation</h3>
          <p className="mb-4 text-gray-600 dark:text-gray-400">
            Test the connection to your configured n8n instance to ensure
            document processing workflows can be triggered.
          </p>
          {isN8nInitializing && (
            <div className="mb-4 p-3 rounded-md text-sm bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200">
              <p className="font-semibold">n8n Service Starting</p>
              <p>Please wait while the n8n integration service initializes...</p>
            </div>
          )}
          <Button 
            onClick={testN8nConnection} 
            disabled={isTestingN8n || isN8nInitializing}
          >
            {isN8nInitializing 
              ? "Initializing..." 
              : isTestingN8n 
                ? "Testing..." 
                : "Test n8n Connection"
            }
          </Button>
          {n8nStatus && (
            <div
              className={`mt-4 p-3 rounded-md text-sm ${
                n8nStatus.success
                  ? "bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200"
                  : "bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200"
              }`}
            >
              <p className="font-semibold">
                {n8nStatus.success ? "Connection Successful" : "Connection Failed"}
              </p>
              <p>{n8nStatus.message}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Settings;