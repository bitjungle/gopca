# @gopca/ui-components

Shared UI components and utilities for GoPCA and GoCSV applications.

## Installation

This package is included as part of the GoPCA monorepo and is automatically available to both GoPCA and GoCSV frontends via npm workspaces.

## Components

### ThemeToggle

A toggle button for switching between light and dark themes.

```tsx
import { ThemeToggle } from '@gopca/ui-components';

function Header() {
  return (
    <header>
      <ThemeToggle />
    </header>
  );
}
```

### ThemeProvider

Context provider for theme management with localStorage persistence.

```tsx
import { ThemeProvider } from '@gopca/ui-components';

function App() {
  return (
    <ThemeProvider storageKey="my-app-theme" defaultTheme="dark">
      {/* Your app content */}
    </ThemeProvider>
  );
}
```

Props:
- `storageKey?: string` - LocalStorage key for persisting theme preference (default: 'gopca-theme')
- `defaultTheme?: 'light' | 'dark'` - Default theme when no preference is saved (default: 'dark')

## Hooks

### useTheme

Access the current theme and toggle function.

```tsx
import { useTheme } from '@gopca/ui-components';

function MyComponent() {
  const { theme, toggleTheme } = useTheme();
  
  return (
    <div className={theme === 'dark' ? 'dark-mode' : 'light-mode'}>
      Current theme: {theme}
      <button onClick={toggleTheme}>Toggle</button>
    </div>
  );
}
```

### useLoadingState

Manage loading states with built-in async operation handling.

```tsx
import { useLoadingState } from '@gopca/ui-components';

function DataLoader() {
  const { isLoading, withLoading } = useLoadingState();
  
  const loadData = async () => {
    const result = await withLoading(async () => {
      const response = await fetch('/api/data');
      return response.json();
    });
    
    if (result) {
      // Handle successful data
    }
  };
  
  return (
    <div>
      {isLoading ? 'Loading...' : 'Ready'}
      <button onClick={loadData} disabled={isLoading}>
        Load Data
      </button>
    </div>
  );
}
```

### useMultipleLoadingStates

Manage multiple loading states independently.

```tsx
import { useMultipleLoadingStates } from '@gopca/ui-components';

function MultiLoader() {
  const { loadingStates, isAnyLoading, withLoading } = useMultipleLoadingStates();
  
  const loadUserData = () => withLoading('user', fetchUserData);
  const loadPostsData = () => withLoading('posts', fetchPostsData);
  
  return (
    <div>
      <button onClick={loadUserData} disabled={loadingStates.user}>
        Load User {loadingStates.user && '(Loading...)'}
      </button>
      <button onClick={loadPostsData} disabled={loadingStates.posts}>
        Load Posts {loadingStates.posts && '(Loading...)'}
      </button>
      {isAnyLoading && <p>Something is loading...</p>}
    </div>
  );
}
```

### useChartTheme

Get theme-appropriate colors for Recharts visualizations.

```tsx
import { useChartTheme } from '@gopca/ui-components';
import { LineChart, Line, XAxis, YAxis, CartesianGrid } from 'recharts';

function ThemedChart({ data }) {
  const chartTheme = useChartTheme();
  
  return (
    <LineChart data={data}>
      <CartesianGrid stroke={chartTheme.gridColor} />
      <XAxis stroke={chartTheme.axisColor} />
      <YAxis stroke={chartTheme.axisColor} />
      <Line stroke={chartTheme.textColor} />
    </LineChart>
  );
}
```

## Utilities

### Error Handling

Centralized error handling utilities with configurable display.

```tsx
import { handleAsync, showError, configureErrorHandling } from '@gopca/ui-components';

// Configure custom error display (optional)
configureErrorHandling({
  showError: (error) => {
    // Use your preferred toast/notification system
    toast.error(typeof error === 'string' ? error : error.message);
  }
});

// Handle async operations with automatic error handling
async function saveData() {
  const result = await handleAsync(
    async () => {
      const response = await fetch('/api/save', { method: 'POST' });
      if (!response.ok) throw new Error('Save failed');
      return response.json();
    },
    {
      errorPrefix: 'Save operation',
      showUserError: true,
      onError: (error) => console.error('Additional error handling', error)
    }
  );
  
  if (result) {
    // Handle successful save
  }
}

// Show error manually
showError({
  message: 'Something went wrong',
  context: 'Data validation'
});
```

### Chart Theme

Get theme-appropriate colors for charts.

```tsx
import { getChartTheme } from '@gopca/ui-components';

const isDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
const colors = getChartTheme(isDarkMode);

// Use colors in your custom chart implementation
```

## Development

### Building

```bash
npm run build -w @gopca/ui-components
```

### Watch Mode

```bash
npm run dev -w @gopca/ui-components
```

## Architecture

This package follows the core development principles from CLAUDE.md:

- **DRY (Don't Repeat Yourself)**: Consolidates shared UI logic used by both applications
- **KISS (Keep It Simple, Stupid)**: Simple, focused components with clear APIs
- **SoC (Separation of Concerns)**: Clear separation between theming, loading states, and error handling

## Adding New Components

1. Create component in appropriate directory:
   - `src/components/` for UI components
   - `src/contexts/` for React contexts
   - `src/hooks/` for custom hooks
   - `src/utils/` for utility functions

2. Export from the main index.ts file

3. Build the package: `npm run build -w @gopca/ui-components`

4. Import in consuming applications: `import { NewComponent } from '@gopca/ui-components'`

## TypeScript Support

All components and utilities are fully typed. Type definitions are automatically generated during build and included in the package.