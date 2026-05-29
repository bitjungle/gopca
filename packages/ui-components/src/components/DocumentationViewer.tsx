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

import React, { useEffect, useState } from 'react';
import { MarkdownRenderer } from './MarkdownRenderer';

export interface DocumentationViewerProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  markdownPath: string;
}

/**
 * Shared documentation viewer component for displaying markdown files in a modal.
 * Handles loading, escape key, scroll prevention, and consistent styling.
 * Used by both GoPCA Desktop and GoCSV Desktop applications.
 */
export const DocumentationViewer: React.FC<DocumentationViewerProps> = ({
  isOpen,
  onClose,
  title,
  markdownPath
}) => {
  const [markdownContent, setMarkdownContent] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (isOpen) {
      // Reset state to show loading and clear stale content
      setIsLoading(true);
      setMarkdownContent('');

      // Load the markdown file
      fetch(markdownPath)
        .then(response => {
          if (!response.ok) {
            throw new Error(`Failed to load documentation: ${response.statusText}`);
          }
          return response.text();
        })
        .then(text => {
          setMarkdownContent(text);
          setIsLoading(false);
        })
        .catch(error => {
          console.error('Error loading documentation:', error);
          setMarkdownContent(`# Error\n\nFailed to load documentation from ${markdownPath}.\n\nPlease try again later.`);
          setIsLoading(false);
        });
    }
  }, [isOpen, markdownPath]);

  // Handle escape key to close
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      // Prevent body scroll when modal is open
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [isOpen, onClose]);

  if (!isOpen) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 bg-white dark:bg-gray-900">
      {/* Header with exit button */}
      <div className="sticky top-0 z-10 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-4xl mx-auto px-6 py-4 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
            {title}
          </h2>
          <button
            onClick={onClose}
            className="px-3 py-1 text-sm rounded-lg bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 transition-colors flex items-center gap-2"
            aria-label="Close documentation"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
            Exit
          </button>
        </div>
      </div>

      {/* Content area */}
      <div className="overflow-y-auto" style={{ height: 'calc(100vh - 73px)' }}>
        <div className="max-w-4xl mx-auto px-6 py-8 text-left">
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="text-gray-500 dark:text-gray-400">Loading documentation...</div>
            </div>
          ) : (
            <MarkdownRenderer content={markdownContent} />
          )}
        </div>
      </div>
    </div>
  );
};