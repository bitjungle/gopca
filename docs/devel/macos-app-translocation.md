# macOS App Translocation (Gatekeeper Path Randomization)

## Overview

macOS App Translocation is a security feature introduced in macOS Sierra (10.12) that protects users from malicious software. When an app is downloaded from the internet (has a quarantine flag) and launched by double-clicking, macOS may run it from a randomized, read-only location instead of its actual location.

## How It Affects GoPCA/GoCSV

When users download our release bundles from GitHub and double-click to launch:
1. macOS moves the app to `/private/var/folders/.../AppTranslocation/[random-id]/`
2. The app runs from this temporary location
3. Relative path detection between GoPCA and GoCSV fails
4. The apps cannot find each other even when in the same folder

## Detection

You can detect if an app is translocated by checking the executable path:

```go
if strings.Contains(execPath, "/AppTranslocation/") {
    // App is translocated
}
```

## Solutions Implemented

### 1. Smart Detection (pkg/integration/app_integration.go)
When we detect App Translocation:
- Search common locations where both apps might be together
- Check /Applications, ~/Applications, ~/Downloads, ~/Desktop
- Look for both apps in the same location before using that path

### 2. User Workarounds
Users can avoid translocation by:
- Moving apps to /Applications before first launch
- Using command line: `open /path/to/GoPCA.app`
- Right-clicking and choosing "Open" (sometimes helps)
- Removing quarantine: `xattr -cr GoPCA.app GoCSV.app`

## Testing

To test translocation handling:
1. Download apps from GitHub releases
2. Unzip to ~/Downloads
3. Double-click to launch (triggers translocation)
4. Set `GOPCA_DEBUG=1` to see detection logic

To verify translocation is active:
```bash
ps aux | grep GoPCA
# Look for /AppTranslocation/ in the path
```

## References
- [Apple Developer: App Translocation](https://developer.apple.com/library/archive/technotes/tn2206/_index.html)
- [macOS Security Guide](https://support.apple.com/guide/security/app-security-overview-sec35dd877d0/web)