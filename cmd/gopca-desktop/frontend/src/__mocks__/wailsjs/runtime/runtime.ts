// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Mock for Wails runtime — used in Vitest.
import { vi } from 'vitest';

export const EventsOn = vi.fn().mockReturnValue(() => {}); // returns unsubscribe fn
export const EventsOff = vi.fn();
export const EventsEmit = vi.fn();
export const BrowserOpenURL = vi.fn();
