// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useEffect, useState } from 'react';
import { PCAResult } from '../types';
import { CalculateModelMetrics } from '../../wailsjs/go/main/App';
import { HelpWrapper } from './HelpWrapper';

interface ModelOverviewProps {
  pcaResult: PCAResult;
  selectedPC?: number;
}

interface ModelMetrics {
  mostInfluentialVariable: string;
  loadingValue: number;
  recommendedComponents: number;
}

export const ModelOverview: React.FC<ModelOverviewProps> = ({ pcaResult, selectedPC = 0 }) => {
  const [metrics, setMetrics] = useState<ModelMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMetrics = async () => {
      if (!pcaResult || !pcaResult.loadings || pcaResult.method === 'kernel') {
        setLoading(false);
        return;
      }

      setLoading(true);
      setError(null);

      try {
        const response = await CalculateModelMetrics({
          loadings: pcaResult.loadings,
          explainedVariance: pcaResult.explained_variance_ratio,
          variableLabels: pcaResult.variable_labels || [],
          selectedPC: selectedPC
        });

        if (response.success) {
          setMetrics({
            mostInfluentialVariable: response.mostInfluentialVariable,
            loadingValue: response.loadingValue,
            recommendedComponents: response.recommendedComponents
          });
        } else {
          setError(response.error || 'Failed to calculate metrics');
        }
      } catch (err) {
        setError('Failed to calculate model metrics');
        console.error('Error calculating model metrics:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchMetrics();
  }, [pcaResult, selectedPC]);

  // Don't render for Kernel PCA as it doesn't have loadings
  if (pcaResult?.method === 'kernel') {
    return null;
  }

  if (loading) {
    return (
      <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4">
        <h3 className="text-lg font-semibold mb-3">Model Overview</h3>
        <div className="animate-pulse">
          <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded mb-2"></div>
          <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded"></div>
        </div>
      </div>
    );
  }

  if (error || !metrics) {
    return (
      <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4">
        <h3 className="text-lg font-semibold mb-3">Model Overview</h3>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          {error || 'Unable to calculate metrics'}
        </p>
      </div>
    );
  }

  return (
    <HelpWrapper helpKey="model-overview">
      <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4">
        <div className="mb-2">
          <h3 className="text-lg font-semibold">Model Overview</h3>
        </div>
        <div className="space-y-2">
          <HelpWrapper helpKey="most-influential-variable">
            <div className="flex justify-between items-start">
              <span className="text-sm text-gray-600 dark:text-gray-300">
                Most influential variable:
              </span>
              <div className="text-right">
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {metrics.mostInfluentialVariable}
                </span>
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  Loading: {metrics.loadingValue.toFixed(3)}
                </div>
              </div>
            </div>
          </HelpWrapper>
          
          <HelpWrapper helpKey="recommended-components">
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600 dark:text-gray-300">
                Recommended components:
              </span>
              <div className="text-right">
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {metrics.recommendedComponents}
                </span>
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  Kaiser criterion
                </div>
              </div>
            </div>
          </HelpWrapper>
        </div>
      </div>
    </HelpWrapper>
  );
};