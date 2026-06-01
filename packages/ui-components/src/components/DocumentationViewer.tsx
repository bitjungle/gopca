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

import React, { useEffect, useRef, useState } from 'react';
import { MarkdownRenderer } from './MarkdownRenderer';
import { TableOfContents } from './TableOfContents';
import { TocEntry, extractHeadings } from '../utils/tocUtils';

export interface DocumentationViewerProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  markdownPath: string;
}

/**
 * Full-screen documentation viewer with a sticky sidebar table of contents.
 * The TOC is auto-generated from H2 and H3 headings; the active section is
 * highlighted via IntersectionObserver as the user scrolls.
 *
 * Used by both GoPCA Desktop and GoCSV Desktop.
 */
export const DocumentationViewer: React.FC<DocumentationViewerProps> = ({
  isOpen,
  onClose,
  title,
  markdownPath
}) => {
  const [markdownContent, setMarkdownContent] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [tocEntries, setTocEntries] = useState<TocEntry[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);

  // Load markdown when the viewer opens or the path changes.
  useEffect(() => {
    if (isOpen) {
      setIsLoading(true);
      setMarkdownContent('');
      setTocEntries([]);
      setActiveId(null);

      fetch(markdownPath)
        .then(response => {
          if (!response.ok) throw new Error(`Failed to load: ${response.statusText}`);
          return response.text();
        })
        .then(text => {
          setMarkdownContent(text);
          setTocEntries(extractHeadings(text));
          setIsLoading(false);
        })
        .catch(error => {
          console.error('Error loading documentation:', error);
          setMarkdownContent(`# Error\n\nFailed to load documentation from ${markdownPath}.`);
          setIsLoading(false);
        });
    }
  }, [isOpen, markdownPath]);

  // Escape key closes the viewer; body scroll is suppressed while open.
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) onClose();
    };
    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }
    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [isOpen, onClose]);

  // IntersectionObserver: highlight the TOC entry for the heading nearest
  // the top of the scroll container.
  useEffect(() => {
    if (!scrollRef.current || tocEntries.length === 0 || isLoading) return;

    const root = scrollRef.current;
    const elements = tocEntries
      .map(e => root.querySelector(`#${CSS.escape(e.id)}`) as HTMLElement | null)
      .filter((el): el is HTMLElement => el !== null);

    if (elements.length === 0) return;

    // Set the first heading as active initially.
    setActiveId(elements[0].id);

    const observer = new IntersectionObserver(
      entries => {
        const visible = entries.filter(e => e.isIntersecting);
        if (visible.length > 0) setActiveId(visible[0].target.id);
      },
      {
        root,
        // Fire when a heading enters the upper quarter of the scroll area.
        rootMargin: '-10% 0px -75% 0px',
        threshold: 0,
      }
    );

    elements.forEach(el => observer.observe(el));
    return () => observer.disconnect();
  }, [tocEntries, isLoading]);

  const handleTocClick = (id: string) => {
    const el = scrollRef.current?.querySelector(`#${CSS.escape(id)}`);
    el?.scrollIntoView({ behavior: 'smooth', block: 'start' });
    setActiveId(id);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 bg-white dark:bg-gray-900 flex flex-col">
      {/* Header */}
      <div className="sticky top-0 z-10 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 shrink-0">
        <div className="px-6 py-4 flex items-center justify-between">
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

      {/* Body: TOC sidebar + scrollable content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sticky TOC sidebar */}
        <TableOfContents
          entries={tocEntries}
          activeId={activeId}
          onEntryClick={handleTocClick}
        />

        {/* Scrollable content column */}
        <div ref={scrollRef} className="flex-1 overflow-y-auto">
          <div className="max-w-4xl mx-auto px-6 py-8 text-left">
            {isLoading ? (
              <div className="flex items-center justify-center h-64">
                <div className="text-gray-500 dark:text-gray-400">Loading documentation…</div>
              </div>
            ) : (
              <MarkdownRenderer content={markdownContent} />
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
