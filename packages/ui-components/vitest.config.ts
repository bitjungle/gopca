// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

import { defineConfig } from 'vitest/config';

// The shared components are a compiled package: both desktop applications get
// them from dist/, so a defect here reaches every app at once and is invisible
// in each app's own test run. Two chart bugs shipped that way before this file
// existed — reference lines pinned at +/-10 regardless of the data, and an
// identity line computed from Math.min() of an empty list.
//
// Only the plain-TypeScript units are covered so far. Rendering the Plotly
// components needs jsdom plus a WebGL stub, which is a larger undertaking; the
// pure geometry and formatting helpers they depend on are testable now and are
// where both of those defects actually lived.
export default defineConfig({
    test: {
        environment: 'node',
        include: ['src/**/*.test.ts']
    }
});
