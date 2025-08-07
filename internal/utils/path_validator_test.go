// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package utils

import (
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid relative path",
			path:    "data/test.csv",
			wantErr: false,
		},
		{
			name:    "valid absolute path",
			path:    "/home/user/data.csv",
			wantErr: false,
		},
		{
			name:    "directory traversal attempt",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "hidden directory traversal",
			path:    "data/../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "complex path with traversal",
			path:    "./data/../../../secret.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOutputPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid output path",
			path:    "output/results.csv",
			wantErr: false,
		},
		{
			name:    "system directory",
			path:    "/etc/important.conf",
			wantErr: true,
		},
		{
			name:    "bin directory",
			path:    "/bin/malicious",
			wantErr: true,
		},
		{
			name:    "valid user directory",
			path:    "/home/user/output.csv",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutputPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
