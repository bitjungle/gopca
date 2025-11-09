# MSIX Packaging for Microsoft Store

This document explains the MSIX packaging process for the GoPCA Suite, bundling both GoPCA Desktop and GoCSV Desktop in a single MSIX package for Microsoft Store distribution.

## Overview

The GoPCA Suite MSIX package bundles both GoPCA Desktop and GoCSV Desktop applications together, enabling their tight integration features on Microsoft Store installations.

### Why Bundle Both Apps?

**Integration Benefits:**
- "Open GoCSV" button in GoPCA works automatically
- "Open in GoPCA" button in GoCSV works seamlessly
- No manual installation of companion app required
- Single install for complete PCA workflow

**Technical Benefits:**
- Both apps installed to same directory
- Existing integration code works without changes
- Apps appear as separate entries in Start Menu
- Clean single-package uninstall

### MSIX Advantages

MSIX is the modern Windows app package format that provides:
- **Trusted installation** - No SmartScreen warnings when distributed via Microsoft Store
- **Sandboxed execution** - Enhanced security through containerization
- **Clean uninstall** - Complete removal with no registry/file remnants
- **Automatic updates** - Seamless updates through the Store
- **Enterprise deployment** - Sideloading support for IT administrators

## Architecture

### Files and Structure

```
cmd/gopca-desktop/
├── AppxManifest.template.xml    # Template with both app definitions
└── Assets/                      # GoPCA MSIX icon assets
    ├── StoreLogo.png           # 50x50 Store logo
    ├── Square44x44Logo.png     # 44x44 GoPCA app list icon
    └── Square150x150Logo.png   # 150x150 GoPCA start menu tile

cmd/gocsv/
└── Assets/                      # GoCSV MSIX icon assets
    ├── GoCSV_Square44x44Logo.png      # 44x44 GoCSV app list icon
    ├── GoCSV_Square150x150Logo.png    # 150x150 GoCSV start menu tile
    └── README.md               # Asset specifications

scripts/
├── windows/
│   └── build-msix.sh           # Core MSIX build script (supports dual apps)
└── test-msix-build.sh          # Local testing script (uses .env secrets)

.github/workflows/release.yml
├── build-msix job              # Automated bundled MSIX build in CI/CD
└── create-release job          # Includes MSIX in GitHub releases
```

### Build Process Flow

1. **Checkout code** - Get repository with templates and assets
2. **Download executables** - Both GoPCA.exe and GoCSV.exe from build jobs
3. **Copy assets** - GoPCA and GoCSV icon assets to package
4. **Generate manifest** - Substitute version and publisher, includes both app definitions
5. **Create package structure** - Copy both executables, manifest, and all assets
6. **Build MSIX** - Use MakeAppx.exe to create bundled package
7. **Sign MSIX** - Self-signed certificate (Store re-signs on publish)
8. **Upload artifact** - Make available for release

## Secrets Configuration

### Required GitHub Secrets

Add these secrets to your GitHub repository settings:

- **MS_ENTRA_TENANT_ID** - Your Microsoft Entra tenant ID
- **MS_PARTNER_ID** - Partner Center partner ID
- **MS_WINDOWS_PUBLISHER_ID** - Publisher GUID (without `CN=` prefix)

### Getting Publisher ID

1. Log into [Microsoft Partner Center](https://partner.microsoft.com)
2. Navigate to **Apps → GoPCA → Product Identity**
3. Copy the **Publisher** value (format: `CN=12345678-90AB-CDEF-1234-567890ABCDEF`)
4. Store WITHOUT the `CN=` prefix as `MS_WINDOWS_PUBLISHER_ID`
   - Example: If Publisher is `CN=ABCD1234-...`, store `ABCD1234-...`
   - The `CN=` prefix is added programmatically during build

### Local Development

For local testing, create a `.env` file with:

```bash
MS_ENTRA_TENANT_ID=your-entra-tenant-id
MS_PARTNER_ID=your-partner-id
MS_WINDOWS_PUBLISHER_ID=your-publisher-guid-without-CN
```

**Never commit `.env` to the repository!**

## Version Format

MSIX requires version in `X.X.X.X` format:

- **Input**: `v1.1.3` (Git tag)
- **Conversion**: Strip `v`, append `.0`
- **Output**: `1.1.3.0` (MSIX version)

This conversion happens automatically in the build scripts.

## AppxManifest Template

The template (`cmd/gopca-desktop/AppxManifest.template.xml`) defines both applications and uses placeholders:

```xml
<Identity
    Name="bitjungle.GoPCA"
    Publisher="{{PUBLISHER}}"
    Version="{{VERSION}}" />

<Applications>
  <Application Id="GoPCA" Executable="GoPCA.exe" ...>
    <!-- GoPCA Desktop configuration -->
  </Application>
  <Application Id="GoCSV" Executable="GoCSV.exe" ...>
    <!-- GoCSV Desktop configuration -->
  </Application>
</Applications>
```

During build:
- `{{VERSION}}` → `1.1.3.0`
- `{{PUBLISHER}}` → `CN=12345678-90AB-CDEF-1234-567890ABCDEF`

Both applications appear as separate entries in the Windows Start Menu, but share the same installation directory for seamless integration.

## Local Testing

### Prerequisites

- Go 1.24+
- Node.js 24+
- `.env` file with Publisher ID configured

### Test Package Structure (macOS/Linux)

```bash
./scripts/test-msix-build.sh
```

This script:
1. Loads Publisher ID from `.env`
2. Creates package structure in `build/msix-test/`
3. Generates AppxManifest.xml with substitutions
4. Validates manifest content
5. Skips MakeAppx.exe (Windows SDK not available)

Inspect the output:
```bash
cat build/msix-test/package/AppxManifest.xml
ls -la build/msix-test/package/
```

### Build MSIX (Windows Only)

On a Windows machine with Windows SDK installed:

```bash
# Ensure both applications are built
make pca-build   # GoPCA Desktop
make csv-build   # GoCSV Desktop

# Ensure GoCSV MSIX assets exist
# See cmd/gocsv/Assets/README.md for requirements

# Run build script with both executables
./scripts/windows/build-msix.sh \
    --version 1.1.4 \
    --publisher "CN=YOUR-PUBLISHER-GUID" \
    --gopca-exe cmd/gopca-desktop/build/bin/GoPCA.exe \
    --gocsv-exe cmd/gocsv/build/bin/GoCSV.exe \
    --output build/msix
```

This creates:
- Package structure: `build/msix/package/`
- MSIX file: `build/msix/GoPCA_1.1.4.0_x64.msix`
- Package contains: Both GoPCA.exe and GoCSV.exe

### Install and Test MSIX

On Windows 10/11:

1. **Install certificate** (for self-signed MSIX):
   ```powershell
   # Right-click MSIX → Properties → Digital Signatures
   # Install certificate to Trusted Root
   ```

2. **Install MSIX**:
   ```powershell
   Add-AppxPackage -Path .\GoPCA_1.1.4.0_x64.msix
   ```

3. **Verify both apps installed**:
   - Open Start Menu
   - Search for "GoPCA" - should appear
   - Search for "GoCSV" - should also appear
   - Both apps installed from single package

4. **Test integration**:
   - Launch GoPCA Desktop
   - Click "Open GoCSV" → should launch GoCSV
   - In GoCSV, click "Open in GoPCA" → should launch GoPCA with data

5. **Uninstall** (removes both apps):
   ```powershell
   Get-AppxPackage *GoPCA* | Remove-AppxPackage
   ```

## CI/CD Workflow

### Automated Build

The `build-msix` job in `.github/workflows/release.yml`:

1. Runs on `windows-latest` runner
2. Depends on `build-desktop`, `build-gocsv`, and `sign-windows-binaries` jobs
3. Downloads both GoPCA.exe and GoCSV.exe artifacts
4. Uses GitHub Secrets for Publisher ID
5. Copies both sets of MSIX assets
6. Builds bundled MSIX with MakeAppx.exe
7. Signs with self-signed certificate
8. Uploads as workflow artifact

### Testing Workflow

Use `workflow_dispatch` to test MSIX build:

```bash
# Trigger manually from GitHub Actions UI
# Or via CLI:
gh workflow run release.yml \
  -f test_version=v1.1.4-test
```

Download the `msix-package` artifact and test installation.

### Release Integration

On release (tag push):

1. MSIX builds automatically
2. Included in GitHub release assets
3. Added to `checksums.txt`
4. Mentioned in release notes

## Bundled Package Considerations

### Start Menu Behavior

The bundled MSIX package creates **two separate Start Menu entries**:
- **GoPCA** - Primary PCA analysis application
- **GoCSV** - CSV data preparation tool

Users can launch either app independently. The apps share the same installation directory, enabling automatic discovery for integration features.

### Integration Detection

The existing integration code in `pkg/integration/app_integration.go` automatically detects both apps when installed from the bundled MSIX:

```go
// Windows same-directory detection (lines 204-208)
case "windows":
    paths = append(paths,
        filepath.Join(execDir, "GoCSV.exe"),
        filepath.Join(execDir, "gocsv.exe"),
    )
```

No code changes were needed - the bundled installation naturally satisfies the same-directory check.

### Package Size

The bundled MSIX package is approximately **30-40MB**, which is acceptable by Microsoft Store standards. This includes:
- GoPCA.exe (~15-20MB)
- GoCSV.exe (~15-20MB)
- MSIX assets for both apps
- AppxManifest.xml

### Asset Requirements

Before building the MSIX package, ensure GoCSV assets exist:

```bash
# Required files (see cmd/gocsv/Assets/README.md for specifications)
cmd/gocsv/Assets/GoCSV_Square150x150Logo.png  # 150x150px Start Menu tile
cmd/gocsv/Assets/GoCSV_Square44x44Logo.png    # 44x44px App list icon
```

The build script will fail with clear error messages if these assets are missing.

## Microsoft Store Submission

### Prerequisites

1. Microsoft Partner Center account
2. App reservation: "GoPCA"
3. Store listing prepared (description, screenshots, etc.)

### Submission Process

1. **Download MSIX** from GitHub release:
   ```bash
   wget https://github.com/bitjungle/gopca/releases/download/v1.1.4/GoPCA_1.1.4.0_x64.msix
   ```

2. **Log into Partner Center**:
   - Go to [https://partner.microsoft.com](https://partner.microsoft.com)
   - Navigate to **Apps → GoPCA**

3. **Create New Submission**:
   - Click "Start new submission"
   - Fill out all required sections

4. **Upload MSIX**:
   - In **Packages** section
   - Upload `GoPCA_1.1.4.0_x64.msix`
   - Microsoft will validate the package

5. **Complete Store Listing**:
   - **Description**: Mention both GoPCA and GoCSV in the app description
     - Highlight the integrated workflow
     - Explain that one install provides both tools
     - Emphasize seamless data preparation → analysis workflow
   - **Screenshots** (1280x720 or 1920x1080 recommended):
     - Include screenshots of both GoPCA and GoCSV
     - Show integration features (Open GoCSV button, etc.)
     - Demonstrate the workflow from CSV editing to PCA analysis
   - **App category**: Developer tools / Education / Productivity
   - **Privacy policy URL**
   - **Support contact info**

6. **Submit for Certification**:
   - Review all sections
   - Click "Submit to the Store"
   - Wait 1-3 business days for review

7. **Monitor Status**:
   - Check certification progress in Partner Center
   - Address any issues if certification fails
   - App goes live automatically upon approval

### Post-Approval

After Microsoft Store approval:

1. **Update README.md**:
   - Add Microsoft Store badge
   - Link to Store listing

2. **Announce**:
   - Update release notes
   - Social media/blog posts
   - Notify users

## Signing Details

### Why Self-Signed is OK

The MSIX generated in CI/CD is signed with a self-signed certificate. This is acceptable because:

1. **Microsoft Re-Signs**: When you submit to Partner Center, Microsoft replaces the signature with their trusted certificate
2. **No SmartScreen Warnings**: Store-distributed apps are automatically trusted
3. **Simpler Build Process**: No need for expensive code signing certificates

### Certificate Creation

The CI/CD workflow creates a temporary self-signed certificate:

```powershell
New-SelfSignedCertificate -Type Custom \
  -Subject "CN=${{ secrets.MS_WINDOWS_PUBLISHER_ID }}" \
  -KeyUsage DigitalSignature \
  -FriendlyName 'GoPCA MSIX Test' \
  -CertStoreLocation 'Cert:\CurrentUser\My'
```

This certificate:
- Matches the Publisher in AppxManifest.xml
- Is valid for package creation and validation
- Is NOT trusted by Windows (until Store re-signs)
- Is deleted after signing

## Troubleshooting

### Common Issues

#### 1. Publisher Mismatch

**Error**: "Publisher doesn't match certificate"

**Solution**: Ensure `MS_WINDOWS_PUBLISHER_ID` secret matches Partner Center exactly (without `CN=` prefix)

#### 2. Version Format Invalid

**Error**: "Invalid version format"

**Solution**: Version must be `X.X.X.X`. Check version conversion logic in workflow.

#### 3. MakeAppx.exe Not Found

**Error**: "MakeAppx.exe is not recognized"

**Solution**:
- Ensure Windows SDK is installed
- Use full path: `C:\Program Files (x86)\Windows Kits\10\bin\10.0.22621.0\x64\MakeAppx.exe`
- Check SDK version matches workflow

#### 4. SignTool Fails

**Error**: "SignTool error: No certificates were found"

**Solution**:
- Verify certificate was created successfully
- Check Publisher ID matches manifest
- This is OK for Store submission (Store re-signs)

#### 5. Package Installation Fails

**Error**: "Can't install this app package"

**Solution**:
- Install certificate to Trusted Root (for sideloading)
- Or submit to Store (users won't need certificate)

#### 6. GoCSV Assets Missing

**Error**: "GoCSV Assets directory not found" or "GoCSV MSIX assets are missing"

**Solution**:
- Create required assets: `cmd/gocsv/Assets/GoCSV_Square150x150Logo.png` and `GoCSV_Square44x44Logo.png`
- See `cmd/gocsv/Assets/README.md` for asset specifications
- Generate from existing icons: `magick cmd/gocsv/build/icons/icon-256.png -resize 150x150 cmd/gocsv/Assets/GoCSV_Square150x150Logo.png`

#### 7. Only GoPCA Appears in Start Menu

**Error**: GoCSV app not showing in Start Menu after installation

**Solution**:
- Verify manifest includes both Application entries: `grep 'Id="GoCSV"' msix-build/package/AppxManifest.xml`
- Check GoCSV.exe was copied to package: `ls msix-build/package/GoCSV.exe`
- Reinstall package after ensuring both apps in manifest

### Debugging

Enable verbose output in build scripts:

```bash
# Local testing
./scripts/windows/build-msix.sh --version 1.1.4 ... --verbose

# Check manifest substitution
cat build/msix/package/AppxManifest.xml | grep -E "(Version|Publisher)"

# Validate package
MakeAppx.exe pack /v /h SHA256 /d build/msix/package /p test.msix
```

Check CI/CD logs:
- Go to GitHub Actions run
- Open "Build MSIX Package" job
- Expand each step for detailed output

## Best Practices

1. **Test Before Release**:
   - Use `workflow_dispatch` to test MSIX build
   - Install and test on clean Windows VM
   - Verify **both apps appear in Start Menu**
   - Test integration features:
     - "Open GoCSV" button in GoPCA
     - "Open in GoPCA" button in GoCSV
     - Data export/import between apps
   - Verify all app functions work in MSIX container

2. **Version Consistency**:
   - Ensure version matches across:
     - Git tag
     - Both `wails.json` files (GoPCA and GoCSV)
     - Both `package.json` files
     - AppxManifest.xml (auto-generated)

3. **Asset Quality**:
   - Use high-quality icons (150x150 minimum)
   - Ensure assets are PNG format
   - **Test both GoPCA and GoCSV icons** appear correctly in Start Menu
   - Verify visual distinction between the two apps

4. **Store Listing**:
   - Prepare screenshots ahead of time **for both apps**
   - Write clear, compelling description **mentioning integrated suite**
   - Highlight the bundled nature as a feature
   - Have privacy policy ready
   - Test search keywords (both "GoPCA" and "GoCSV")

5. **Update Cadence**:
   - Submit MSIX updates regularly
   - Keep in sync with GitHub releases
   - **Coordinate versions** - both apps in bundle should match
   - Note: Store certification takes 1-3 days

6. **Pre-Submission Checklist**:
   - [ ] Both GoPCA.exe and GoCSV.exe built
   - [ ] GoCSV MSIX assets created and validated
   - [ ] Package builds without errors
   - [ ] Both apps appear in test installation
   - [ ] Integration features tested and working
   - [ ] Package size under 50MB
   - [ ] Screenshots show both apps
   - [ ] Store listing mentions bundled apps

## Resources

- [MSIX Documentation](https://learn.microsoft.com/en-us/windows/msix/)
- [MakeAppx Tool Reference](https://learn.microsoft.com/en-us/windows/msix/package/create-app-package-with-makeappx-tool)
- [Partner Center Documentation](https://learn.microsoft.com/en-us/windows/apps/publish/)
- [SignTool Documentation](https://learn.microsoft.com/en-us/windows/msix/package/sign-app-package-using-signtool)

## See Also

- [Release Guide](release-guide.md) - Full release process
- [Windows Installer Guide](windows-installer-guide.md) - Traditional installer
- [Git Workflow](git-workflow-simple.md) - Branch and PR process
