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
import './SkipLinks.css';

export interface SkipLink {
  href: string;
  label: string;
}

export interface SkipLinksProps {
  links?: SkipLink[];
}

/**
 * SkipLinks component provides keyboard navigation shortcuts to jump to main content areas
 * These links are only visible when focused (for keyboard users)
 */
export const SkipLinks: React.FC<SkipLinksProps> = ({
  links = [
    { href: '#main-content', label: 'Skip to main content' },
    { href: '#navigation', label: 'Skip to navigation' }
  ]
}) => {
  return (
    <div className="skip-links" role="navigation" aria-label="Skip links">
      {links.map((link) => (
        <a
          key={link.href}
          href={link.href}
          className="skip-link"
          onClick={(e) => {
            e.preventDefault();
            const target = document.querySelector(link.href);
            if (target) {
              (target as HTMLElement).focus();
              (target as HTMLElement).scrollIntoView();
            }
          }}
        >
          {link.label}
        </a>
      ))}
    </div>
  );
};