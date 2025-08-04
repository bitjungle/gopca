import React from 'react';

interface ImportProgressProps {
    progress: number;
}

export const ImportProgress: React.FC<ImportProgressProps> = ({ progress }) => {
    return (
        <div className="flex flex-col items-center justify-center py-12">
            <div className="mb-8">
                <svg className="w-16 h-16 text-blue-600 dark:text-blue-400 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
            </div>
            
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
                Importing Data...
            </h3>
            
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-6">
                Please wait while we import your file
            </p>
            
            <div className="w-full max-w-md">
                <div className="relative pt-1">
                    <div className="flex mb-2 items-center justify-between">
                        <div>
                            <span className="text-xs font-semibold inline-block py-1 px-2 uppercase rounded-full text-blue-600 bg-blue-200 dark:text-blue-200 dark:bg-blue-800">
                                Progress
                            </span>
                        </div>
                        <div className="text-right">
                            <span className="text-xs font-semibold inline-block text-blue-600 dark:text-blue-400">
                                {progress}%
                            </span>
                        </div>
                    </div>
                    <div className="overflow-hidden h-2 mb-4 text-xs flex rounded bg-gray-200 dark:bg-gray-700">
                        <div 
                            style={{ width: `${progress}%` }} 
                            className="shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-blue-600 dark:bg-blue-400 transition-all duration-300"
                        />
                    </div>
                </div>
            </div>
            
            <div className="mt-4 text-xs text-gray-500 dark:text-gray-400">
                {progress < 20 && "Reading file..."}
                {progress >= 20 && progress < 40 && "Parsing data..."}
                {progress >= 40 && progress < 60 && "Validating columns..."}
                {progress >= 60 && progress < 80 && "Processing rows..."}
                {progress >= 80 && progress < 100 && "Finalizing import..."}
                {progress === 100 && "Import complete!"}
            </div>
        </div>
    );
};