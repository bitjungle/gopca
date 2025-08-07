# Release v0.9.0

## Overview
First pre-release of the GoPCA toolkit, featuring three integrated applications:
- **GoPCA CLI**: Command-line tool for scriptable PCA analysis
- **GoPCA Desktop**: Professional GUI application for interactive PCA exploration
- **GoCSV**: Companion tool for CSV data preparation and editing

## Key Features

### GoPCA CLI
- Complete PCA implementation with SVD and NIPALS algorithms
- Multiple preprocessing options (mean center, standard scale, SNV, etc.)
- Export results in JSON, CSV, and text formats
- Cross-platform support (macOS, Linux, Windows)

### GoPCA Desktop
- Interactive visualization with scores plots, loadings plots, and biplots
- Real-time plot customization and theming
- Support for group analysis with confidence ellipses
- Dark/light mode support
- Export plots as PNG images

### GoCSV
- CSV file editing and preparation
- Data cleaning and transformation tools
- Seamless integration with GoPCA workflow

## Platform Support
- macOS (Intel and Apple Silicon) - Signed and notarized
- Linux (x64 and ARM64)
- Windows (x64)

## Infrastructure
- Self-hosted runner integration for cost-effective CI/CD
- Automated signing and notarization for macOS binaries
- Comprehensive test coverage (>85% for core PCA engine)

## Notes
This is a pre-release version (0.9.x series) leading up to the stable 1.0.0 release.