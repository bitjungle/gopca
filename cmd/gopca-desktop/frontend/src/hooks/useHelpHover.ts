import { useEffect, useRef } from 'react';
import { useHelp } from '../contexts/HelpContext';

export function useHelpHover(helpKey: string) {
  const ref = useRef<HTMLDivElement>(null);
  const { registerHelpElement, unregisterHelpElement } = useHelp();

  useEffect(() => {
    const element = ref.current;
    if (element && helpKey) {
      registerHelpElement(element, helpKey);
      return () => {
        unregisterHelpElement(element);
      };
    }
  }, [helpKey, registerHelpElement, unregisterHelpElement]);

  return ref;
}