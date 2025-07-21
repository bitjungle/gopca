import React, { useRef } from 'react';
import { ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from 'recharts';
import { PCAResult } from '../../types';
import { ExportButton } from '../ExportButton';

interface ScoresPlotProps {
  pcaResult: PCAResult;
  rowNames: string[];
  xComponent?: number; // 0-based index
  yComponent?: number; // 0-based index
}

export const ScoresPlot: React.FC<ScoresPlotProps> = ({ 
  pcaResult, 
  rowNames,
  xComponent = 0, 
  yComponent = 1 
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  // Transform scores data for Recharts
  const data = pcaResult.scores.map((row, index) => {
    const xVal = row[xComponent] || 0;
    const yVal = row[yComponent] || 0;
    
    // Check for invalid values
    if (!isFinite(xVal) || !isFinite(yVal)) {
      console.warn(`Invalid values at index ${index}: x=${xVal}, y=${yVal}`);
      return null;
    }
    
    return {
      x: xVal,
      y: yVal,
      name: rowNames[index] || `Sample ${index + 1}`,
    };
  }).filter(point => point !== null);


  // Get variance percentages for axis labels
  const xVariance = pcaResult.explained_variance[xComponent]?.toFixed(1) || '0';
  const yVariance = pcaResult.explained_variance[yComponent]?.toFixed(1) || '0';

  const xLabel = `PC${xComponent + 1} (${xVariance}%)`;
  const yLabel = `PC${yComponent + 1} (${yVariance}%)`;

  // Calculate data range to ensure 0 is included and axes cross at origin
  const xValues = data.map(d => d!.x);
  const yValues = data.map(d => d!.y);
  const xMin = Math.min(0, ...xValues);
  const xMax = Math.max(0, ...xValues);
  const yMin = Math.min(0, ...yValues);
  const yMax = Math.max(0, ...yValues);
  
  // Add padding to the range
  const xPadding = (xMax - xMin) * 0.1 || 1;
  const yPadding = (yMax - yMin) * 0.1 || 1;

  // Handle case where there's no data
  if (data.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-400">
        <p>No data to display</p>
      </div>
    );
  }

  return (
    <div className="w-full h-full">
      <div className="flex justify-end mb-2">
        <ExportButton 
          chartRef={containerRef} 
          fileName={`scores-plot-PC${xComponent + 1}-vs-PC${yComponent + 1}`}
        />
      </div>
      <div ref={containerRef} className="w-full" style={{ height: 'calc(100% - 40px)' }}>
        <ResponsiveContainer width="100%" height="100%">
        <ScatterChart
          data={data}
          margin={{ top: 20, right: 20, bottom: 60, left: 80 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
          <XAxis 
            type="number" 
            dataKey="x" 
            name={xLabel}
            label={{ value: xLabel, position: 'insideBottom', offset: -10 }}
            stroke="#9CA3AF"
            domain={[xMin - xPadding, xMax + xPadding]}
            axisLine={{ stroke: '#9CA3AF' }}
            tickLine={{ stroke: '#9CA3AF' }}
            tickFormatter={(value) => value.toFixed(1)}
          />
          <YAxis 
            type="number" 
            dataKey="y" 
            name={yLabel}
            label={{ value: yLabel, angle: -90, position: 'insideLeft' }}
            stroke="#9CA3AF"
            domain={[yMin - yPadding, yMax + yPadding]}
            axisLine={{ stroke: '#9CA3AF' }}
            tickLine={{ stroke: '#9CA3AF' }}
            tickFormatter={(value) => value.toFixed(1)}
          />
          <ReferenceLine x={0} stroke="#6B7280" strokeWidth={2} />
          <ReferenceLine y={0} stroke="#6B7280" strokeWidth={2} />
          <Tooltip 
            cursor={{ strokeDasharray: '3 3' }}
            content={({ active, payload }) => {
              if (active && payload && payload.length) {
                const data = payload[0].payload;
                return (
                  <div className="bg-gray-800 p-2 rounded shadow-lg border border-gray-600">
                    <p className="text-white font-semibold">{data.name}</p>
                    <p className="text-gray-300">{xLabel}: {data.x.toFixed(3)}</p>
                    <p className="text-gray-300">{yLabel}: {data.y.toFixed(3)}</p>
                  </div>
                );
              }
              return null;
            }}
          />
          <Scatter 
            name="Scores" 
            fill="#3B82F6"
            fillOpacity={0.8}
            strokeWidth={1}
            stroke="#1E40AF"
          />
        </ScatterChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};