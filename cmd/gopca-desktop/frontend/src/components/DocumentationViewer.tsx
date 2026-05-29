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