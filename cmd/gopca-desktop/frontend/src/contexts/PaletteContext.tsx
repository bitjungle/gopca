// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { 
  QualitativePaletteName, 
  SequentialPaletteName 
} from '../utils/colorPalettes';

type PaletteMode = 'categorical' | 'continuous' | 'none';

interface PaletteContextType {
  // Current mode based on selected column type
  mode: PaletteMode;
  setMode: (mode: PaletteMode) => void;
  
  // Selected palettes for each mode
  qualitativePalette: QualitativePaletteName;
  setQualitativePalette: (palette: QualitativePaletteName) => void;
  
  sequentialPalette: SequentialPaletteName;
  setSequentialPalette: (palette: SequentialPaletteName) => void;
  
  // Legacy support - maps to current active palette
  paletteType: 'qualitative' | 'sequential';
  setPaletteType: (type: 'qualitative' | 'sequential') => void;
}

const PaletteContext = createContext<PaletteContextType | undefined>(undefined);

const QUALITATIVE_STORAGE_KEY = 'gopca-qualitative-palette';
const SEQUENTIAL_STORAGE_KEY = 'gopca-sequential-palette';

export const PaletteProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  // Current mode based on selected column
  const [mode, setMode] = useState<PaletteMode>('none');
  
  // Selected palette for each mode
  const [qualitativePalette, setQualitativePalette] = useState<QualitativePaletteName>(() => {
    const stored = localStorage.getItem(QUALITATIVE_STORAGE_KEY);
    return (stored as QualitativePaletteName) || 'deep';
  });
  
  const [sequentialPalette, setSequentialPalette] = useState<SequentialPaletteName>(() => {
    const stored = localStorage.getItem(SEQUENTIAL_STORAGE_KEY);
    return (stored as SequentialPaletteName) || 'rocket';
  });
  
  // Save palette selections to localStorage
  useEffect(() => {
    localStorage.setItem(QUALITATIVE_STORAGE_KEY, qualitativePalette);
  }, [qualitativePalette]);
  
  useEffect(() => {
    localStorage.setItem(SEQUENTIAL_STORAGE_KEY, sequentialPalette);
  }, [sequentialPalette]);
  
  // Legacy paletteType getter - returns the type based on current mode
  const paletteType = mode === 'continuous' ? 'sequential' : 'qualitative';
  
  // Legacy setPaletteType - does nothing as mode is now controlled by column selection
  const setPaletteType = () => {
    // No-op for backward compatibility
    console.warn('setPaletteType is deprecated. Palette type is now automatically determined by column type.');
  };
  
  const value: PaletteContextType = {
    mode,
    setMode,
    qualitativePalette,
    setQualitativePalette,
    sequentialPalette,
    setSequentialPalette,
    paletteType,
    setPaletteType,
  };
  
  return <PaletteContext.Provider value={value}>{children}</PaletteContext.Provider>;
};

export const usePalette = () => {
  const context = useContext(PaletteContext);
  if (context === undefined) {
    throw new Error('usePalette must be used within a PaletteProvider');
  }
  return context;
};