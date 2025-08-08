# Code Signing Guide

This document describes how code signing is implemented for GoPCA binaries across different platforms.

## Overview

GoPCA implements code signing to ensure users can trust our binaries and avoid security warnings:
- **macOS**: Automated signing and notarization using Apple Developer certificates
- **Windows**: Optional SignPath.io integration for digital signatures
- **Linux**: No signing currently (planned for future)

## macOS Code Signing

### Setup
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
5. Generate API token and add to GitHub secrets

### Process
The `sign-windows-binaries` job in the release workflow:
1. Checks if SignPath is configured
2. Uploads binaries to SignPath for signing
3. Downloads signed versions
4. Release job prefers signed binaries when available

### Certificate Types
- **Test Certificate**: Self-signed, for workflow testing
- **Production Certificate**: Apply for free open source certificate from SignPath

## Linux Code Signing

Currently not implemented. Future options include:
- GPG signing for package repositories
- Reproducible builds for verification

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