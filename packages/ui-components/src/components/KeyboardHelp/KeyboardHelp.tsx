import React from 'react';
import { KeyboardShortcut, formatShortcut, getModifierKey } from '../../hooks/useKeyboardShortcuts';
import './KeyboardHelp.css';

export interface KeyboardHelpProps {
  shortcuts: KeyboardShortcut[];
  isVisible: boolean;
  onClose: () => void;
}

/**
 * KeyboardHelp component displays available keyboard shortcuts in a modal
 */
export const KeyboardHelp: React.FC<KeyboardHelpProps> = ({
  shortcuts,
  isVisible,
  onClose,
}) => {
  if (!isVisible) return null;

  // Group shortcuts by category (based on description patterns)
  const groupedShortcuts = shortcuts.reduce((groups, shortcut) => {
    let category = 'General';
    
    if (shortcut.description.toLowerCase().includes('file') || 
        shortcut.description.toLowerCase().includes('open') ||
        shortcut.description.toLowerCase().includes('save')) {
      category = 'File Operations';
    } else if (shortcut.description.toLowerCase().includes('edit') ||
               shortcut.description.toLowerCase().includes('undo') ||
               shortcut.description.toLowerCase().includes('redo')) {
      category = 'Edit';
    } else if (shortcut.description.toLowerCase().includes('view') ||
               shortcut.description.toLowerCase().includes('theme') ||
               shortcut.description.toLowerCase().includes('zoom')) {
      category = 'View';
    } else if (shortcut.description.toLowerCase().includes('navigate') ||
               shortcut.description.toLowerCase().includes('focus')) {
      category = 'Navigation';
    }

    if (!groups[category]) {
      groups[category] = [];
    }
    groups[category].push(shortcut);
    return groups;
  }, {} as Record<string, KeyboardShortcut[]>);

  return (
    <>
      {/* Backdrop */}
      <div 
        className="keyboard-help-backdrop" 
        onClick={onClose}
        role="presentation"
        aria-hidden="true"
      />
      
      {/* Modal */}
      <div
        className="keyboard-help-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="keyboard-help-title"
        tabIndex={-1}
      >
        <div className="keyboard-help-header">
          <h2 id="keyboard-help-title">Keyboard Shortcuts</h2>
          <button
            className="keyboard-help-close"
            onClick={onClose}
            aria-label="Close keyboard shortcuts help"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={1.5}
              stroke="currentColor"
              className="w-5 h-5"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        <div className="keyboard-help-content">
          <div className="keyboard-help-note">
            <strong>Note:</strong> Use {getModifierKey()} for keyboard shortcuts on your platform.
          </div>

          {Object.entries(groupedShortcuts).map(([category, categoryShortcuts]) => (
            <div key={category} className="keyboard-help-category">
              <h3 className="keyboard-help-category-title">{category}</h3>
              <table className="keyboard-help-table">
                <tbody>
                  {categoryShortcuts.map((shortcut, index) => (
                    <tr key={index}>
                      <td className="keyboard-help-description">
                        {shortcut.description}
                      </td>
                      <td className="keyboard-help-keys">
                        <kbd>{formatShortcut(shortcut)}</kbd>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ))}

          <div className="keyboard-help-footer">
            <p>Press <kbd>?</kbd> to toggle this help menu</p>
            <p>Press <kbd>Escape</kbd> to close</p>
          </div>
        </div>
      </div>
    </>
  );
};