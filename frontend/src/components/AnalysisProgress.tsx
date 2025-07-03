import React from 'react';
import { 
  Loader2, 
  FileText, 
  Search, 
  Brain, 
  CheckCircle,
  Clock
} from 'lucide-react';

interface AnalysisProgressProps {
  isVisible: boolean;
  progress: number;
}

interface ProgressStage {
  id: string;
  name: string;
  icon: React.ReactNode;
  minProgress: number;
  maxProgress: number;
}

export function AnalysisProgress({ isVisible, progress }: AnalysisProgressProps) {
  const stages: ProgressStage[] = [
    {
      id: 'prepare',
      name: 'Preparing Analysis',
      icon: <Clock className="h-4 w-4" />,
      minProgress: 0,
      maxProgress: 20
    },
    {
      id: 'discover',
      name: 'Discovering Templates',
      icon: <Search className="h-4 w-4" />,
      minProgress: 20,
      maxProgress: 40
    },
    {
      id: 'extract',
      name: 'Extracting Document Data',
      icon: <FileText className="h-4 w-4" />,
      minProgress: 40,
      maxProgress: 60
    },
    {
      id: 'analyze',
      name: 'AI Analysis & Field Mapping',
      icon: <Brain className="h-4 w-4" />,
      minProgress: 60,
      maxProgress: 80
    },
    {
      id: 'populate',
      name: 'Populating Templates',
      icon: <FileText className="h-4 w-4" />,
      minProgress: 80,
      maxProgress: 95
    },
    {
      id: 'finalize',
      name: 'Finalizing Analysis',
      icon: <CheckCircle className="h-4 w-4" />,
      minProgress: 95,
      maxProgress: 100
    }
  ];

  const getCurrentStage = () => {
    return stages.find(stage => progress >= stage.minProgress && progress < stage.maxProgress) || stages[stages.length - 1];
  };

  const currentStage = getCurrentStage();

  if (!isVisible) return null;

  return (
    <div className="bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 border-b border-blue-200 dark:border-blue-800">
      <div className="px-6 py-4">
        {/* Main Progress Header */}
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center space-x-3">
            <div className="relative">
              <Loader2 className="h-6 w-6 text-blue-600 animate-spin" />
              <div className="absolute inset-0 rounded-full border-2 border-blue-200 dark:border-blue-700"></div>
            </div>
            <div>
              <h3 className="font-semibold text-blue-900 dark:text-blue-100">
                Template Analysis in Progress
              </h3>
              <p className="text-sm text-blue-700 dark:text-blue-300">
                {currentStage.name}
              </p>
            </div>
          </div>
          <div className="text-right">
            <div className="text-2xl font-bold text-blue-600">
              {progress}%
            </div>
            <div className="text-xs text-blue-500 dark:text-blue-400">
              Complete
            </div>
          </div>
        </div>

        {/* Progress Bar */}
        <div className="relative">
          <div className="w-full bg-blue-200 dark:bg-blue-800 rounded-full h-3 overflow-hidden">
            <div 
              className="bg-gradient-to-r from-blue-500 to-indigo-600 h-3 rounded-full transition-all duration-700 ease-out relative"
              style={{ width: `${progress}%` }}
            >
              {/* Animated shine effect */}
              <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent animate-pulse"></div>
            </div>
          </div>
        </div>

        {/* Stage Indicators */}
        <div className="flex justify-between mt-4 px-1">
          {stages.map((stage, index) => {
            const isActive = progress >= stage.minProgress;
            const isCurrent = currentStage.id === stage.id;
            
            return (
              <div 
                key={stage.id}
                className={`flex flex-col items-center space-y-1 transition-all duration-300 ${
                  isActive ? 'opacity-100' : 'opacity-40'
                }`}
              >
                <div className={`p-2 rounded-full transition-all duration-300 ${
                  isCurrent 
                    ? 'bg-blue-600 text-white scale-110 shadow-lg' 
                    : isActive 
                      ? 'bg-blue-100 dark:bg-blue-800 text-blue-600 dark:text-blue-300' 
                      : 'bg-gray-100 dark:bg-gray-700 text-gray-400'
                }`}>
                  {isCurrent ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    stage.icon
                  )}
                </div>
                <span className={`text-xs text-center max-w-16 leading-tight ${
                  isActive 
                    ? 'text-blue-700 dark:text-blue-300 font-medium' 
                    : 'text-gray-500 dark:text-gray-400'
                }`}>
                  {stage.name.split(' ').map((word, i) => (
                    <span key={i} className="block">
                      {word}
                    </span>
                  ))}
                </span>
              </div>
            );
          })}
        </div>

        {/* Additional Info */}
        <div className="mt-4 p-3 bg-blue-100/50 dark:bg-blue-800/30 rounded-lg">
          <p className="text-sm text-blue-800 dark:text-blue-200">
            <span className="font-medium">What's happening:</span> The system is analyzing your documents, 
            discovering relevant templates, and automatically populating them with extracted data. 
            This process includes AI-powered field mapping and formula preservation.
          </p>
        </div>
      </div>
    </div>
  );
} 