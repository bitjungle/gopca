# Developer Documentation

This directory contains developer documentation for the GoPCA Suite. These documents cover development workflows, standards, and guidelines for contributors.

## Quick Links

### Getting Started
- [Git Workflow (Simple)](git-workflow-simple.md) - The simple develop→main workflow used for most development
- [Pre-commit Checklist](pre-commit-checklist.md) - What to check before committing code
- [Documentation Standards](documentation-standards.md) - How to write and maintain documentation

### Development Guidelines
- [API Guidelines](api-guidelines.md) - REST API design principles and standards
- [Accessibility Standards](accessibility-standards.md) - WCAG 2.1 AA compliance requirements
- [Security Policy](security.md) - Security features and vulnerability reporting
- [Cross-Platform Guide](cross-platform-guide.md) - Building for Windows, macOS, and Linux

### Testing & Validation
- [Validation Methodology](validation-methodology.md) - PCA validation against scikit-learn
- [Integration Testing](integration-testing.md) - End-to-end testing strategies
- [Reliability Audit Progress](reliability-audit-progress.md) - Tracking PCA reliability improvements
- [UX Testing Checklist](ux-testing-checklist.md) - User experience testing procedures

### Technical Specifications
- [JSON Schemas](json-schemas.md) - Schema definitions for PCA models
- [CSV Format](csv-format.md) - Supported CSV file formats and specifications
- [Configuration](configuration.md) - Application configuration options
- [GUI Visualization Patterns](gui-visualization-patterns.md) - Frontend visualization architecture
- [Plotting](plotting.md) - Data visualization implementation details

### Release & Deployment
- [Release Guide](release-guide.md) - Step-by-step release process
- [Code Signing](code-signing.md) - Code signing for macOS and Windows
- [Windows Installer Guide](windows-installer-guide.md) - NSIS installer configuration
- [macOS App Translocation](macos-app-translocation.md) - Handling Gatekeeper restrictions

### UI/UX
- [UI Consistency Checklist](ui-consistency-checklist.md) - Ensuring consistent user interface

## Workflow Overview

The GoPCA project uses a simplified Git workflow:

```
main (production) ← develop (integration) ← feature branches
```

- All development happens in feature branches created from `develop`
- Feature branches are merged to `develop` via Pull Requests
- Releases are created from `develop` and merged to `main`
- Hotfixes (rare) go directly to `main` and are back-merged to `develop`

For detailed workflow instructions, see [Git Workflow (Simple)](git-workflow-simple.md).

## Key Principles

1. **As-Built Documentation** - Documentation reflects actual implementation
2. **Developer Experience** - Clear, concise, and helpful documentation
3. **Continuous Improvement** - Documentation evolves with the codebase
4. **Accessibility First** - All features must be accessible to all users

## Contributing

When adding new developer documentation:
1. Place files in this directory
2. Update this README.md with a link and description
3. Follow the [Documentation Standards](documentation-standards.md)
4. Ensure accuracy - documentation must reflect actual implementation

## Version Information

Current stable version: v1.0.2
Next planned version: v1.1.0

See [CHANGELOG.md](../../CHANGELOG.md) for version history.