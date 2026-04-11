// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

/**
 * Structured logger that suppresses all output in production builds.
 *
 * In development (import.meta.env.DEV === true) all calls pass through to the
 * browser console.  In production builds Vite tree-shakes the console calls
 * away entirely, so no user-visible output leaks.
 *
 * Usage:
 *   import { logger } from '../utils/logger';
 *   logger.error('Something went wrong:', err);
 *   logger.warn('Unexpected state:', value);
 *   logger.debug('Trace point reached');
 */
const isDev = import.meta.env.DEV;

export const logger = {
    error: (...args: unknown[]): void => { if (isDev) console.error(...args); },
    warn:  (...args: unknown[]): void => { if (isDev) console.warn(...args); },
    debug: (...args: unknown[]): void => { if (isDev) console.log(...args); },
};
