/**
 * Standardized error message formatting for consistent UX across applications
 */

export type ErrorSeverity = 'error' | 'warning' | 'info';

export interface FormattedError {
  severity: ErrorSeverity;
  title: string;
  message: string;
  suggestion?: string;
  code?: string;
  timestamp: Date;
}

/**
 * Standard error message templates
 */
export const ErrorTemplates = {
  // File errors
  FILE_NOT_FOUND: (filename: string): FormattedError => ({
    severity: 'error',
    title: 'File Not Found',
    message: `The file "${filename}" could not be found.`,
    suggestion: 'Please check the file path and try again.',
    code: 'E001',
    timestamp: new Date()
  }),

  FILE_READ_ERROR: (filename: string, error?: string): FormattedError => ({
    severity: 'error',
    title: 'File Read Error',
    message: `Failed to read file "${filename}".${error ? ` ${error}` : ''}`,
    suggestion: 'Ensure the file exists and you have permission to read it.',
    code: 'E002',
    timestamp: new Date()
  }),

  FILE_TOO_LARGE: (maxSize: string): FormattedError => ({
    severity: 'error',
    title: 'File Too Large',
    message: `The selected file exceeds the maximum size of ${maxSize}.`,
    suggestion: 'Please select a smaller file or split your data.',
    code: 'E003',
    timestamp: new Date()
  }),

  // CSV errors
  CSV_PARSE_ERROR: (line?: number, detail?: string): FormattedError => ({
    severity: 'error',
    title: 'CSV Parse Error',
    message: line
      ? `Error parsing CSV at line ${line}.${detail ? ` ${detail}` : ''}`
      : `Failed to parse CSV file.${detail ? ` ${detail}` : ''}`,
    suggestion: 'Check that your file is properly formatted CSV with consistent columns.',
    code: 'E010',
    timestamp: new Date()
  }),

  NO_NUMERIC_DATA: (): FormattedError => ({
    severity: 'error',
    title: 'No Numeric Data Found',
    message: 'The CSV file contains no numeric columns suitable for PCA.',
    suggestion: 'PCA requires at least 2 numeric columns. Check your data format.',
    code: 'E011',
    timestamp: new Date()
  }),

  INSUFFICIENT_DATA: (required: number, found: number): FormattedError => ({
    severity: 'error',
    title: 'Insufficient Data',
    message: `PCA requires at least ${required} samples, but only ${found} were found.`,
    suggestion: 'Add more data rows or check for excluded samples.',
    code: 'E012',
    timestamp: new Date()
  }),

  // PCA errors
  PCA_COMPUTATION_ERROR: (detail?: string): FormattedError => ({
    severity: 'error',
    title: 'PCA Computation Failed',
    message: `Failed to compute PCA.${detail ? ` ${detail}` : ''}`,
    suggestion: 'Try different preprocessing options or check for invalid data values.',
    code: 'E020',
    timestamp: new Date()
  }),

  INVALID_COMPONENTS: (requested: number, max: number): FormattedError => ({
    severity: 'error',
    title: 'Invalid Component Count',
    message: `Cannot extract ${requested} components. Maximum available is ${max}.`,
    suggestion: `Choose a value between 1 and ${max}.`,
    code: 'E021',
    timestamp: new Date()
  }),

  SINGULAR_MATRIX: (): FormattedError => ({
    severity: 'error',
    title: 'Singular Matrix',
    message: 'The data matrix is singular or nearly singular.',
    suggestion: 'Remove constant columns or highly correlated features.',
    code: 'E022',
    timestamp: new Date()
  }),

  // Validation warnings
  MISSING_VALUES_DETECTED: (count: number): FormattedError => ({
    severity: 'warning',
    title: 'Missing Values Detected',
    message: `Found ${count} missing values in the dataset.`,
    suggestion: 'Consider using NIPALS method or imputation strategy.',
    code: 'W001',
    timestamp: new Date()
  }),

  OUTLIERS_DETECTED: (count: number): FormattedError => ({
    severity: 'warning',
    title: 'Potential Outliers',
    message: `Detected ${count} potential outlier${count > 1 ? 's' : ''} in the data.`,
    suggestion: 'Review the diagnostic plots to identify and handle outliers.',
    code: 'W002',
    timestamp: new Date()
  }),

  LOW_VARIANCE_FEATURES: (features: string[]): FormattedError => ({
    severity: 'warning',
    title: 'Low Variance Features',
    message: `Features with very low variance: ${features.slice(0, 3).join(', ')}${features.length > 3 ? '...' : ''}`,
    suggestion: 'Consider removing constant or near-constant features.',
    code: 'W003',
    timestamp: new Date()
  }),

  // Info messages
  ANALYSIS_COMPLETE: (time: number): FormattedError => ({
    severity: 'info',
    title: 'Analysis Complete',
    message: `PCA analysis completed successfully in ${time.toFixed(2)} seconds.`,
    timestamp: new Date()
  }),

  DATA_LOADED: (rows: number, cols: number): FormattedError => ({
    severity: 'info',
    title: 'Data Loaded',
    message: `Loaded ${rows} samples with ${cols} features.`,
    timestamp: new Date()
  }),

  // Network errors
  NETWORK_ERROR: (): FormattedError => ({
    severity: 'error',
    title: 'Network Error',
    message: 'Failed to connect to the server.',
    suggestion: 'Check your internet connection and try again.',
    code: 'E030',
    timestamp: new Date()
  }),

  // Export errors
  EXPORT_FAILED: (format: string): FormattedError => ({
    severity: 'error',
    title: 'Export Failed',
    message: `Failed to export data as ${format}.`,
    suggestion: 'Try a different format or check disk space.',
    code: 'E040',
    timestamp: new Date()
  })
};

/**
 * Format error for display
 */
export function formatErrorMessage(error: FormattedError): string {
  let message = error.message;
  if (error.suggestion) {
    message += ` ${error.suggestion}`;
  }
  if (error.code) {
    message += ` (${error.code})`;
  }
  return message;
}

/**
 * Get icon for error severity
 */
export function getErrorIcon(severity: ErrorSeverity): string {
  switch (severity) {
    case 'error':
      return '❌';
    case 'warning':
      return '⚠️';
    case 'info':
      return 'ℹ️';
    default:
      return '';
  }
}

/**
 * Get color class for error severity (Tailwind CSS)
 */
export function getErrorColorClass(severity: ErrorSeverity): string {
  switch (severity) {
    case 'error':
      return 'text-red-600 dark:text-red-400';
    case 'warning':
      return 'text-yellow-600 dark:text-yellow-400';
    case 'info':
      return 'text-blue-600 dark:text-blue-400';
    default:
      return 'text-gray-600 dark:text-gray-400';
  }
}

/**
 * Get background color class for error severity (Tailwind CSS)
 */
export function getErrorBgColorClass(severity: ErrorSeverity): string {
  switch (severity) {
    case 'error':
      return 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800';
    case 'warning':
      return 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800';
    case 'info':
      return 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800';
    default:
      return 'bg-gray-50 dark:bg-gray-900/20 border-gray-200 dark:border-gray-800';
  }
}

/**
 * Parse error object and return formatted error
 */
export function parseError(error: unknown): FormattedError {
  if (error instanceof Error) {
    // Check for specific error types
    if (error.message.includes('ENOENT')) {
      const match = error.message.match(/ENOENT.*'(.+)'/);
      const filename = match ? match[1] : 'unknown';
      return ErrorTemplates.FILE_NOT_FOUND(filename);
    }

    if (error.message.includes('CSV')) {
      return ErrorTemplates.CSV_PARSE_ERROR(undefined, error.message);
    }

    if (error.message.includes('Network')) {
      return ErrorTemplates.NETWORK_ERROR();
    }

    // Generic error
    return {
      severity: 'error',
      title: 'Error',
      message: error.message,
      timestamp: new Date()
    };
  }

  // Unknown error type
  return {
    severity: 'error',
    title: 'Unknown Error',
    message: String(error),
    timestamp: new Date()
  };
}