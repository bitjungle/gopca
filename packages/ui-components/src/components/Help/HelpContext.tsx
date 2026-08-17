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
        handleMouseLeave
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
        unregisterHelpElement
      }}
    >
      {children}
    </HelpContext.Provider>
  );
};
