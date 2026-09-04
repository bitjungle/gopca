// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
//
// Mock for the generated Wails model classes, used in Vitest.
//
// The real wailsjs/go/models.ts is produced by `wails generate module` at build
// time and is gitignored, so it does not exist in a fresh checkout or in CI.
// Most imports of it are type-only (config.GUIConfig, main.PCRResponse), but
// they are written as value imports, so the bundler still has to resolve the
// module; and PCAContext constructs a real ExportPCAModelRequest.
//
// Only runtime shape is needed here. Types continue to resolve against the real
// generated file through tsconfig, which is what `npm run build` type-checks
// against.

class ExportPCAModelRequest {
    constructor(source: Record<string, unknown> = {}) {
        Object.assign(this, source);
    }
}

export const main = {
    ExportPCAModelRequest
};

export const config = {};
