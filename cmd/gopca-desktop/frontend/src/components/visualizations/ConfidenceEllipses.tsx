import React from 'react';

interface EllipseParams {
  centerX: number;
  centerY: number;
  majorAxis: number;
  minorAxis: number;
  angle: number;
  confidenceLevel: number;
}

interface ConfidenceEllipsesProps {
  ellipses: Record<string, EllipseParams>;
  colorMap: Map<string, string>;
  xDomain: [number, number];
  yDomain: [number, number];
  // These are injected by Recharts Customized component
  xAxisMap?: any;
  yAxisMap?: any;
  width?: number;
  height?: number;
  offset?: { top: number; right: number; bottom: number; left: number };
}

export const ConfidenceEllipses: React.FC<ConfidenceEllipsesProps> = (props) => {
  const { 
    ellipses, 
    colorMap,
    xDomain,
    yDomain,
    xAxisMap,
    yAxisMap,
    width = 0,
    height = 0,
    offset = { top: 20, right: 20, bottom: 60, left: 80 }
  } = props;
  
  // If Recharts doesn't provide the necessary props, return null
  if (!xAxisMap || !yAxisMap || !width || !height) {
    console.log('ConfidenceEllipses: Missing required Recharts props', { xAxisMap, yAxisMap, width, height });
    return null;
  }
  
  // Get the first axis from the map (there's usually only one)
  const xAxis = xAxisMap && Object.values(xAxisMap)[0] as any;
  const yAxis = yAxisMap && Object.values(yAxisMap)[0] as any;
  
  if (!xAxis || !yAxis) {
    console.log('ConfidenceEllipses: No axis found in map');
    return null;
  }
  
  // Use the axis scale functions if available
  const xScale = xAxis.scale || ((value: number) => {
    const plotWidth = width - offset.left - offset.right;
    const xRange = xDomain[1] - xDomain[0];
    const xRatio = (value - xDomain[0]) / xRange;
    return offset.left + xRatio * plotWidth;
  });
  
  const yScale = yAxis.scale || ((value: number) => {
    const plotHeight = height - offset.top - offset.bottom;
    const yRange = yDomain[1] - yDomain[0];
    const yRatio = (value - yDomain[0]) / yRange;
    // Y axis is inverted in SVG
    return offset.top + plotHeight - yRatio * plotHeight;
  });

  if (!ellipses || Object.keys(ellipses).length === 0) {
    return null;
  }

  // Convert ellipse parameters to SVG path
  const ellipseToPath = (ellipse: EllipseParams, group: string): string => {
    const { centerX, centerY, majorAxis, minorAxis, angle } = ellipse;
    
    // Convert data coordinates to pixel coordinates
    const cx = xScale(centerX);
    const cy = yScale(centerY);
    
    // Scale the axes - need to account for the scale transformation
    // Calculate the scale factor from data units to pixels
    const xScaleFactor = Math.abs(xScale(1) - xScale(0));
    const yScaleFactor = Math.abs(yScale(0) - yScale(1)); // Y is inverted
    
    const rx = majorAxis * xScaleFactor;
    const ry = minorAxis * yScaleFactor;
    
    // Convert angle from radians to degrees
    const angleDegrees = (angle * 180) / Math.PI;
    
    // Create SVG path for rotated ellipse
    // Using parametric equations for an ellipse
    const steps = 50;
    const points: [number, number][] = [];
    
    for (let i = 0; i <= steps; i++) {
      const t = (i / steps) * 2 * Math.PI;
      // Ellipse in local coordinates
      const x = rx * Math.cos(t);
      const y = ry * Math.sin(t);
      
      // Apply rotation
      const rotatedX = x * Math.cos(angle) - y * Math.sin(angle);
      const rotatedY = x * Math.sin(angle) + y * Math.cos(angle);
      
      // Translate to center
      points.push([cx + rotatedX, cy + rotatedY]);
    }
    
    // Create path string
    const pathData = points
      .map((point, index) => (index === 0 ? `M ${point[0]} ${point[1]}` : `L ${point[0]} ${point[1]}`))
      .join(' ') + ' Z';
    
    return pathData;
  };

  return (
    <g className="confidence-ellipses">
      {Object.entries(ellipses).map(([group, ellipse]) => {
        const color = colorMap.get(group) || '#888888';
        const path = ellipseToPath(ellipse, group);
        
        return (
          <g key={`ellipse-${group}`}>
            <path
              d={path}
              fill={color}
              fillOpacity={0.1}
              stroke={color}
              strokeWidth={2}
              strokeOpacity={0.8}
              strokeDasharray="5,5"
            />
          </g>
        );
      })}
    </g>
  );
};