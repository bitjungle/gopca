// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

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