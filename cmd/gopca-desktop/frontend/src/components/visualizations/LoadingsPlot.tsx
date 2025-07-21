import React, { useState, useEffect } from 'react';
import { 
  BarChart, 
  Bar, 
  LineChart,
  Line,
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  ReferenceLine,
  Cell
} from 'recharts';
import { PCAResult } from '../../types';

interface LoadingsPlotProps {
  pcaResult: PCAResult;
  selectedComponent?: number; // 0-based index
  variableThreshold?: number; // Threshold for auto-switching between bar and line
}

export const LoadingsPlot: React.FC<LoadingsPlotProps> = ({ 
  pcaResult, 
  selectedComponent = 0,
  variableThreshold = 50
}) => {
  // Auto-detect plot type based on number of variables
  const numVariables = pcaResult.loadings[0]?.length || 0;
  const autoPlotType = numVariables > variableThreshold ? 'line' : 'bar';
  
  const [plotType, setPlotType] = useState<'bar' | 'line'>(autoPlotType);

  // Update plot type when data changes
  useEffect(() => {
    setPlotType(numVariables > variableThreshold ? 'line' : 'bar');
  }, [numVariables, variableThreshold]);

  // Extract loadings for selected component
  const loadingsData = pcaResult.loadings.map((row, index) => {
    const loading = row[selectedComponent] || 0;
    return {
      variable: pcaResult.variable_labels?.[index] || `Var${index + 1}`,
      loading: loading,
      absLoading: Math.abs(loading),
      variableIndex: index
    };
  });

  // Sort by absolute loading value for bar chart
  const sortedData = plotType === 'bar' 
    ? [...loadingsData].sort((a, b) => b.absLoading - a.absLoading)
    : loadingsData;

  // Calculate min/max for axis domain
  const loadingValues = sortedData.map(d => d.loading);
  const minLoading = Math.min(...loadingValues);
  const maxLoading = Math.max(...loadingValues);
  const absMax = Math.max(Math.abs(minLoading), Math.abs(maxLoading));
  const padding = absMax * 0.1;

  // Get component label and variance
  const componentLabel = pcaResult.component_labels?.[selectedComponent] || `PC${selectedComponent + 1}`;
  const variance = pcaResult.explained_variance[selectedComponent]?.toFixed(1) || '0';

  // Handle edge cases
  if (!loadingsData || loadingsData.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-400">
        <p>No loadings data to display</p>
      </div>
    );
  }

  // Custom tooltip
  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      return (
        <div className="bg-gray-800 p-3 rounded shadow-lg border border-gray-600">
          <p className="text-white font-semibold">{data.variable}</p>
          <p className="text-blue-400">
            Loading: {data.loading.toFixed(4)}
          </p>
        </div>
      );
    }
    return null;
  };

  return (
    <div className="w-full h-full">
      {/* Header with plot type selector */}
      <div className="flex items-center justify-between mb-4">
        <h4 className="text-md font-medium text-gray-300">
          {componentLabel} Loadings ({variance}% variance)
        </h4>
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-400">Plot type:</span>
          <select
            value={plotType}
            onChange={(e) => setPlotType(e.target.value as 'bar' | 'line')}
            className="px-2 py-1 bg-gray-700 rounded text-sm"
          >
            <option value="bar">Bar Chart</option>
            <option value="line">Line Chart</option>
          </select>
          {plotType !== autoPlotType && (
            <span className="text-xs text-yellow-500">(manual)</span>
          )}
        </div>
      </div>

      <ResponsiveContainer width="100%" height="90%">
        {plotType === 'bar' ? (
          <BarChart 
            data={sortedData} 
            margin={{ top: 20, right: 20, bottom: 60, left: 80 }}
          >
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            
            <XAxis 
              dataKey="variable" 
              stroke="#9CA3AF"
              angle={-45}
              textAnchor="end"
              height={60}
              interval={numVariables <= 20 ? 0 : 'preserveStartEnd'}
            />
            
            <YAxis 
              stroke="#9CA3AF"
              domain={[-absMax - padding, absMax + padding]}
              tickFormatter={(value) => value.toFixed(2)}
            />
            
            <ReferenceLine y={0} stroke="#6B7280" strokeWidth={2} />
            
            <Tooltip content={<CustomTooltip />} />
            
            <Bar 
              dataKey="loading" 
              radius={[4, 4, 0, 0]}
            >
              {sortedData.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.loading >= 0 ? '#3B82F6' : '#EF4444'} />
              ))}
            </Bar>
          </BarChart>
        ) : (
          <LineChart 
            data={sortedData} 
            margin={{ top: 20, right: 20, bottom: 60, left: 80 }}
          >
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            
            <XAxis 
              dataKey="variableIndex" 
              stroke="#9CA3AF"
              label={{ 
                value: 'Variable Index', 
                position: 'insideBottom', 
                offset: -10,
                style: { fill: '#9CA3AF' }
              }}
              domain={[0, numVariables - 1]}
            />
            
            <YAxis 
              stroke="#9CA3AF"
              domain={[-absMax - padding, absMax + padding]}
              tickFormatter={(value) => value.toFixed(2)}
              label={{ 
                value: 'Loading Value', 
                angle: -90, 
                position: 'insideLeft',
                style: { fill: '#9CA3AF' }
              }}
            />
            
            <ReferenceLine y={0} stroke="#6B7280" strokeWidth={2} />
            
            <Tooltip content={<CustomTooltip />} />
            
            <Line 
              type="monotone" 
              dataKey="loading" 
              stroke="#3B82F6" 
              strokeWidth={2}
              dot={numVariables <= 50}
              activeDot={{ r: 6 }}
            />
          </LineChart>
        )}
      </ResponsiveContainer>
    </div>
  );
};