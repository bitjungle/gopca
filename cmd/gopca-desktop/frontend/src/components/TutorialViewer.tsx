// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useEffect, useState } from 'react';
import { MarkdownRenderer } from '@gopca/ui-components';

/**
 * Maps a dataset name to the public path of its tutorial markdown file.
 * Tutorial files live under frontend/public/tutorials/<dataset>/.
 * Only datasets that have a tutorial file are listed here; the others
 * are planned but not yet written (issues #654–#657).
 */
const TUTORIAL_PATHS: Record<string, string> = {
  iris:       '/tutorials/iris/iris_exploration.md',
  wine:       '/tutorials/wine/wine_exploration.md',
  corn:       '/tutorials/corn/corn_exploration.md',
  swiss_roll: '/tutorials/swiss_roll/swiss_roll_exploration.md',
  stocks:     '/tutorials/stocks/stocks_exploration.md',
};

interface TutorialViewerProps {
  /** Dataset name as returned by GetAppMode().Dataset */
  dataset: string;
}

/**
 * TutorialViewer renders a dataset exploration tutorial in the tutorial window.
 *
 * It fetches the markdown file from the embedded asset server, rewrites
 * relative image paths to absolute public paths, then renders using the
 * shared MarkdownRenderer component.
 *
 * This component is rendered by App.tsx when GetAppMode() returns
 * mode "tutorial". It is never shown in the main application window.
 */
export const TutorialViewer: React.FC<TutorialViewerProps> = ({ dataset }) => {
  const [content, setContent] = useState<string>('');
  const [error, setError] = useState<string | null>(null);

  const tutorialPath = TUTORIAL_PATHS[dataset];
  // Base directory for resolving relative assets (e.g. images)
  const assetBase = `/tutorials/${dataset}/`;

  useEffect(() => {
    if (!tutorialPath) {
      setError(`No tutorial available for dataset: ${dataset}`);
      return;
    }

    fetch(tutorialPath)
      .then(response => {
        if (!response.ok) {
          throw new Error(`Failed to load tutorial (HTTP ${response.status})`);
        }
        return response.text();
      })
      .then(text => {
        // Rewrite relative image references like ./foo.png to absolute paths
        // so MarkdownRenderer can load them from the embedded asset server.
        const resolved = text.replace(/\(\.\/([^)]+)\)/g, `(${assetBase}$1)`);
        setContent(resolved);
      })
      .catch(err => {
        setError(`Could not load tutorial: ${err.message}`);
      });
  }, [tutorialPath, assetBase]);

  if (error) {
    return (
      <div className="min-h-screen bg-white dark:bg-gray-900 flex items-center justify-center">
        <p className="text-red-600 dark:text-red-400">{error}</p>
      </div>
    );
  }

  if (!content) {
    return (
      <div className="min-h-screen bg-white dark:bg-gray-900 flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">Loading tutorial…</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white dark:bg-gray-900 overflow-auto">
      <div className="max-w-3xl mx-auto px-8 py-10">
        <MarkdownRenderer content={content} />
      </div>
    </div>
  );
};
