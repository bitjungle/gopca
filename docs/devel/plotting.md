# State-of-the-Art PCA Visualizations

## Overview

GoPCA uses **Plotly.js** as its visualization library, providing interactive, high-performance charts for PCA analysis. The migration from Recharts to Plotly.js was completed to provide:
- Better performance with large datasets through WebGL rendering
- More advanced visualization features (3D plots, density overlays)
- Professional export capabilities (PNG, SVG)
- Consistent mathematical accuracy with academic references

## Architecture

### Current Structure (Post-Migration)
```typescript
packages/ui-components/src/charts/
├── PlotlyBarChart.tsx               // General bar chart
├── pca/                             // PCA-specific visualizations
│   ├── PlotlyScoresPlot.tsx        // 2D scores with WebGL optimization
│   ├── Plotly3DScoresPlot.tsx      // 3D visualization
│   ├── PlotlyBiplot.tsx            // Vector scaling with filtering
│   ├── PlotlyScreePlot.tsx         // Explained variance
│   ├── PlotlyLoadingsPlot.tsx      // Component loadings
│   ├── PlotlyCircleOfCorrelations.tsx // Correlation circle
│   ├── PlotlyDiagnosticPlot.tsx    // Outlier detection
│   └── PlotlyEigencorrelationPlot.tsx // Correlation heatmap
├── core/
│   └── PlotlyVisualization.tsx     // Base class with features
└── utils/
    ├── plotlyTheme.ts               // Dark/light theme support
    ├── plotlyPerformance.ts         // WebGL optimization utilities
    ├── plotlyExport.ts              // High-quality export
    └── plotlyMath.ts                // Mathematical functions with references
```

### Core PlotlyVisualization Base Class

```typescript
// Base class for all Plotly visualizations
export abstract class PlotlyVisualization {
  // Performance features
  protected useWebGL: boolean = true;
  protected dataThreshold: number = 1000;
  
  // Interactivity features
  protected enableLasso: boolean = true;
  protected enableCrosshair: boolean = true;
  protected customModebar: PlotlyButton[] = [];
  
  // Statistical overlays
  protected showDensity: boolean = false;
  protected densityType: 'contour' | 'heatmap' | 'kde' = 'contour';
  
  // Mathematical references
  protected references: MathReference[] = [];
  
  // Render method with automatic optimization
  render(): React.ReactElement {
    const traces = this.useWebGL && this.data.length > this.dataThreshold
      ? this.getWebGLTraces()
      : this.getStandardTraces();
    
    return <PlotlyCore 
      traces={traces}
      layout={this.getEnhancedLayout()}
      config={this.getAdvancedConfig()}
    />;
  }
}
```

## Performance Optimizations

### Automatic WebGL Rendering
All scatter-based visualizations automatically switch to WebGL rendering (`scattergl`) when datasets exceed 100 points:

```typescript
import { optimizeTraceType } from '../utils/plotlyPerformance';

// Automatically chooses 'scatter' or 'scattergl'
const traceType = optimizeTraceType(data, 100);
```

### Data Decimation
For very large datasets (>10,000 points), intelligent decimation preserves visual fidelity while improving performance:

```typescript
import { decimateTrace } from '../utils/plotlyPerformance';

// Reduces points while preserving data extremes
const optimizedTrace = decimateTrace(trace, 5000);
```

### Loading Vector Filtering
The Biplot visualization now supports automatic filtering of loading vectors for datasets with many variables:

```typescript
// Configured in internal/config/gui_config.go
biplot_max_variables: 100  // Show top 100 variables by magnitude
```

When filtering is active, users see an indicator: "Showing top 100 of 1000 variables"

## State-of-the-art features

All 8 PCA visualizations have been implemented with state-of-the-art features:

1. **PlotlyScoresPlot** 
   - WebGL optimization for datasets >1000 points
   - Smart label selection (beloved feature preserved)
   - Confidence ellipses with chi-square distribution
   - Density overlays with KDE
   - Group coloring support

2. **Plotly3DScoresPlot** 
   - Interactive 3D visualization
   - WebGL rendering by default
   - Camera controls and rotation
   - Optional 2D projections

3. **PlotlyScreePlot** 
   - Dual y-axis implementation
   - Bar chart for explained variance
   - Line chart for cumulative variance
   - Customizable threshold line
   - Color-coded components

4. **PlotlyLoadingsPlot** 
   - Three modes: bar, line, grouped
   - Sort by magnitude option
   - Threshold visualization
   - Variable filtering

5. **PlotlyBiplot** 
   - Gabriel (1971) vector scaling
   - Three scaling types: correlation, symmetric, PCA
   - Smart label selection
   - Loading vectors with arrows

6. **PlotlyCircleOfCorrelations** 
   - Unit circle visualization
   - Color-coded by magnitude
   - Vector arrows with labels
   - Quadrant annotations

7. **PlotlyDiagnosticPlot** 
   - Mahalanobis distance vs RSS
   - Outlier detection (4 categories)
   - Chi-square based thresholds
   - Robust MAD-based thresholds

8. **PlotlyEigencorrelationPlot** 
   - Correlation heatmap
   - Variable clustering option
   - Value annotations
   - Responsive sizing

## Code Quality Improvements

### ESLint Configuration
A comprehensive ESLint configuration has been added to ensure code quality:
- TypeScript-specific rules with `@typescript-eslint`
- React best practices with `eslint-plugin-react`
- React Hooks rules with `eslint-plugin-react-hooks`
- Consistent code formatting rules

Run linting with:
```bash
npm run lint        # Check for issues
npm run lint:fix    # Auto-fix issues
```

### Strict TypeScript
TypeScript has been configured with strict mode for better type safety:
- `strict: true` - Enables all strict type checking options
- `noImplicitAny: true` - Disallows implicit any types
- `strictNullChecks: true` - Ensures null/undefined handling
- `noUnusedLocals: true` - Catches unused variables

### Component Consolidation
The ExportButton component has been refactored to use the shared component from `ui-components`, reducing code duplication and improving maintainability.

## Migration Notes

### From Recharts to Plotly
The migration from Recharts to Plotly.js involved:
1. Complete replacement of all chart components
2. Preservation of user-beloved features (smart labels, etc.)
3. Addition of new capabilities (3D plots, WebGL)
4. Removal of the unused chart abstraction layer
5. Consolidation of shared code in `packages/ui-components`

### GoCSV Integration
GoCSV has also been migrated to use Plotly.js:
- `DistributionChart` replaced with `PlotlyDistributionChart`
- Reuses `PlotlyBarChart` from shared components
- Consistent theming and export capabilities

