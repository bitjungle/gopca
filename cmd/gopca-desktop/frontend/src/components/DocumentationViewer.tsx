// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { DocumentationViewer as SharedDocumentationViewer } from '@gopca/ui-components';

interface DocumentationViewerProps {
  isOpen: boolean;
  onClose: () => void;
}

/**
 * GoPCA Desktop's documentation viewer.
 * Uses the shared DocumentationViewer component with GoPCA-specific configuration.
 */
export const DocumentationViewer: React.FC<DocumentationViewerProps> = ({ isOpen, onClose }) => {
  return (
    <SharedDocumentationViewer
      isOpen={isOpen}
      onClose={onClose}
      title="GoPCA Documentation"
      markdownPath="/docs/intro_to_pca.md"
    />
  );
};