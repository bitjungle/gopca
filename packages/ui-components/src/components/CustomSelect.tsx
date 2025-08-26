// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import React, { useState, useRef, useEffect, useCallback } from 'react';

export interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
  group?: string;
  icon?: React.ReactNode;
}

interface CustomSelectProps {
  value: string;
  onChange: (value: string) => void;
  options: SelectOption[];
  placeholder?: string;
  disabled?: boolean;
  className?: string;
  label?: string;
  helpText?: string;
  id?: string;
  name?: string;
}

export const CustomSelect: React.FC<CustomSelectProps> = ({
  value,
  onChange,
  options,
  placeholder = 'Select an option',
  disabled = false,
  className = '',
  label,
  helpText,
  id,
  name
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [highlightedIndex, setHighlightedIndex] = useState(-1);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout>>();

  // Get selected option
  const selectedOption = options.find(opt => opt.value === value);

  // Group options by category if groups exist
  const groupedOptions = React.useMemo(() => {
    const groups: { [key: string]: SelectOption[] } = {};
    const ungrouped: SelectOption[] = [];

    options.forEach(option => {
      if (option.group) {
        if (!groups[option.group]) {
          groups[option.group] = [];
        }
        groups[option.group].push(option);
      } else {
        ungrouped.push(option);
      }
    });

    return { groups, ungrouped };
  }, [options]);

  // Filter options based on search term
  const filteredOptions = React.useMemo(() => {
    if (!searchTerm) return options;
    
    return options.filter(option =>
      option.label.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [options, searchTerm]);

  // Handle click outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
        setSearchTerm('');
        setHighlightedIndex(-1);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  // Handle keyboard navigation
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (disabled) return;

    switch (e.key) {
      case 'Enter':
      case ' ':
        e.preventDefault();
        if (!isOpen) {
          setIsOpen(true);
        } else if (highlightedIndex >= 0 && highlightedIndex < filteredOptions.length) {
          const option = filteredOptions[highlightedIndex];
          if (!option.disabled) {
            onChange(option.value);
            setIsOpen(false);
            setSearchTerm('');
            setHighlightedIndex(-1);
          }
        }
        break;

      case 'Escape':
        e.preventDefault();
        setIsOpen(false);
        setSearchTerm('');
        setHighlightedIndex(-1);
        buttonRef.current?.focus();
        break;

      case 'ArrowDown':
        e.preventDefault();
        if (!isOpen) {
          setIsOpen(true);
        } else {
          setHighlightedIndex(prev => {
            const next = prev + 1;
            if (next >= filteredOptions.length) return 0;
            // Skip disabled options
            if (filteredOptions[next]?.disabled) {
              let nextValid = next + 1;
              while (nextValid < filteredOptions.length && filteredOptions[nextValid]?.disabled) {
                nextValid++;
              }
              return nextValid >= filteredOptions.length ? 0 : nextValid;
            }
            return next;
          });
        }
        break;

      case 'ArrowUp':
        e.preventDefault();
        if (!isOpen) {
          setIsOpen(true);
        } else {
          setHighlightedIndex(prev => {
            const next = prev - 1;
            if (next < 0) return filteredOptions.length - 1;
            // Skip disabled options
            if (filteredOptions[next]?.disabled) {
              let nextValid = next - 1;
              while (nextValid >= 0 && filteredOptions[nextValid]?.disabled) {
                nextValid--;
              }
              return nextValid < 0 ? filteredOptions.length - 1 : nextValid;
            }
            return next;
          });
        }
        break;

      default:
        // Type-to-search functionality
        if (isOpen && e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
          e.preventDefault();
          const newSearchTerm = searchTerm + e.key;
          setSearchTerm(newSearchTerm);
          
          // Clear search after 1.5 seconds of inactivity
          if (searchTimeoutRef.current) {
            clearTimeout(searchTimeoutRef.current);
          }
          searchTimeoutRef.current = setTimeout(() => {
            setSearchTerm('');
          }, 1500);

          // Highlight first matching option
          const firstMatch = options.findIndex(option =>
            option.label.toLowerCase().startsWith(newSearchTerm.toLowerCase())
          );
          if (firstMatch >= 0) {
            setHighlightedIndex(firstMatch);
          }
        }
        break;
    }
  }, [disabled, isOpen, highlightedIndex, filteredOptions, onChange, searchTerm, options]);

  const handleOptionClick = (option: SelectOption) => {
    if (!option.disabled) {
      onChange(option.value);
      setIsOpen(false);
      setSearchTerm('');
      setHighlightedIndex(-1);
      buttonRef.current?.focus();
    }
  };

  const renderOption = (option: SelectOption, index: number) => {
    const isSelected = option.value === value;
    const isHighlighted = index === highlightedIndex;

    return (
      <div
        key={option.value}
        role="option"
        aria-selected={isSelected}
        aria-disabled={option.disabled}
        onClick={() => handleOptionClick(option)}
        onMouseEnter={() => !option.disabled && setHighlightedIndex(index)}
        className={`
          relative flex items-center px-3 py-2 cursor-pointer select-none text-left
          ${option.disabled 
            ? 'opacity-50 cursor-not-allowed bg-gray-50 dark:bg-gray-800 text-gray-400 dark:text-gray-500' 
            : isHighlighted
              ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400'
              : isSelected
                ? 'bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-white font-medium'
                : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800'
          }
        `}
      >
        {option.icon && (
          <span className="mr-2 flex-shrink-0">{option.icon}</span>
        )}
        <span className="flex-1 truncate">{option.label}</span>
        {isSelected && (
          <svg className="w-4 h-4 ml-2 flex-shrink-0 text-blue-600 dark:text-blue-400" fill="none" strokeWidth="2" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
          </svg>
        )}
      </div>
    );
  };

  const renderOptions = () => {
    if (filteredOptions.length === 0) {
      return (
        <div className="px-3 py-2 text-gray-500 dark:text-gray-400 text-sm">
          No options found
        </div>
      );
    }

    // If we have groups, render grouped options
    if (Object.keys(groupedOptions.groups).length > 0) {
      return (
        <>
          {groupedOptions.ungrouped.map((option) => 
            renderOption(option, options.indexOf(option))
          )}
          {Object.entries(groupedOptions.groups).map(([groupName, groupOptions]) => (
            <div key={groupName}>
              <div className="px-3 py-1 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider bg-gray-50 dark:bg-gray-800/50">
                {groupName}
              </div>
              {groupOptions.map(option => 
                renderOption(option, options.indexOf(option))
              )}
            </div>
          ))}
        </>
      );
    }

    // Otherwise render flat list
    return filteredOptions.map((option) => 
      renderOption(option, options.indexOf(option))
    );
  };

  return (
    <div className={`relative ${className}`} ref={dropdownRef}>
      {label && (
        <label 
          htmlFor={id}
          className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
        >
          {label}
        </label>
      )}
      
      <button
        ref={buttonRef}
        id={id}
        name={name}
        type="button"
        role="combobox"
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        aria-controls="select-dropdown"
        disabled={disabled}
        onClick={() => !disabled && setIsOpen(!isOpen)}
        onKeyDown={handleKeyDown}
        className={`
          relative w-full px-3 py-2 text-left border rounded-lg shadow-sm
          transition-all duration-200 ease-in-out
          ${disabled
            ? 'bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700 text-gray-400 dark:text-gray-500 cursor-not-allowed'
            : isOpen
              ? 'bg-white dark:bg-gray-700 border-blue-500 dark:border-blue-400 ring-2 ring-blue-500/20 dark:ring-blue-400/20'
              : 'bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500'
          }
          ${!disabled && 'focus:outline-none focus:ring-2 focus:ring-blue-500/20 dark:focus:ring-blue-400/20 focus:border-blue-500 dark:focus:border-blue-400'}
        `}
      >
        <span className={`block truncate ${!selectedOption ? 'text-gray-400 dark:text-gray-500' : 'text-gray-900 dark:text-white'}`}>
          {selectedOption ? selectedOption.label : placeholder}
        </span>
        <span className="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
          <svg 
            className={`w-5 h-5 text-gray-400 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`} 
            fill="none" 
            strokeWidth="2" 
            stroke="currentColor" 
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
          </svg>
        </span>
      </button>

      {helpText && (
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {helpText}
        </p>
      )}

      {/* Dropdown */}
      {isOpen && (
        <div
          id="select-dropdown"
          role="listbox"
          aria-label={label || 'Select options'}
          className={`
            absolute z-50 w-full mt-1 bg-white dark:bg-gray-700 
            border border-gray-200 dark:border-gray-600 
            rounded-lg shadow-lg dark:shadow-xl
            max-h-60 overflow-auto
            animate-in fade-in-0 zoom-in-95 duration-200
          `}
          style={{
            animation: 'slideDown 0.2s ease-out'
          }}
        >
          {searchTerm && (
            <div className="px-3 py-1 text-xs text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-600">
              Searching: "{searchTerm}"
            </div>
          )}
          {renderOptions()}
        </div>
      )}

      <style>{`
        @keyframes slideDown {
          from {
            opacity: 0;
            transform: translateY(-10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </div>
  );
};