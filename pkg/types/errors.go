// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package types

import (
	"fmt"
)

// ErrorType represents categories of errors that can occur in the system
type ErrorType string

const (
	// ErrValidation indicates invalid input or parameters
	ErrValidation ErrorType = "validation"
	// ErrComputation indicates numerical or algorithmic errors
	ErrComputation ErrorType = "computation"
	// ErrIO indicates file or data I/O errors
	ErrIO ErrorType = "io"
	// ErrPreprocessing indicates errors during data preprocessing
	ErrPreprocessing ErrorType = "preprocessing"
	// ErrConfiguration indicates invalid configuration
	ErrConfiguration ErrorType = "configuration"
	// ErrNotFitted indicates model hasn't been fitted yet
	ErrNotFitted ErrorType = "not_fitted"
	// ErrDimension indicates dimension mismatch errors
	ErrDimension ErrorType = "dimension"
	// ErrMissingData indicates issues with missing data handling
	ErrMissingData ErrorType = "missing_data"
	// ErrConvergence indicates algorithm convergence failures
	ErrConvergence ErrorType = "convergence"
	// ErrMemory indicates memory allocation errors
	ErrMemory ErrorType = "memory"
)

// PCAError represents a structured error for PCA operations
type PCAError struct {
	Type    ErrorType
	Message string
	Context map[string]interface{}
	Cause   error
}

// Error implements the error interface
func (e *PCAError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s error: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause
func (e *PCAError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a new validation error
func NewValidationError(message string, cause error) *PCAError {
	return &PCAError{
		Type:    ErrValidation,
		Message: message,
		Cause:   cause,
	}
}

// NewComputationError creates a new computation error
func NewComputationError(message string, cause error) *PCAError {
	return &PCAError{
		Type:    ErrComputation,
		Message: message,
		Cause:   cause,
	}
}

// NewIOError creates a new I/O error
func NewIOError(message string, cause error) *PCAError {
	return &PCAError{
		Type:    ErrIO,
		Message: message,
		Cause:   cause,
	}
}

// NewPreprocessingError creates a new preprocessing error
func NewPreprocessingError(message string, cause error) *PCAError {
	return &PCAError{
		Type:    ErrPreprocessing,
		Message: message,
		Cause:   cause,
	}
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(message string, cause error) *PCAError {
	return &PCAError{
		Type:    ErrConfiguration,
		Message: message,
		Cause:   cause,
	}
}

// NewNotFittedError creates a new not fitted error
func NewNotFittedError(message string) *PCAError {
	return &PCAError{
		Type:    ErrNotFitted,
		Message: message,
	}
}

// NewDimensionError creates a new dimension mismatch error
func NewDimensionError(message string, expected, actual int) *PCAError {
	return &PCAError{
		Type:    ErrDimension,
		Message: message,
		Context: map[string]interface{}{
			"expected": expected,
			"actual":   actual,
		},
	}
}

// NewMissingDataError creates a new missing data error
func NewMissingDataError(message string, location map[string]int) *PCAError {
	ctx := make(map[string]interface{})
	for k, v := range location {
		ctx[k] = v
	}
	return &PCAError{
		Type:    ErrMissingData,
		Message: message,
		Context: ctx,
	}
}

// NewConvergenceError creates a new convergence error
func NewConvergenceError(message string, iterations int) *PCAError {
	return &PCAError{
		Type:    ErrConvergence,
		Message: message,
		Context: map[string]interface{}{
			"iterations": iterations,
		},
	}
}

// NewMemoryError creates a new memory error
func NewMemoryError(message string, size int64) *PCAError {
	return &PCAError{
		Type:    ErrMemory,
		Message: message,
		Context: map[string]interface{}{
			"size_bytes": size,
		},
	}
}
