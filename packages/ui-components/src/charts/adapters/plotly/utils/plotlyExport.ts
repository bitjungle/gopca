// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Export utilities for high-quality Plotly visualizations

import { PlotlyHTMLElement } from 'plotly.js';

export interface ExportConfig {
  format: 'png' | 'jpeg' | 'svg' | 'pdf';
  filename: string;
  width: number;
  height: number;
  scale: number;
}

export const PUBLICATION_EXPORT_CONFIG: ExportConfig = {
  format: 'svg',
  filename: 'pca-publication',
  width: 3200,
  height: 2400,
  scale: 4
};

export const PRESENTATION_EXPORT_CONFIG: ExportConfig = {
  format: 'png',
  filename: 'pca-presentation',
  width: 1920,
  height: 1080,
  scale: 2
};

export const WEB_EXPORT_CONFIG: ExportConfig = {
  format: 'png',
  filename: 'pca-web',
  width: 1200,
  height: 800,
  scale: 1
};

/**
 * Export plot to high-quality image
 * 
 * @param plotElement - Plotly element to export
 * @param config - Export configuration
 * @returns Promise resolving to blob URL
 */
export async function exportPlot(
  plotElement: PlotlyHTMLElement,
  config: Partial<ExportConfig> = {}
): Promise<string> {
  const finalConfig = {
    ...WEB_EXPORT_CONFIG,
    ...config
  };
  
  // Use Plotly's built-in export
  const Plotly = (window as any).Plotly;
  if (!Plotly) {
    throw new Error('Plotly not loaded');
  }
  
  return Plotly.toImage(plotElement, {
    format: finalConfig.format,
    width: finalConfig.width,
    height: finalConfig.height,
    scale: finalConfig.scale
  });
}

/**
 * Download plot as image
 * 
 * @param plotElement - Plotly element to export
 * @param config - Export configuration
 */
export async function downloadPlot(
  plotElement: PlotlyHTMLElement,
  config: Partial<ExportConfig> = {}
): Promise<void> {
  const finalConfig = {
    ...WEB_EXPORT_CONFIG,
    ...config
  };
  
  const Plotly = (window as any).Plotly;
  if (!Plotly) {
    throw new Error('Plotly not loaded');
  }
  
  return Plotly.downloadImage(plotElement, {
    format: finalConfig.format,
    width: finalConfig.width,
    height: finalConfig.height,
    scale: finalConfig.scale,
    filename: finalConfig.filename
  });
}

/**
 * Export plot data to CSV
 * 
 * @param data - Plot data
 * @param filename - Output filename
 */
export function exportToCSV(data: any[], filename: string = 'plot-data.csv'): void {
  if (!data || data.length === 0) {
    console.warn('No data to export');
    return;
  }
  
  // Extract headers from first trace
  const headers = new Set<string>();
  data.forEach(trace => {
    if (trace.x) headers.add('x');
    if (trace.y) headers.add('y');
    if (trace.z) headers.add('z');
    if (trace.text) headers.add('label');
    if (trace.name) headers.add('group');
  });
  
  // Build CSV content
  const rows: string[] = [];
  rows.push(Array.from(headers).join(','));
  
  // Combine all traces
  data.forEach(trace => {
    const length = trace.x?.length || trace.y?.length || 0;
    for (let i = 0; i < length; i++) {
      const row: string[] = [];
      if (headers.has('x')) row.push(trace.x?.[i] ?? '');
      if (headers.has('y')) row.push(trace.y?.[i] ?? '');
      if (headers.has('z')) row.push(trace.z?.[i] ?? '');
      if (headers.has('label')) row.push(trace.text?.[i] ?? '');
      if (headers.has('group')) row.push(trace.name ?? '');
      rows.push(row.join(','));
    }
  });
  
  // Create and download file
  const csvContent = rows.join('\n');
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  link.href = URL.createObjectURL(blob);
  link.download = filename;
  link.click();
  URL.revokeObjectURL(link.href);
}

/**
 * Export plot configuration to JSON
 * 
 * @param layout - Plot layout
 * @param config - Plot config
 * @param filename - Output filename
 */
export function exportConfig(
  layout: any,
  config: any,
  filename: string = 'plot-config.json'
): void {
  const exportData = {
    layout,
    config,
    timestamp: new Date().toISOString(),
    version: '2.0.0'
  };
  
  const jsonContent = JSON.stringify(exportData, null, 2);
  const blob = new Blob([jsonContent], { type: 'application/json' });
  const link = document.createElement('a');
  link.href = URL.createObjectURL(blob);
  link.download = filename;
  link.click();
  URL.revokeObjectURL(link.href);
}

/**
 * Copy plot to clipboard as image
 * 
 * @param plotElement - Plotly element to copy
 * @param config - Export configuration
 */
export async function copyToClipboard(
  plotElement: PlotlyHTMLElement,
  config: Partial<ExportConfig> = {}
): Promise<void> {
  const finalConfig = {
    ...WEB_EXPORT_CONFIG,
    format: 'png' as const,
    ...config
  };
  
  const Plotly = (window as any).Plotly;
  if (!Plotly) {
    throw new Error('Plotly not loaded');
  }
  
  // Convert to blob
  const dataUrl = await Plotly.toImage(plotElement, {
    format: 'png',
    width: finalConfig.width,
    height: finalConfig.height,
    scale: finalConfig.scale
  });
  
  // Convert data URL to blob
  const response = await fetch(dataUrl);
  const blob = await response.blob();
  
  // Copy to clipboard using Clipboard API
  if (navigator.clipboard && (navigator.clipboard as any).write) {
    const item = new ClipboardItem({ 'image/png': blob });
    await (navigator.clipboard as any).write([item]);
  } else {
    throw new Error('Clipboard API not supported');
  }
}

/**
 * Generate export menu items for modebar
 */
export function getExportMenuItems(): any[] {
  return [
    {
      name: 'Export as Publication Quality SVG',
      icon: null,
      click: (gd: PlotlyHTMLElement) => {
        downloadPlot(gd, PUBLICATION_EXPORT_CONFIG);
      }
    },
    {
      name: 'Export as Presentation PNG',
      icon: null,
      click: (gd: PlotlyHTMLElement) => {
        downloadPlot(gd, PRESENTATION_EXPORT_CONFIG);
      }
    },
    {
      name: 'Export Data as CSV',
      icon: null,
      click: (gd: PlotlyHTMLElement) => {
        const data = (gd as any).data;
        exportToCSV(data);
      }
    },
    {
      name: 'Copy to Clipboard',
      icon: null,
      click: async (gd: PlotlyHTMLElement) => {
        try {
          await copyToClipboard(gd);
          console.log('Copied to clipboard');
        } catch (error) {
          console.error('Failed to copy to clipboard:', error);
        }
      }
    }
  ];
}