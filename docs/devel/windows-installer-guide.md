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
- `cmd/gopca-desktop/build/bin/GoPCA-amd64.exe` or `GoPCA.exe` - Build with `make pca-build`
- `cmd/gocsv/build/bin/GoCSV-amd64.exe` or `GoCSV.exe` - Build with `make csv-build`

**IMPORTANT**: 
- All three components are REQUIRED - the installer will fail if any are missing
- Desktop applications can be cross-compiled using Wails (the filename will include `-amd64` suffix when cross-compiled)
- The installer enforces complete packages to ensure users always get all components

## Building the Installer

### Quick Build
```bash
# Build pca CLI binary (can be done on any platform)
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
│   └── pca.exe        # pca CLI
└── uninstall.exe      # Uninstaller
```

### Registry Entries
The installer creates registry entries for:
- Uninstall information
- Installation directory
- Version information

## Cross-Platform Building

The installer can be built on any platform with NSIS installed. This enables CI/CD pipelines and developers on non-Windows systems to create Windows installers.

### Platform-Specific Considerations

#### On macOS
```bash
# Install NSIS via Homebrew
brew install nsis

# Build Windows CLI (cross-compilation)
GOOS=windows GOARCH=amd64 make build

# Build Windows Desktop apps with Wails
cd cmd/gopca-desktop && wails build -platform windows/amd64
cd cmd/gocsv && wails build -platform windows/amd64

# Create installer
make windows-installer
```

**Note:** macOS NSIS via Homebrew includes all standard plugins and works identically to Windows version.

#### On Linux
```bash
# Install NSIS (Ubuntu/Debian)
sudo apt-get update
sudo apt-get install -y nsis nsis-pluginapi

# For other distributions
# Fedora: sudo dnf install mingw32-nsis
# Arch: yay -S nsis

# Build Windows CLI
GOOS=windows GOARCH=amd64 make build

# Build Windows Desktop apps with Wails
cd cmd/gopca-desktop && wails build -platform windows/amd64
cd cmd/gocsv && wails build -platform windows/amd64

# Create installer
make windows-installer
```

**Note:** Some Linux distributions may have older NSIS versions. Ensure version 3.0+ for full compatibility.

#### On Windows
```bash
# Native builds (faster than cross-compilation)
make build-windows-amd64
make pca-build
make csv-build

# Create installer
make windows-installer
```

**Note:** Windows builds create native binaries without the `-amd64` suffix.

### Cross-Compilation Notes

1. **Binary Naming**: 
   - Cross-compiled: `GoPCA-amd64.exe`, `GoCSV-amd64.exe`
   - Native Windows: `GoPCA.exe`, `GoCSV.exe`
   - The installer handles both naming conventions

2. **Build Performance**:
   - Native builds are faster than cross-compilation
   - Use CI/CD for automated cross-platform builds
   - Cache dependencies for faster subsequent builds

3. **Testing**:
   - Always test installer on actual Windows system
   - Use Windows VMs or containers for validation
   - Check both installation and uninstallation processes

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
Desktop applications can be built using Wails:
1. Run `make pca-build` and `make csv-build` (works on any platform with Wails)
2. The executables will be created at:
   - `cmd/gopca-desktop/build/bin/GoPCA-amd64.exe` (when cross-compiled)
   - `cmd/gopca-desktop/build/bin/GoPCA.exe` (when built on Windows)
   - `cmd/gocsv/build/bin/GoCSV-amd64.exe` (when cross-compiled)
   - `cmd/gocsv/build/bin/GoCSV.exe` (when built on Windows)
3. Run `make windows-installer`

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

## Testing the Installer Build

You can test the Windows installer build without creating a release:

1. **Via GitHub Actions UI** (workflow_dispatch):
   - Go to Actions → Release workflow
   - Click "Run workflow"
   - Select branch and test version (e.g., v0.9.5-test)
   - Download the installer artifact after completion

2. **Locally** (requires NSIS):
   ```bash
   make windows-installer
   # or for signed version (requires certificates):
   make windows-installer-signed
   ```

The installer build is automatically tested in CI/CD when:
- Pushing version tags (production releases)
- Using workflow_dispatch trigger (test builds)

### Version Handling
- Production versions: `v0.9.5` → creates `GoPCA-Setup-v0.9.5.exe`
- Test versions: `v0.9.5-test` → creates `GoPCA-Setup-v0.9.5-test.exe`
- The 'v' prefix is stripped internally for NSIS processing

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