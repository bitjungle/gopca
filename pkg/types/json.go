// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

package types

import (
	"encoding/json"
	"math"
)

// JSONFloat64 is a float64 that marshals NaN and Inf values as null in JSON.
// This ensures compatibility with JavaScript and other JSON consumers that
// don't support these special float values.
type JSONFloat64 float64

// MarshalJSON implements the json.Marshaler interface.
// NaN and Inf values are marshaled as null to ensure JSON compatibility.
func (f JSONFloat64) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
		return []byte("null"), nil
	}
	return json.Marshal(float64(f))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// null values are unmarshaled as NaN.
func (f *JSONFloat64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*f = JSONFloat64(math.NaN())
		return nil
	}

	var val float64
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	*f = JSONFloat64(val)
	return nil
}

// Float64 returns the underlying float64 value.
func (f JSONFloat64) Float64() float64 {
	return float64(f)
}

// IsNaN returns true if the value is NaN.
func (f JSONFloat64) IsNaN() bool {
	return math.IsNaN(float64(f))
}

// IsInf returns true if the value is infinite.
func (f JSONFloat64) IsInf() bool {
	return math.IsInf(float64(f), 0)
}
