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
        // Most tests here are pure functions and need no DOM. The documentation
        // viewer is the exception: its table of contents was wired correctly on
        // every reading of the code and still did nothing when clicked, which is
        // a failure only a rendered tree can show. Files opting in with
        // `@vitest-environment jsdom` get one.
        environment: 'node',
        include: ['src/**/*.test.ts', 'src/**/*.test.tsx']
    }
});
