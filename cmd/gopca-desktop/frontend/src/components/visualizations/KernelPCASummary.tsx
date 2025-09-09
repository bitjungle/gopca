// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Kernel PCA Summary Component - Model overview for Kernel PCA

import React from 'react';
import { useTheme } from '@gopca/ui-components';
import { PCAResult } from '../../types';

interface KernelPCASummaryProps {
  pcaResult: PCAResult;
}

export const KernelPCASummary: React.FC<KernelPCASummaryProps> = ({ pcaResult }) => {
  const { theme } = useTheme();
  const isDark = theme === 'dark';

  // Format kernel type for display
  const formatKernelType = (type: string): string => {
    switch (type) {
      case 'rbf':
        return 'Radial Basis Function (RBF)';
      case 'linear':
        return 'Linear';
      case 'poly':
        return 'Polynomial';
      default:
        return type;
    }
  };

  // Format parameter value
  const formatParamValue = (value: number): string => {
    if (Number.isInteger(value)) {
      return value.toString();
    }
    return value.toExponential(4);
  };

  // Calculate total variance explained
  const totalVarianceExplained = pcaResult.cumulative_variance?.[pcaResult.cumulative_variance.length - 1] || 0;

  return (
    <div className={`p-6 ${isDark ? 'bg-gray-800 text-gray-100' : 'bg-white text-gray-900'}`}>
      <h2 className="text-2xl font-bold mb-6">Kernel PCA Model Summary</h2>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Kernel Configuration */}
        <div className={`p-4 rounded-lg ${isDark ? 'bg-gray-700' : 'bg-gray-50'}`}>
          <h3 className="text-lg font-semibold mb-3">Kernel Configuration</h3>
          <div className="space-y-2">
            <div>
              <span className="font-medium">Kernel Type:</span>{' '}
              <span className={isDark ? 'text-blue-400' : 'text-blue-600'}>
                {formatKernelType(pcaResult.kernel_type || '')}
              </span>
            </div>
            
            {pcaResult.kernel_params && Object.keys(pcaResult.kernel_params).length > 0 && (
              <div>
                <span className="font-medium">Parameters:</span>
                <div className="ml-4 mt-1 space-y-1">
                  {Object.entries(pcaResult.kernel_params).map(([key, value]) => (
                    <div key={key} className="text-sm">
                      <span className="font-mono">{key}:</span>{' '}
                      <span className={isDark ? 'text-green-400' : 'text-green-600'}>
                        {formatParamValue(value)}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Model Information */}
        <div className={`p-4 rounded-lg ${isDark ? 'bg-gray-700' : 'bg-gray-50'}`}>
          <h3 className="text-lg font-semibold mb-3">Model Information</h3>
          <div className="space-y-2">
            <div>
              <span className="font-medium">Components Extracted:</span>{' '}
              <span className={isDark ? 'text-blue-400' : 'text-blue-600'}>
                {pcaResult.components_computed}
              </span>
            </div>
            <div>
              <span className="font-medium">Total Variance Explained:</span>{' '}
              <span className={isDark ? 'text-green-400' : 'text-green-600'}>
                {totalVarianceExplained.toFixed(2)}%
              </span>
            </div>
            <div>
              <span className="font-medium">Preprocessing Applied:</span>{' '}
              <span className={isDark ? 'text-yellow-400' : 'text-yellow-600'}>
                {pcaResult.preprocessing_applied ? 'Yes' : 'No'}
              </span>
            </div>
          </div>
        </div>

        {/* Component Details */}
        <div className={`p-4 rounded-lg ${isDark ? 'bg-gray-700' : 'bg-gray-50'} md:col-span-2`}>
          <h3 className="text-lg font-semibold mb-3">Component Details</h3>
          <div className="overflow-x-auto">
            <table className="min-w-full">
              <thead>
                <tr className={`border-b ${isDark ? 'border-gray-600' : 'border-gray-200'}`}>
                  <th className="text-left py-2 px-3">Component</th>
                  <th className="text-right py-2 px-3">Eigenvalue</th>
                  <th className="text-right py-2 px-3">Variance (%)</th>
                  <th className="text-right py-2 px-3">Cumulative (%)</th>
                </tr>
              </thead>
              <tbody>
                {pcaResult.component_labels?.map((label, i) => (
                  <tr key={label} className={`border-b ${isDark ? 'border-gray-600' : 'border-gray-200'}`}>
                    <td className="py-2 px-3 font-medium">{label}</td>
                    <td className="text-right py-2 px-3 font-mono text-sm">
                      {pcaResult.explained_variance[i]?.toFixed(4) || '0.0000'}
                    </td>
                    <td className="text-right py-2 px-3">
                      {pcaResult.explained_variance_ratio[i]?.toFixed(2) || '0.00'}%
                    </td>
                    <td className="text-right py-2 px-3 font-medium">
                      {pcaResult.cumulative_variance[i]?.toFixed(2) || '0.00'}%
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Data Information */}
        {pcaResult.kernel_matrix && (
          <div className={`p-4 rounded-lg ${isDark ? 'bg-gray-700' : 'bg-gray-50'} md:col-span-2`}>
            <h3 className="text-lg font-semibold mb-3">Data Information</h3>
            <div className="space-y-2">
              <div>
                <span className="font-medium">Number of Samples:</span>{' '}
                <span className={isDark ? 'text-blue-400' : 'text-blue-600'}>
                  {pcaResult.scores?.length || 0}
                </span>
              </div>
              <div>
                <span className="font-medium">Kernel Matrix Size:</span>{' '}
                <span className={isDark ? 'text-green-400' : 'text-green-600'}>
                  {pcaResult.kernel_matrix?.length || 0} × {pcaResult.kernel_matrix?.[0]?.length || 0}
                </span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};