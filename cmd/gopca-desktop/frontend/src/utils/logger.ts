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
    debug: (...args: unknown[]): void => { if (isDev) console.log(...args); }
};
