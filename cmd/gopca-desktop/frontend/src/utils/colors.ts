// Color palette for group visualization
export const GROUP_COLORS = [
  '#3B82F6', // blue
  '#EF4444', // red
  '#10B981', // green
  '#F59E0B', // amber
  '#8B5CF6', // violet
  '#EC4899', // pink
  '#06B6D4', // cyan
  '#F97316', // orange
  '#6366F1', // indigo
  '#84CC16', // lime
  '#14B8A6', // teal
  '#A855F7', // purple
];

// Get color for a group
export function getGroupColor(groupIndex: number): string {
  return GROUP_COLORS[groupIndex % GROUP_COLORS.length];
}

// Create a mapping from unique group labels to colors
export function createGroupColorMap(groupLabels: string[]): Map<string, string> {
  const uniqueGroups = [...new Set(groupLabels)].sort();
  const colorMap = new Map<string, string>();
  
  uniqueGroups.forEach((group, index) => {
    colorMap.set(group, getGroupColor(index));
  });
  
  return colorMap;
}