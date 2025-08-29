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
  standardScale?: boolean;
  originalData?: number[][];
}

interface ModelMetrics {
  mostInfluentialVariable: string;
  loadingValue: number;
  recommendedComponents: number;
  varianceCaptured: number;
  kaiserComponents: number;
  scaleRatio: number;
  scaleWarning?: string;
}

export const ModelOverview: React.FC<ModelOverviewProps> = ({ pcaResult, selectedPC = 0, standardScale = false, originalData }) => {
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
          selectedPC: selectedPC,
          standardScale: standardScale,
          originalData: originalData || []
        });

        if (response.success) {
          setMetrics({
            mostInfluentialVariable: response.mostInfluentialVariable,
            loadingValue: response.loadingValue,
            recommendedComponents: response.recommendedComponents,
            varianceCaptured: response.varianceCaptured,
            kaiserComponents: response.kaiserComponents,
            scaleRatio: response.scaleRatio,
            scaleWarning: response.scaleWarning
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
  }, [pcaResult, selectedPC, standardScale, originalData]);

  // Don't render for Kernel PCA as it doesn't have loadings
  if (pcaResult?.method === 'kernel') {
    return null;
  }

  if (loading) {
    return (
      <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4 h-full">
        <h3 className="text-lg font-semibold mb-3">Model Overview</h3>
        <div className="animate-pulse space-y-3">
          <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded"></div>
          <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded"></div>
          <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded"></div>
        </div>
      </div>
    );
  }

  if (error || !metrics) {
    return (
      <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4 h-full">
        <h3 className="text-lg font-semibold mb-3">Model Overview</h3>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          {error || 'Unable to calculate metrics'}
        </p>
      </div>
    );
  }

  // Format the recommendation subtitle
  const getRecommendationSubtitle = () => {
    if (!metrics) return '';
    
    const varianceText = `${metrics.varianceCaptured.toFixed(1)}% variance`;
    
    // If standardized and Kaiser is available and matches, show it
    if (standardScale && metrics.kaiserComponents > 0 && metrics.kaiserComponents === metrics.recommendedComponents) {
      return `${varianceText} (Kaiser agrees)`;
    } else if (standardScale && metrics.kaiserComponents > 0) {
      return `${varianceText} (Kaiser: ${metrics.kaiserComponents})`;
    }
    
    return varianceText;
  };

  return (
    <HelpWrapper helpKey="model-overview">
      <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4 h-full flex flex-col">
        <div className="mb-2">
          <h3 className="text-lg font-semibold">Model Overview</h3>
        </div>
        <div className="space-y-2 flex-grow">
          <HelpWrapper helpKey="most-influential-variable">
            <div className="flex justify-between items-start">
              <span>Top variable:</span>
              <div className="text-right">
                <span className="font-medium">
                  {metrics.mostInfluentialVariable}
                </span>
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  Loading: {metrics.loadingValue.toFixed(3)}
                </div>
              </div>
            </div>
          </HelpWrapper>
          
          <HelpWrapper helpKey="recommended-components">
            <div className="flex justify-between items-start">
              <span>
                Recommended:
                {metrics.scaleWarning && (
                  <span className="ml-1 text-yellow-600 dark:text-yellow-500" title={metrics.scaleWarning}>
                    ⚠️
                  </span>
                )}
              </span>
              <div className="text-right">
                <span className="font-medium">
                  {metrics.recommendedComponents} components
                </span>
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  {getRecommendationSubtitle()}
                </div>
              </div>
            </div>
          </HelpWrapper>

          <HelpWrapper helpKey="variance-captured">
            <div className="flex justify-between items-start">
              <span>Variance captured:</span>
              <div className="text-right">
                <span className="font-medium">
                  {metrics.varianceCaptured.toFixed(1)}%
                </span>
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  by {metrics.recommendedComponents} PC{metrics.recommendedComponents !== 1 ? 's' : ''}
                </div>
              </div>
            </div>
          </HelpWrapper>

          {/* Show scale warning if present */}
          {metrics.scaleWarning && (
            <div className="mt-2 p-2 bg-yellow-100 dark:bg-yellow-900/30 rounded text-xs text-yellow-800 dark:text-yellow-200">
              ⚠️ {metrics.scaleWarning}
            </div>
          )}
        </div>
      </div>
    </HelpWrapper>
  );
};