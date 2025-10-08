# Windows Installation Guide

This guide explains the different ways to install GoPCA on Windows and how to handle security warnings.

## Recommended Installation Methods

### Option 1: Microsoft Store (Coming Soon - Recommended)

**Benefits:**
- ✅ Zero security warnings
- ✅ Automatic updates
- ✅ One-click installation
- ✅ Clean uninstallation
- ✅ No administrator rights needed

**How to Install:**

1. Open Microsoft Store
2. Search for "GoPCA"
3. Click "Install"
4. Launch from Start Menu

> **Note**: Microsoft Store distribution is currently under review. This option will be available soon!

---

### Option 2: Windows Installer (Current Recommended)

**Benefits:**
- ✅ Traditional Windows installation
- ✅ Start Menu shortcuts
- ✅ Optional CLI PATH configuration
- ✅ Works offline
- ⚠️ May show SmartScreen warning (see below)

**How to Install:**

1. Download [`GoPCA-Setup-vX.X.X.exe`](https://github.com/bitjungle/gopca/releases/latest) from GitHub Releases
2. Double-click the installer
3. **If you see a SmartScreen warning** (see [Understanding SmartScreen](#understanding-smartscreen-warnings)):
   - Click "More info"
   - Click "Run anyway"
4. Follow the installation wizard
5. Choose components:
   - ☑ GoPCA Desktop (required)
   - ☑ GoCSV Editor (required)
   - ☑ PCA Command Line Tool (required)
   - ☑ Add CLI to System PATH (optional, recommended for command-line users)
6. Click "Install"

The installer will create:
- `C:\Program Files\GoPCA\` - Application directory
- Start Menu shortcuts for GoPCA Desktop and GoCSV
- Desktop shortcut for GoPCA Desktop (optional)
- `pca.exe` in PATH (if selected)

**To Uninstall:**
- Settings → Apps → GoPCA → Uninstall
- Or use the Start Menu uninstaller shortcut

---

### Option 3: Portable ZIP Archive

**Benefits:**
- ✅ No installation required
- ✅ Run from USB drive
- ✅ No registry changes
- ⚠️ No Start Menu shortcuts
- ⚠️ Manual PATH configuration

**How to Use:**

1. Download [`gopca-windows-x64.zip`](https://github.com/bitjungle/gopca/releases/latest) from GitHub Releases
2. Extract to your preferred location (e.g., `C:\Tools\GoPCA\`)
3. **If Windows blocks the files**:
   - Right-click the ZIP file
   - Properties → Check "Unblock" → OK
   - Extract again
4. Run applications:
   - Double-click `GoPCA.exe` for the desktop app
   - Double-click `GoCSV.exe` for the CSV editor
   - Run `bin\pca.exe` from command line for CLI

**To Add CLI to PATH (Optional):**

1. Open Start Menu → Search "Environment Variables"
2. Click "Edit the system environment variables"
3. Click "Environment Variables" button
4. Under "System variables", select "Path" → Edit
5. Click "New"
6. Add: `C:\Tools\GoPCA\bin` (adjust to your extraction location)
7. Click OK on all dialogs
8. Restart Command Prompt

---

### Option 4: MSIX Package (Advanced Users / IT Admins)

**Benefits:**
- ✅ Modern Windows package format
- ✅ Sandboxed execution
- ✅ Clean uninstall
- ⚠️ Requires certificate installation (sideloading)
- ⚠️ Not recommended for most users

**How to Install (Sideloading):**

1. Download [`GoPCA_X.X.X.0_x64.msix`](https://github.com/bitjungle/gopca/releases/latest)
2. Right-click MSIX file → Properties → Digital Signatures
3. Select signature → Details → View Certificate
4. Install Certificate → Local Machine → Place in: Trusted Root Certification Authorities
5. Double-click MSIX file → Install
6. Launch from Start Menu

**PowerShell Installation:**

```powershell
# Install MSIX package
Add-AppxPackage -Path ".\GoPCA_1.1.4.0_x64.msix"

# Uninstall (if needed)
Get-AppxPackage *GoPCA* | Remove-AppxPackage
```

**Note:** MSIX packages are primarily intended for:
- Enterprise deployment via Intune/SCCM
- Testing before Microsoft Store submission
- Users who prefer the MSIX package format

Most users should use the **Windows Installer** or wait for **Microsoft Store** availability.

---

## Understanding SmartScreen Warnings

### Why You See Warnings

When downloading software from the internet, Windows SmartScreen may show warnings like:

```
Windows protected your PC
Windows Defender SmartScreen prevented an unrecognized app from starting.
```

This happens because:

1. **Lack of Reputation**: The software is new or not downloaded frequently enough to build "reputation" with Microsoft
2. **No Extended Validation Certificate**: The current installer uses a standard code signing certificate (or is unsigned during testing)

**This is NOT because the software is malicious.** It's a normal part of Windows trying to protect users from unknown software.

### Is GoPCA Safe?

**Yes!** GoPCA is:

- ✅ **Open source**: All code is publicly available on [GitHub](https://github.com/bitjungle/gopca)
- ✅ **Transparent build process**: Built automatically via GitHub Actions
- ✅ **Checksums provided**: Verify file integrity with SHA-256 checksums
- ✅ **Community vetted**: Used by researchers and data scientists

### How to Proceed Safely

**Option 1: Click "More info" → "Run anyway"**

If you downloaded from the official [GitHub Releases](https://github.com/bitjungle/gopca/releases) page:

1. Click "More info" on the SmartScreen dialog
2. Click "Run anyway"
3. The installer will proceed

**Option 2: Verify the Download**

Before running, verify the file integrity:

1. Download `checksums.txt` from the same release
2. Open PowerShell in the download folder
3. Run:
   ```powershell
   Get-FileHash .\GoPCA-Setup-vX.X.X.exe -Algorithm SHA256
   ```
4. Compare the hash with `checksums.txt`
5. If they match, the file is authentic and unmodified

**Option 3: Wait for Microsoft Store (Coming Soon)**

The Microsoft Store version will have ZERO warnings because:
- Microsoft verifies and signs all Store apps
- Store apps are automatically trusted by Windows
- No SmartScreen warnings ever appear

### What About Antivirus?

Some antivirus software may flag unfamiliar executables. This is a "false positive" due to:
- **Heuristic detection**: AV software being overly cautious with new files
- **Lack of widespread use**: Not enough users have run the software yet

If your antivirus blocks GoPCA:
1. Verify the download with checksums (see above)
2. Add an exception in your antivirus for GoPCA
3. Report the false positive to your antivirus vendor

---

## System Requirements

- **Operating System**: Windows 10 (version 1809 or later) or Windows 11
- **Architecture**: 64-bit (x64) only
- **RAM**: 4 GB minimum, 8 GB recommended
- **Disk Space**: 100 MB for installation

---

## Troubleshooting

### Installer Won't Run

**Problem**: Double-clicking does nothing or shows error.

**Solutions**:
1. Right-click installer → "Run as administrator"
2. Check download wasn't corrupted (verify checksum)
3. Ensure Windows 10/11 (64-bit)
4. Temporarily disable antivirus and retry

### App Won't Launch After Installation

**Problem**: Clicking Start Menu shortcut does nothing.

**Solutions**:
1. Check Windows Event Viewer for error details
2. Try running from installation directory: `C:\Program Files\GoPCA\GoPCA.exe`
3. Reinstall using installer with "Run as administrator"
4. Ensure .NET Desktop Runtime is installed (usually automatic)

### SmartScreen Blocks Every Time

**Problem**: SmartScreen warning appears repeatedly.

**Solutions**:
1. Use "Run anyway" option (safe if downloaded from official releases)
2. Wait for Microsoft Store version (no warnings)
3. Contact your IT admin if on managed PC

### Portable Version Won't Extract

**Problem**: ZIP file won't extract or shows errors.

**Solutions**:
1. Right-click ZIP → Properties → "Unblock" checkbox → OK
2. Use 7-Zip instead of Windows Explorer
3. Try extracting to a different location (e.g., Desktop)

### CLI Not Found After Adding to PATH

**Problem**: `pca` command not recognized in Command Prompt.

**Solutions**:
1. Restart Command Prompt (new PATH not loaded)
2. Verify PATH entry: `echo %PATH%` should include GoPCA directory
3. Try full path: `C:\Program Files\GoPCA\bin\pca.exe --version`

---

## Getting Help

- **Documentation**: [https://github.com/bitjungle/gopca/tree/main/docs](https://github.com/bitjungle/gopca/tree/main/docs)
- **Issues**: [https://github.com/bitjungle/gopca/issues](https://github.com/bitjungle/gopca/issues)
- **Discussions**: [https://github.com/bitjungle/gopca/discussions](https://github.com/bitjungle/gopca/discussions)

---

## See Also

- [Getting Started Guide](getting-started.md) - Using GoPCA after installation
- [CLI Reference](cli_reference.md) - Command-line documentation
- [FAQ](faq.md) - Frequently asked questions
