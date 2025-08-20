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
    if (trace.x) {
headers.add('x');
}
    if (trace.y) {
headers.add('y');
}
    if (trace.z) {
headers.add('z');
}
    if (trace.text) {
headers.add('label');
}
    if (trace.name) {
headers.add('group');
}
  });

  // Build CSV content
  const rows: string[] = [];
  rows.push(Array.from(headers).join(','));

  // Combine all traces
  data.forEach(trace => {
    const length = trace.x?.length || trace.y?.length || 0;
    for (let i = 0; i < length; i++) {
      const row: string[] = [];
      if (headers.has('x')) {
row.push(trace.x?.[i] ?? '');
}
      if (headers.has('y')) {
row.push(trace.y?.[i] ?? '');
}
      if (headers.has('z')) {
row.push(trace.z?.[i] ?? '');
}
      if (headers.has('label')) {
row.push(trace.text?.[i] ?? '');
}
      if (headers.has('group')) {
row.push(trace.name ?? '');
}
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
 * Returns empty array - we'll patch Plotly's download mechanism instead
 */
export function getExportMenuItems(): any[] {
  return [];
}

/**
 * Setup Plotly to work with Wails SaveFile API
 * This intercepts all Plotly export mechanisms to use Wails when available
 */
export function setupPlotlyWailsIntegration(): void {
  console.info('Starting Plotly-Wails integration setup...');

  // Track blob URLs and their associated blobs
  const blobUrlMap = new Map<string, Blob>();

  // Intercept URL.createObjectURL to track blob URLs
  const originalCreateObjectURL = URL.createObjectURL;
  URL.createObjectURL = function(blob: Blob | MediaSource): string {
    const url = originalCreateObjectURL.call(URL, blob);
    if (blob instanceof Blob) {
      blobUrlMap.set(url, blob);
      console.info('Tracked blob URL:', url);
    }
    return url;
  };

  // Intercept URL.revokeObjectURL to clean up tracking
  const originalRevokeObjectURL = URL.revokeObjectURL;
  URL.revokeObjectURL = function(url: string): void {
    blobUrlMap.delete(url);
    originalRevokeObjectURL.call(URL, url);
  };

  // Helper function to convert blob to data URL
  const blobToDataURL = (blob: Blob): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result as string);
      reader.onerror = reject;
      reader.readAsDataURL(blob);
    });
  };

  // Intercept anchor element clicks to catch downloads
  const originalAnchorClick = HTMLAnchorElement.prototype.click;
  HTMLAnchorElement.prototype.click = function() {
    const anchor = this;
    
    // Check if this is a download link
    if (anchor.download && anchor.href) {
      const saveFileFunc = (window as any).go?.main?.App?.SaveFile;
      
      if (saveFileFunc) {
        console.info('Intercepting anchor download click');
        console.info('Download attribute:', anchor.download);
        console.info('Href:', anchor.href.substring(0, 100) + '...');
        
        // Handle the download through Wails
        (async () => {
          try {
            let dataUrl = anchor.href;
            
            // If it's a blob URL, convert to data URL
            if (dataUrl.startsWith('blob:')) {
              const blob = blobUrlMap.get(dataUrl);
              if (blob) {
                console.info('Converting blob URL to data URL');
                dataUrl = await blobToDataURL(blob);
              } else {
                console.warn('Blob not found for URL:', dataUrl);
                // Try to fetch the blob
                try {
                  const response = await fetch(dataUrl);
                  const fetchedBlob = await response.blob();
                  dataUrl = await blobToDataURL(fetchedBlob);
                } catch (fetchError) {
                  console.error('Failed to fetch blob:', fetchError);
                  // Fall back to original behavior
                  originalAnchorClick.call(anchor);
                  return;
                }
              }
            }
            
            // Use Wails SaveFile
            console.info('Calling Wails SaveFile with filename:', anchor.download);
            await saveFileFunc(anchor.download, dataUrl);
            console.info('File saved via Wails:', anchor.download);
            
            // Clean up blob URL if needed
            if (anchor.href.startsWith('blob:')) {
              URL.revokeObjectURL(anchor.href);
            }
          } catch (error) {
            console.error('Failed to save via Wails:', error);
            // Fall back to original behavior
            originalAnchorClick.call(anchor);
          }
        })();
        
        // Don't call the original click to prevent browser download
        return;
      }
    }
    
    // For non-download clicks, use original behavior
    originalAnchorClick.call(anchor);
  };

  // Wait for Plotly to be available
  const checkPlotly = () => {
    const Plotly = (window as any).Plotly;
    if (Plotly && Plotly.downloadImage) {
      console.info('Plotly found, patching export functions...');

      // Store original functions
      const originalDownloadImage = Plotly.downloadImage.bind(Plotly);
      const originalToImage = Plotly.toImage ? Plotly.toImage.bind(Plotly) : null;

      // Custom download handler that uses Wails
      const handlePlotlyExport = async (gd: any, opts: any = {}) => {
        console.info('Handling Plotly export with opts:', opts);

        try {
          // Generate the image data URL
          const imageOpts = {
            format: opts.format || 'png',
            width: opts.width || 1600,
            height: opts.height || 1200,
            scale: opts.scale || 2
          };

          console.info('Generating image with options:', imageOpts);
          const dataUrl = await (originalToImage || Plotly.toImage)(gd, imageOpts);

          // Check if we're in Wails environment and have SaveFile available
          let saveFileFunc = null;

          // Method 1: Direct window.go access
          if ((window as any).go?.main?.App?.SaveFile) {
            saveFileFunc = (window as any).go.main.App.SaveFile;
            console.info('Found SaveFile via window.go.main.App');
          }
          // Method 2: Check if SaveFile was injected globally
          else if ((window as any).SaveFile) {
            saveFileFunc = (window as any).SaveFile;
            console.info('Found SaveFile via window.SaveFile');
          }

          if (saveFileFunc) {
            // Use Wails SaveFile API
            const format = imageOpts.format;
            const filename = opts.filename || 'pca-plot';
            const fullFilename = `${filename}.${format}`;

            console.info(`Calling Wails SaveFile with filename: ${fullFilename}`);
            await saveFileFunc(fullFilename, dataUrl);
            console.info(`Plot saved via Wails: ${fullFilename}`);

            // Return without calling original to prevent double download
            return;
          } else {
            console.info('Wails SaveFile not found, falling back to browser download');
            // Fall back to original browser download
            return originalDownloadImage(gd, opts);
          }
        } catch (error) {
          console.error('Failed to save plot:', error);
          // Fall back to original on error
          return originalDownloadImage(gd, opts);
        }
      };

      // Replace downloadImage
      Plotly.downloadImage = handlePlotlyExport;


      console.info('Plotly-Wails integration setup complete');
    } else {
      // Retry after a short delay
      setTimeout(checkPlotly, 100);
    }
  };

  // Start checking for Plotly
  checkPlotly();
}

/**
 * Setup keyboard shortcuts for plot export
 * Ctrl/Cmd+C for copy to clipboard
 * Returns a cleanup function to remove the event listener
 */
export function setupExportShortcuts(plotElement: PlotlyHTMLElement): () => void {
  const handleKeyDown = async (e: KeyboardEvent) => {
    // Check if plot element is visible and focused area
    if (!plotElement || !document.contains(plotElement)) {
return;
}

    // Ctrl/Cmd + C for copy (when not in text input)
    if ((e.ctrlKey || e.metaKey) && e.key === 'c') {
      const activeElement = document.activeElement;
      const isTextInput = activeElement && (
        activeElement.tagName === 'INPUT' ||
        activeElement.tagName === 'TEXTAREA' ||
        (activeElement as HTMLElement).contentEditable === 'true'
      );

      if (!isTextInput) {
        e.preventDefault();
        try {
          await copyToClipboard(plotElement);
          // Optional: Show success feedback
          console.info('Plot copied to clipboard');
        } catch (error) {
          console.error('Failed to copy plot:', error);
        }
      }
    }
  };

  // Add event listener
  document.addEventListener('keydown', handleKeyDown);

  // Return cleanup function
  return () => {
    document.removeEventListener('keydown', handleKeyDown);
  };
}