# GoPCA Desktop

Professional PCA analysis application with interactive visualizations.

## Architecture

- **Backend**: Go with Wails v2 bindings (`app.go`)
- **Frontend**: React + TypeScript + Vite
- **Visualizations**: Recharts for all plots (scores, loadings, biplot, scree)
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

- `app.go` - Wails bindings, PCA execution, data I/O
- `frontend/src/App.tsx` - Main React component with 3-step workflow
- `frontend/src/components/ScoresPlot.tsx` - Interactive 2D/3D scores visualization
- `frontend/src/components/BiplotDisplay.tsx` - Combined scores/loadings plot
- `frontend/src/contexts/HelpContext.tsx` - Contextual help system

## Features

- SVD, NIPALS, and Kernel PCA methods
- Multiple preprocessing options (SNV, scaling, centering)
- Interactive 2D/3D visualizations with confidence ellipses
- Biplot with variable contributions
- Eigencorrelation analysis
- Export to multiple formats (PNG, SVG, CSV, JSON)

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