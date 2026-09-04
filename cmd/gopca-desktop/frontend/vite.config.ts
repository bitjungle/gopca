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

/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  optimizeDeps: {
    // Pre-bundle plotly so Vite doesn't re-optimize it mid-session when the
    // first lazy-loaded plot component is mounted. Without this, Vite discovers
    // plotly.js-dist-min at runtime, forces a reload, and the Wails dev-server
    // proxy panics on the cancelled request.
    include: ['plotly.js-dist-min'],
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
    alias: {
      // Redirect Wails bindings and ui-components to mocks in test environment
      '../wailsjs/go/main/App': path.resolve(__dirname, 'src/__mocks__/wailsjs/go/main/App.ts'),
      '../../wailsjs/go/main/App': path.resolve(__dirname, 'src/__mocks__/wailsjs/go/main/App.ts'),
      '../wailsjs/go/models': path.resolve(__dirname, 'src/__mocks__/wailsjs/go/models.ts'),
      '../../wailsjs/go/models': path.resolve(__dirname, 'src/__mocks__/wailsjs/go/models.ts'),
      '../../../wailsjs/go/models': path.resolve(__dirname, 'src/__mocks__/wailsjs/go/models.ts'),
      '../wailsjs/runtime/runtime': path.resolve(__dirname, 'src/__mocks__/wailsjs/runtime/runtime.ts'),
      '../../wailsjs/runtime/runtime': path.resolve(__dirname, 'src/__mocks__/wailsjs/runtime/runtime.ts'),
      '@gopca/ui-components': path.resolve(__dirname, 'src/__mocks__/@gopca/ui-components.tsx'),
    },
  },
});
