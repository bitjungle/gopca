# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2026-04-11

### Added
- React Context state management for GoPCA Desktop, eliminating prop drilling from App.tsx (#630, #505)
  - 5 new context providers: `FileDataContext`, `PCAContext`, `VisualizationContext`, `UIContext`, `GoCSVContext`
  - Extracted `DataLoadSection`, `PCAConfigSection`, and `ResultsSection` as focused sub-components
  - App.tsx reduced to a lean provider stack (~191 lines)
- Transform and preprocessing validation test suite (#625, #480)
  - 33 new tests validating the complete PCA transform operation and preprocessing pipeline
  - `internal/core` test coverage: 88.1% (requirement: >80%)

### Fixed
- Arrow colors in Circle of Correlations now follow the palette correctly (#626, #513)
  - Arrowhead color was hardcoded to `#3b82f6` instead of reading from the active color scheme

### Changed
- Upgrade Plotly.js from v2.35.3 to v3.5.0 (#622, #618)
  - Includes fix for `customdata` handling and other upstream improvements
- Upgrade Wails framework from v2.10.2 to v2.12.0 (#623, #619)
- Refactor App.tsx: extract 7 custom hooks from monolithic `AppContent` component (#629, #503, #507)
  - `useAppInit`, `useFileData`, `useGoCSVIntegration`, `usePCAConfig`, `usePCARunner`, `useVisualization`, `useUIState`
  - Adds production-safe `logger.ts` — suppresses all console output in production builds
- Optimize React context rendering in GoPCA Desktop (#631, #506)
  - `useMemo` on all 5 context value objects prevents unnecessary consumer re-renders on parent re-renders
  - Functional `setConfig` updater form eliminates stale closure bugs in `PCAConfigSection`
  - `useMemo` for derived arrays (`plotOptions`, `groupColumnOptions`) in `ResultsSection`
- Centralize NIPALS and Kernel PCA algorithm constants (#628, #484)
  - `NIPALSConfig` (tolerance: 1e-8, maxIterations: 1000) and `KernelPCAConfig` (minEigenvalue: 1e-10) in `internal/config`
  - Eliminates duplicated `const` blocks across PCA algorithm files
- Remove dead fields from `PCAResult` struct (#627, #459)
  - Six fields never set or read in production removed: `ExplainedVariance`, `CumulativeVariance`, `Mean`, `Scale`, `Components`, `Config`
- Update copyright year range to 2025-2026 across all source files and LICENSE (#632)

## [1.1.8] - 2025-11-09

### Added
- **Bundled MSIX Package** for Microsoft Store (#608, #567)
  - Single installation package containing both GoPCA Desktop and GoCSV
  - Both apps installed in same directory enabling tight integration
  - Enables "Open in GoPCA" and "Open GoCSV" cross-app features
  - Includes comprehensive MSIX packaging documentation
  - Suitable for Microsoft Store submission and enterprise sideloading

### Fixed
- CustomSelect component keyboard navigation with type-to-search filtering (#607)
- Preprocessor state now used for PreprocessingApplied metadata field (#606)
- Non-comma delimiter handling in mixed parsing and CLI validation (#605, #599, #600)
- --target-columns CLI flag now parsed and applied correctly (#604)
- Missing-strategy zero implementation for imputation (#603)

## [1.1.7] - 2025-11-07

### Fixed
- RFC 4180 compliant CSV field escaping to prevent data corruption (#593, #588)
- DocumentationViewer modal state reset to prevent stale content flash (#595, #589)
- GoCSV async history clearing with proper await (#592, #587)
- CLI command generator uses actual file paths instead of dataset names (#590, #585)

### Changed
- Removed dead code for non-existent kernel types (code cleanup) (#591, #586)

## [1.1.6] - 2025-11-07

### Added
- Microsoft Store distribution with automated MSIX packaging (#561, #560)
- CSV format tooltip to data table in Step 1 (#547)

### Fixed
- Race conditions in EventBus tests preventing pre-commit hooks (#581, #580)
- Nil dereference in CalculateGroupEllipses function (#578, #571)
- Student's t-distribution CDF implementation for accurate p-values (#577, #570)
- Kernel PCA method name normalization ('polynomial' → 'poly') (#576, #569)
- CLI double-preprocessing and incorrect diagnostics data (#575, #572, #573)
- Windows SDK tools dynamic location for MSIX builds (#562)
- Z-axis component selector for 3D Biplot visualization (#535)
- LaTeX math duplication by importing ui-components CSS (#534)
- Preserve unselected rows in table after plot selection (#544)

### Changed
- Removed legacy missing-value helper functions (technical debt cleanup) (#579, #574)
- Default coef0 value for polynomial kernel from 0 to 1 for better defaults (#542)
- Simplified embedded stock dataset for accessibility to non-traders (#539)
- Removed redundant CSV format text in Step 1 UI (#548)

### Documentation
- Optimized intro_to_pca.md illustration images for faster loading (#556)
- Improved temporal PCA documentation and figures (#527, #455)

### Testing
- Comprehensive test coverage improvements for pkg packages (#553)
  - pkg/csv: 26% → 73.8% coverage
  - pkg/security: 50.6% → 85.8% coverage
  - pkg/integration: 34.1% → 46.6% coverage
- Fixed data races in EventBus tests with proper channel synchronization (#581)

### Refactoring
- Code quality improvements following DRY/KISS principles (#557)
  - Consolidated JSON conversion helpers (~150 lines eliminated)
  - Extracted NIPALS algorithm helpers (72% and 54% complexity reduction)
  - Improved testability with comprehensive unit tests

## [1.1.5] - 2025-10-09

### Fixed
- Corrected MSIX package identity name to match Microsoft Partner Center reservation
  - Changed from `BitjungleGoPCA` to `bitjungle.GoPCA`
  - Fixes Partner Center upload validation error
  - Required for successful Microsoft Store submission

## [1.1.4] - 2025-10-09

### Added
- **Microsoft Store Distribution** (#563, #560)
  - Automated MSIX package generation in CI/CD workflow
  - Self-signed temporary certificate (Microsoft Store re-signs on publish)
  - Dynamic Windows SDK tool discovery (no hardcoded paths)
  - Smart version format conversion (strips pre-release suffixes)
  - Complete infrastructure with AppxManifest template and MSIX assets
  - Comprehensive documentation in `docs/devel/msix-packaging.md`
  - End-user installation guide in `docs/windows-installation.md`
  - Resolves Windows SmartScreen warning showstopper

### Documentation
- Added MSIX package to release verification checklist
- Updated release guide with Microsoft Store submission process
- Created comprehensive MSIX technical documentation
- Added Windows installation guide for end users

## [1.1.3] - 2025-01-10

### Changed
- **Code Quality Improvements**: Significant refactoring to improve maintainability (#557)
  - Consolidated JSON conversion helpers to eliminate ~150 lines of duplication across apps
  - Extracted NIPALS algorithm helpers into reusable functions (72% and 54% complexity reduction)
  - Improved testability with comprehensive unit tests for helper functions
  - All changes maintain mathematical correctness and performance
  - Better adherence to DRY and KISS principles from development guidelines

### Documentation
- Optimized intro_to_pca.md illustration images for faster loading (#556)

## [1.1.2] - 2025-01-28

### Added
- Comprehensive test coverage improvements (#553)
  - pkg/csv: 26% → 73.8% coverage with validation and writer tests
  - pkg/security: 50.6% → 85.8% coverage with command, path, and input validation tests
  - pkg/integration: 34.1% → 46.6% coverage with event bus tests
- CSV format tooltip to data table in Step 1 (#547)

### Fixed
- Preserve unselected rows in table after plot selection (#544)
- ReDoS vulnerability in error parsing (#537)
- Z-axis component selector for 3D Biplot visualization (#535)
- LaTeX math duplication by importing ui-components CSS (#534)
- Linter errors in path security tests and output.go

### Changed
- Default coef0 value for polynomial kernel from 0 to 1 (#542)
- Simplified embedded stock dataset for non-traders (#539)
- Removed redundant CSV format text in Step 1 (#548)

## [1.1.1] - 2025-09-25

### Fixed
- Fixed Z-axis component selector missing in 3D Biplot visualization (#535)
- Fixed LaTeX math duplication in UI components by importing CSS correctly (#534)
- Resolved ReDoS vulnerability in ENOENT error parsing (#537)
- Fixed multiple CI/CD configuration issues with golangci-lint v2
- Fixed Wails embed pattern exclusions in linter checks

### Changed
- Simplified embedded stock dataset for better accessibility to non-traders (#539)

### Added
- Experimental non-linear datasets for testing kernel PCA methods

### Documentation
- Updated release guide with lessons learned from v1.1.0 release

## [1.1.0] - 2025-09-23

### Added
- **Temporal PCA (SSA/Time-Delay PCA)**: Complete implementation for time series analysis (#424)
  - Variable Importance Plot for temporal analysis (#502)
  - Biplot support for temporal PCA (#442)
  - SSA documentation and examples (#525, #527)
- **Kernel PCA enhancements**: Additional visualizations and model summary (#483)
- **Validation framework**: sklearn validation tests for mathematical correctness (#460)
- **Shared UI components package**: Modular architecture for frontend components (#453)
- **Stock market datasets**: US technology stocks and enriched market factor data (#514, #475)
- **Privacy audit system**: Comprehensive privacy verification framework (#518)
- **Code quality tools**: golangci-lint integration for Go code analysis (#523)
- **Cross-method validation**: Comprehensive validation between PCA methods (#482)
- **Numerical stability tests**: Tests for algorithm robustness (#463)
- **Error boundaries**: Standardized error handling in frontend (#454)
- **Group coloring in diagnostic plots** (#447)
- **Range syntax for exclusions**: Support for ranges in --exclude-rows/columns (#417)
- **Row index coloring option** for sample visualization (#429)
- **Automatic documentation sync** between main docs and desktop app (#434)

### Fixed
- Font scaling in 3D plot axis labels (#526)
- Color palette control visibility issues (#512)
- Watermark missing in Temporal Variable Importance plot (#510)
- TypeScript/React linting errors (#520)
- 3D plot label display feature (#498)
- File loading in GoPCA Desktop (#494)
- Preprocessing settings when switching from Kernel PCA (#488)
- CLI feedback for variance explained criterion (#478)
- Temporal PCA explained variance display (#476)
- Biplot crash with Kernel PCA (#469)
- Error dialog alignment and overflow (#466)
- sklearn validation test handling (#474)
- CSV loading performance for large datasets (#427)
- Table selection state reset after filtering (#420)

### Changed
- Standardized file dialogs and button naming across apps (#492)
- Removed MET dataset from embedded resources (#490)
- Improved developer documentation structure (#524)
- Enhanced Plotly selection tools in BiPlot and DiagnosticPlot (#445)

### Documentation
- Improved temporal PCA documentation with figures (#527)
- Added SSA/temporal analysis info to README (#525)
- Added PCA literature summaries for reference (#444)
- Fixed markdown table rendering issues (#439)
- Comprehensive frontend code audit and roadmap (#452)

## [1.0.2] - 2025-09-02

### Fixed
- Fixed critical bug where row and column exclusions were not being applied before PCA analysis (#419)
  - Exclusions are now correctly applied to the data matrix before any PCA computations
  - This ensures accurate results when users exclude specific rows or columns from analysis

### Documentation
- Added comprehensive Git workflow documentation (#414)
  - Created detailed guides for the project's Git workflow
  - Helps contributors understand branch strategies and contribution process

## [1.0.1] - 2025-08-31

### Fixed
- Fixed 3D Scores Plot to properly color samples when using target columns (continuous data) (#410)
- Previously only categorical coloring worked; now continuous coloring matches the behavior of 3D Biplot

### Added
- Added missing help text entries for GUI elements in GoPCA Desktop (#411)
  - Added help entries for logo click, documentation button, font size control, 3D plots, and diagnostic threshold
  - Added HelpWrapper components to logo, documentation button, and theme toggle
- Created comprehensive help-content.json for GoCSV Desktop as foundation for future help system implementation
- Created new issue #412 for implementing help system in GoCSV Desktop

### Documentation
- Significantly enhanced CLAUDE.md development guide with:
  - Common Issues & Solutions section
  - Project Structure Quick Reference
  - Release Checklist
  - Performance Optimization Guidelines
  - Security Considerations
  - Testing Strategy
  - Git Aliases
  - MCP Tool Best Practices

## [1.0.0] - 2025-08-31

### Added
- Initial public release of GoPCA Suite
- Three integrated applications:
  - pca CLI for command-line PCA analysis
  - GoPCA Desktop for interactive data exploration
  - GoCSV Desktop for data preparation
- Multiple PCA algorithms: SVD, NIPALS, and Kernel PCA
- Comprehensive preprocessing options
- Rich visualization suite including 2D and 3D plots
- Cross-platform support (Windows, macOS, Linux)
- JSON model export/import capability
- Extensive documentation and help system

[1.2.0]: https://github.com/bitjungle/gopca/compare/v1.1.8...v1.2.0
[1.1.8]: https://github.com/bitjungle/gopca/releases/tag/v1.1.8
[1.1.7]: https://github.com/bitjungle/gopca/releases/tag/v1.1.7
[1.1.6]: https://github.com/bitjungle/gopca/releases/tag/v1.1.6
[1.1.5]: https://github.com/bitjungle/gopca/releases/tag/v1.1.5
[1.1.4]: https://github.com/bitjungle/gopca/releases/tag/v1.1.4
[1.1.3]: https://github.com/bitjungle/gopca/releases/tag/v1.1.3
[1.1.2]: https://github.com/bitjungle/gopca/releases/tag/v1.1.2
[1.1.1]: https://github.com/bitjungle/gopca/releases/tag/v1.1.1
[1.1.0]: https://github.com/bitjungle/gopca/releases/tag/v1.1.0
[1.0.2]: https://github.com/bitjungle/gopca/releases/tag/v1.0.2
[1.0.1]: https://github.com/bitjungle/gopca/releases/tag/v1.0.1
[1.0.0]: https://github.com/bitjungle/gopca/releases/tag/v1.0.0