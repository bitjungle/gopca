// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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
      '../wailsjs/runtime/runtime': path.resolve(__dirname, 'src/__mocks__/wailsjs/runtime/runtime.ts'),
      '../../wailsjs/runtime/runtime': path.resolve(__dirname, 'src/__mocks__/wailsjs/runtime/runtime.ts'),
      '@gopca/ui-components': path.resolve(__dirname, 'src/__mocks__/@gopca/ui-components.tsx'),
    },
  },
});
