# CompLab Desktop

Cross-platform desktop application for PCA (Principal Component Analysis) built with Wails.

## Features

- Load CSV data files or use built-in iris dataset
- Configure PCA parameters (components, scaling options, method)
- View input data and PCA results in interactive tables
- Display explained variance for each component
- Simple workflow UI guiding through the analysis process

## Technology Stack

- **Backend**: Go with Wails v2
- **Frontend**: React 19 + TypeScript + Vite
- **UI**: Tailwind CSS v3
- **Data Tables**: TanStack Table v8

## Development

### Prerequisites

- Go 1.24.5+
- Node.js 18+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Running in Development

```bash
wails dev
```

### Building

```bash
wails build
```

The built application will be in `build/bin/`.

## Usage

1. **Load Data**: Upload a CSV file or click "Load Iris Dataset"
2. **Configure PCA**: Set number of components and preprocessing options
3. **Run Analysis**: Click "Run PCA Analysis" to compute results
4. **View Results**: Examine the scores matrix and explained variance

## Architecture

The desktop app is part of the main CompLab module and reuses the core PCA implementation from `internal/core`.