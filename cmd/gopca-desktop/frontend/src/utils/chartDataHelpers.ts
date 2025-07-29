import { 
  createQualitativeColorMap, 
  getSequentialColorScale,
  QualitativePaletteName,
  SequentialPaletteName 
} from './colorPalettes';

export interface ColorMappingResult {
  color: string;
  group: string;
  value?: number;
}

export const getColorForDataPoint = (
  index: number,
  groupType: 'categorical' | 'continuous',
  groupLabels?: string[],
  groupValues?: number[],
  qualitativePalette?: QualitativePaletteName,
  sequentialPalette?: SequentialPaletteName
): ColorMappingResult => {
  let color = '#3B82F6'; // Default color
  let group = 'Unknown';
  let value: number | undefined;
  
  if (groupType === 'categorical' && groupLabels) {
    group = groupLabels[index] || 'Unknown';
    if (group && qualitativePalette) {
      const colorMap = createQualitativeColorMap(groupLabels, qualitativePalette);
      color = colorMap.get(group) || color;
    }
  } else if (groupType === 'continuous' && groupValues && sequentialPalette) {
    const val = groupValues[index];
    value = val;
    if (!isNaN(val) && isFinite(val)) {
      const validValues = groupValues.filter(v => !isNaN(v) && isFinite(v));
      if (validValues.length > 0) {
        const min = Math.min(...validValues);
        const max = Math.max(...validValues);
        color = getSequentialColorScale(val, min, max, sequentialPalette);
        group = val.toFixed(2); // For display purposes
      }
    }
  }
  
  return { color, group, value };
};