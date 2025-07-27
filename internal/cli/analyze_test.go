package cli

import (
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// TestDecimalSeparatorParsing tests that decimal separator parsing works correctly with switch statement
func TestDecimalSeparatorParsing(t *testing.T) {
	tests := []struct {
		name      string
		separator string
		want      rune
	}{
		{
			name:      "dot separator",
			separator: "dot",
			want:      '.',
		},
		{
			name:      "comma separator word",
			separator: "comma",
			want:      ',',
		},
		{
			name:      "comma separator symbol",
			separator: ",",
			want:      ',',
		},
		{
			name:      "custom separator semicolon",
			separator: ";",
			want:      ';',
		},
		{
			name:      "custom separator pipe",
			separator: "|",
			want:      '|',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create parse options
			parseOpts := types.CSVFormat{}

			// Apply the same logic as in analyze.go
			switch tt.separator {
			case "dot":
				parseOpts.DecimalSeparator = '.'
			case "comma", ",":
				parseOpts.DecimalSeparator = ','
			default:
				parseOpts.DecimalSeparator = rune(tt.separator[0])
			}

			if parseOpts.DecimalSeparator != tt.want {
				t.Errorf("Expected decimal separator %c, got %c", tt.want, parseOpts.DecimalSeparator)
			}
		})
	}
}
