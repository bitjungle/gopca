// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package security

import (
	"testing"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "valid email",
			email: "user@example.com",
			want:  true,
		},
		{
			name:  "valid email with subdomain",
			email: "user@mail.example.com",
			want:  true,
		},
		{
			name:  "valid email with plus",
			email: "user+tag@example.com",
			want:  true,
		},
		{
			name:  "invalid no at sign",
			email: "userexample.com",
			want:  false,
		},
		{
			name:  "invalid no domain",
			email: "user@",
			want:  false,
		},
		{
			name:  "invalid no user",
			email: "@example.com",
			want:  false,
		},
		{
			name:  "empty string",
			email: "",
			want:  false,
		},
		{
			name:  "multiple at signs",
			email: "user@@example.com",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmail(tt.email)
			if got != tt.want {
				t.Errorf("IsValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}
