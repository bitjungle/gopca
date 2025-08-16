import { useEffect, useCallback, useState } from 'react';

export interface KeyboardShortcut {
  key: string;
  ctrl?: boolean;
  cmd?: boolean;
  shift?: boolean;
  alt?: boolean;
  description: string;
  handler: () => void;
  enabled?: boolean;
}

/**
 * Hook for managing keyboard shortcuts in the application
 */
export const useKeyboardShortcuts = (shortcuts: KeyboardShortcut[]) => {
  const [isHelpVisible, setIsHelpVisible] = useState(false);

  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      // Check for help shortcut (?)
      if (event.key === '?' && !event.ctrlKey && !event.metaKey && !event.altKey) {
        // Don't show help if user is typing in an input
        const target = event.target as HTMLElement;
        if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') {
          return;
        }
        event.preventDefault();
        setIsHelpVisible(prev => !prev);
        return;
      }

      // Check each shortcut
      for (const shortcut of shortcuts) {
        if (shortcut.enabled === false) continue;

        // Check if the key matches
        if (event.key.toLowerCase() !== shortcut.key.toLowerCase()) continue;

        // Check modifiers
        const ctrlOrCmd = navigator.platform.includes('Mac') ? event.metaKey : event.ctrlKey;
        const expectedCtrlOrCmd = shortcut.ctrl || shortcut.cmd;

        if (expectedCtrlOrCmd && !ctrlOrCmd) continue;
        if (!expectedCtrlOrCmd && ctrlOrCmd) continue;
        if (shortcut.shift && !event.shiftKey) continue;
        if (!shortcut.shift && event.shiftKey) continue;
        if (shortcut.alt && !event.altKey) continue;
        if (!shortcut.alt && event.altKey) continue;

        // Don't trigger shortcuts when typing in inputs (unless it's a global shortcut)
        const target = event.target as HTMLElement;
        const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA';
        const isGlobalShortcut = expectedCtrlOrCmd || shortcut.alt;

        if (isInput && !isGlobalShortcut) {
          return;
        }

        event.preventDefault();
        shortcut.handler();
        return;
      }
    },
    [shortcuts]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return {
    isHelpVisible,
    setIsHelpVisible,
  };
};

/**
 * Get platform-appropriate modifier key symbol
 */
export const getModifierKey = (): string => {
  return navigator.platform.includes('Mac') ? '⌘' : 'Ctrl';
};

/**
 * Format shortcut for display
 */
export const formatShortcut = (shortcut: KeyboardShortcut): string => {
  const parts: string[] = [];
  
  if (shortcut.ctrl || shortcut.cmd) {
    parts.push(getModifierKey());
  }
  if (shortcut.shift) {
    parts.push('Shift');
  }
  if (shortcut.alt) {
    parts.push('Alt');
  }
  
  // Format the key nicely
  const key = shortcut.key.length === 1 
    ? shortcut.key.toUpperCase() 
    : shortcut.key.charAt(0).toUpperCase() + shortcut.key.slice(1);
  
  parts.push(key);
  
  return parts.join('+');
};

/**
 * Common keyboard shortcuts for GoPCA applications
 */
export const commonShortcuts = {
  openFile: {
    key: 'o',
    ctrl: true,
    cmd: true,
    description: 'Open file',
  },
  saveResults: {
    key: 's',
    ctrl: true,
    cmd: true,
    description: 'Save results',
  },
  exportData: {
    key: 'e',
    ctrl: true,
    cmd: true,
    description: 'Export data',
  },
  settings: {
    key: ',',
    ctrl: true,
    cmd: true,
    description: 'Open settings',
  },
  toggleTheme: {
    key: 't',
    alt: true,
    description: 'Toggle theme',
  },
  closeDialog: {
    key: 'Escape',
    description: 'Close dialog/modal',
  },
  help: {
    key: '?',
    description: 'Show keyboard shortcuts',
  },
  refresh: {
    key: 'r',
    ctrl: true,
    cmd: true,
    description: 'Refresh data',
  },
  undo: {
    key: 'z',
    ctrl: true,
    cmd: true,
    description: 'Undo',
  },
  redo: {
    key: 'z',
    ctrl: true,
    cmd: true,
    shift: true,
    description: 'Redo',
  },
};

/**
 * Hook for single keyboard shortcut
 */
export const useKeyboardShortcut = (
  key: string,
  handler: () => void,
  options?: {
    ctrl?: boolean;
    cmd?: boolean;
    shift?: boolean;
    alt?: boolean;
    enabled?: boolean;
  }
) => {
  const shortcut: KeyboardShortcut = {
    key,
    ...options,
    description: '',
    handler,
  };

  useKeyboardShortcuts([shortcut]);
};

/**
 * Hook for escape key handling (common for closing modals)
 */
export const useEscapeKey = (handler: () => void, enabled = true) => {
  useKeyboardShortcut('Escape', handler, { enabled });
};