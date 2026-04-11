// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Test setup: extends expect with jest-dom matchers.
import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock localStorage for tests — PaletteContext uses it in useState initializers
const localStorageMock = {
    getItem: vi.fn().mockReturnValue(null),
    setItem: vi.fn(),
    removeItem: vi.fn(),
    clear: vi.fn(),
};
Object.defineProperty(window, 'localStorage', { value: localStorageMock });
