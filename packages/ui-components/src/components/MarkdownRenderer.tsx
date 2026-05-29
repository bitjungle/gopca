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

import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkMath from 'remark-math';
import remarkGfm from 'remark-gfm';
import rehypeKatex from 'rehype-katex';
import 'katex/dist/katex.min.css';

export interface MarkdownRendererProps {
  content: string;
  className?: string;
  /** Called when the user clicks an external (http/https) link. If omitted,
   *  external links open in a new tab via target="_blank". Provide this in
   *  Wails-hosted windows to open links via BrowserOpenURL so they launch
   *  in the system browser rather than inside the embedded WebView. */
  onExternalLink?: (url: string) => void;
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
  className = '',
  onExternalLink,
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
          // Custom link component: external links are opened via onExternalLink
          // callback (e.g. Wails BrowserOpenURL) when provided, so they launch
          // in the system browser rather than navigating the embedded WebView.
          // Falls back to target="_blank" when running in a regular browser.
          a: ({ children, href, ...props }) => {
            const isExternal = href?.startsWith('http');
            const handleClick = isExternal && onExternalLink
              ? (e: React.MouseEvent<HTMLAnchorElement>) => {
                  e.preventDefault();
                  onExternalLink(href!);
                }
              : undefined;
            return (
              <a
                href={href}
                target={isExternal && !onExternalLink ? '_blank' : undefined}
                rel={isExternal && !onExternalLink ? 'noopener noreferrer' : undefined}
                onClick={handleClick}
                className="text-blue-600 dark:text-blue-400 hover:underline"
                {...props}
              >
                {children}
              </a>
            );
          },
          // Custom code block styling
          code: ({ className, children, ...props }) => {
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
          blockquote: ({ children, ...props }) => (
            <blockquote
              className="border-l-4 border-blue-500 pl-4 my-4 italic text-gray-700 dark:text-gray-300"
              {...props}
            >
              {children}
            </blockquote>
          ),
          // Enhanced table styling for better readability
          table: ({ children, ...props }) => (
            <div className="overflow-x-auto my-6">
              <table className="min-w-full divide-y divide-gray-300 dark:divide-gray-600" {...props}>
                {children}
              </table>
            </div>
          ),
          thead: ({ children, ...props }) => (
            <thead className="bg-gray-50 dark:bg-gray-800" {...props}>
              {children}
            </thead>
          ),
          tbody: ({ children, ...props }) => (
            <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700" {...props}>
              {children}
            </tbody>
          ),
          tr: ({ children, ...props }) => (
            <tr className="hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors" {...props}>
              {children}
            </tr>
          ),
          th: ({ children, ...props }) => (
            <th
              className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
              {...props}
            >
              {children}
            </th>
          ),
          td: ({ children, ...props }) => (
            <td className="px-4 py-3 text-sm text-gray-900 dark:text-gray-100" {...props}>
              {children}
            </td>
          ),
          // Custom image component to handle relative paths
          img: ({ src, alt, ...props }) => {
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