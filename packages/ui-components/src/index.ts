// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Components
export { ThemeToggle } from './components/ThemeToggle';
export { ExportButton } from './components/ExportButton';
export { FileSelector } from './components/FileSelector';
export { ConfirmDialog } from './components/ConfirmDialog';
export { ProgressIndicator } from './components/ProgressIndicator';
export { Dialog, DialogFooter, DialogBody } from './components/Dialog';
export { InputDialog } from './components/InputDialog';
export { SkipLinks } from './components/SkipLinks';
export { KeyboardHelp } from './components/KeyboardHelp';

// Component Types
export type { ExportButtonProps, ExportConfig, ExportFormat } from './components/ExportButton';
export type { FileSelectorProps } from './components/FileSelector';
export type { ConfirmDialogProps } from './components/ConfirmDialog';
export type { ProgressIndicatorProps } from './components/ProgressIndicator';
export type { DialogProps } from './components/Dialog';
export type { InputDialogProps } from './components/InputDialog';
export type { SkipLinksProps, SkipLink } from './components/SkipLinks';
export type { KeyboardHelpProps } from './components/KeyboardHelp';

// Contexts
export { ThemeProvider, useTheme } from './contexts/ThemeContext';

// Hooks
export { useLoadingState, useMultipleLoadingStates } from './hooks/useLoadingState';
export { useChartTheme } from './hooks/useChartTheme';
export { 
  useFocusManagement, 
  useFocusRestore, 
  useFocusTrap 
} from './hooks/useFocusManagement';
export { 
  useKeyboardShortcuts, 
  useKeyboardShortcut,
  useEscapeKey,
  getModifierKey,
  formatShortcut,
  commonShortcuts,
  type KeyboardShortcut 
} from './hooks/useKeyboardShortcuts';

// Utils
export { 
  showError, 
  handleAsync, 
  getErrorMessage, 
  configureErrorHandling,
  type ErrorInfo,
  type ErrorConfig 
} from './utils/errorHandling';
export { 
  getChartTheme,
  type ChartTheme 
} from './utils/chartTheme';
export {
  ErrorTemplates,
  formatErrorMessage,
  getErrorIcon,
  getErrorColorClass,
  getErrorBgColorClass,
  parseError,
  type FormattedError,
  type ErrorSeverity
} from './utils/errorMessages';

// Charts - Removed as part of Plotly migration
// Chart components have been replaced with Plotly PCA visualizations below

// Plotly PCA Visualizations
export {
  // Components
  PCAScoresPlot,
  PCA3DScoresPlot,
  PCAScreePlot,
  PCALoadingsPlot,
  PCABiplot,
  PCACircleOfCorrelations,
  PCADiagnosticPlot,
  PCAEigencorrelationPlot,
  // Classes for advanced usage
  PlotlyScoresPlot,
  Plotly3DScoresPlot,
  PlotlyScreePlot,
  PlotlyLoadingsPlot,
  PlotlyBiplot,
  PlotlyCircleOfCorrelations,
  PlotlyDiagnosticPlot,
  PlotlyEigencorrelationPlot,
  // Types
  type ScoresPlotData,
  type ScoresPlotConfig,
  type Scores3DPlotData,
  type Scores3DPlotConfig,
  type ScreePlotData,
  type ScreePlotConfig,
  type LoadingsPlotData,
  type LoadingsPlotConfig,
  type BiplotData,
  type BiplotConfig,
  type CircleOfCorrelationsData,
  type CircleOfCorrelationsConfig,
  type DiagnosticPlotData,
  type DiagnosticPlotConfig,
  type EigencorrelationPlotData,
  type EigencorrelationPlotConfig,
} from './charts/adapters/plotly/pca';

// Plotly Export Utils
export {
  setupPlotlyWailsIntegration
} from './charts/adapters/plotly/utils/plotlyExport';