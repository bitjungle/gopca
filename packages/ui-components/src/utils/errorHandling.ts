// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

/**
 * Centralized error handling utilities for GoPCA applications
 * Following CLAUDE.md principles: KISS, DRY, and readability
 */

export interface ErrorInfo {
  message: string;
  context?: string;
  isUserError?: boolean;
}

export interface ErrorConfig {
  showError?: (error: ErrorInfo | string) => void;
}

// Default error display function (can be overridden)
let defaultShowError = (error: ErrorInfo | string): void => {
  const errorMessage = typeof error === 'string' 
    ? error 
    : error.context 
      ? `${error.context}: ${error.message}`
      : error.message;
  
  alert(errorMessage);
};

/**
 * Configure the error handling behavior
 */
export function configureErrorHandling(config: ErrorConfig): void {
  if (config.showError) {
    defaultShowError = config.showError;
  }
}

/**
 * Display error to user in a consistent manner
 */
export function showError(error: ErrorInfo | string): void {
  defaultShowError(error);
}

/**
 * Handle async operations with consistent error handling
 */
export async function handleAsync<T>(
  operation: () => Promise<T>,
  options: {
    errorPrefix?: string;
    onError?: (error: any) => void;
    showUserError?: boolean;
  } = {}
): Promise<T | null> {
  try {
    return await operation();
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error(`${options.errorPrefix || 'Error'}:`, error);
    
    if (options.showUserError !== false) {
      showError({
        message: errorMessage,
        context: options.errorPrefix
      });
    }
    
    if (options.onError) {
      options.onError(error);
    }
    
    return null;
  }
}

/**
 * Extract error message from various error types
 */
export function getErrorMessage(error: any): string {
  if (error instanceof Error) {
    return error.message;
  } else if (typeof error === 'string') {
    return error;
  } else if (error?.message) {
    return error.message;
  } else if (error?.toString) {
    return error.toString();
  } else {
    return 'Unknown error';
  }
}