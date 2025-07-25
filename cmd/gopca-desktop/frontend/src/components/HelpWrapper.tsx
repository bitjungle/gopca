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