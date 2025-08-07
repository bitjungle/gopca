# GoCSV

Data preparation tool for GoPCA. Provides visual CSV editing with validation specific to PCA requirements.

## Architecture

- **Backend**: Go with Wails v2 bindings (`app.go`, `commands.go`)
- **Frontend**: React + TypeScript + Vite
- **Components**: Custom grid editor, import wizard, data quality tools
- **Shared**: Uses `@gopca/ui-components` via npm workspace

## Development

```bash
# From repository root
make csv-dev          # Run in development mode (hot reload)
make csv-build        # Build for current platform
make csv-build-all    # Build for all platforms

# Or directly with Wails
cd cmd/gocsv
wails dev            # Development mode
wails build          # Production build
```

## Key Files

- `app.go` - Main application logic, file I/O, data operations
- `commands.go` - Data transformation and analysis functions
- `frontend/src/App.tsx` - Main React component with 3-step workflow
- `frontend/src/components/CSVGrid.tsx` - Editable data grid implementation

## Features

- Grid-based CSV editor with undo/redo
- Missing value detection and filling strategies
- Data quality analysis and reporting
- Column type detection (numeric/categorical)
- Excel import/export support
- Direct integration with GoPCA Desktop

## Testing

```bash
cd cmd/gocsv
go test -v           # Backend tests
npm test --prefix frontend  # Frontend tests
```

## Build Output

```
cmd/gocsv/build/bin/
├── gocsv.app/       # macOS application
└── gocsv-amd64.exe  # Windows executable
```