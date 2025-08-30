# Security Policy

## Supported Versions

As a personal project maintained in spare time, security updates are provided on a best-effort basis for the latest release only.

| Version | Supported          |
| ------- | ------------------ |
| Latest release | ✅ Best effort |
| Older releases | ❌ Not supported |

## Reporting a Vulnerability

If you discover a security vulnerability in GoPCA Suite, I appreciate your help in disclosing it responsibly.

### How to Report

1. **Do NOT create a public issue** for security vulnerabilities
2. Send an email to: devel@bitjungle.com
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if you have one)

### What to Expect

- **Acknowledgment**: I'll try to acknowledge receipt within a week, but as this is a spare-time project, it may take longer
- **Investigation**: I'll investigate as time permits (this could take several weeks)
- **Resolution**: Critical vulnerabilities will be prioritized, but fix timeline depends on complexity and my available time
- **Disclosure**: Once fixed, the vulnerability will be disclosed in the release notes

### Scope

This security policy applies to:
- The `pca` CLI tool
- GoPCA Desktop application
- GoCSV Desktop application
- Core PCA algorithms and data handling

### Out of Scope

The following are not considered security vulnerabilities:
- Issues requiring physical access to the user's machine
- Social engineering attacks
- Issues in dependencies (report these to the dependency maintainer)
- Missing security headers in the desktop applications (they run locally)

## General Security Practices

GoPCA Suite is designed with security in mind:
- **100% local processing** - No data is sent to external servers
- **No telemetry** - No usage data or metrics are collected
- **No network dependencies** - Works completely offline
- **Open source** - All code is publicly auditable

However, users should:
- Keep their installation updated to the latest version
- Only load CSV files from trusted sources
- Be aware that PCA models may contain information about the training data

## Disclaimer

This is a personal open-source project maintained in spare time. While I take security seriously and will do my best to address vulnerabilities, there are no guarantees or SLAs for security fixes. Organizations requiring guaranteed security response times should consider alternative solutions or maintain their own fork.

Thank you for helping keep GoPCA Suite secure!