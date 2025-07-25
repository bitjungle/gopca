import { useState, useCallback, useRef } from 'react';

interface ZoomPanState {
  x: [number, number] | null;
  y: [number, number] | null;
}

interface UseZoomPanProps {
  defaultXDomain: [number, number];
  defaultYDomain: [number, number];
  zoomFactor?: number;
  maintainAspectRatio?: boolean;
}

export const useZoomPan = ({
  defaultXDomain,
  defaultYDomain,
  zoomFactor = 0.7,
  maintainAspectRatio = false
}: UseZoomPanProps) => {
  const [zoomDomain, setZoomDomain] = useState<ZoomPanState>({ x: null, y: null });
  const [isPanning, setIsPanning] = useState(false);
  const panStartRef = useRef<{ x: number; y: number } | null>(null);
  const domainStartRef = useRef<ZoomPanState | null>(null);
  
  const handleZoomIn = useCallback(() => {
    setZoomDomain(prevDomain => {
      const currentXDomain = prevDomain.x || defaultXDomain;
      const currentYDomain = prevDomain.y || defaultYDomain;
    
      const xCenter = (currentXDomain[0] + currentXDomain[1]) / 2;
      const yCenter = (currentYDomain[0] + currentYDomain[1]) / 2;
      
      const xRange = (currentXDomain[1] - currentXDomain[0]) * zoomFactor;
      const yRange = (currentYDomain[1] - currentYDomain[0]) * zoomFactor;
      
      if (maintainAspectRatio) {
        const avgRange = (xRange + yRange) / 2;
        return {
          x: [xCenter - avgRange / 2, xCenter + avgRange / 2] as [number, number],
          y: [yCenter - avgRange / 2, yCenter + avgRange / 2] as [number, number]
        };
      } else {
        return {
          x: [xCenter - xRange / 2, xCenter + xRange / 2] as [number, number],
          y: [yCenter - yRange / 2, yCenter + yRange / 2] as [number, number]
        };
      }
    });
  }, [defaultXDomain, defaultYDomain, zoomFactor, maintainAspectRatio]);
  
  const handleZoomOut = useCallback(() => {
    setZoomDomain(prevDomain => {
      const currentXDomain = prevDomain.x || defaultXDomain;
      const currentYDomain = prevDomain.y || defaultYDomain;
      
      const xCenter = (currentXDomain[0] + currentXDomain[1]) / 2;
      const yCenter = (currentYDomain[0] + currentYDomain[1]) / 2;
      
      const zoomOutFactor = 1 / zoomFactor;
      const xRange = (currentXDomain[1] - currentXDomain[0]) * zoomOutFactor;
      const yRange = (currentYDomain[1] - currentYDomain[0]) * zoomOutFactor;
      
      // Check if we're zooming out beyond default domain
      const defaultXRange = defaultXDomain[1] - defaultXDomain[0];
      const defaultYRange = defaultYDomain[1] - defaultYDomain[0];
      
      if (xRange >= defaultXRange && yRange >= defaultYRange) {
        return { x: null, y: null };
      }
      
      if (maintainAspectRatio) {
        const avgRange = (xRange + yRange) / 2;
        const maxRange = Math.max(defaultXRange, defaultYRange);
        const clampedRange = Math.min(avgRange, maxRange);
        
        return {
          x: [xCenter - clampedRange / 2, xCenter + clampedRange / 2],
          y: [yCenter - clampedRange / 2, yCenter + clampedRange / 2]
        };
      } else {
        const newXDomain: [number, number] = [
          Math.max(defaultXDomain[0], xCenter - xRange / 2),
          Math.min(defaultXDomain[1], xCenter + xRange / 2)
        ];
        const newYDomain: [number, number] = [
          Math.max(defaultYDomain[0], yCenter - yRange / 2),
          Math.min(defaultYDomain[1], yCenter + yRange / 2)
        ];
        
        return {
          x: newXDomain,
          y: newYDomain
        };
      }
    });
  }, [defaultXDomain, defaultYDomain, zoomFactor, maintainAspectRatio]);
  
  const handleResetView = useCallback(() => {
    setZoomDomain({ x: null, y: null });
  }, []);
  
  const handlePanStart = useCallback((e: React.MouseEvent) => {
    if (!zoomDomain.x) return; // Only allow panning when zoomed
    
    setIsPanning(true);
    panStartRef.current = { x: e.clientX, y: e.clientY };
    domainStartRef.current = { ...zoomDomain };
    
    // Prevent text selection while panning
    e.preventDefault();
  }, [zoomDomain]);
  
  const handlePanMove = useCallback((e: React.MouseEvent) => {
    if (!isPanning || !panStartRef.current || !domainStartRef.current) return;
    
    const deltaX = e.clientX - panStartRef.current.x;
    const deltaY = e.clientY - panStartRef.current.y;
    
    // Get the chart dimensions from the event target
    const chartElement = e.currentTarget;
    const chartWidth = chartElement.clientWidth;
    const chartHeight = chartElement.clientHeight;
    
    // Use the domain from when panning started
    const currentXDomain = domainStartRef.current.x || defaultXDomain;
    const currentYDomain = domainStartRef.current.y || defaultYDomain;
    
    // Calculate pan distance in data coordinates
    const xRange = currentXDomain[1] - currentXDomain[0];
    const yRange = currentYDomain[1] - currentYDomain[0];
    
    const xPanDistance = -(deltaX / chartWidth) * xRange;
    const yPanDistance = (deltaY / chartHeight) * yRange; // Inverted for y-axis
    
    // Apply pan with bounds checking
    const newXDomain: [number, number] = [
      currentXDomain[0] + xPanDistance,
      currentXDomain[1] + xPanDistance
    ];
    const newYDomain: [number, number] = [
      currentYDomain[0] + yPanDistance,
      currentYDomain[1] + yPanDistance
    ];
    
    setZoomDomain({
      x: newXDomain,
      y: newYDomain
    });
  }, [isPanning, defaultXDomain, defaultYDomain]);
  
  const handlePanEnd = useCallback(() => {
    setIsPanning(false);
    panStartRef.current = null;
    domainStartRef.current = null;
  }, []);
  
  return {
    zoomDomain,
    isPanning,
    handleZoomIn,
    handleZoomOut,
    handleResetView,
    handlePanStart,
    handlePanMove,
    handlePanEnd,
    isZoomed: !!zoomDomain.x
  };
};