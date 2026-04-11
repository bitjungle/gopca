// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Mock for @gopca/ui-components — used in Vitest.
import React from 'react';
import { vi } from 'vitest';

// Simple pass-through components
export const ThemeProvider = ({ children }: { children: React.ReactNode }) => <>{children}</>;
export const ThemeToggle = () => <button>theme</button>;
export const ErrorBoundary = ({ children }: { children: React.ReactNode; onError?: () => void }) => <>{children}</>;
export const ErrorAlert = ({ message }: { message: string; type?: string; title?: string; onDismiss?: () => void }) =>
    <div data-testid="error-alert">{message}</div>;
export const ConfirmDialog = ({ isOpen, children }: { isOpen: boolean; children?: React.ReactNode; onClose?: () => void; onConfirm?: () => void; title?: string; message?: string; confirmText?: string; cancelText?: string }) =>
    isOpen ? <div data-testid="confirm-dialog">{children}</div> : null;
export const CustomSelect = ({ value, onChange, options }: { value: string; onChange: (v: string) => void; options: { value: string; label: string }[]; className?: string }) =>
    <select data-testid="custom-select" value={value} onChange={(e) => onChange(e.target.value)}>
        {options.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
    </select>;
export const DocumentationViewer = ({ isOpen }: { isOpen: boolean; onClose?: () => void }) =>
    isOpen ? <div data-testid="documentation-viewer" /> : null;
export const AboutDialog = ({ isOpen, version }: { isOpen: boolean; onClose?: () => void; version?: string }) =>
    isOpen ? <div data-testid="about-dialog">{version}</div> : null;

// Hooks
export const useTheme = vi.fn().mockReturnValue({ theme: 'light', toggleTheme: vi.fn() });
export const useChartTheme = vi.fn().mockReturnValue({ backgroundColor: '#fff', gridColor: '#eee' });

// Utility
export const setupPlotlyWailsIntegration = vi.fn();
export const PCAScoresPlot = vi.fn().mockReturnValue(null);
