import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { PaletteType } from '../utils/colorPalettes';

interface PaletteContextType {
  paletteType: PaletteType;
  setPaletteType: (type: PaletteType) => void;
}

const PaletteContext = createContext<PaletteContextType | undefined>(undefined);

const PALETTE_STORAGE_KEY = 'gopca-palette-type';

export const PaletteProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  // Initialize from localStorage or default to qualitative
  const [paletteType, setPaletteType] = useState<PaletteType>(() => {
    const stored = localStorage.getItem(PALETTE_STORAGE_KEY);
    return (stored === 'qualitative' || stored === 'sequential') ? stored : 'qualitative';
  });

  // Save to localStorage whenever palette changes
  useEffect(() => {
    localStorage.setItem(PALETTE_STORAGE_KEY, paletteType);
  }, [paletteType]);

  const value = {
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