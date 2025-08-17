// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import { useTheme } from '../contexts/ThemeContext';
import { getChartTheme } from '../utils/chartTheme';

export const useChartTheme = () => {
  const { theme } = useTheme();
  return {
    ...getChartTheme(theme === 'dark'),
    theme: theme as 'light' | 'dark'
  };
};