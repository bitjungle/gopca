// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkMath from 'remark-math';
import rehypeKatex from 'rehype-katex';
import { useTheme } from '@gopca/ui-components';
import 'katex/dist/katex.min.css';

interface DocumentationViewerProps {
  isOpen: boolean;
  onClose: () => void;
}

export const DocumentationViewer: React.FC<DocumentationViewerProps> = ({ isOpen, onClose }) => {
  const [markdownContent, setMarkdownContent] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const { theme } = useTheme();

  useEffect(() => {
    if (isOpen) {
      // Load the markdown file
      fetch('/docs/intro_to_pca.md')
        .then(response => response.text())
        .then(text => {
          setMarkdownContent(text);
          setIsLoading(false);
        })
        .catch(error => {
          console.error('Error loading documentation:', error);
          setMarkdownContent('# Error\n\nFailed to load documentation. Please try again later.');
          setIsLoading(false);
        });
    }
  }, [isOpen]);

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

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 bg-white dark:bg-gray-900">
      {/* Header with exit button */}
      <div className="sticky top-0 z-10 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-4xl mx-auto px-6 py-4 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
            GoPCA Documentation
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
            <div className="prose prose-lg dark:prose-invert max-w-none text-left
              prose-headings:text-gray-900 dark:prose-headings:text-gray-100 prose-headings:text-left
              prose-p:text-gray-700 dark:prose-p:text-gray-300 prose-p:text-justify
              prose-a:text-blue-600 dark:prose-a:text-blue-400
              prose-strong:text-gray-900 dark:prose-strong:text-gray-100
              prose-code:text-gray-800 dark:prose-code:text-gray-200
              prose-pre:bg-gray-100 dark:prose-pre:bg-gray-800
              prose-blockquote:text-gray-700 dark:prose-blockquote:text-gray-300
              prose-blockquote:border-blue-500
              prose-li:text-gray-700 dark:prose-li:text-gray-300 prose-li:text-left">
              <ReactMarkdown
                remarkPlugins={[remarkMath]}
                rehypePlugins={[rehypeKatex]}
                components={{
                  // Custom link component to open external links in new tab
                  a: ({ node, children, href, ...props }) => (
                    <a
                      href={href}
                      target={href?.startsWith('http') ? '_blank' : undefined}
                      rel={href?.startsWith('http') ? 'noopener noreferrer' : undefined}
                      className="text-blue-600 dark:text-blue-400 hover:underline"
                      {...props}
                    >
                      {children}
                    </a>
                  ),
                  // Custom code block styling
                  code: ({ node, className, children, ...props }) => {
                    const match = /language-(\w+)/.exec(className || '');
                    const isInline = !match && !className;
                    return isInline ? (
                      <code
                        className="px-1 py-0.5 rounded bg-gray-100 dark:bg-gray-800 text-sm"
                        {...props}
                      >
                        {children}
                      </code>
                    ) : (
                      <pre className="overflow-x-auto">
                        <code
                          className={`block p-4 rounded-lg bg-gray-100 dark:bg-gray-800 text-sm ${className || ''}`}
                          {...props}
                        >
                          {children}
                        </code>
                      </pre>
                    );
                  },
                  // Custom blockquote styling
                  blockquote: ({ node, children, ...props }) => (
                    <blockquote
                      className="border-l-4 border-blue-500 pl-4 my-4 italic text-gray-700 dark:text-gray-300"
                      {...props}
                    >
                      {children}
                    </blockquote>
                  ),
                }}
              >
                {markdownContent}
              </ReactMarkdown>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};