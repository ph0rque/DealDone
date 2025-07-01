import React, { useState, useEffect } from 'react';
import { CheckFolderWritePermission, GetDefaultDealDonePath, SetDealDoneRoot } from '../../wailsjs/go/main/App';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { FolderOpen, CheckCircle, AlertCircle } from 'lucide-react';
import { Dialog } from './ui/dialog';

interface FirstRunSetupProps {
  onComplete: () => void;
}

export const FirstRunSetup: React.FC<FirstRunSetupProps> = ({ onComplete }) => {
  const [selectedPath, setSelectedPath] = useState('');
  const [isValidPath, setIsValidPath] = useState<boolean | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Get default path on mount
    GetDefaultDealDonePath().then((path: string) => {
      setSelectedPath(path);
      checkPath(path);
    });
  }, []);

  const checkPath = async (path: string) => {
    if (!path) {
      setIsValidPath(null);
      return;
    }
    
    try {
      const hasPermission = await CheckFolderWritePermission(path);
      setIsValidPath(hasPermission);
    } catch (err) {
      setIsValidPath(false);
    }
  };

  const handlePathChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newPath = e.target.value;
    setSelectedPath(newPath);
    setError(null);
    checkPath(newPath);
  };

  const handleSelectFolder = async () => {
    try {
      // Note: In a real implementation, you'd use the Wails dialog API
      // For now, we'll just use the default path
      const result = await (window as any).runtime.OpenDirectoryDialog({
        Title: "Select DealDone Location",
        DefaultDirectory: selectedPath,
      });
      
      if (result) {
        setSelectedPath(result);
        checkPath(result);
      }
    } catch (err) {
      console.error('Error selecting folder:', err);
    }
  };

  const handleCreateStructure = async () => {
    if (!selectedPath || !isValidPath) return;

    setIsCreating(true);
    setError(null);

    try {
      await SetDealDoneRoot(selectedPath);
      onComplete();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create folder structure');
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800 flex items-center justify-center p-8">
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-2xl w-full p-8">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-4">
            Welcome to DealDone
          </h1>
          <p className="text-lg text-gray-600 dark:text-gray-300">
            Let's set up your DealDone workspace
          </p>
        </div>

        <div className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              DealDone Folder Location
            </label>
            <div className="flex gap-2">
              <Input
                type="text"
                value={selectedPath}
                onChange={handlePathChange}
                className="flex-1"
                placeholder="Enter path or browse..."
              />
              <Button
                onClick={handleSelectFolder}
                variant="outline"
                className="px-3"
              >
                <FolderOpen className="h-4 w-4" />
              </Button>
            </div>
            
            {isValidPath !== null && (
              <div className={`mt-2 flex items-center gap-2 text-sm ${
                isValidPath ? 'text-green-600' : 'text-red-600'
              }`}>
                {isValidPath ? (
                  <>
                    <CheckCircle className="h-4 w-4" />
                    Valid location with write permissions
                  </>
                ) : (
                  <>
                    <AlertCircle className="h-4 w-4" />
                    Cannot write to this location
                  </>
                )}
              </div>
            )}
          </div>

          <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
            <h3 className="font-medium text-gray-900 dark:text-white mb-2">
              What will be created:
            </h3>
            <ul className="space-y-1 text-sm text-gray-600 dark:text-gray-300">
              <li>📁 {selectedPath}/</li>
              <li className="ml-4">📁 Templates/ - Store your analysis templates here</li>
              <li className="ml-4">📁 Deals/ - Your deals will be organized here</li>
            </ul>
          </div>

          {error && (
            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
              <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
            </div>
          )}

          <div className="pt-4">
            <Button
              onClick={handleCreateStructure}
              disabled={!isValidPath || isCreating}
              className="w-full py-3 text-lg"
            >
              {isCreating ? 'Creating Structure...' : 'Create DealDone Workspace'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}; 