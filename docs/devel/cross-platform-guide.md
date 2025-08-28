# Cross-Platform Development Guide

This guide documents cross-platform compatibility considerations for GoPCA Suite development, ensuring consistent behavior across Windows, macOS, and Linux.

## Overview

GoPCA Suite is designed to work seamlessly across all major operating systems. This guide outlines platform-specific considerations, common pitfalls, and best practices for maintaining cross-platform compatibility.

## Platform Detection

### Runtime Detection
```go
import "runtime"

switch runtime.GOOS {
case "windows":
    // Windows-specific code
case "darwin":
    // macOS-specific code
case "linux":
    // Linux-specific code
default:
    // Fallback for other Unix-like systems
}
```

### Build Tags
Use build tags for platform-specific source files:
```go
//go:build windows
// +build windows

// Code specific to Windows
```

## File Path Handling

### Always Use filepath Package
```go
import "path/filepath"

// CORRECT - uses proper separator for the OS
path := filepath.Join("dir", "subdir", "file.txt")

// WRONG - hardcoded separator
path := "dir/subdir/file.txt"  // Breaks on Windows
```

### Path Separators
- **Windows**: `\` (backslash)
- **Unix/macOS**: `/` (forward slash)
- Use `filepath.Separator` when needed
- Use `filepath.ToSlash()` for URL-style paths
- Use `filepath.FromSlash()` to convert from URL-style

### Absolute vs Relative Paths
```go
// Convert to absolute path for consistency
absPath, err := filepath.Abs(relativePath)

// Check if path is absolute
if filepath.IsAbs(path) {
    // Path is absolute
}
```

### Case Sensitivity
- **Windows/macOS**: Case-insensitive (by default)
- **Linux**: Case-sensitive
- Always use exact case in code
- Use `strings.EqualFold()` for case-insensitive comparison when needed

## Platform-Specific Paths

### System Directories
```go
// Windows system directories (pkg/security/path_security.go)
var WindowsSystemDirectories = []string{
    `C:\Windows`, `C:\Program Files`, `C:\Program Files (x86)`,
    `C:\ProgramData`, `C:\System32`, `C:\SysWOW64`,
}

// Unix system directories
var SystemDirectories = []string{
    "/etc", "/bin", "/sbin", "/usr/bin", "/usr/sbin",
    "/sys", "/proc", "/dev", "/boot", "/lib", "/lib64",
}
```

### User Directories
```go
import "os"

// Get user home directory (cross-platform)
home, err := os.UserHomeDir()

// Get temp directory (cross-platform)
temp := os.TempDir()
```

### Application Data
```go
// Platform-specific app data locations
func GetAppDataDir() string {
    switch runtime.GOOS {
    case "windows":
        return filepath.Join(os.Getenv("APPDATA"), "GoPCA")
    case "darwin":
        home, _ := os.UserHomeDir()
        return filepath.Join(home, "Library", "Application Support", "GoPCA")
    default: // Linux and others
        home, _ := os.UserHomeDir()
        return filepath.Join(home, ".config", "gopca")
    }
}
```

## File Permissions

### Unix Permissions
```go
// Only set permissions on Unix-like systems
if runtime.GOOS != "windows" {
    err := os.Chmod(file, 0600)  // Owner read/write only
}
```

### Windows Considerations
- Windows doesn't support Unix-style permissions
- Use Windows ACLs for advanced permissions (not implemented)
- Basic read/write permissions work across platforms

## Line Endings

### Text File Handling
- **Windows**: CRLF (`\r\n`)
- **Unix/macOS**: LF (`\n`)
- Git configuration: `core.autocrlf = false`
- Use `strings.ReplaceAll()` to normalize when needed

### CSV Files
```go
// Always use \n for CSV writing (normalized by encoding/csv)
writer := csv.NewWriter(file)
writer.UseCRLF = false  // Use LF on all platforms
```

## Process and Command Execution

### Command Names
```go
// Platform-specific command names
func getProcessCommand() string {
    switch runtime.GOOS {
    case "windows":
        return "tasklist"
    case "darwin", "linux":
        return "pgrep"
    default:
        return ""
    }
}
```

### Executable Extensions
```go
// Add .exe extension on Windows
execName := "gopca"
if runtime.GOOS == "windows" {
    execName += ".exe"
}
```

### Shell Commands
```go
// Avoid shell execution for security
cmd := exec.Command(program, args...)  // Direct execution

// If shell is needed (avoid if possible)
switch runtime.GOOS {
case "windows":
    cmd = exec.Command("cmd", "/c", command)
default:
    cmd = exec.Command("sh", "-c", command)
}
```

## UI and User Experience

### Keyboard Shortcuts
```javascript
// Frontend detection (TypeScript/React)
const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
const modifierKey = isMac ? 'Cmd' : 'Ctrl';

// Display appropriate shortcuts
const undoShortcut = `${modifierKey}+Z`;
const redoShortcut = isMac ? `${modifierKey}+Shift+Z` : `${modifierKey}+Y`;
```

### File Dialogs
- Use Wails runtime for native file dialogs
- Respect platform conventions for file filters
- Default to user's home or documents directory

### Application Icons
- **Windows**: `.ico` format (multiple resolutions)
- **macOS**: `.icns` format (multiple resolutions)
- **Linux**: `.png` format (multiple sizes)

## Building and Packaging

### Cross-Compilation
```bash
# Build for specific platform
GOOS=windows GOARCH=amd64 go build
GOOS=darwin GOARCH=arm64 go build
GOOS=linux GOARCH=amd64 go build
```

### Platform-Specific Assets
```
build/
├── darwin/
│   ├── Info.plist       # macOS app metadata
│   └── icons.icns       # macOS icon
├── windows/
│   ├── info.json        # Windows metadata
│   └── icon.ico         # Windows icon
└── linux/
    └── icon.png         # Linux icon
```

### Installers
- **Windows**: NSIS installer (`.exe`)
- **macOS**: DMG or signed .app bundle
- **Linux**: AppImage, .deb, or .rpm

## Testing Across Platforms

### CI/CD Matrix
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
```

### Platform-Specific Test Skips
```go
func TestUnixPermissions(t *testing.T) {
    if runtime.GOOS == "windows" {
        t.Skip("Unix permissions not applicable on Windows")
    }
    // Test code
}
```

### Test Data Considerations
- Use well-conditioned test data to avoid numerical differences
- Normalize line endings in test files
- Use `filepath.Join()` for test file paths

## Security Considerations

### Path Validation
- Check for directory traversal (`../`)
- Validate against system directories
- Handle Windows reserved names (CON, PRN, AUX, etc.)
- Normalize paths before comparison

### Command Injection
- Never pass user input directly to shell
- Use exec.Command() with separate arguments
- Validate and whitelist commands
- Escape special characters appropriately

## Common Pitfalls and Solutions

### Issue: Hardcoded Path Separators
**Wrong:**
```go
path := "data/files/test.csv"
```
**Correct:**
```go
path := filepath.Join("data", "files", "test.csv")
```

### Issue: Case-Sensitive File Names
**Wrong:**
```go
// Works on Windows/macOS, fails on Linux
file := "MyFile.txt"  // Actual file is "myfile.txt"
```
**Correct:**
```go
// Use exact case or normalize
file := "myfile.txt"
```

### Issue: Platform-Specific Features
**Wrong:**
```go
// Assumes Unix signals exist
signal.Notify(c, syscall.SIGUSR1)  // Fails on Windows
```
**Correct:**
```go
// Check platform first
if runtime.GOOS != "windows" {
    signal.Notify(c, syscall.SIGUSR1)
}
```

### Issue: File Locking
**Note:** File locking behavior differs:
- **Windows**: Exclusive locks by default
- **Unix**: Advisory locks
- Use appropriate locking strategy per platform

## Platform-Specific Features

### macOS
- **App Translocation**: Handle randomized paths for downloaded apps
- **Code Signing**: Required for distribution
- **Notarization**: Required for macOS 10.15+
- **Universal Binaries**: Support both Intel and Apple Silicon

### Windows
- **UAC**: May require elevation for Program Files
- **Antivirus**: May flag unsigned binaries
- **Registry**: For file associations and uninstall info
- **Long Path Support**: Enable for paths > 260 chars

### Linux
- **Distributions**: Test on major distros (Ubuntu, Fedora, Arch)
- **Desktop Environments**: Support GNOME, KDE, XFCE
- **Package Managers**: Provide .deb, .rpm, AppImage
- **Wayland vs X11**: Test on both display servers

## Debugging Platform Issues

### Debug Logging
```go
if os.Getenv("GOPCA_DEBUG") == "1" {
    fmt.Printf("[DEBUG] Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
    fmt.Printf("[DEBUG] Path separator: %c\n", filepath.Separator)
    fmt.Printf("[DEBUG] Temp dir: %s\n", os.TempDir())
}
```

### Platform-Specific Tests
```bash
# Test on different platforms locally
docker run --rm -v $(pwd):/app golang:1.24 bash -c "cd /app && go test ./..."
```

## Best Practices

1. **Always use filepath package** for path operations
2. **Test on all target platforms** before release
3. **Handle platform differences explicitly** rather than assuming
4. **Document platform-specific behavior** in code comments
5. **Use CI/CD matrix testing** to catch issues early
6. **Provide platform-specific installers** for better UX
7. **Follow platform conventions** for UI and file locations
8. **Validate inputs** considering platform differences
9. **Use abstraction layers** for platform-specific features
10. **Keep platform-specific code isolated** and well-documented

## References

- [Go filepath package](https://pkg.go.dev/path/filepath)
- [Go build constraints](https://pkg.go.dev/go/build#hdr-Build_Constraints)
- [Wails cross-platform guide](https://wails.io/docs/guides/crossplatform)
- [NSIS Windows installer](https://nsis.sourceforge.io)
- [macOS Code Signing](https://developer.apple.com/documentation/security/code_signing_services)
- [Linux Desktop Entry Specification](https://specifications.freedesktop.org/desktop-entry-spec/latest/)