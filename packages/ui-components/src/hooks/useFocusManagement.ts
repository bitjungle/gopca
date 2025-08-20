import { useRef, useCallback, useEffect } from 'react';

/**
 * Hook for managing focus in accessible components
 * Provides utilities for saving/restoring focus and trapping focus within containers
 */
export const useFocusManagement = () => {
  const previousFocusRef = useRef<HTMLElement | null>(null);

  /**
   * Save the currently focused element
   */
  const saveFocus = useCallback(() => {
    previousFocusRef.current = document.activeElement as HTMLElement;
  }, []);

  /**
   * Restore focus to the previously saved element
   */
  const restoreFocus = useCallback(() => {
    if (previousFocusRef.current && previousFocusRef.current.focus) {
      previousFocusRef.current.focus();
    }
  }, []);

  /**
   * Focus the first focusable element within a container
   */
  const focusFirst = useCallback((container: HTMLElement) => {
    const focusableElements = getFocusableElements(container);
    if (focusableElements.length > 0) {
      (focusableElements[0] as HTMLElement).focus();
    }
  }, []);

  /**
   * Trap focus within a container (useful for modals)
   * Returns a cleanup function to remove the event listener
   */
  const trapFocus = useCallback((container: HTMLElement) => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') {
return;
}

      const focusableElements = getFocusableElements(container);
      if (focusableElements.length === 0) {
return;
}

      const firstElement = focusableElements[0] as HTMLElement;
      const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;

      if (e.shiftKey) {
        // Shift + Tab
        if (document.activeElement === firstElement) {
          lastElement.focus();
          e.preventDefault();
        }
      } else {
        // Tab
        if (document.activeElement === lastElement) {
          firstElement.focus();
          e.preventDefault();
        }
      }
    };

    container.addEventListener('keydown', handleKeyDown);
    return () => container.removeEventListener('keydown', handleKeyDown);
  }, []);

  /**
   * Set up arrow key navigation within a list of elements
   */
  const setupArrowNavigation = useCallback((
    container: HTMLElement,
    options?: {
      vertical?: boolean;
      horizontal?: boolean;
      wrap?: boolean;
    }
  ) => {
    const { vertical = true, horizontal = false, wrap = true } = options || {};

    const handleKeyDown = (e: KeyboardEvent) => {
      const focusableElements = getFocusableElements(container);
      const currentIndex = Array.from(focusableElements).indexOf(
        document.activeElement as HTMLElement
      );

      if (currentIndex === -1) {
return;
}

      let nextIndex = currentIndex;

      if (vertical) {
        if (e.key === 'ArrowDown') {
          e.preventDefault();
          nextIndex = currentIndex + 1;
          if (nextIndex >= focusableElements.length) {
            nextIndex = wrap ? 0 : focusableElements.length - 1;
          }
        } else if (e.key === 'ArrowUp') {
          e.preventDefault();
          nextIndex = currentIndex - 1;
          if (nextIndex < 0) {
            nextIndex = wrap ? focusableElements.length - 1 : 0;
          }
        }
      }

      if (horizontal) {
        if (e.key === 'ArrowRight') {
          e.preventDefault();
          nextIndex = currentIndex + 1;
          if (nextIndex >= focusableElements.length) {
            nextIndex = wrap ? 0 : focusableElements.length - 1;
          }
        } else if (e.key === 'ArrowLeft') {
          e.preventDefault();
          nextIndex = currentIndex - 1;
          if (nextIndex < 0) {
            nextIndex = wrap ? focusableElements.length - 1 : 0;
          }
        }
      }

      if (nextIndex !== currentIndex) {
        (focusableElements[nextIndex] as HTMLElement).focus();
      }
    };

    container.addEventListener('keydown', handleKeyDown);
    return () => container.removeEventListener('keydown', handleKeyDown);
  }, []);

  return {
    saveFocus,
    restoreFocus,
    focusFirst,
    trapFocus,
    setupArrowNavigation
  };
};

/**
 * Get all focusable elements within a container
 */
function getFocusableElements(container: HTMLElement): NodeListOf<Element> {
  return container.querySelectorAll(
    'button:not([disabled]), ' +
    '[href], ' +
    'input:not([disabled]), ' +
    'select:not([disabled]), ' +
    'textarea:not([disabled]), ' +
    '[tabindex]:not([tabindex="-1"]), ' +
    'details, ' +
    '[contenteditable]:not([contenteditable="false"])'
  );
}

/**
 * Hook to manage focus restoration when a component unmounts
 */
export const useFocusRestore = () => {
  const { saveFocus, restoreFocus } = useFocusManagement();

  useEffect(() => {
    saveFocus();
    return () => {
      restoreFocus();
    };
  }, [saveFocus, restoreFocus]);
};

/**
 * Hook to trap focus within a ref element (useful for modals/dialogs)
 */
export const useFocusTrap = (isActive: boolean) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const { trapFocus, focusFirst } = useFocusManagement();

  useEffect(() => {
    if (!isActive || !containerRef.current) {
return;
}

    // Focus first element when trap becomes active
    focusFirst(containerRef.current);

    // Set up focus trap
    const cleanup = trapFocus(containerRef.current);

    return cleanup;
  }, [isActive, trapFocus, focusFirst]);

  return containerRef;
};