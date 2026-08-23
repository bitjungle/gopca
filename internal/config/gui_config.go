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

package config

// GUIConfig holds values the Go backend supplies to GoPCA Desktop's frontend.
//
// Only add a field here when the frontend actually reads it. This struct once
// carried twelve fields of which ten were read by nothing: declared, defaulted,
// unit-tested for their default values and serialised across the Wails boundary,
// while the values genuinely in force were hardcoded at the point of use.
// Editing them changed nothing, with no way to discover that short of running
// the application. See issue #778.
//
// Note also that GetGUIConfig returns DefaultGUIConfig() unconditionally: no
// configuration file is read anywhere. These are constants delivered to the
// frontend, not settings a user can change. A field added here and then read by
// the frontend still yields the compiled-in default, so it earns its place only
// when the value must genuinely be shared with Go -- not as a gesture toward a
// configurability that does not exist.
type GUIConfig struct {
	// Visualization configuration
	Visualization VisualizationConfig `json:"visualization"`
}

// VisualizationConfig holds visualization-related configuration
type VisualizationConfig struct {
	// Above this many variables the loadings plot switches from bars to a line.
	// Read in ResultsSection.tsx.
	LoadingsVariableThreshold int `json:"loadings_variable_threshold"`

	// Maximum variables to draw in the 2D and 3D biplots.
	// Read in ResultsSection.tsx.
	BiplotMaxVariables int `json:"biplot_max_variables"`
}

// DefaultGUIConfig returns the default GUI configuration
func DefaultGUIConfig() *GUIConfig {
	return &GUIConfig{
		Visualization: VisualizationConfig{
			LoadingsVariableThreshold: 100,
			BiplotMaxVariables:        100,
		},
	}
}
