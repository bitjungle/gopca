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