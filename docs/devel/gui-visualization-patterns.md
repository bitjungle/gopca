# GUI Visualization Patterns

This document provides detailed guidelines for creating consistent visualization plots in the GoPCA Desktop application.

## Design System and Icons

### Icon Library
Both GoPCA Desktop and GoCSV applications use **[Heroicons](https://heroicons.com/)** for all UI icons to maintain visual consistency. 

**Icon Guidelines:**
- Use outline style icons (not solid/filled)
- Standard stroke width: `strokeWidth={1.5}`
- Header icons: `className="w-5 h-5"`
- Inline icons: `className="w-4 h-4"`
- Icons automatically adapt to theme colors using `currentColor`
- Always include appropriate ARIA labels for accessibility

**Example Icon Usage:**
```tsx
<svg
  xmlns="http://www.w3.org/2000/svg"
  fill="none"
  viewBox="0 0 24 24"
  strokeWidth={1.5}
  stroke="currentColor"
  className="w-5 h-5"
  aria-label="Icon description"
>
  <path strokeLinecap="round" strokeLinejoin="round" d="..." />
</svg>
```

## Required Imports and Structure

```tsx
import React, { useRef, useState, useCallback } from 'react';
import { PCAResult } from '../../types';
import { ExportButton } from '../ExportButton';
import { PlotControls } from '../PlotControls';
import { useChartTheme } from '../../hooks/useChartTheme';
// Additional imports for your specific chart type
```

## Component Setup

Every plot component should include:

```tsx
const YourPlot: React.FC<YourPlotProps> = ({ pcaResult, ...otherProps }) => {
  // Essential refs and state
  const chartRef = useRef<HTMLDivElement>(null);
  const fullscreenRef = useRef<HTMLDivElement>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const chartTheme = useChartTheme();
  
  // Fullscreen handler (standard implementation)
  const handleToggleFullscreen = useCallback(() => {
    if (!fullscreenRef.current) return;
    
    if (!isFullscreen) {
      if (fullscreenRef.current.requestFullscreen) {
        fullscreenRef.current.requestFullscreen();
      }
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      }
    }
    
    setIsFullscreen(!isFullscreen);
  }, [isFullscreen]);

  const handleResetView = useCallback(() => {
    // Implement zoom reset if applicable, otherwise leave empty
  }, []);
```

## Layout Structure

Use this consistent layout pattern:

```tsx
return (
  <div ref={fullscreenRef} className={`w-full h-full ${isFullscreen ? 'fixed inset-0 z-50 bg-white dark:bg-gray-900 p-4' : ''}`}>
    <div className="w-full h-full" ref={chartRef}>
      {/* Header with title and controls */}
      <div className="flex items-center justify-between mb-4">
        <h4 className="text-md font-medium text-gray-700 dark:text-gray-300">
          Plot Title: Component Details
        </h4>
        <div className="flex items-center gap-2">
          <PlotControls 
            onResetView={handleResetView}
            onToggleFullscreen={handleToggleFullscreen}
            isFullscreen={isFullscreen}
            // Add zoom handlers if applicable:
            // onZoomIn={handleZoomIn}
            // onZoomOut={handleZoomOut}
          />
          <ExportButton 
            chartRef={chartRef} 
            fileName="descriptive-filename"
          />
        </div>
      </div>
      
      {/* Chart content */}
      <div style={{ height: isFullscreen ? 'calc(100vh - 80px)' : 'calc(100% - 40px)' }}>
        {/* Your chart implementation */}
      </div>
    </div>
  </div>
);
```

## Control Button Ordering and Palette Selection

Always maintain this button order (left to right):
1. Any plot-specific controls (selectors, toggles)
2. PaletteSelector (when color coding is active)
3. PlotControls (zoom buttons if applicable, reset, fullscreen)
4. ExportButton (always rightmost)

### Important Spacing Rules
- Use `gap-2` (8px) between PlotControls and ExportButton
- Use `gap-4` (16px) between different control groups
- Wrap PlotControls and ExportButton in their own div with `gap-2`

### Palette Selector Integration

The PaletteSelector is managed at the application level (App.tsx) and appears globally when a group column is selected. Individual plot components don't need to include it directly - they access the selected palette through the PaletteContext:

```tsx
// In your plot component, access palette settings:
import { usePalette } from '../../contexts/PaletteContext';

const { mode, qualitativePalette, sequentialPalette } = usePalette();
```

Note: The PaletteSelector automatically appears in the main UI when `selectedGroupColumn` is set.

## Theme and Color Palette Integration

### Chart Theme

Always use the chart theme for consistent styling:

```tsx
const chartTheme = useChartTheme();

// Use theme colors in your charts:
<CartesianGrid stroke={chartTheme.gridColor} />
<XAxis stroke={chartTheme.axisColor} />
<text fill={chartTheme.textColor}>Label</text>
```

### Color Palette System

The application uses a unified context-aware palette system that automatically switches between categorical and continuous palettes based on the selected column:

```tsx
import { usePalette } from '../../contexts/PaletteContext';
import { 
  getQualitativeColor, 
  getSequentialColorScale, 
  createQualitativeColorMap 
} from '../../utils/colorPalettes';

// In your component:
const { mode, qualitativePalette, sequentialPalette } = usePalette();

// For categorical data:
const colorMap = createQualitativeColorMap(groupLabels, qualitativePalette);
const color = colorMap.get(groupLabel);

// For continuous data (including #target columns):
const color = getSequentialColorScale(value, min, max, sequentialPalette);
```

### Target Column Detection
- Columns ending with `#target` (with or without space) are automatically detected as continuous target columns
- These columns are available for visualization coloring but not included in PCA calculations
- The palette context automatically switches to 'continuous' mode when a target column is selected

## Error States

Handle missing data consistently:

```tsx
if (!data || data.length === 0) {
  return (
    <div className="w-full h-full flex items-center justify-center text-gray-400">
      <p>Descriptive message about what's missing</p>
    </div>
  );
}
```

## Row Labels Pattern

For plots that display individual data points (Scores, Biplot, Diagnostic), implement optional row labels using shared components:

### Using Row Labels in Your Plot

```tsx
import { CustomPointWithLabel } from '../CustomPointWithLabel';
import { calculateTopPoints } from '../../utils/labelUtils';

interface YourPlotProps {
  // ... other props
  showRowLabels?: boolean;
  maxRowLabelsToShow?: number;
}

// In your component:
const [hoveredPoint, setHoveredPoint] = useState<number | null>(null);

// Calculate which points should have labels (furthest from origin)
const topPoints = useMemo(() => 
  calculateTopPoints(data, showRowLabels, maxRowLabelsToShow),
  [data, showRowLabels, maxRowLabelsToShow]
);

// Create custom dot component
const CustomDot = useCallback((props: any) => (
  <CustomPointWithLabel
    {...props}
    topPoints={topPoints}
    hoveredPoint={hoveredPoint}
    showLabels={showRowLabels}
    onMouseEnter={setHoveredPoint}
    onMouseLeave={() => setHoveredPoint(null)}
    chartTheme={chartTheme}
    fontSize={11}  // Optional, defaults to 11
  />
), [topPoints, hoveredPoint, showRowLabels, chartTheme]);

// Use in your Scatter component
<Scatter 
  data={data}
  shape={showRowLabels ? <CustomDot /> : 'circle'}
/>
```

### Key Features
- Labels are positioned based on quadrant to avoid overlaps
- Only the N furthest points from origin are labeled (configurable)
- Hovered points always show labels regardless of distance
- Shared utility functions ensure consistent behavior across plots

## Confidence Ellipses Pattern

For plots showing grouped data (Scores, Biplot), use the shared ellipse components:

### Using Ellipses in Your Plot

```tsx
import { useEllipses } from '../../hooks/useEllipses';
import { EllipseOverlay } from '../EllipseOverlay';

// Track container size for scaling
const [containerSize, setContainerSize] = useState({ width: 0, height: 0 });

// Use ResizeObserver to track container dimensions
React.useEffect(() => {
  if (!containerRef.current) return;
  
  const resizeObserver = new ResizeObserver((entries) => {
    for (const entry of entries) {
      setContainerSize({
        width: entry.contentRect.width,
        height: entry.contentRect.height
      });
    }
  });
  
  resizeObserver.observe(containerRef.current);
  return () => resizeObserver.disconnect();
}, []);

// Calculate ellipses if not provided
const { 
  ellipses90, ellipses95, ellipses99,
  isLoading: ellipsesLoading,
  error: ellipsesError
} = useEllipses({
  scores: pcaResult.scores,
  groupLabels: groupLabels || [],
  xComponent,
  yComponent,
  enabled: shouldCalculateEllipses
});

// In your render, add the overlay
<div ref={containerRef} className="w-full relative">
  {showEllipses && effectiveGroupEllipses && groupColorMap && (
    <EllipseOverlay
      groupEllipses={effectiveGroupEllipses}
      groupColorMap={groupColorMap}
      xDomain={zoomDomain.x || defaultXDomain}
      yDomain={zoomDomain.y || defaultYDomain}
      containerSize={containerSize}
      margins={{ top: 20, right: 20, bottom: 60, left: 80 }}  // Optional
    />
  )}
  <ResponsiveContainer>
    {/* Your chart */}
  </ResponsiveContainer>
</div>
```

## Additional Guidelines

- **Zoom/Pan Support**: Use the `useZoomPan` hook if applicable
- **Export Functionality**: Provide meaningful filenames (e.g., `scores-plot-PC1-vs-PC2`)
- **Responsive Design**: Use `ResponsiveContainer` from Recharts
- **TypeScript Types**: Define clear prop interfaces
- **Tooltips**: Use React Portals for consistent tooltips (never use native `title` attribute)
- **Shared Components**: Prefer shared utilities (`labelUtils`, `ellipseUtils`) over duplicated code
- **Container Tracking**: Use ResizeObserver when you need pixel-perfect overlays

### Tooltip Implementation Pattern

```tsx
// Tooltip implementation pattern
const [tooltip, setTooltip] = useState<{ show: boolean; text: string; x: number; y: number }>({
  show: false, text: '', x: 0, y: 0
});

// Mouse event handlers
const handleMouseEnter = (e: React.MouseEvent, text: string) => {
  const rect = e.currentTarget.getBoundingClientRect();
  setTooltip({
    show: true,
    text,
    x: rect.left + rect.width / 2,
    y: rect.top - 10
  });
};

// Render tooltip with Portal
{tooltip.show && ReactDOM.createPortal(
  <div
    className="fixed z-50 px-2 py-1 text-xs rounded shadow-lg border pointer-events-none"
    style={{
      backgroundColor: chartTheme.tooltipBackgroundColor,
      borderColor: chartTheme.tooltipBorderColor,
      color: chartTheme.tooltipTextColor,
      left: tooltip.x,
      top: tooltip.y - 30,
      transform: 'translateX(-50%)'
    }}
  >
    {tooltip.text}
  </div>,
  document.body
)}
```

## Examples

For real-world examples, see the following components in `cmd/gopca-desktop/frontend/src/components/visualizations/`:
- `ScoresPlot.tsx` - Scatter plot with group coloring, confidence ellipses, and row labels
- `Biplot.tsx` - Complex plot combining scores and loadings with row labels and ellipses
- `ScreePlot.tsx` - Bar chart showing explained variance
- `LoadingsPlot.tsx` - Heatmap visualization of component loadings
- `CircleOfCorrelations.tsx` - Unit circle plot for variable correlations
- `DiagnosticScatterPlot.tsx` - Diagnostic plots for outlier detection with row labels
- `EigencorrelationPlot.tsx` - Correlation plots for eigenvectors

### Shared Utilities and Components
- `utils/labelUtils.ts` - Functions for calculating top points and label positioning
- `utils/ellipseUtils.ts` - Ellipse path generation and coordinate scaling
- `components/CustomPointWithLabel.tsx` - Reusable point component with smart labeling
- `components/EllipseOverlay.tsx` - Reusable SVG overlay for confidence ellipses