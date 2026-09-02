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

// Package crossval generates resampling designs for model validation.
//
// The package produces index sets only. It knows nothing about PCA, regression,
// or any particular estimator, which keeps it testable in isolation and reusable
// by any validation work that follows.
//
// # Designs
//
// GroupKFold is the general splitter. With the default grouping of one row per
// group it covers the whole family of partition-based designs:
//
//	Requested design          GroupKFold parameters
//	------------------------  ------------------------------------------
//	K-fold, random            Groups nil, K as given, Shuffle true
//	Leave-one-out             Groups nil, K = 0 (means "as many as groups")
//	Grouped K-fold            Groups set, K as given, Shuffle true
//	Leave-one-group-out       Groups set, K = 0
//	Contiguous blocks         Groups nil, K as given, Shuffle false
//
// Leave-one-out is therefore not a separate algorithm and does not get its own
// type: it is K-fold at K = n. Writing it separately would duplicate the fold
// bookkeeping and give the two paths independent opportunities to disagree.
//
// ForwardChaining is the exception that genuinely cannot be expressed as a
// partition, because its training sets are nested prefixes rather than the
// complements of disjoint test folds. It is the correct design for time-ordered
// data, where training on observations later than those being predicted would
// invent information the model will not have in deployment.
//
// # Grouping
//
// A group is the independent unit of observation: an object measured several
// times, a production batch, an instrument, a site. When rows within a group are
// near duplicates, splitting them across folds lets a model interpolate between
// twins and reports an error lower than deployment will deliver. Grouping keeps
// them together. The effective sample size is the number of groups, not the
// number of rows, so K is capped at the group count rather than the row count.
//
// # Determinism
//
// Every splitter is deterministic given its seed, so a recorded design can be
// reproduced exactly from the seed alone. Shuffling uses a locally seeded
// generator and never the global source, so callers cannot perturb one another.
package crossval
