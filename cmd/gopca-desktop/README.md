# GoPCA Desktop

Professional PCA analysis application with interactive visualizations.

## Architecture

- **Backend**: Go with Wails v2 bindings (`app.go`)
- **Frontend**: React + TypeScript + Vite
- **Visualizations**: Plotly.js for all plots (scores, loadings, biplot, scree)
- **Shared**: Uses `@gopca/ui-components` via npm workspace
- **Core**: Reuses PCA engine from `internal/core`

## Development

```bash
# From repository root
make pca-dev          # Run in development mode (hot reload)
make pca-build        # Build for current platform  
make pca-build-all    # Build for all platforms

# Or directly with Wails
cd cmd/gopca-desktop
wails dev            # Development mode
wails build          # Production build
```

## Key Files

### Backend
- `app.go` - Wails bindings, PCA execution, data I/O

### Frontend Components
- `frontend/src/App.tsx` - Main React component with 3-step workflow
- `frontend/src/components/ScoresPlot.tsx` - Interactive 2D/3D scores visualization
- `frontend/src/components/BiplotDisplay.tsx` - Combined scores/loadings plot with ellipses
- `frontend/src/components/CustomPointWithLabel.tsx` - Reusable component for points with labels
- `frontend/src/components/EllipseOverlay.tsx` - SVG overlay for confidence ellipses

### Frontend Utilities
- `frontend/src/utils/labelUtils.ts` - Functions for label positioning and top point selection
- `frontend/src/utils/ellipseUtils.ts` - Ellipse path generation and coordinate scaling

### Contexts
- `frontend/src/contexts/HelpContext.tsx` - Contextual help system

## Features

- SVD, NIPALS, and Kernel PCA methods
- Multiple preprocessing options (SNV, scaling, centering)
- Interactive 2D/3D visualizations with confidence ellipses
- Row labels for identifying specific data points
- Biplot with variable contributions and ellipse support
- Eigencorrelation analysis
- Export to multiple formats (PNG, SVG, CSV, JSON)
- Two-tier control layout for better UI organization

## Testing

```bash
cd cmd/gopca-desktop
go test -v           # Backend tests
npm test --prefix frontend  # Frontend tests
```

## Build Output

```
cmd/gopca-desktop/build/bin/
├── gopca-desktop.app/       # macOS application
└── gopca-desktop-amd64.exe  # Windows executable
```