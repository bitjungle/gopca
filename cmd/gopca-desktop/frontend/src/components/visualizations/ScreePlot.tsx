import React, { useRef } from 'react';
import { 
  ComposedChart, 
  Bar, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  Legend,
  ReferenceLine
} from 'recharts';
import { PCAResult } from '../../types';
import { ExportButton } from '../ExportButton';

interface ScreePlotProps {
  pcaResult: PCAResult;
  showCumulative?: boolean;
  elbowThreshold?: number; // Optional: highlight components explaining this % variance
}

export const ScreePlot: React.FC<ScreePlotProps> = ({ 
  pcaResult, 
  showCumulative = true,
  elbowThreshold = 80 
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  // Transform variance data for Recharts
  const data = pcaResult.explained_variance.map((variance, index) => ({
    component: pcaResult.component_labels?.[index] || `PC${index + 1}`,
    variance: variance,
    cumulative: pcaResult.cumulative_variance[index],
    componentNumber: index + 1
  }));

  // Find elbow point (where cumulative variance crosses threshold)
  const elbowIndex = pcaResult.cumulative_variance.findIndex(cv => cv >= elbowThreshold);
  const elbowComponent = elbowIndex >= 0 ? elbowIndex + 1 : data.length;

  // Handle edge cases
  if (!data || data.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center text-gray-400">
        <p>No variance data to display</p>
      </div>
    );
  }

  return (
    <div className="w-full h-full">
      <div className="flex justify-end mb-2">
        <ExportButton 
          chartRef={containerRef} 
          fileName="scree-plot"
        />
      </div>
      <div ref={containerRef} className="w-full" style={{ height: 'calc(100% - 40px)' }}>
        <ResponsiveContainer width="100%" height="100%">
        <ComposedChart 
          data={data} 
          margin={{ top: 20, right: 20, bottom: 60, left: 80 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
          
          <XAxis 
            dataKey="component" 
            stroke="#9CA3AF"
            label={{ 
              value: 'Principal Component', 
              position: 'insideBottom', 
              offset: -10,
              style: { fill: '#9CA3AF' }
            }}
          />
          
          <YAxis 
            yAxisId="variance"
            stroke="#9CA3AF"
            label={{ 
              value: 'Explained Variance (%)', 
              angle: -90, 
              position: 'insideLeft',
              style: { fill: '#9CA3AF' }
            }}
            domain={[0, 'auto']}
            tickFormatter={(value) => value.toFixed(0)}
          />
          
          {showCumulative && (
            <YAxis 
              yAxisId="cumulative"
              orientation="right"
              stroke="#10B981"
              label={{ 
                value: 'Cumulative Variance (%)', 
                angle: 90, 
                position: 'insideRight',
                style: { fill: '#10B981' }
              }}
              domain={[0, 100]}
              tickFormatter={(value) => value.toFixed(0)}
            />
          )}
          
          <Tooltip 
            content={({ active, payload }) => {
              if (active && payload && payload.length) {
                const data = payload[0].payload;
                return (
                  <div className="bg-gray-800 p-3 rounded shadow-lg border border-gray-600">
                    <p className="text-white font-semibold">{data.component}</p>
                    <p className="text-blue-400">
                      Variance: {data.variance.toFixed(2)}%
                    </p>
                    {showCumulative && (
                      <p className="text-green-400">
                        Cumulative: {data.cumulative.toFixed(2)}%
                      </p>
                    )}
                  </div>
                );
              }
              return null;
            }}
          />
          
          <Legend 
            wrapperStyle={{ paddingTop: '20px' }}
            iconType="rect"
          />
          
          {/* Reference line at elbow threshold */}
          {showCumulative && (
            <ReferenceLine 
              y={elbowThreshold} 
              yAxisId="cumulative"
              stroke="#EF4444" 
              strokeDasharray="5 5"
              label={{ 
                value: `${elbowThreshold}% threshold`, 
                position: 'right',
                style: { fill: '#EF4444' }
              }}
            />
          )}
          
          {/* Vertical line at elbow component */}
          {elbowComponent <= data.length && (
            <ReferenceLine 
              x={`PC${elbowComponent}`} 
              stroke="#EF4444" 
              strokeDasharray="3 3"
              strokeOpacity={0.5}
            />
          )}
          
          <Bar 
            yAxisId="variance"
            dataKey="variance" 
            fill="#3B82F6"
            name="Explained Variance"
            radius={[4, 4, 0, 0]}
          />
          
          {showCumulative && (
            <Line 
              yAxisId="cumulative"
              type="monotone" 
              dataKey="cumulative" 
              stroke="#10B981"
              strokeWidth={2}
              name="Cumulative Variance"
              dot={{ fill: '#10B981', r: 4 }}
            />
          )}
        </ComposedChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};