// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Mock for Wails Go bindings — used in Vitest (browser backend unavailable).
import { vi } from 'vitest';

export const GetVersion = vi.fn().mockResolvedValue('test-version');
export const GetGUIConfig = vi.fn().mockResolvedValue(null);
export const SaveFile = vi.fn().mockResolvedValue(undefined);
export const CheckGoCSVStatus = vi.fn().mockResolvedValue({ installed: false });
export const LoadCSVFile = vi.fn().mockResolvedValue(null);
export const LoadDatasetFile = vi.fn().mockResolvedValue(null);
export const SelectCSVFile = vi.fn().mockResolvedValue(null);
export const RunPCA = vi.fn().mockResolvedValue({ success: false, error: 'mock' });
export const ExportPCAModel = vi.fn().mockResolvedValue(undefined);
export const OpenInGoCSV = vi.fn().mockResolvedValue(undefined);
export const LaunchGoCSV = vi.fn().mockResolvedValue(undefined);
export const DownloadGoCSV = vi.fn().mockResolvedValue(undefined);
export const CalculateModelMetrics = vi.fn().mockResolvedValue({});
