import React, { useState, useEffect } from 'react';
import { 
  CheckCircle2, 
  AlertCircle, 
  Clock, 
  FileText,
  Loader2,
  X,
  ChevronDown,
  ChevronUp
} from 'lucide-react';
import { Button } from './ui/button';

interface ProcessingStep {
  id: string;
  name: string;
  status: 'pending' | 'processing' | 'success' | 'error';
  progress?: number;
  message?: string;
  startTime?: Date;
  endTime?: Date;
}

interface ProcessingProgressProps {
  isOpen: boolean;
  onClose?: () => void;
  title?: string;
}

export function ProcessingProgress({ isOpen, onClose, title = "Processing Documents" }: ProcessingProgressProps) {
  const [steps, setSteps] = useState<ProcessingStep[]>([
    { id: 'upload', name: 'Uploading Documents', status: 'pending' },
    { id: 'classify', name: 'Classifying Documents', status: 'pending' },
    { id: 'extract', name: 'Extracting Text', status: 'pending' },
    { id: 'analyze', name: 'AI Analysis', status: 'pending' },
    { id: 'route', name: 'Routing to Folders', status: 'pending' },
    { id: 'complete', name: 'Finalizing', status: 'pending' },
  ]);
  
  const [currentStep, setCurrentStep] = useState(0);
  const [isMinimized, setIsMinimized] = useState(false);
  const [overallProgress, setOverallProgress] = useState(0);

  useEffect(() => {
    if (isOpen) {
      // Simulate processing progress
      simulateProgress();
    }
  }, [isOpen]);

  const simulateProgress = async () => {
    for (let i = 0; i < steps.length; i++) {
      // Update current step to processing
      setSteps(prev => prev.map((step, idx) => 
        idx === i ? { ...step, status: 'processing', startTime: new Date() } : step
      ));
      setCurrentStep(i);
      
      // Simulate progress for current step
      for (let progress = 0; progress <= 100; progress += 10) {
        await new Promise(resolve => setTimeout(resolve, 200));
        
        setSteps(prev => prev.map((step, idx) => 
          idx === i ? { ...step, progress } : step
        ));
        
        // Update overall progress
        const overallProg = ((i * 100 + progress) / steps.length);
        setOverallProgress(Math.round(overallProg));
      }
      
      // Mark step as complete
      setSteps(prev => prev.map((step, idx) => 
        idx === i ? { 
          ...step, 
          status: 'success', 
          endTime: new Date(),
          message: `Completed in ${Math.random() * 2 + 1}s`
        } : step
      ));
    }
  };

  const getStepIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-5 w-5 text-gray-400" />;
      case 'processing':
        return <Loader2 className="h-5 w-5 text-blue-600 animate-spin" />;
      case 'success':
        return <CheckCircle2 className="h-5 w-5 text-green-600" />;
      case 'error':
        return <AlertCircle className="h-5 w-5 text-red-600" />;
      default:
        return <FileText className="h-5 w-5 text-gray-400" />;
    }
  };

  const getStepColor = (status: string) => {
    switch (status) {
      case 'processing':
        return 'text-blue-600 bg-blue-50 dark:bg-blue-900/20';
      case 'success':
        return 'text-green-600 bg-green-50 dark:bg-green-900/20';
      case 'error':
        return 'text-red-600 bg-red-50 dark:bg-red-900/20';
      default:
        return 'text-gray-600 bg-gray-50 dark:bg-gray-800';
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed bottom-4 right-4 w-96 bg-white dark:bg-gray-900 rounded-lg shadow-xl border border-gray-200 dark:border-gray-700 z-50">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center space-x-2">
          <Loader2 className="h-5 w-5 text-blue-600 animate-spin" />
          <h3 className="font-semibold">{title}</h3>
        </div>
        
        <div className="flex items-center space-x-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsMinimized(!isMinimized)}
          >
            {isMinimized ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
          </Button>
          {onClose && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onClose}
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      </div>

      {!isMinimized && (
        <>
          {/* Overall Progress */}
          <div className="p-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-gray-600 dark:text-gray-400">Overall Progress</span>
              <span className="text-sm font-medium">{overallProgress}%</span>
            </div>
            <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
              <div 
                className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                style={{ width: `${overallProgress}%` }}
              />
            </div>
          </div>

          {/* Steps */}
          <div className="p-4 max-h-80 overflow-y-auto">
            <div className="space-y-3">
              {steps.map((step, idx) => (
                <div
                  key={step.id}
                  className={`flex items-start space-x-3 p-3 rounded-lg transition-all ${
                    step.status === 'processing' ? getStepColor(step.status) : ''
                  }`}
                >
                  <div className="mt-0.5">
                    {getStepIcon(step.status)}
                  </div>
                  
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <p className={`text-sm font-medium ${
                        step.status === 'pending' ? 'text-gray-500' : ''
                      }`}>
                        {step.name}
                      </p>
                      {step.message && (
                        <span className="text-xs text-gray-500">{step.message}</span>
                      )}
                    </div>
                    
                    {step.status === 'processing' && step.progress !== undefined && (
                      <div className="mt-2">
                        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1">
                          <div 
                            className="bg-blue-600 h-1 rounded-full transition-all duration-300"
                            style={{ width: `${step.progress}%` }}
                          />
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Summary */}
          {overallProgress === 100 && (
            <div className="p-4 border-t border-gray-200 dark:border-gray-700 bg-green-50 dark:bg-green-900/20">
              <div className="flex items-center space-x-2">
                <CheckCircle2 className="h-5 w-5 text-green-600" />
                <div>
                  <p className="text-sm font-medium text-green-800 dark:text-green-200">
                    Processing Complete
                  </p>
                  <p className="text-xs text-green-600 dark:text-green-400">
                    All documents have been successfully processed
                  </p>
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}

// Inline Progress Bar Component
export function InlineProgress({ 
  progress, 
  label, 
  showPercentage = true,
  size = 'md',
  color = 'blue' 
}: {
  progress: number;
  label?: string;
  showPercentage?: boolean;
  size?: 'sm' | 'md' | 'lg';
  color?: 'blue' | 'green' | 'yellow' | 'red';
}) {
  const heights = {
    sm: 'h-1',
    md: 'h-2',
    lg: 'h-3'
  };

  const colors = {
    blue: 'bg-blue-600',
    green: 'bg-green-600',
    yellow: 'bg-yellow-600',
    red: 'bg-red-600'
  };

  return (
    <div className="w-full">
      {(label || showPercentage) && (
        <div className="flex items-center justify-between mb-1">
          {label && <span className="text-sm text-gray-600 dark:text-gray-400">{label}</span>}
          {showPercentage && <span className="text-sm font-medium">{progress}%</span>}
        </div>
      )}
      <div className={`w-full bg-gray-200 dark:bg-gray-700 rounded-full ${heights[size]}`}>
        <div 
          className={`${colors[color]} ${heights[size]} rounded-full transition-all duration-300`}
          style={{ width: `${Math.min(100, Math.max(0, progress))}%` }}
        />
      </div>
    </div>
  );
}

// Circular Progress Component
export function CircularProgress({ 
  progress, 
  size = 80, 
  strokeWidth = 8,
  showPercentage = true,
  color = 'blue'
}: {
  progress: number;
  size?: number;
  strokeWidth?: number;
  showPercentage?: boolean;
  color?: 'blue' | 'green' | 'yellow' | 'red';
}) {
  const radius = (size - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const offset = circumference - (progress / 100) * circumference;

  const colors = {
    blue: 'stroke-blue-600',
    green: 'stroke-green-600',
    yellow: 'stroke-yellow-600',
    red: 'stroke-red-600'
  };

  return (
    <div className="relative inline-flex items-center justify-center">
      <svg
        className="transform -rotate-90"
        width={size}
        height={size}
      >
        <circle
          className="text-gray-200 dark:text-gray-700"
          strokeWidth={strokeWidth}
          stroke="currentColor"
          fill="transparent"
          r={radius}
          cx={size / 2}
          cy={size / 2}
        />
        <circle
          className={`${colors[color]} transition-all duration-300`}
          strokeWidth={strokeWidth}
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          strokeLinecap="round"
          fill="transparent"
          r={radius}
          cx={size / 2}
          cy={size / 2}
        />
      </svg>
      {showPercentage && (
        <span className="absolute text-sm font-semibold">
          {progress}%
        </span>
      )}
    </div>
  );
} 