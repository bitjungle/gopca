// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { createContext, useContext, useState, ReactNode } from 'react';

export type ChartLibrary = 'recharts' | 'plotly' | 'd3';

interface ChartConfig {
  provider: ChartLibrary;
  zoomEnabled: boolean;
  panEnabled: boolean;
}

interface ChartContextType {
  config: ChartConfig;
  setProvider: (provider: ChartLibrary) => void;
  setZoomEnabled: (enabled: boolean) => void;
  setPanEnabled: (enabled: boolean) => void;
}

const ChartContext = createContext<ChartContextType | undefined>(undefined);

const DEFAULT_CONFIG: ChartConfig = {
  provider: 'plotly',
  zoomEnabled: true,
  panEnabled: true,
};

export const ChartProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [config, setConfig] = useState<ChartConfig>(DEFAULT_CONFIG);

  const setProvider = (provider: ChartLibrary) => {
    setConfig(prev => ({ ...prev, provider }));
  };

  const setZoomEnabled = (enabled: boolean) => {
    setConfig(prev => ({ ...prev, zoomEnabled: enabled }));
  };

  const setPanEnabled = (enabled: boolean) => {
    setConfig(prev => ({ ...prev, panEnabled: enabled }));
  };

  return (
    <ChartContext.Provider 
      value={{
        config,
        setProvider,
        setZoomEnabled,
        setPanEnabled,
      }}
    >
      {children}
    </ChartContext.Provider>
  );
};

export const useChartConfig = () => {
  const context = useContext(ChartContext);
  if (!context) {
    throw new Error('useChartConfig must be used within a ChartProvider');
  }
  return context;
};