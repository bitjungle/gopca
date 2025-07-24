import React from 'react';

export const MatrixIllustration: React.FC = () => {
  return (
    <div className="flex flex-col items-center justify-center h-full">
      <svg 
        width="320" 
        height="200" 
        viewBox="0 0 320 200" 
        className="w-full max-w-[320px] h-auto"
      >
        {/* Background grid */}
        <rect x="70" y="40" width="200" height="120" fill="none" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        
        {/* Column headers */}
        <text x="120" y="30" textAnchor="middle" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Feature 1</text>
        <text x="170" y="30" textAnchor="middle" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Feature 2</text>
        <text x="220" y="30" textAnchor="middle" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Feature 3</text>
        <text x="260" y="30" textAnchor="middle" className="text-xs fill-gray-500 dark:fill-gray-500">...</text>
        
        {/* Row headers */}
        <text x="65" y="65" textAnchor="end" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Sample 1</text>
        <text x="65" y="95" textAnchor="end" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Sample 2</text>
        <text x="65" y="125" textAnchor="end" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Sample 3</text>
        <text x="65" y="155" textAnchor="end" className="text-xs fill-gray-500 dark:fill-gray-500">...</text>
        
        {/* Grid lines */}
        <line x1="70" y1="70" x2="270" y2="70" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="70" y1="100" x2="270" y2="100" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="70" y1="130" x2="270" y2="130" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        
        <line x1="145" y1="40" x2="145" y2="160" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="195" y1="40" x2="195" y2="160" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="245" y1="40" x2="245" y2="160" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        
        {/* Data values */}
        <text x="107" y="60" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">5.1</text>
        <text x="170" y="60" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">3.5</text>
        <text x="220" y="60" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">1.4</text>
        
        <text x="107" y="90" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">4.9</text>
        <text x="170" y="90" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">3.0</text>
        <text x="220" y="90" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">1.4</text>
        
        <text x="107" y="120" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">4.7</text>
        <text x="170" y="120" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">3.2</text>
        <text x="220" y="120" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">1.3</text>
        
        {/* Arrow indicators */}
        <path d="M 280 100 L 290 100 L 285 95 M 290 100 L 285 105" stroke="#3b82f6" strokeWidth="2" fill="none" className="dark:stroke-blue-400"/>
        <text x="295" y="103" className="text-xs fill-blue-600 dark:fill-blue-400 font-medium">Rows</text>
        
        <path d="M 170 170 L 170 180 L 165 175 M 170 180 L 175 175" stroke="#3b82f6" strokeWidth="2" fill="none" className="dark:stroke-blue-400"/>
        <text x="170" y="195" textAnchor="middle" className="text-xs fill-blue-600 dark:fill-blue-400 font-medium">Columns</text>
      </svg>
      
      <div className="mt-4 text-center">
        <p className="text-sm text-gray-600 dark:text-gray-400">
          CSV format: first row contains feature names,<br/>
          first column contains sample names
        </p>
      </div>
    </div>
  );
};