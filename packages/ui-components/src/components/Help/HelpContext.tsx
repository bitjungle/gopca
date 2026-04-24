// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { createContext, useContext, useState, useCallback, useEffect } from 'react';

export interface HelpItem {
  title: string;
  text: string;
  category: string;
}

interface HelpContent {
  help: Record<string, HelpItem>;
}

interface HelpContextType {
  currentHelp: HelpItem | null;
  currentHelpKey: string | null;
  setHelpKey: (key: string | null) => void;
  registerHelpElement: (element: HTMLElement, helpKey: string) => void;
  unregisterHelpElement: (element: HTMLElement) => void;
}

const HelpContext = createContext<HelpContextType | undefined>(undefined);

export const useHelp = () => {
  const context = useContext(HelpContext);
  if (!context) {
    throw new Error('useHelp must be used within a HelpProvider');
  }
  return context;
};

// Interface to properly type the event handlers stored on elements
interface HelpHandlers {
  handleMouseEnter: () => void;
  handleMouseLeave: () => void;
}

// Extend HTMLElement to include our custom property
interface HTMLElementWithHelpHandlers extends HTMLElement {
  _helpHandlers?: HelpHandlers;
}

interface HelpProviderProps {
  /** Application-specific help content. Must follow { help: Record<string, HelpItem> } shape. */
  content: HelpContent;
  children: React.ReactNode;
}

export const HelpProvider: React.FC<HelpProviderProps> = ({ content, children }) => {
  const [currentHelpKey, setCurrentHelpKey] = useState<string | null>(null);
  const [currentHelp, setCurrentHelp] = useState<HelpItem | null>(null);
  const [helpElements] = useState<Map<HTMLElement, string>>(new Map());

  const setHelpKey = useCallback(
    (key: string | null) => {
      setCurrentHelpKey(key);
      if (key && content.help[key]) {
        setCurrentHelp(content.help[key]);
      } else {
        setCurrentHelp(null);
      }
    },
    [content]
  );

  const registerHelpElement = useCallback(
    (element: HTMLElement, helpKey: string) => {
      helpElements.set(element, helpKey);

      const handleMouseEnter = () => setHelpKey(helpKey);
      const handleMouseLeave = () => setHelpKey(null);

      element.addEventListener('mouseenter', handleMouseEnter);
      element.addEventListener('mouseleave', handleMouseLeave);

      // Store handlers for cleanup
      (element as HTMLElementWithHelpHandlers)._helpHandlers = {
        handleMouseEnter,
        handleMouseLeave,
      };
    },
    [helpElements, setHelpKey]
  );

  const unregisterHelpElement = useCallback(
    (element: HTMLElement) => {
      helpElements.delete(element);

      const handlers = (element as HTMLElementWithHelpHandlers)._helpHandlers;
      if (handlers) {
        element.removeEventListener('mouseenter', handlers.handleMouseEnter);
        element.removeEventListener('mouseleave', handlers.handleMouseLeave);
        delete (element as HTMLElementWithHelpHandlers)._helpHandlers;
      }
    },
    [helpElements]
  );

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      helpElements.forEach((_, element) => {
        unregisterHelpElement(element);
      });
    };
  }, [helpElements, unregisterHelpElement]);

  return (
    <HelpContext.Provider
      value={{
        currentHelp,
        currentHelpKey,
        setHelpKey,
        registerHelpElement,
        unregisterHelpElement,
      }}
    >
      {children}
    </HelpContext.Provider>
  );
};
