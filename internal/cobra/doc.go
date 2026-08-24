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

// Package cobra defines the Cobra command tree for the pca command-line
// interface: the root command plus the analyze, transform, validate, version,
// and shell-completion subcommands.
//
// Commands are thin wrappers that parse flags, delegate the numerical work to
// the shared engine in [github.com/bitjungle/gopca/internal/core], and format
// results for the terminal. The exported Version, BuildTime, and Commit
// variables are populated at startup from the build-time version information.
package cobra
