# Accessibility Standards

This document defines accessibility standards for GoPCA applications to ensure they are usable by everyone, including users with disabilities.

## WCAG 2.1 AA Compliance

GoPCA targets **WCAG 2.1 Level AA** compliance for all user interfaces.

### Key Requirements

#### 1. Perceivable
- **Color Contrast**: Minimum 4.5:1 for normal text, 3:1 for large text
- **Non-Color Indicators**: Never rely solely on color to convey information
- **Text Alternatives**: Provide alt text for all informative images and icons
- **Responsive Design**: Support zoom up to 200% without horizontal scrolling

#### 2. Operable
- **Keyboard Accessible**: All functionality available via keyboard
- **Focus Indicators**: Visible focus indicators for all interactive elements
- **Skip Links**: Provide skip navigation links for repetitive content
- **Timing**: No time limits or provide user control over timing

#### 3. Understandable
- **Clear Labels**: All form controls have descriptive labels
- **Error Messages**: Clear, specific error messages with recovery suggestions
- **Consistent Navigation**: Navigation patterns consistent across applications
- **Help Text**: Contextual help available for complex operations

#### 4. Robust
- **Semantic HTML**: Use proper HTML elements for their intended purpose
- **ARIA Support**: Proper ARIA attributes where semantic HTML insufficient
- **Cross-Browser**: Test with major browsers and screen readers

## Implementation Guidelines

### Semantic HTML Structure

```html
<!-- Use semantic landmarks -->
<header role="banner">
  <nav role="navigation" aria-label="Main navigation">
    <!-- Skip link for keyboard users -->
    <a href="#main-content" class="skip-link">Skip to main content</a>
  </nav>
</header>

<main id="main-content" role="main">
  <!-- Main application content -->
</main>

<aside role="complementary" aria-label="Sidebar">
  <!-- Supplementary content -->
</aside>

<footer role="contentinfo">
  <!-- Footer content -->
</footer>
```

### Keyboard Navigation Patterns

#### Global Shortcuts
| Key | Action |
|-----|--------|
| `Tab` | Navigate forward through focusable elements |
| `Shift+Tab` | Navigate backward through focusable elements |
| `Enter` | Activate buttons, links, and submit forms |
| `Space` | Toggle checkboxes, activate buttons |
| `Escape` | Close modals, cancel operations |
| `Arrow Keys` | Navigate within components (menus, tabs, etc.) |

#### Application Shortcuts
| Key | Action |
|-----|--------|
| `Ctrl/Cmd+O` | Open file |
| `Ctrl/Cmd+S` | Save results |
| `Ctrl/Cmd+E` | Export data |
| `Ctrl/Cmd+,` | Open settings |
| `?` | Show keyboard shortcuts help |
| `Alt+T` | Toggle theme (dark/light) |

### Focus Management

```typescript
// Example focus management hook
export const useFocusManagement = () => {
  const previousFocusRef = useRef<HTMLElement | null>(null);

  const saveFocus = () => {
    previousFocusRef.current = document.activeElement as HTMLElement;
  };

  const restoreFocus = () => {
    if (previousFocusRef.current) {
      previousFocusRef.current.focus();
    }
  };

  const trapFocus = (container: HTMLElement) => {
    const focusableElements = container.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    const firstElement = focusableElements[0] as HTMLElement;
    const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;

    const handleTabKey = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;

      if (e.shiftKey) {
        if (document.activeElement === firstElement) {
          lastElement.focus();
          e.preventDefault();
        }
      } else {
        if (document.activeElement === lastElement) {
          firstElement.focus();
          e.preventDefault();
        }
      }
    };

    container.addEventListener('keydown', handleTabKey);
    return () => container.removeEventListener('keydown', handleTabKey);
  };

  return { saveFocus, restoreFocus, trapFocus };
};
```

### ARIA Attributes

#### Common ARIA Patterns

```tsx
// Loading state
<div role="status" aria-live="polite" aria-busy="true">
  <span className="sr-only">Loading...</span>
  <ProgressIndicator />
</div>

// Error message
<div role="alert" aria-live="assertive">
  {errorMessage}
</div>

// Form field with error
<div>
  <label htmlFor="email">Email Address</label>
  <input
    id="email"
    type="email"
    aria-invalid={hasError}
    aria-describedby={hasError ? "email-error" : undefined}
  />
  {hasError && (
    <span id="email-error" role="alert">
      Please enter a valid email address
    </span>
  )}
</div>

// Modal dialog
<div
  role="dialog"
  aria-modal="true"
  aria-labelledby="dialog-title"
  aria-describedby="dialog-description"
>
  <h2 id="dialog-title">Confirm Action</h2>
  <p id="dialog-description">Are you sure you want to proceed?</p>
</div>
```

### Color and Contrast

#### Color Palette Requirements
- **Primary Text**: #1a1a1a on white background (21:1 ratio) ✅
- **Secondary Text**: #666666 on white background (5.74:1 ratio) ✅
- **Link Color**: #0066cc on white background (5.07:1 ratio) ✅
- **Error Color**: #dc2626 on white background (5.87:1 ratio) ✅
- **Success Color**: #16a34a on white background (4.54:1 ratio) ✅

#### Dark Mode Requirements
- **Primary Text**: #f3f4f6 on #1a1a1a background (13.94:1 ratio) ✅
- **Secondary Text**: #9ca3af on #1a1a1a background (5.31:1 ratio) ✅
- Ensure all color combinations meet WCAG AA standards

### Screen Reader Support

#### Best Practices
1. **Announce Dynamic Changes**: Use `aria-live` regions for updates
2. **Descriptive Labels**: Provide context in labels and descriptions
3. **Hidden Helper Text**: Use `.sr-only` class for screen reader only content
4. **Landmark Roles**: Use ARIA landmarks to define page structure
5. **State Changes**: Announce state changes (loading, success, error)

```css
/* Screen reader only class */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}
```

## Testing Requirements

### Automated Testing
- Use axe-core for automated accessibility testing
- Integrate with CI/CD pipeline
- Test color contrast ratios programmatically
- Validate ARIA attribute usage

### Manual Testing
1. **Keyboard Navigation**: Test all features using only keyboard
2. **Screen Reader Testing**: Test with NVDA (Windows), JAWS, or VoiceOver (macOS)
3. **Browser Zoom**: Test at 200% zoom level
4. **Color Blindness**: Test with color blindness simulators
5. **Focus Order**: Verify logical tab order

### Testing Checklist
- [ ] All interactive elements reachable via keyboard
- [ ] Focus indicators visible and clear
- [ ] Skip links functional
- [ ] Forms have proper labels and error messages
- [ ] ARIA attributes correctly implemented
- [ ] Color contrast meets WCAG AA standards
- [ ] Dynamic content announced to screen readers
- [ ] Modals trap focus appropriately
- [ ] Escape key closes modals/dropdowns
- [ ] Keyboard shortcuts documented and functional

## Component Requirements

### Buttons
- Minimum size: 44x44 pixels (touch targets)
- Clear focus indicator
- Descriptive text or aria-label
- Disabled state clearly indicated

### Forms
- All inputs have associated labels
- Required fields marked with both visual and ARIA indicators
- Error messages associated with fields
- Submit button indicates loading state

### Modals/Dialogs
- Focus trapped within modal
- Focus returns to trigger element on close
- Escape key closes modal
- Background content marked as inert

### Charts/Visualizations
- Provide text alternatives for data
- Ensure sufficient color contrast
- Don't rely solely on color to distinguish data
- Provide data tables as alternatives

## CLI Accessibility

### Terminal Output
- Use clear, consistent formatting
- Provide verbose output options
- Support screen reader friendly output formats
- Use semantic color coding sparingly

### Error Messages
- Clear, actionable error messages
- Suggest corrections when possible
- Use consistent formatting
- Avoid relying solely on color

## Resources

### Tools
- [axe DevTools](https://www.deque.com/axe/devtools/) - Browser extension for testing
- [WAVE](https://wave.webaim.org/) - Web accessibility evaluation tool
- [Contrast Checker](https://webaim.org/resources/contrastchecker/) - Color contrast validation
- [NVDA](https://www.nvaccess.org/) - Free screen reader for Windows
- [VoiceOver](https://support.apple.com/guide/voiceover/welcome/mac) - Built-in macOS screen reader

### References
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [ARIA Authoring Practices](https://www.w3.org/WAI/ARIA/apg/)
- [WebAIM Resources](https://webaim.org/resources/)
- [A11y Project Checklist](https://www.a11yproject.com/checklist/)

## Compliance Statement

GoPCA is committed to providing accessible software that can be used by everyone. We continuously work to improve accessibility and welcome feedback from users about their experiences.

For accessibility issues or suggestions, please open an issue on our [GitHub repository](https://github.com/bitjungle/gopca/issues) with the "accessibility" label.