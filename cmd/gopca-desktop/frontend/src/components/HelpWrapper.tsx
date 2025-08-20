// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { useHelpHover } from '../hooks/useHelpHover';

interface HelpWrapperProps {
  helpKey: string;
  children: React.ReactNode;
  className?: string;
}

export const HelpWrapper: React.FC<HelpWrapperProps> = ({
  helpKey,
  children,
  className
}) => {
  const helpRef = useHelpHover(helpKey);

  return (
    <div ref={helpRef as React.RefObject<HTMLDivElement>} className={className}>
      {children}
    </div>
  );
};