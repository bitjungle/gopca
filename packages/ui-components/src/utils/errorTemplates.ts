// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

export interface ErrorTemplate {
  title: string;
  message: string;
  suggestion?: string;
  code?: string;
}

export const ErrorTemplates = {
  // File operations
  FILE_LOAD: {
    title: 'File Loading Error',
    message: 'Unable to load the selected file',
    suggestion: 'Please check the file format and try again',
    code: 'FILE_001'
  },
  FILE_PARSE: {
    title: 'File Parsing Error',
    message: 'Unable to parse the file contents',
    suggestion: 'Ensure the file is in a valid CSV format',
    code: 'FILE_002'
  },
  FILE_TOO_LARGE: {
    title: 'File Too Large',
    message: 'The selected file exceeds the maximum size limit',
    suggestion: 'Please select a smaller file or split your data',
    code: 'FILE_003'
  },
  FILE_EMPTY: {
    title: 'Empty File',
    message: 'The selected file appears to be empty',
    suggestion: 'Please select a file with data',
    code: 'FILE_004'
  },

  // PCA execution
  PCA_EXECUTION: {
    title: 'Analysis Error',
    message: 'An error occurred during PCA analysis',
    suggestion: 'Please check your data and parameters',
    code: 'PCA_001'
  },
  PCA_INSUFFICIENT_DATA: {
    title: 'Insufficient Data',
    message: 'Not enough data points for PCA analysis',
    suggestion: 'PCA requires at least 2 samples and 2 variables',
    code: 'PCA_002'
  },
  PCA_INVALID_PARAMS: {
    title: 'Invalid Parameters',
    message: 'The specified parameters are invalid',
    suggestion: 'Please check the number of components and other settings',
    code: 'PCA_003'
  },
  PCA_NUMERIC_DATA: {
    title: 'Non-Numeric Data',
    message: 'PCA requires numeric data',
    suggestion: 'Please ensure all selected columns contain numeric values',
    code: 'PCA_004'
  },

  // Visualization
  VIZ_RENDER: {
    title: 'Visualization Error',
    message: 'Unable to render the visualization',
    suggestion: 'Try refreshing the page or selecting different parameters',
    code: 'VIZ_001'
  },
  VIZ_NO_DATA: {
    title: 'No Data Available',
    message: 'No data available for visualization',
    suggestion: 'Please run PCA analysis first',
    code: 'VIZ_002'
  },
  VIZ_INCOMPATIBLE: {
    title: 'Incompatible Visualization',
    message: 'This visualization is not available for the current data',
    suggestion: 'Try a different visualization type',
    code: 'VIZ_003'
  },

  // Export
  EXPORT_FAILED: {
    title: 'Export Failed',
    message: 'Unable to export the data',
    suggestion: 'Please try again or use a different format',
    code: 'EXP_001'
  },
  EXPORT_NO_DATA: {
    title: 'No Data to Export',
    message: 'There is no data available to export',
    suggestion: 'Please load data and run analysis first',
    code: 'EXP_002'
  },

  // Network
  NETWORK_ERROR: {
    title: 'Network Error',
    message: 'Unable to connect to the server',
    suggestion: 'Please check your internet connection',
    code: 'NET_001'
  },
  NETWORK_TIMEOUT: {
    title: 'Request Timeout',
    message: 'The request took too long to complete',
    suggestion: 'Please try again',
    code: 'NET_002'
  },

  // Generic
  UNKNOWN_ERROR: {
    title: 'Unexpected Error',
    message: 'An unexpected error occurred',
    suggestion: 'Please try again or contact support if the problem persists',
    code: 'GEN_001'
  },
  COMPONENT_ERROR: {
    title: 'Component Error',
    message: 'An error occurred in this component',
    suggestion: 'Try refreshing the page',
    code: 'GEN_002'
  },
  MEMORY_ERROR: {
    title: 'Memory Error',
    message: 'The application is running low on memory',
    suggestion: 'Try closing other applications or reducing data size',
    code: 'GEN_003'
  }
} as const;

/**
 * Get an error template by code
 */
export function getErrorTemplate(code: string): ErrorTemplate | undefined {
  return Object.values(ErrorTemplates).find(template => template.code === code);
}

/**
 * Format an error message with custom values
 */
export function formatErrorMessage(
  template: ErrorTemplate,
  customValues?: Partial<ErrorTemplate>
): ErrorTemplate {
  return {
    ...template,
    ...customValues
  };
}