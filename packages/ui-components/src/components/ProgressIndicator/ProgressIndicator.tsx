// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';

export interface ProgressIndicatorProps {
  progress?: number;
  isIndeterminate?: boolean;
  title?: string;
  subtitle?: string;
  message?: string;
  showPercentage?: boolean;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  progressBarClassName?: string;
  getStatusMessage?: (progress: number) => string;
}

const defaultStatusMessages = (progress: number): string => {
  if (progress < 20) {
return 'Starting...';
}
  if (progress < 40) {
return 'Processing...';
}
  if (progress < 60) {
return 'Working...';
}
  if (progress < 80) {
return 'Almost done...';
}
  if (progress < 100) {
return 'Finalizing...';
}
  return 'Complete!';
};

export const ProgressIndicator: React.FC<ProgressIndicatorProps> = ({
  progress = 0,
  isIndeterminate = false,
  title,
  subtitle,
  message,
  showPercentage = true,
  size = 'md',
  className = '',
  progressBarClassName = '',
  getStatusMessage = defaultStatusMessages
}) => {
  const sizeConfig = {
    sm: {
      spinner: 'w-8 h-8',
      title: 'text-base',
      subtitle: 'text-xs',
      container: 'py-6'
    },
    md: {
      spinner: 'w-12 h-12',
      title: 'text-lg',
      subtitle: 'text-sm',
      container: 'py-8'
    },
    lg: {
      spinner: 'w-16 h-16',
      title: 'text-xl',
      subtitle: 'text-base',
      container: 'py-12'
    }
  };

  const config = sizeConfig[size];

  if (isIndeterminate) {
    return (
      <div className={`flex flex-col items-center justify-center ${config.container} ${className}`}>
        <div className="mb-4">
          <svg
            className={`${config.spinner} text-blue-600 dark:text-blue-400 animate-spin`}
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        </div>

        {title && (
          <h3 className={`${config.title} font-medium text-gray-900 dark:text-white mb-2`}>
            {title}
          </h3>
        )}

        {subtitle && (
          <p className={`${config.subtitle} text-gray-600 dark:text-gray-400`}>
            {subtitle}
          </p>
        )}

        {message && (
          <div className={`mt-4 ${config.subtitle} text-gray-500 dark:text-gray-400`}>
            {message}
          </div>
        )}
      </div>
    );
  }

  const clampedProgress = Math.min(100, Math.max(0, progress));

  return (
    <div className={`flex flex-col items-center justify-center ${config.container} ${className}`}>
      {title && (
        <h3 className={`${config.title} font-medium text-gray-900 dark:text-white mb-2`}>
          {title}
        </h3>
      )}

      {subtitle && (
        <p className={`${config.subtitle} text-gray-600 dark:text-gray-400 mb-6`}>
          {subtitle}
        </p>
      )}

      <div className="w-full max-w-md">
        <div className="relative pt-1">
          {showPercentage && (
            <div className="flex mb-2 items-center justify-between">
              <div>
                <span className="text-xs font-semibold inline-block py-1 px-2 uppercase rounded-full text-blue-600 bg-blue-200 dark:text-blue-200 dark:bg-blue-800">
                  Progress
                </span>
              </div>
              <div className="text-right">
                <span className="text-xs font-semibold inline-block text-blue-600 dark:text-blue-400">
                  {clampedProgress}%
                </span>
              </div>
            </div>
          )}

          <div className="overflow-hidden h-2 mb-4 text-xs flex rounded bg-gray-200 dark:bg-gray-700">
            <div
              style={{ width: `${clampedProgress}%` }}
              className={`shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-blue-600 dark:bg-blue-400 transition-all duration-300 ${progressBarClassName}`}
            />
          </div>
        </div>
      </div>

      <div className={`mt-4 ${config.subtitle} text-gray-500 dark:text-gray-400`}>
        {message || getStatusMessage(clampedProgress)}
      </div>
    </div>
  );
};