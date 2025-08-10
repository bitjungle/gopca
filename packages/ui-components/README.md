# GoPCA UI Components

Shared React components library for GoPCA Desktop and GoCSV applications. This package provides reusable UI components that ensure consistency and reduce code duplication across both frontend applications.

## Overview

`@gopca/ui-components` is a TypeScript-based React component library that implements common UI patterns used throughout the GoPCA ecosystem. All components are designed with dark mode support, accessibility, and customization in mind.

## Installation

This package is automatically available to both applications through npm workspaces:

```json
// In package.json
"dependencies": {
  "@gopca/ui-components": "workspace:*"
}
```

## Available Components

### ExportButton

A unified export button supporting multiple file formats with customizable handlers.

```tsx
import { ExportButton, type ExportConfig } from '@gopca/ui-components';

const exportConfigs: ExportConfig[] = [
  {
    format: 'png',
    handler: async (format) => {
      // Export logic for PNG
    }
  },
  {
    format: 'svg',
    handler: async (format) => {
      // Export logic for SVG
    }
  }
];

<ExportButton 
  formats={exportConfigs}
  label="Export Chart"
  size="md"
/>
```

**Props:**
- `formats`: Array of export configurations with format and handler
- `label`: Button label (default: "Export")
- `size`: Button size - 'sm' | 'md' | 'lg' (default: 'md')
- `disabled`: Disable the button
- `className`: Additional CSS classes

**Supported Formats:** `png`, `svg`, `csv`, `json`, `xlsx`

### FileSelector

Consistent file selection component with drag-and-drop support.

```tsx
import { FileSelector } from '@gopca/ui-components';

<FileSelector
  onFileSelect={(filePath) => handleFileSelect(filePath)}
  onBrowseClick={async () => {
    const file = await selectFile();
    return file;
  }}
  isLoading={loading}
  acceptedFormats={['.csv', '.xlsx']}
  showDragDrop={true}
  showFormatInfo={true}
/>
```

**Props:**
- `onFileSelect`: Callback when file is selected
- `onBrowseClick`: Async function to trigger file browser
- `isLoading`: Show loading state
- `acceptedFormats`: Array of accepted file extensions
- `recentFiles`: Array of recent file paths
- `showDragDrop`: Show drag-and-drop area (default: true)
- `showFormatInfo`: Show supported formats info (default: true)
- `browseLabel`: Custom browse button label
- `dragDropMessage`: Custom drag-drop message

### ConfirmDialog

Standard confirmation dialog for destructive actions with keyboard shortcuts.

```tsx
import { ConfirmDialog } from '@gopca/ui-components';

<ConfirmDialog
  isOpen={showDialog}
  onClose={() => setShowDialog(false)}
  onConfirm={() => handleDelete()}
  title="Delete File"
  message="Are you sure you want to delete this file?"
  confirmText="Delete"
  cancelText="Cancel"
  destructive={true}
/>
```

**Props:**
- `isOpen`: Control dialog visibility
- `onClose`: Callback when dialog is closed
- `onConfirm`: Callback when action is confirmed
- `title`: Dialog title (default: "Confirm")
- `message`: Confirmation message (required)
- `confirmText`: Confirm button text (default: "Confirm")
- `cancelText`: Cancel button text (default: "Cancel")
- `destructive`: Use destructive styling for dangerous actions

**Keyboard Shortcuts:**
- `Esc`: Cancel
- `Cmd/Ctrl + Enter`: Confirm

### ProgressIndicator

Versatile progress indicator for long-running operations.

```tsx
import { ProgressIndicator } from '@gopca/ui-components';

// Determinate progress
<ProgressIndicator
  progress={75}
  title="Processing Data"
  subtitle="Please wait..."
  showPercentage={true}
/>

// Indeterminate progress
<ProgressIndicator
  isIndeterminate={true}
  title="Loading..."
  message="Fetching data from server"
/>
```

**Props:**
- `progress`: Progress percentage (0-100)
- `isIndeterminate`: Show indeterminate spinner
- `title`: Main title text
- `subtitle`: Subtitle text
- `message`: Additional message
- `showPercentage`: Show percentage value (default: true)
- `size`: Size variant - 'sm' | 'md' | 'lg' (default: 'md')
- `getStatusMessage`: Custom function to generate status messages

## Shared Contexts

### ThemeProvider & useTheme

Provides consistent theming across applications.

```tsx
import { ThemeProvider, useTheme } from '@gopca/ui-components';

// Wrap your app
<ThemeProvider>
  <App />
</ThemeProvider>

// Use in components
const { theme, setTheme } = useTheme();
```

## Shared Hooks

### useLoadingState

Manage loading states with automatic cleanup.

```tsx
import { useLoadingState } from '@gopca/ui-components';

const { isLoading, withLoading } = useLoadingState();

const handleAction = async () => {
  await withLoading(async () => {
    // Your async operation
    await fetchData();
  });
};
```

### useChartTheme

Get theme-aware chart colors and styles.

```tsx
import { useChartTheme } from '@gopca/ui-components';

const chartTheme = useChartTheme();
// Returns colors and styles based on current theme
```

## Utilities

### Error Handling

```tsx
import { showError, handleAsync } from '@gopca/ui-components';

// Show error to user
showError('Failed to load data', errorDetails);

// Wrap async operations
await handleAsync(
  async () => await riskyOperation(),
  'Operation failed'
);
```

### Chart Themes

```tsx
import { getChartTheme } from '@gopca/ui-components';

const theme = getChartTheme('dark');
// Returns chart colors for dark mode
```

## Development

### Building the Package

```bash
cd packages/ui-components
npm run build
```

### Testing Changes

1. Make changes to components
2. Run `npm run build` in the ui-components directory
3. Test in either application:
   ```bash
   make pca-dev  # Test in GoPCA Desktop
   make csv-dev  # Test in GoCSV
   ```

### Adding New Components

1. Create component directory: `src/components/ComponentName/`
2. Add component file: `ComponentName.tsx`
3. Add index file: `index.ts`
4. Export from main index: `src/index.ts`
5. Add TypeScript types and documentation
6. Test in both applications

## Design Principles

1. **TypeScript First**: All components have full type definitions
2. **Dark Mode Support**: Every component works in light and dark themes
3. **Accessibility**: Keyboard navigation and ARIA labels where appropriate
4. **Customization**: className props and style overrides supported
5. **Performance**: Optimized re-renders and lazy loading where beneficial
6. **Consistency**: Unified look and feel across both applications

## Component Guidelines

When creating new shared components:

- ✅ **DO**: Make components generic and reusable
- ✅ **DO**: Provide sensible defaults
- ✅ **DO**: Support both controlled and uncontrolled modes
- ✅ **DO**: Include comprehensive TypeScript types
- ✅ **DO**: Test in both light and dark themes
- ❌ **DON'T**: Include application-specific logic
- ❌ **DON'T**: Make assumptions about the runtime environment
- ❌ **DON'T**: Use hard-coded colors (use theme system)

## Version History

### 0.1.0 (Current)
- Initial release with Phase 3 consolidation components
- ExportButton, FileSelector, ConfirmDialog, ProgressIndicator
- Theme system and shared hooks
- Error handling utilities

## Future Enhancements

Planned components for future releases:
- DataGrid - Unified table/grid component
- ChartContainer - Consistent chart wrapper with controls
- Toolbar - Reusable toolbar component
- StatusBar - Application status display
- TabPanel - Consistent tab navigation

## Contributing

When modifying shared components:
1. Ensure changes don't break existing usage in either app
2. Add/update TypeScript types
3. Test in both GoPCA Desktop and GoCSV
4. Update this documentation as needed
5. Follow project standards from CLAUDE.md

## License

Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
Licensed under the MIT License.