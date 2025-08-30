// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package core

import (
	"strings"
	"testing"

	"github.com/bitjungle/gopca/pkg/security"
	"github.com/bitjungle/gopca/pkg/types"
)

// TestKernelPCAMemoryLimit tests that kernel PCA properly enforces memory limits
func TestKernelPCAMemoryLimit(t *testing.T) {
	tests := []struct {
		name        string
		nSamples    int
		shouldFail  bool
		errContains string
	}{
		{
			name:        "Small dataset - should work",
			nSamples:    100,
			shouldFail:  false,
			errContains: "",
		},
		{
			name:        "Medium dataset - should work",
			nSamples:    500,
			shouldFail:  false,
			errContains: "",
		},
		{
			name:        "Over limit - should fail",
			nSamples:    security.MaxKernelPCASamples + 1,
			shouldFail:  true,
			errContains: "kernel PCA limited to",
		},
		{
			name:        "Large dataset - should fail",
			nSamples:    50000,
			shouldFail:  true,
			errContains: "memory safety",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate dummy data with specified number of samples
			data := make(types.Matrix, tt.nSamples)
			for i := 0; i < tt.nSamples; i++ {
				data[i] = []float64{float64(i), float64(i) * 2, float64(i) * 3}
			}

			// Create kernel PCA engine
			kpca := NewKernelPCAEngine()

			// Configure for RBF kernel PCA
			config := types.PCAConfig{
				Components:  2,
				KernelType:  "rbf",
				KernelGamma: 0.1,
			}

			// Try to fit the model
			_, err := kpca.Fit(data, config)

			// Check if error matches expectation
			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error for %d samples, but got none", tt.nSamples)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %d samples: %v", tt.nSamples, err)
				}
			}
		})
	}
}
