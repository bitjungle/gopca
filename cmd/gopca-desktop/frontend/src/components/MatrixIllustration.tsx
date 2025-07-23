import React from 'react';

export const MatrixIllustration: React.FC = () => {
  return (
    <div className="flex flex-col items-center justify-center h-full">
      <svg 
        width="280" 
        height="200" 
        viewBox="0 0 280 200" 
        className="w-full max-w-[280px] h-auto"
      >
        {/* Background grid */}
        <rect x="40" y="40" width="200" height="120" fill="none" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        
        {/* Column headers */}
        <text x="90" y="30" textAnchor="middle" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Feature 1</text>
        <text x="140" y="30" textAnchor="middle" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Feature 2</text>
        <text x="190" y="30" textAnchor="middle" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Feature 3</text>
        <text x="230" y="30" textAnchor="middle" className="text-xs fill-gray-500 dark:fill-gray-500">...</text>
        
        {/* Row headers */}
        <text x="35" y="65" textAnchor="end" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Sample 1</text>
        <text x="35" y="95" textAnchor="end" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Sample 2</text>
        <text x="35" y="125" textAnchor="end" className="text-xs fill-gray-600 dark:fill-gray-400 font-medium">Sample 3</text>
        <text x="35" y="155" textAnchor="end" className="text-xs fill-gray-500 dark:fill-gray-500">...</text>
        
        {/* Grid lines */}
        <line x1="40" y1="70" x2="240" y2="70" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="40" y1="100" x2="240" y2="100" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="40" y1="130" x2="240" y2="130" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        
        <line x1="115" y1="40" x2="115" y2="160" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="165" y1="40" x2="165" y2="160" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        <line x1="215" y1="40" x2="215" y2="160" stroke="#e5e7eb" strokeWidth="1" className="dark:stroke-gray-600"/>
        
        {/* Data values */}
        <text x="77" y="60" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">5.1</text>
        <text x="140" y="60" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">3.5</text>
        <text x="190" y="60" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">1.4</text>
        
        <text x="77" y="90" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">4.9</text>
        <text x="140" y="90" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">3.0</text>
        <text x="190" y="90" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">1.4</text>
        
        <text x="77" y="120" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">4.7</text>
        <text x="140" y="120" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">3.2</text>
        <text x="190" y="120" textAnchor="middle" className="text-xs fill-gray-700 dark:fill-gray-300">1.3</text>
        
        {/* Arrow indicators */}
        <path d="M 250 100 L 260 100 L 255 95 M 260 100 L 255 105" stroke="#3b82f6" strokeWidth="2" fill="none" className="dark:stroke-blue-400"/>
        <text x="265" y="103" className="text-xs fill-blue-600 dark:fill-blue-400 font-medium">Rows</text>
        
        <path d="M 140 170 L 140 180 L 135 175 M 140 180 L 145 175" stroke="#3b82f6" strokeWidth="2" fill="none" className="dark:stroke-blue-400"/>
        <text x="140" y="195" textAnchor="middle" className="text-xs fill-blue-600 dark:fill-blue-400 font-medium">Columns</text>
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