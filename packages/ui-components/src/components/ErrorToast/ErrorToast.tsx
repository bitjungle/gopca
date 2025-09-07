// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useEffect, useState } from 'react';

interface ErrorToastProps {
  type?: 'error' | 'warning' | 'info' | 'success';
  message: string;
  duration?: number; // milliseconds
  onDismiss?: () => void;
  position?: 'top' | 'bottom' | 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left';
  className?: string;
}

export const ErrorToast: React.FC<ErrorToastProps> = ({
  type = 'info',
  message,
  duration = 5000,
  onDismiss,
  position = 'top-right',
  className = ''
}) => {
  const [isVisible, setIsVisible] = useState(true);
  const [isExiting, setIsExiting] = useState(false);

  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        handleDismiss();
      }, duration);
      return () => clearTimeout(timer);
    }
  }, [duration]);

  const handleDismiss = () => {
    setIsExiting(true);
    setTimeout(() => {
      setIsVisible(false);
      onDismiss?.();
    }, 300); // Animation duration
  };

  if (!isVisible) return null;

  const iconPaths = {
    error: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z',
    warning: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
    info: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    success: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'
  };

  const colorClasses = {
    error: {
      bg: 'bg-red-600 dark:bg-red-500',
      text: 'text-white'
    },
    warning: {
      bg: 'bg-yellow-500 dark:bg-yellow-600',
      text: 'text-white'
    },
    info: {
      bg: 'bg-blue-600 dark:bg-blue-500',
      text: 'text-white'
    },
    success: {
      bg: 'bg-green-600 dark:bg-green-500',
      text: 'text-white'
    }
  };

  const positionClasses = {
    'top': 'top-4 left-1/2 transform -translate-x-1/2',
    'bottom': 'bottom-4 left-1/2 transform -translate-x-1/2',
    'top-right': 'top-4 right-4',
    'top-left': 'top-4 left-4',
    'bottom-right': 'bottom-4 right-4',
    'bottom-left': 'bottom-4 left-4'
  };

  const colors = colorClasses[type];
  const positionClass = positionClasses[position];

  return (
    <div
      className={`fixed ${positionClass} z-50 transition-all duration-300 ${
        isExiting ? 'opacity-0 transform scale-95' : 'opacity-100 transform scale-100'
      } ${className}`}
    >
      <div className={`${colors.bg} rounded-lg shadow-lg p-4 max-w-sm`}>
        <div className="flex items-center">
          <svg
            className={`h-5 w-5 ${colors.text} flex-shrink-0`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d={iconPaths[type]}
            />
          </svg>
          <p className={`ml-3 text-sm font-medium ${colors.text} flex-1`}>
            {message}
          </p>
          <button
            onClick={handleDismiss}
            className={`ml-4 flex-shrink-0 rounded-md ${colors.text} hover:opacity-75 focus:outline-none focus:ring-2 focus:ring-white`}
            aria-label="Dismiss"
          >
            <svg
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>
      </div>
    </div>
  );
};