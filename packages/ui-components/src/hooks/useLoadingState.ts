// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

/**
 * Loading state management utilities
 * Following DRY principle to avoid repetitive loading state patterns
 */

import { useState, useCallback } from 'react';

/**
 * Custom hook for managing loading states
 */
export function useLoadingState(initialState = false) {
  const [isLoading, setIsLoading] = useState(initialState);

  const withLoading = useCallback(async <T,>(
    operation: () => Promise<T>
  ): Promise<T | null> => {
    setIsLoading(true);
    try {
      const result = await operation();
      return result;
    } finally {
      setIsLoading(false);
    }
  }, []);

  return { isLoading, setIsLoading, withLoading };
}

/**
 * Combine multiple loading states
 */
export function useMultipleLoadingStates() {
  const [loadingStates, setLoadingStates] = useState<Record<string, boolean>>({});

  const setLoading = useCallback((key: string, value: boolean) => {
    setLoadingStates(prev => ({ ...prev, [key]: value }));
  }, []);

  const isAnyLoading = Object.values(loadingStates).some(state => state);

  const withLoading = useCallback(async <T,>(
    key: string,
    operation: () => Promise<T>
  ): Promise<T | null> => {
    setLoading(key, true);
    try {
      const result = await operation();
      return result;
    } finally {
      setLoading(key, false);
    }
  }, [setLoading]);

  return { loadingStates, isAnyLoading, setLoading, withLoading };
}