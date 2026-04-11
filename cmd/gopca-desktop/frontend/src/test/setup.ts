// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Test setup: extends expect with jest-dom matchers.
import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock localStorage for tests — PaletteContext uses it in useState initializers.
// vi.stubGlobal is used instead of Object.defineProperty so the property remains
// configurable and Vitest can restore it cleanly between test files.
vi.stubGlobal('localStorage', {
    getItem: vi.fn().mockReturnValue(null),
    setItem: vi.fn(),
    removeItem: vi.fn(),
    clear: vi.fn(),
});
