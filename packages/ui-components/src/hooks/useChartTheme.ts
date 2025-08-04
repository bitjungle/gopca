import { useTheme } from '../contexts/ThemeContext';
import { getChartTheme } from '../utils/chartTheme';

export const useChartTheme = () => {
  const { theme } = useTheme();
  return getChartTheme(theme === 'dark');
};