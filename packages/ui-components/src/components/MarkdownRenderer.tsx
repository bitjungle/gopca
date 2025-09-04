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

interface MarkdownRendererProps {
  content: string;
  className?: string;
}

/**
 * Shared markdown renderer component for documentation display.
 * Supports GitHub Flavored Markdown tables, math notation, and standard markdown.
 * 
 * Note: This component expects react-markdown and its plugins to be installed
 * in the consuming application as peer dependencies.
 */
export const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ 
  content, 
  className = '' 
}) => {
  return (
    <div className={`prose prose-lg dark:prose-invert max-w-none text-left ${className}`}>
      <ReactMarkdown
        remarkPlugins={[
          remarkGfm,  // Enables tables, strikethrough, task lists, etc.
          remarkMath  // Enables math notation with $ and $$
        ]}
        rehypePlugins={[
          rehypeKatex  // Renders math notation using KaTeX
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
          // Custom table styling with better dark mode support
          table: ({ node, children, ...props }) => (
            <div className="overflow-x-auto my-4">
              <table className="min-w-full border-collapse" {...props}>
                {children}
              </table>
            </div>
          ),
          thead: ({ node, children, ...props }) => (
            <thead className="border-b border-gray-300 dark:border-gray-600" {...props}>
              {children}
            </thead>
          ),
          th: ({ node, children, ...props }) => (
            <th
              className="text-left font-semibold px-4 py-2 text-gray-900 dark:text-gray-100"
              {...props}
            >
              {children}
            </th>
          ),
          tbody: ({ node, children, ...props }) => (
            <tbody {...props}>{children}</tbody>
          ),
          tr: ({ node, children, ...props }) => (
            <tr className="border-b border-gray-200 dark:border-gray-700" {...props}>
              {children}
            </tr>
          ),
          td: ({ node, children, ...props }) => (
            <td
              className="px-4 py-2 text-gray-700 dark:text-gray-300"
              {...props}
            >
              {children}
            </td>
          )
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
};