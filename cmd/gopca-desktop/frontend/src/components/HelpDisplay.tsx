import React from 'react';

interface HelpDisplayProps {
  helpKey: string | null;
  title: string;
  text: string;
}

export const HelpDisplay: React.FC<HelpDisplayProps> = ({ helpKey, title, text }) => {
  if (!helpKey) {
    return (
      <div className="flex items-center justify-center text-gray-500 dark:text-gray-400">
        <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span className="text-sm">Hover over any element for help</span>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center justify-center max-w-md mx-auto animate-fadeIn">
      <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-1">
        {title}
      </h3>
      <p className="text-xs text-gray-600 dark:text-gray-300 text-center leading-relaxed">
        {text}
      </p>
    </div>
  );
};