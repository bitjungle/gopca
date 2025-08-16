// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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

// Constants for zoom limits and behavior
const MIN_ZOOM_LEVEL = 0.1;  // Maximum zoom in (10x magnification)
const MAX_ZOOM_LEVEL = 10;   // Maximum zoom out (0.1x scale)
const ZOOM_STEP = 0.8;       // Zoom step factor (smaller = more zoom per click)

export const useZoomPan = ({
  defaultXDomain,
  defaultYDomain,
  zoomFactor = ZOOM_STEP,
  maintainAspectRatio = false
}: UseZoomPanProps) => {
  const [zoomDomain, setZoomDomain] = useState<ZoomPanState>({ x: null, y: null });
  const [zoomLevel, setZoomLevel] = useState(1.0); // Track zoom level multiplier
  const [isPanning, setIsPanning] = useState(false);
  const panStartRef = useRef<{ x: number; y: number } | null>(null);
  const domainStartRef = useRef<ZoomPanState | null>(null);
  
  const handleZoomIn = useCallback(() => {
    setZoomLevel(prevLevel => {
      // Calculate new zoom level (zoom in = multiply by zoom factor)
      const newLevel = Math.max(MIN_ZOOM_LEVEL, prevLevel * zoomFactor);
      
      // Update zoom domain based on new level
      setZoomDomain(prevDomain => {
        const currentXDomain = prevDomain.x || defaultXDomain;
        const currentYDomain = prevDomain.y || defaultYDomain;
        
        // Keep current center point
        const xCenter = (currentXDomain[0] + currentXDomain[1]) / 2;
        const yCenter = (currentYDomain[0] + currentYDomain[1]) / 2;
        
        // Calculate new ranges based on zoom level
        const defaultXRange = defaultXDomain[1] - defaultXDomain[0];
        const defaultYRange = defaultYDomain[1] - defaultYDomain[0];
        const xRange = defaultXRange * newLevel;
        const yRange = defaultYRange * newLevel;
        
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
      
      return newLevel;
    });
  }, [defaultXDomain, defaultYDomain, zoomFactor, maintainAspectRatio]);
  
  const handleZoomOut = useCallback(() => {
    setZoomLevel(prevLevel => {
      // Calculate new zoom level (zoom out = divide by zoom factor)
      const newLevel = Math.min(MAX_ZOOM_LEVEL, prevLevel / zoomFactor);
      
      // Update zoom domain based on new level
      setZoomDomain(prevDomain => {
        const currentXDomain = prevDomain.x || defaultXDomain;
        const currentYDomain = prevDomain.y || defaultYDomain;
        
        // Keep current center point
        const xCenter = (currentXDomain[0] + currentXDomain[1]) / 2;
        const yCenter = (currentYDomain[0] + currentYDomain[1]) / 2;
        
        // Calculate new ranges based on zoom level
        const defaultXRange = defaultXDomain[1] - defaultXDomain[0];
        const defaultYRange = defaultYDomain[1] - defaultYDomain[0];
        const xRange = defaultXRange * newLevel;
        const yRange = defaultYRange * newLevel;
        
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
      
      return newLevel;
    });
  }, [defaultXDomain, defaultYDomain, zoomFactor, maintainAspectRatio]);
  
  const handleResetView = useCallback(() => {
    setZoomDomain({ x: null, y: null });
    setZoomLevel(1.0); // Reset zoom level to default
  }, []);
  
  const handlePanStart = useCallback((e: React.MouseEvent) => {
    // Allow panning at any zoom level
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
    
    // Calculate new domains with pan applied
    let newXDomain: [number, number] = [
      currentXDomain[0] + xPanDistance,
      currentXDomain[1] + xPanDistance
    ];
    let newYDomain: [number, number] = [
      currentYDomain[0] + yPanDistance,
      currentYDomain[1] + yPanDistance
    ];
    
    // Apply bounds checking to prevent panning too far from data
    // Allow panning up to 50% of the range beyond the default domain
    const maxPanDistance = 0.5;
    const defaultXRange = defaultXDomain[1] - defaultXDomain[0];
    const defaultYRange = defaultYDomain[1] - defaultYDomain[0];
    
    // Check X bounds
    if (newXDomain[0] < defaultXDomain[0] - defaultXRange * maxPanDistance) {
      const shift = defaultXDomain[0] - defaultXRange * maxPanDistance - newXDomain[0];
      newXDomain = [newXDomain[0] + shift, newXDomain[1] + shift];
    }
    if (newXDomain[1] > defaultXDomain[1] + defaultXRange * maxPanDistance) {
      const shift = newXDomain[1] - (defaultXDomain[1] + defaultXRange * maxPanDistance);
      newXDomain = [newXDomain[0] - shift, newXDomain[1] - shift];
    }
    
    // Check Y bounds
    if (newYDomain[0] < defaultYDomain[0] - defaultYRange * maxPanDistance) {
      const shift = defaultYDomain[0] - defaultYRange * maxPanDistance - newYDomain[0];
      newYDomain = [newYDomain[0] + shift, newYDomain[1] + shift];
    }
    if (newYDomain[1] > defaultYDomain[1] + defaultYRange * maxPanDistance) {
      const shift = newYDomain[1] - (defaultYDomain[1] + defaultYRange * maxPanDistance);
      newYDomain = [newYDomain[0] - shift, newYDomain[1] - shift];
    }
    
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
    zoomLevel,
    isPanning,
    handleZoomIn,
    handleZoomOut,
    handleResetView,
    handlePanStart,
    handlePanMove,
    handlePanEnd,
    isZoomed: zoomLevel !== 1.0 || !!zoomDomain.x
  };
};