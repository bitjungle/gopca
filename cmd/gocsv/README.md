# GoCSV - GoPCA CSV Editor

A focused, standalone companion application for preparing and editing CSV files that conform to the GoPCA data format.

## Overview

GoCSV addresses immediate user pain points in data preparation while maintaining GoPCA's commitment to simplicity and quality. It provides a visual interface for editing CSV files with validation specific to GoPCA requirements.

## Features

### Core Features
- **Visual Data Editing**: Grid-based CSV editor with row/column operations
- **Column Type Detection**: Visual indicators for numeric, categorical, and target columns
- **Real-time Validation**: Validates data against GoPCA requirements
- **Missing Value Detection**: Highlights and handles missing values

### File Format Support
- CSV (comma and semicolon variants)
- Excel files (.xlsx, .xls) - planned
- TSV (Tab-separated values) - planned
- JSON (tabular format) - planned

### GoPCA Integration
- One-click "Open in GoPCA" functionality
- Shared validation logic from internal/io
- Consistent preprocessing preview

## Development

### Prerequisites
- Go 1.24+
- Node.js 24+
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Setup
```bash
cd cmd/gocsv
npm install --prefix frontend
```

### Running in Development
```bash
wails dev
```

The application will be available at http://localhost:34115

### Building
```bash
wails build
```

## Architecture

GoCSV follows the same architectural principles as GoPCA Desktop:
- Wails framework for desktop application
- React + TypeScript frontend
- Tailwind CSS for styling
- Reuses GoPCA's validation logic

## UI Design

The UI is designed to match GoPCA Desktop's visual style:
- Dark/light theme support
- Card-based layout
- Consistent color scheme and typography
- Smooth animations and transitions

## Status

This is the initial implementation focusing on UI consistency with GoPCA Desktop. The core editing functionality (ag-Grid integration) is pending implementation.