# Security Policy

## Overview

GoPCA takes security seriously. This document outlines our security policies, best practices, and how to report security vulnerabilities.

## Supported Versions

Security updates are provided for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.9.x   | :white_check_mark: |
| < 0.9.0 | :x:                |

## Security Features

### Input Validation
- All user inputs are validated and sanitized
- Numeric inputs have defined bounds and type checking
- String inputs are checked for dangerous characters
- File paths undergo comprehensive validation

### Path Security
- Protection against directory traversal attacks
- Validation of both input and output paths
- System directory write protection
- Symbolic link resolution and validation
- Platform-specific path validation (Windows reserved names, etc.)

### CSV Parsing Security
- Maximum file size limits (500MB)
- Field length restrictions (100K characters)
- Row and column count limits (1M rows, 10K columns)
- Memory usage monitoring and caps
- Protection against CSV injection attacks

### Command Execution Security
- Command whitelist enforcement
- Argument validation and sanitization
- Protection against command injection
- Minimal environment variable exposure
- No shell interpretation of commands

### Data Limits
- Components: 1-1000
- Kernel gamma: 1e-6 to 1e6
- Maximum memory usage: 2GB per operation
- Maximum iterations: 10,000

## Security Best Practices

### For Users
1. Always download GoPCA from official sources
2. Verify checksums of downloaded binaries
3. Keep your installation up to date
4. Use appropriate file permissions on sensitive data
5. Be cautious with CSV files from untrusted sources

### For Developers
1. Always validate user inputs before processing
2. Use the `pkg/security` package for validation
3. Never execute shell commands with user input
4. Follow the principle of least privilege
5. Keep dependencies up to date
6. Run security tests before committing

## Threat Model

### In Scope
- Local file system access control
- Input validation and sanitization
- Memory exhaustion prevention
- Command injection prevention
- Path traversal prevention
- CSV parsing vulnerabilities

### Out of Scope
- Network security (GoPCA is offline-only)
- Authentication/authorization (single-user application)
- Cryptographic operations (not used)
- Database security (no database)

## Security Checklist for PRs

Before submitting a PR, ensure:

- [ ] All user inputs are validated using `pkg/security`
- [ ] File paths use `security.ValidateInputPath()` or `security.ValidateOutputPath()`
- [ ] CSV parsing uses size and memory limits
- [ ] No direct `exec.Command()` with user input
- [ ] No use of `fmt.Sprintf()` for building commands
- [ ] Security tests pass
- [ ] No new dependencies with known vulnerabilities

## Reporting Security Vulnerabilities

### Do NOT
- Open a public GitHub issue for security vulnerabilities
- Post details on social media or forums

### DO
1. Email security details to the maintainer
2. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)
3. Allow reasonable time for a fix before disclosure

### Response Timeline
- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 7 days
- **Fix Development**: Based on severity
  - Critical: Within 14 days
  - High: Within 30 days
  - Medium/Low: Next release

## Security Testing

### Automated Testing
```bash
# Run security tests
go test ./pkg/security/...

# Run all tests including security
make test

# Static analysis
golangci-lint run
go vet ./...
```

### Manual Testing
Test with:
- Malformed CSV files
- Path traversal attempts (`../../etc/passwd`)
- Large files exceeding limits
- Special characters in inputs
- Command injection patterns

## Dependencies

### Checking for Vulnerabilities
```bash
# Check Go dependencies
go list -m -u -json all | nancy sleuth

# Check npm dependencies
npm audit

# Update dependencies
go get -u ./...
npm update
```

### Trusted Dependencies
- `gonum.org/v1/gonum` - Numerical computing
- `github.com/wailsapp/wails/v2` - Desktop framework
- Standard library preferred for everything else

## Security Updates

Security updates are announced through:
- GitHub Security Advisories
- Release notes
- CHANGELOG.md

## Incident Response

In case of a security incident:

1. **Containment**: Isolate affected systems
2. **Assessment**: Determine scope and impact
3. **Mitigation**: Apply temporary fixes
4. **Resolution**: Develop and deploy permanent fix
5. **Documentation**: Update security documentation
6. **Disclosure**: Notify users if needed

## Security Contacts

- **Primary**: Create a private security advisory on GitHub
- **Email**: Contact repository maintainer

## Compliance

GoPCA aims to follow:
- OWASP Secure Coding Practices
- CWE Top 25 Most Dangerous Software Weaknesses
- Go security best practices

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-08-15 | Initial security policy |

---

*This security policy is a living document and will be updated as needed.*