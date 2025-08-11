# Windows Installer Build Guide

This guide explains how to build a Windows installer for the GoPCA suite using NSIS (Nullsoft Scriptable Install System).

## Overview

The Windows installer packages all three GoPCA components into a single installer:
- **GoPCA Desktop** - GUI application for PCA analysis
- **GoCSV** - CSV editor and data preparation tool
- **PCA CLI** - Command-line tool for automation

## Prerequisites

### Required Software
- **NSIS 3.0+** - The installer creation system
  - macOS: `brew install nsis`
  - Ubuntu/Debian: `sudo apt-get install nsis`
  - Windows: Download from [nsis.sourceforge.io](https://nsis.sourceforge.io)

### Required Binaries
The installer requires ALL three Windows executables to be present:
- `build/pca-windows-amd64.exe` - Build with `make build-windows-amd64`
- `cmd/gopca-desktop/build/bin/GoPCA.exe` - Build with `make pca-build` on Windows
- `cmd/gocsv/build/bin/GoCSV.exe` - Build with `make csv-build` on Windows

**IMPORTANT**: 
- All three components are REQUIRED - the installer will fail if any are missing
- Desktop applications (GoPCA.exe and GoCSV.exe) must be built on Windows as Wails requires the target platform for GUI apps
- The installer enforces complete packages to ensure users always get all components

## Building the Installer

### Quick Build
```bash
# Build CLI binary (can be done on any platform)
make build-windows-amd64

# Build installer with available binaries
make windows-installer
```

### Build with Signed Binaries
```bash
# Sign binaries first (requires certificates)
make sign-windows

# Build installer with signed binaries
make windows-installer-signed
```

### Build Everything
```bash
# Build all components and create installer
make windows-installer-all
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `windows-installer` | Build installer with current binaries |
| `windows-installer-signed` | Sign binaries then build installer |
| `windows-installer-all` | Build all binaries and create installer |

## Installer Features

### Components
The installer includes all components (no selection required):
- **GoPCA Desktop** (required) - Main application
- **GoCSV Editor** (required) - CSV manipulation tool
- **PCA CLI Tool** (required) - Command-line interface
- **Add to PATH** (optional) - System PATH configuration for CLI
- **Start Menu Shortcuts** (optional) - Program shortcuts

All three main components are always installed to ensure users have the complete GoPCA suite.

### Installation Locations
```
C:\Program Files\GoPCA\
├── GoPCA.exe          # Desktop application
├── GoCSV.exe          # CSV editor
├── bin\
│   └── pca.exe        # CLI tool
└── uninstall.exe      # Uninstaller
```

### Registry Entries
The installer creates registry entries for:
- Uninstall information
- Installation directory
- Version information

## Cross-Platform Building

The installer can be built on any platform with NSIS installed:

### On macOS/Linux
```bash
# Install NSIS
brew install nsis        # macOS
sudo apt-get install nsis  # Ubuntu/Debian

# Build Windows CLI
make build-windows-amd64

# Create installer
make windows-installer
```

### On Windows
```bash
# Build all components
make build-windows-amd64
make pca-build
make csv-build

# Create installer with all components
make windows-installer
```

## Installer Script

The NSIS script is located at `scripts/windows/installer.nsi` and includes:
- Component selection logic
- PATH environment variable management
- Start Menu shortcut creation
- Uninstaller generation
- Version information embedding

### Customization
To modify the installer behavior, edit `scripts/windows/installer.nsi`:
- Change installation directory defaults
- Modify component descriptions
- Add file associations
- Customize UI elements

## Output

The installer is created at:
```
build/windows-installer/GoPCA-Setup-v{VERSION}.exe
```

Where `{VERSION}` is determined from git tags or the VERSION variable in the Makefile.

## Troubleshooting

### "makensis not found"
Install NSIS for your platform:
- macOS: `brew install nsis`
- Linux: `sudo apt-get install nsis`
- Windows: Download installer from official site

### "GoPCA.exe not found" or "GoCSV.exe not found"
Desktop applications must be built on Windows:
1. Switch to a Windows machine
2. Run `make pca-build` and `make csv-build`
3. Copy the .exe files to the correct locations on your build machine:
   - `cmd/gopca-desktop/build/bin/GoPCA.exe`
   - `cmd/gocsv/build/bin/GoCSV.exe`
4. Run `make windows-installer`

The installer will fail if any component is missing - this is intentional to ensure complete packages.

### Version Format Error
The NSIS script requires version in X.X.X.X format. The script automatically converts semantic versions (X.X.X) by appending .0.

## CI/CD Integration

For automated builds in CI/CD:

```yaml
# Example GitHub Actions workflow
- name: Build Windows CLI
  run: make build-windows-amd64

- name: Build Windows Installer
  run: |
    sudo apt-get update
    sudo apt-get install -y nsis
    make windows-installer

- name: Upload Installer
  uses: actions/upload-artifact@v4
  with:
    name: windows-installer
    path: build/windows-installer/GoPCA-Setup-*.exe
```

## Security Considerations

- The installer requires administrator privileges to install to Program Files
- Signed binaries reduce security warnings
- The uninstaller removes all installed files and registry entries
- No telemetry or user data is collected

## Future Enhancements

Planned improvements for the installer:
- MSI format option for enterprise deployments
- Silent installation mode for automation
- Auto-update functionality
- Chocolatey package integration
- File association for .csv files with GoCSV