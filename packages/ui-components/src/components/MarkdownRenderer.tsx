// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkMath from 'remark-math';
import remarkGfm from 'remark-gfm';
import rehypeKatex from 'rehype-katex';
import 'katex/dist/katex.min.css';

export interface MarkdownRendererProps {
  content: string;
  className?: string;
}

/**
 * Shared markdown renderer component with support for:
 * - GitHub Flavored Markdown tables via remark-gfm
 * - LaTeX math rendering via remark-math and rehype-katex
 * - Consistent styling across GoPCA and GoCSV applications
 * - Dark mode support
 */
export const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ 
  content, 
  className = '' 
}) => {
  return (
    <div className={`prose prose-lg dark:prose-invert max-w-none text-left
      prose-headings:text-gray-900 dark:prose-headings:text-gray-100 prose-headings:text-left
      prose-p:text-gray-700 dark:prose-p:text-gray-300 prose-p:text-justify
      prose-a:text-blue-600 dark:prose-a:text-blue-400
      prose-strong:text-gray-900 dark:prose-strong:text-gray-100
      prose-code:text-gray-800 dark:prose-code:text-gray-200
      prose-pre:bg-gray-100 dark:prose-pre:bg-gray-800
      prose-blockquote:text-gray-700 dark:prose-blockquote:text-gray-300
      prose-blockquote:border-blue-500
      prose-li:text-gray-700 dark:prose-li:text-gray-300 prose-li:text-left
      prose-table:w-full prose-table:border-collapse
      prose-thead:border-b-2 prose-thead:border-gray-300 dark:prose-thead:border-gray-600
      prose-th:text-left prose-th:p-2 prose-th:font-semibold
      prose-td:p-2 prose-td:border-b prose-td:border-gray-200 dark:prose-td:border-gray-700
      prose-tbody:divide-y prose-tbody:divide-gray-200 dark:prose-tbody:divide-gray-700
      ${className}`}>
      <ReactMarkdown
        remarkPlugins={[
          remarkGfm,  // Enables GitHub Flavored Markdown (tables, strikethrough, etc.)
          remarkMath  // Enables math notation
        ]}
        rehypePlugins={[
          rehypeKatex  // Renders math using KaTeX
        ]}
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
          // Enhanced table styling for better readability
          table: ({ node, children, ...props }) => (
            <div className="overflow-x-auto my-6">
              <table className="min-w-full divide-y divide-gray-300 dark:divide-gray-600" {...props}>
                {children}
              </table>
            </div>
          ),
          thead: ({ node, children, ...props }) => (
            <thead className="bg-gray-50 dark:bg-gray-800" {...props}>
              {children}
            </thead>
          ),
          tbody: ({ node, children, ...props }) => (
            <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700" {...props}>
              {children}
            </tbody>
          ),
          tr: ({ node, children, ...props }) => (
            <tr className="hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors" {...props}>
              {children}
            </tr>
          ),
          th: ({ node, children, ...props }) => (
            <th 
              className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider" 
              {...props}
            >
              {children}
            </th>
          ),
          td: ({ node, children, ...props }) => (
            <td className="px-4 py-3 text-sm text-gray-900 dark:text-gray-100" {...props}>
              {children}
            </td>
          ),
          // Custom image component to handle relative paths
          img: ({ node, src, alt, ...props }) => {
            // If the src is a relative path starting with 'images/', prepend the docs path
            const imageSrc = src?.startsWith('images/') ? `/docs/${src}` : src;
            return (
              <img
                src={imageSrc}
                alt={alt}
                className="rounded-lg shadow-md my-4 mx-auto max-w-full h-auto"
                loading="lazy"
                {...props}
              />
            );
          }
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
};