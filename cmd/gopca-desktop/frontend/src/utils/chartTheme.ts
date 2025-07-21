export interface ChartTheme {
  gridColor: string;
  axisColor: string;
  textColor: string;
  backgroundColor: string;
  referenceLineColor: string;
  tooltipBackgroundColor: string;
  tooltipBorderColor: string;
  tooltipTextColor: string;
}

export const getChartTheme = (isDark: boolean): ChartTheme => {
  if (isDark) {
    return {
      gridColor: '#374151',
      axisColor: '#9CA3AF',
      textColor: '#E5E7EB',
      backgroundColor: '#1F2937',
      referenceLineColor: '#6B7280',
      tooltipBackgroundColor: '#1F2937',
      tooltipBorderColor: '#374151',
      tooltipTextColor: '#E5E7EB',
    };
  } else {
    return {
      gridColor: '#E5E7EB',
      axisColor: '#6B7280',
      textColor: '#374151',
      backgroundColor: '#FFFFFF',
      referenceLineColor: '#D1D5DB',
      tooltipBackgroundColor: '#FFFFFF',
      tooltipBorderColor: '#E5E7EB',
      tooltipTextColor: '#374151',
    };
  }
};