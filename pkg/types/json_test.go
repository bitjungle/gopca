package types

import (
	"encoding/json"
	"math"
	"testing"
)

func TestJSONFloat64_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		value    JSONFloat64
		wantJSON string
	}{
		{"regular number", JSONFloat64(3.14), "3.14"},
		{"zero", JSONFloat64(0), "0"},
		{"negative number", JSONFloat64(-42.5), "-42.5"},
		{"NaN", JSONFloat64(math.NaN()), "null"},
		{"positive infinity", JSONFloat64(math.Inf(1)), "null"},
		{"negative infinity", JSONFloat64(math.Inf(-1)), "null"},
		{"very large number", JSONFloat64(1e308), "1e+308"},
		{"very small number", JSONFloat64(1e-308), "1e-308"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.wantJSON {
				t.Errorf("MarshalJSON() = %s, want %s", string(got), tt.wantJSON)
			}
		})
	}
}

func TestJSONFloat64_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		wantValue float64
		wantNaN   bool
		wantErr   bool
	}{
		{"regular number", "3.14", 3.14, false, false},
		{"zero", "0", 0, false, false},
		{"negative number", "-42.5", -42.5, false, false},
		{"null", "null", 0, true, false},
		{"scientific notation", "1e-10", 1e-10, false, false},
		{"invalid JSON", "invalid", 0, false, true},
		{"empty string", `""`, 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f JSONFloat64
			err := json.Unmarshal([]byte(tt.jsonData), &f)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.wantNaN {
				if !math.IsNaN(float64(f)) {
					t.Errorf("UnmarshalJSON() expected NaN, got %v", float64(f))
				}
			} else {
				if float64(f) != tt.wantValue {
					t.Errorf("UnmarshalJSON() = %v, want %v", float64(f), tt.wantValue)
				}
			}
		})
	}
}

func TestJSONFloat64_Methods(t *testing.T) {
	t.Run("Float64", func(t *testing.T) {
		f := JSONFloat64(3.14)
		if f.Float64() != 3.14 {
			t.Errorf("Float64() = %v, want 3.14", f.Float64())
		}
	})

	t.Run("IsNaN", func(t *testing.T) {
		tests := []struct {
			value JSONFloat64
			want  bool
		}{
			{JSONFloat64(3.14), false},
			{JSONFloat64(math.NaN()), true},
			{JSONFloat64(math.Inf(1)), false},
		}

		for _, tt := range tests {
			if got := tt.value.IsNaN(); got != tt.want {
				t.Errorf("IsNaN() = %v, want %v for value %v", got, tt.want, float64(tt.value))
			}
		}
	})

	t.Run("IsInf", func(t *testing.T) {
		tests := []struct {
			value JSONFloat64
			want  bool
		}{
			{JSONFloat64(3.14), false},
			{JSONFloat64(math.NaN()), false},
			{JSONFloat64(math.Inf(1)), true},
			{JSONFloat64(math.Inf(-1)), true},
		}

		for _, tt := range tests {
			if got := tt.value.IsInf(); got != tt.want {
				t.Errorf("IsInf() = %v, want %v for value %v", got, tt.want, float64(tt.value))
			}
		}
	})
}

func TestJSONFloat64_RoundTrip(t *testing.T) {
	// Test that marshaling and unmarshaling preserves values
	// Note: NaN and Inf values are marshaled as null and unmarshaled as NaN
	// This is by design for JSON compatibility
	tests := []struct {
		value   float64
		wantNaN bool // whether the round-trip result should be NaN
	}{
		{0, false},
		{1, false},
		{-1, false},
		{3.14, false},
		{-3.14, false},
		{1e10, false},
		{1e-10, false},
		{math.NaN(), true},   // NaN -> null -> NaN
		{math.Inf(1), true},  // +Inf -> null -> NaN
		{math.Inf(-1), true}, // -Inf -> null -> NaN
	}

	for _, tt := range tests {
		original := JSONFloat64(tt.value)

		// Marshal
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("failed to marshal %v: %v", tt.value, err)
		}

		// Unmarshal
		var decoded JSONFloat64
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Fatalf("failed to unmarshal %v: %v", string(data), err)
		}

		// Check result
		if tt.wantNaN {
			if !math.IsNaN(float64(decoded)) {
				t.Errorf("round trip for %v: expected NaN, got %v", tt.value, float64(decoded))
			}
		} else {
			if float64(decoded) != tt.value {
				t.Errorf("round trip for %v: got %v", tt.value, float64(decoded))
			}
		}
	}
}
