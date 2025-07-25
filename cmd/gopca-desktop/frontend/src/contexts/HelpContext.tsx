import React, { createContext, useContext, useState, useCallback, useEffect } from 'react';
import helpContent from '../help/help-content.json';

interface HelpItem {
  title: string;
  text: string;
  category: string;
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

interface HelpProviderProps {
  children: React.ReactNode;
}

export const HelpProvider: React.FC<HelpProviderProps> = ({ children }) => {
  const [currentHelpKey, setCurrentHelpKey] = useState<string | null>(null);
  const [currentHelp, setCurrentHelp] = useState<HelpItem | null>(null);
  const [helpElements] = useState<Map<HTMLElement, string>>(new Map());

  const setHelpKey = useCallback((key: string | null) => {
    setCurrentHelpKey(key);
    if (key && helpContent.help[key as keyof typeof helpContent.help]) {
      setCurrentHelp(helpContent.help[key as keyof typeof helpContent.help]);
    } else {
      setCurrentHelp(null);
    }
  }, []);

  const registerHelpElement = useCallback((element: HTMLElement, helpKey: string) => {
    helpElements.set(element, helpKey);
    
    const handleMouseEnter = () => setHelpKey(helpKey);
    const handleMouseLeave = () => setHelpKey(null);
    
    element.addEventListener('mouseenter', handleMouseEnter);
    element.addEventListener('mouseleave', handleMouseLeave);
    
    // Store handlers for cleanup
    (element as any)._helpHandlers = { handleMouseEnter, handleMouseLeave };
  }, [helpElements, setHelpKey]);

  const unregisterHelpElement = useCallback((element: HTMLElement) => {
    helpElements.delete(element);
    
    const handlers = (element as any)._helpHandlers;
    if (handlers) {
      element.removeEventListener('mouseenter', handlers.handleMouseEnter);
      element.removeEventListener('mouseleave', handlers.handleMouseLeave);
      delete (element as any)._helpHandlers;
    }
  }, [helpElements]);

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