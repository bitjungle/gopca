# Contributing to GoPCA Suite

First off, thank you for your interest in GoPCA Suite! This document provides guidelines for contributing to the project.

## About This Project

GoPCA Suite is a **personal exploration project** that I maintain in my spare time. It represents my learning journey in creating professional-grade PCA analysis tools. While the code is open source under the MIT license, please understand that this is primarily a personal project with a specific vision and roadmap.

## Before You Contribute

### Consider Forking

If you need specific features or modifications for your use case, I encourage you to **fork the repository**. This gives you complete freedom to adapt the code to your needs without waiting for reviews or approvals.

### Limited Maintenance Bandwidth

As this is a spare-time project, I have very limited bandwidth for reviewing contributions. Response times for issues and pull requests may range from several weeks to months. Please plan accordingly.

## How to Contribute

### 1. Start with a Discussion

**Before writing any code**, please open a discussion in the [GitHub Discussions](https://github.com/bitjungle/gopca/discussions) area to:
- Explain what you'd like to contribute
- Understand if it aligns with the project's roadmap
- Get feedback on your approach

This saves everyone time and ensures we're aligned before you invest effort in code.

### 2. Reporting Bugs

Before reporting a bug:
- Check if it's already reported in [existing issues](https://github.com/bitjungle/gopca/issues)
- Ensure you can reproduce it with the latest version
- Collect all relevant information (OS, version, steps to reproduce)

Use the bug report template when creating an issue.

### 3. Suggesting Features

Feature suggestions are welcome through GitHub Discussions. Please understand that:
- Features must align with the project's core mission of PCA analysis
- I prioritize features that I personally need for my use cases
- Implementation may take considerable time or may not happen

### 4. Code Contributions

If we've agreed through discussion that a contribution makes sense:

#### Prerequisites
- Go 1.24+ and Node.js 24+
- Familiarity with the codebase structure
- Understanding of PCA mathematics (for algorithm contributions)

#### Quality Standards
All contributions must meet these standards:
- **Tests Required**: New features need comprehensive tests (target 80%+ coverage)
- **Documentation**: Update relevant documentation and add inline comments
- **Code Style**: Follow existing patterns in the codebase
- **Commit Messages**: Use conventional commits format (`feat:`, `fix:`, etc.)
- **Mathematical Correctness**: PCA-related code must be mathematically sound with references
- **Cross-platform**: Code must work on Windows, macOS, and Linux
- **No Breaking Changes**: Unless previously discussed and approved

#### Pull Request Process
1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes following the standards above
4. Ensure all tests pass: `make test`
5. Run linters: `make lint`
6. Submit a PR with:
   - Clear description of changes
   - Reference to the discussion thread
   - Test results
   - Screenshots (for UI changes)

#### Review Timeline
- PRs may take **several weeks to months** for review
- Complex changes require more time
- I may request changes or decide not to merge
- No response doesn't mean rejection - it means I haven't had time yet

## What Happens to Contributions

Please understand that:
- Not all PRs will be merged, even if they meet quality standards
- I may implement your idea differently to match my vision
- Merged code becomes part of the project under the MIT license
- I may modify or revert changes in future versions

## Alternative Ways to Contribute

If code contributions don't work out:
- **Report bugs**: Well-documented bug reports are incredibly helpful
- **Improve documentation**: Spot a typo or unclear explanation? Let me know
- **Share your use cases**: Understanding how you use GoPCA helps guide development
- **Star the repository**: If you find it useful, a star is appreciated

## Questions?

For questions about contributing, please use [GitHub Discussions](https://github.com/bitjungle/gopca/discussions) rather than issues.

## Final Note

I deeply appreciate your interest in contributing to GoPCA Suite. While I maintain tight control over the project's direction due to limited time and specific vision, I value the open source community and encourage you to fork and adapt the code for your needs. The MIT license ensures you have complete freedom to do so.

Thank you for understanding and respecting these contribution guidelines.