// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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