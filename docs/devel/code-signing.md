# Code Signing Guide

This document describes how code signing is implemented for GoPCA binaries across different platforms.

## Overview

GoPCA implements code signing to ensure users can trust our binaries and avoid security warnings:
- **macOS**: Automated signing and notarization using Apple Developer certificates
- **Windows**: Optional SignPath.io integration for digital signatures
- **Linux**: No signing currently (planned for future)

## macOS Code Signing

### Overview
macOS requires code signing for distribution and notarization for macOS 10.15+. The process involves:
1. Code signing with a Developer ID certificate
2. Notarization with Apple's notary service
3. Stapling the notarization ticket (for .app bundles)

### Local Development Signing

#### Prerequisites
1. **Apple Developer Account**: Required for Developer ID certificates
2. **Developer ID Certificate**: Install in Keychain Access
3. **App-Specific Password**: Generate at https://appleid.apple.com
4. **Xcode Command Line Tools**: `xcode-select --install`

#### Setup Environment
Create a `.env` file in the project root:
```bash
# Apple Developer credentials
APPLE_IDENTITY="Developer ID Application: Your Name (TEAMID)"
APPLE_ID="your-apple-id@example.com"
APPLE_TEAM_ID="YOUR_TEAM_ID"
APPLE_APP_SPECIFIC_PASSWORD="xxxx-xxxx-xxxx-xxxx"
```

#### Local Signing Process

1. **Build the applications**:
   ```bash
   make build           # CLI binary
   make pca-build       # GoPCA Desktop
   make csv-build       # GoCSV Editor
   ```

2. **Sign all binaries**:
   ```bash
   make sign            # Signs CLI, GoPCA, and GoCSV
   # Or sign individually:
   make sign-cli        # CLI only
   make sign-pca        # GoPCA only
   make sign-csv        # GoCSV only
   ```

3. **Verify signatures**:
   ```bash
   codesign --verify --verbose build/pca
   codesign --verify --verbose cmd/gopca-desktop/build/bin/GoPCA.app
   codesign --verify --verbose cmd/gocsv/build/bin/GoCSV.app
   ```

#### Local Notarization

1. **Notarize all binaries**:
   ```bash
   make notarize        # Notarizes all binaries
   # Or notarize individually:
   make notarize-cli    # CLI only
   make notarize-pca    # GoPCA only
   make notarize-csv    # GoCSV only
   ```

2. **Combined signing and notarization**:
   ```bash
   make sign-and-notarize  # Does everything in one step
   ```

3. **Verify notarization**:
   ```bash
   # For .app bundles (can be verified locally)
   spctl -a -vvv -t install cmd/gopca-desktop/build/bin/GoPCA.app
   spctl -a -vvv -t install cmd/gocsv/build/bin/GoCSV.app
   
   # For CLI binary (verified online when run)
   # Test with quarantine attribute
   xattr -w com.apple.quarantine "0081;00000000;Safari;|" build/pca
   ./build/pca --version
   ```

#### Finding Your Developer ID

To find your signing identity:
```bash
security find-identity -v -p codesigning
```

Look for "Developer ID Application" certificates. The output will show:
```
1) XXXXXXXXXX "Developer ID Application: Your Name (TEAMID)"
```

Use the full string in quotes as your `APPLE_IDENTITY`.

### CI/CD Setup
macOS signing is fully automated in our CI/CD pipeline using Apple Developer certificates.

**Required GitHub Secrets:**
- `APPLE_CERTIFICATE_BASE64`: Base64-encoded Developer ID certificate
- `APPLE_CERTIFICATE_PASSWORD`: Certificate password
- `APPLE_IDENTITY`: Developer ID identity (e.g., "Developer ID Application: Name (TEAMID)")
- `APPLE_ID`: Apple ID for notarization
- `APPLE_APP_SPECIFIC_PASSWORD`: App-specific password for notarization
- `APPLE_TEAM_ID`: Apple Developer Team ID

### Process
1. **Signing**: All macOS binaries are signed with Developer ID certificate
2. **Notarization**: Binaries are submitted to Apple for notarization
3. **Stapling**: Notarization ticket is stapled to .app bundles

### Scripts
- `scripts/ci/sign-macos-ci.sh`: Signs binaries in CI
- `scripts/ci/notarize-macos-ci.sh`: Notarizes binaries in CI
- `scripts/sign-macos.sh`: Local signing script
- `scripts/notarize-macos.sh`: Local notarization script

## Windows Code Signing

### Setup
Windows signing uses SignPath.io and is optional - the workflow functions without it configured.

**Required GitHub Secrets:**
- `SIGNPATH_API_TOKEN`: SignPath API token
- `SIGNPATH_ORG_ID`: SignPath Organization ID

### SignPath Configuration
1. Create account at [SignPath.io](https://signpath.io)
2. Create certificate (test or production)
3. Create project with slug `gopca`
4. Add signing policy with slug `test-signing` (or `release-signing` for production)
5. **Configure GitHub as Trusted Build System**:
   - Navigate to Organization Settings → Trusted Build Systems
   - Click "Add predefined" and select "GitHub.com"
   - Link the trusted build system to your `gopca` project
   - This is required for GitHub Actions to authenticate with SignPath
6. Generate API token and add to GitHub secrets

### Process
The `sign-windows-binaries` job in the release workflow:
1. Checks if SignPath is configured
2. Uploads binaries to SignPath for signing
3. Downloads signed versions
4. Release job prefers signed binaries when available

### Certificate Types
- **Test Certificate**: Self-signed, for workflow testing
- **Production Certificate**: Apply for free open source certificate from SignPath

## Local Windows Testing with Self-Signed Certificates

For local development and testing, you can use self-signed certificates to test the Windows signing process.

⚠️ **WARNING**: Self-signed certificates are for TESTING ONLY. They will trigger Windows security warnings and should never be used for distribution.

### Generating Test Certificates

1. **Run the certificate generation script:**
   ```bash
   ./scripts/generate-test-cert.sh
   ```
   
   This creates:
   - `.certs/test-cert.p12` - PKCS#12 certificate file
   - `.certs/test-cert.key` - Private key (intermediate file)
   - `.certs/test-cert.crt` - Certificate (intermediate file)

2. **Configure environment variables:**
   
   Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```
   
   The default configuration is:
   ```bash
   WINDOWS_CERT_FILE=.certs/test-cert.p12
   WINDOWS_CERT_PASSWORD=test-password
   ```

### Signing Binaries Locally

1. **Build Windows binaries:**
   ```bash
   make build-windows-amd64    # CLI
   make pca-build              # Desktop (on Windows)
   make csv-build              # GoCSV (on Windows)
   ```

2. **Sign the binaries:**
   ```bash
   make sign-windows
   ```
   
   The Makefile will:
   - Load certificate configuration from `.env`
   - Check if certificate files exist
   - Sign all Windows .exe files using osslsigncode
   - Display success messages for each signed binary

### Verifying Signatures

To verify that binaries were signed:

```bash
# Using osslsigncode
osslsigncode verify build/pca-windows-amd64.exe

# Output for self-signed certificate:
# Signature verification: ok
# Number of signers: 1
# Signer #0:
#   Subject: /C=US/ST=Test/L=Test/O=GoPCA Test/CN=GoPCA Test Certificate
#   Issuer: /C=US/ST=Test/L=Test/O=GoPCA Test/CN=GoPCA Test Certificate
```

On Windows, right-click the .exe file → Properties → Digital Signatures tab to view certificate details.

### Security Considerations

- **Never commit certificates**: The `.certs/` directory is gitignored
- **Use unique passwords**: Don't use the default "test-password" for real certificates
- **Test certificates only**: These are only for local testing, not distribution
- **Windows warnings**: Self-signed certificates will show "Unknown Publisher" warnings

### Troubleshooting

**"Certificate file not found"**
- Run `./scripts/generate-test-cert.sh` to generate certificates
- Check that `.env` file exists and has correct paths

**"osslsigncode: error: PKCS#12 parse error"**
- Verify the certificate password in `.env` matches the one used during generation
- Regenerate certificate if password is unknown

**"osslsigncode not found"**
- Install osslsigncode:
  - macOS: `brew install osslsigncode`
  - Ubuntu/Debian: `sudo apt-get install osslsigncode`
  - Windows: Use signtool from Windows SDK instead

### Production Signing

For production releases, use SignPath.io (configured in CI/CD) or obtain a proper code signing certificate from a Certificate Authority. Never use self-signed certificates for distribution.

## Linux Code Signing

Currently not implemented. Future options include:
- GPG signing for package repositories
- Reproducible builds for verification
- AppImage signatures
- Package manager signatures (.deb, .rpm)

### Planned Implementation
1. **GPG Signing**:
   - Sign release artifacts with GPG key
   - Publish public key for verification
   - Include signatures in release assets

2. **Package Signatures**:
   - Debian packages: debsign
   - RPM packages: rpmsign
   - AppImage: embedded signatures

3. **Reproducible Builds**:
   - Deterministic compilation
   - Source-based verification
   - Build environment documentation

## Verification

### macOS
```bash
# Check signature
codesign --verify --verbose <binary>

# Check notarization
spctl -a -vvv -t install <binary>
```

### Windows
```powershell
# Check signature
Get-AuthenticodeSignature <binary.exe>
```

Or right-click → Properties → Digital Signatures tab

## Troubleshooting

### macOS Issues
- **"Certificate not found"**: Check APPLE_IDENTITY secret format
- **"Failed to notarize"**: Verify app-specific password is correct
- **"Stapling failed"**: Only .app bundles can be stapled, not standalone binaries

### Windows Issues
- **Signing skipped**: Verify both SIGNPATH_API_TOKEN and SIGNPATH_ORG_ID are set
- **Authentication error**: Regenerate API token
- **Project not found**: Check project slug is `gopca`

## Security Considerations

1. **Never commit certificates or tokens** to the repository
2. **Use GitHub Secrets** for all sensitive information
3. **Rotate credentials periodically**
4. **Monitor signing logs** for unauthorized attempts

## References

- [Apple Developer - Notarizing macOS Software](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [SignPath Documentation](https://about.signpath.io/documentation)
- [Windows Authenticode](https://docs.microsoft.com/en-us/windows-hardware/drivers/install/authenticode)