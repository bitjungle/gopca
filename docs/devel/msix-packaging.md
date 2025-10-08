# MSIX Packaging for Microsoft Store

This document explains the MSIX packaging process for GoPCA Desktop, enabling distribution through the Microsoft Store.

## Overview

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
├── AppxManifest.template.xml    # Template with {{VERSION}} and {{PUBLISHER}} placeholders
└── Assets/                      # MSIX icon assets
    ├── StoreLogo.png           # 50x50 Store logo
    ├── Square44x44Logo.png     # 44x44 App list icon
    └── Square150x150Logo.png   # 150x150 Start menu tile

scripts/
├── windows/
│   └── build-msix.sh           # Core MSIX build script
└── test-msix-build.sh          # Local testing script (uses .env secrets)

.github/workflows/release.yml
├── build-msix job              # Automated MSIX build in CI/CD
└── create-release job          # Includes MSIX in GitHub releases
```

### Build Process Flow

1. **Checkout code** - Get repository with templates and assets
2. **Download GoPCA.exe** - From build-desktop job (signed or unsigned)
3. **Generate manifest** - Substitute version and publisher in template
4. **Create package structure** - Copy exe, manifest, and assets
5. **Build MSIX** - Use MakeAppx.exe to create package
6. **Sign MSIX** - Self-signed certificate (Store re-signs on publish)
7. **Upload artifact** - Make available for release

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

The template (`cmd/gopca-desktop/AppxManifest.template.xml`) uses placeholders:

```xml
<Identity
    Name="BitjungleGoPCA"
    Publisher="{{PUBLISHER}}"
    Version="{{VERSION}}" />
```

During build:
- `{{VERSION}}` → `1.1.3.0`
- `{{PUBLISHER}}` → `CN=12345678-90AB-CDEF-1234-567890ABCDEF`

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
# Ensure GoPCA.exe exists
make pca-build

# Run build script
./scripts/windows/build-msix.sh \
    --version 1.1.4 \
    --publisher "CN=YOUR-PUBLISHER-GUID" \
    --exe cmd/gopca-desktop/build/bin/GoPCA.exe \
    --output build/msix
```

This creates:
- Package structure: `build/msix/package/`
- MSIX file: `build/msix/GoPCA_1.1.4.0_x64.msix`

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

3. **Launch app**:
   - Open Start Menu
   - Search for "GoPCA"
   - Click to launch

4. **Uninstall** (if needed):
   ```powershell
   Get-AppxPackage *GoPCA* | Remove-AppxPackage
   ```

## CI/CD Workflow

### Automated Build

The `build-msix` job in `.github/workflows/release.yml`:

1. Runs on `windows-latest` runner
2. Depends on `build-desktop` and `sign-windows-binaries` jobs
3. Uses GitHub Secrets for Publisher ID
4. Builds MSIX with MakeAppx.exe
5. Signs with self-signed certificate
6. Uploads as workflow artifact

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
   - Description
   - Screenshots (1280x720 or 1920x1080 recommended)
   - App category: Developer tools / Education
   - Privacy policy URL
   - Support contact info

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
   - Verify all app functions work in MSIX container

2. **Version Consistency**:
   - Ensure version matches across:
     - Git tag
     - `wails.json`
     - `package.json`
     - AppxManifest.xml (auto-generated)

3. **Asset Quality**:
   - Use high-quality icons (150x150 minimum)
   - Ensure assets are PNG format
   - Test icons appear correctly in Start Menu

4. **Store Listing**:
   - Prepare screenshots ahead of time
   - Write clear, compelling description
   - Have privacy policy ready
   - Test search keywords

5. **Update Cadence**:
   - Submit MSIX updates regularly
   - Keep in sync with GitHub releases
   - Note: Store certification takes 1-3 days

## Resources

- [MSIX Documentation](https://learn.microsoft.com/en-us/windows/msix/)
- [MakeAppx Tool Reference](https://learn.microsoft.com/en-us/windows/msix/package/create-app-package-with-makeappx-tool)
- [Partner Center Documentation](https://learn.microsoft.com/en-us/windows/apps/publish/)
- [SignTool Documentation](https://learn.microsoft.com/en-us/windows/msix/package/sign-app-package-using-signtool)

## See Also

- [Release Guide](release-guide.md) - Full release process
- [Windows Installer Guide](windows-installer-guide.md) - Traditional installer
- [Git Workflow](git-workflow-simple.md) - Branch and PR process
