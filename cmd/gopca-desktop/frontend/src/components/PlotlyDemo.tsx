// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState, useMemo } from 'react';
import { useChartConfig } from '@gopca/ui-components';
import { PlotlyScoresPlot } from '../../../../packages/ui-components/src/charts/adapters/plotly/PlotlyScoresPlot';
import { ScoresPlot } from './visualizations/ScoresPlot';

interface PlotlyDemoProps {
  pcaResult: any;
  rowNames: string[];
  groupColumn?: string | null;
  groupLabels?: string[];
  groupEllipses?: any;
  showEllipses?: boolean;
  confidenceLevel?: 0.90 | 0.95 | 0.99;
}

export const PlotlyDemo: React.FC<PlotlyDemoProps> = ({
  pcaResult,
  rowNames,
  groupColumn,
  groupLabels,
  groupEllipses,
  showEllipses = false,
  confidenceLevel = 0.95,
}) => {
  const { config, setProvider } = useChartConfig();
  const [selectedProvider, setSelectedProvider] = useState<'recharts' | 'plotly'>('recharts');
  
  // Transform data for Plotly
  const plotlyData = useMemo(() => {
    if (!pcaResult?.scores) return [];
    
    return pcaResult.scores.map((row: number[], index: number) => ({
      x: row[0] || 0,
      y: row[1] || 0,
      name: rowNames[index] || `Sample ${index + 1}`,
      group: groupLabels?.[index] || 'Default',
      index: index,
    }));
  }, [pcaResult, rowNames, groupLabels]);
  
  // Create color map for groups
  const groupColorMap = useMemo(() => {
    if (!groupLabels) return new Map();
    
    const uniqueGroups = Array.from(new Set(groupLabels));
    const colors = [
      '#3B82F6', // Blue
      '#EF4444', // Red
      '#10B981', // Green
      '#F59E0B', // Amber
      '#8B5CF6', // Purple
      '#EC4899', // Pink
      '#14B8A6', // Teal
      '#F97316', // Orange
    ];
    
    const map = new Map<string, string>();
    uniqueGroups.forEach((group, index) => {
      map.set(group, colors[index % colors.length]);
    });
    
    return map;
  }, [groupLabels]);
  
  const handleProviderChange = (provider: 'recharts' | 'plotly') => {
    setSelectedProvider(provider);
    setProvider(provider);
  };
  
  if (!pcaResult) {
    return <div>No PCA results available</div>;
  }
  
  const xVariance = pcaResult.explained_variance_ratio[0]?.toFixed(1) || '0';
  const yVariance = pcaResult.explained_variance_ratio[1]?.toFixed(1) || '0';
  const xLabel = `PC1 (${xVariance}%)`;
  const yLabel = `PC2 (${yVariance}%)`;
  
  return (
    <div className="w-full h-full flex flex-col">
      <div className="p-4 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Chart Library Comparison</h2>
          <div className="flex gap-2">
            <button
              onClick={() => handleProviderChange('recharts')}
              className={`px-4 py-2 rounded ${
                selectedProvider === 'recharts'
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
              }`}
            >
              Recharts (Current)
            </button>
            <button
              onClick={() => handleProviderChange('plotly')}
              className={`px-4 py-2 rounded ${
                selectedProvider === 'plotly'
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
              }`}
            >
              Plotly.js (New)
            </button>
          </div>
        </div>
        <div className="mt-2 text-sm text-gray-600 dark:text-gray-400">
          {selectedProvider === 'recharts' ? (
            <p>❌ Known issues: Broken zoom/pan, incorrect ellipse positioning</p>
          ) : (
            <p>✅ Features: Scroll wheel zoom, correct ellipse rendering, better performance</p>
          )}
        </div>
      </div>
      
      <div className="flex-1 p-4">
        {selectedProvider === 'recharts' ? (
          <ScoresPlot
            pcaResult={pcaResult}
            rowNames={rowNames}
            xComponent={0}
            yComponent={1}
            groupColumn={groupColumn}
            groupLabels={groupLabels}
            groupEllipses={groupEllipses}
            showEllipses={showEllipses}
            confidenceLevel={confidenceLevel}
            showRowLabels={false}
            maxLabelsToShow={10}
          />
        ) : (
          <PlotlyScoresPlot
            data={plotlyData}
            xLabel={xLabel}
            yLabel={yLabel}
            groupEllipses={groupEllipses}
            showEllipses={showEllipses}
            confidenceLevel={confidenceLevel}
            groupColorMap={groupColorMap}
            showRowLabels={false}
            maxLabelsToShow={10}
          />
        )}
      </div>
    </div>
  );
};