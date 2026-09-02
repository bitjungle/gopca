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

package crossval

import "fmt"

// ForwardChaining validates time-ordered data by rolling the origin forward:
// each fold trains on a prefix of the series and tests on the block immediately
// after it.
//
// This is the one design that cannot be expressed as a partition, because the
// training sets are nested prefixes rather than the complements of disjoint test
// folds. Every training set is a subset of the next, and the earliest block is
// never tested because nothing precedes it to train on.
//
// The reason to prefer it over K-fold for ordered data is not statistical
// fastidiousness. Training on observations later than those being predicted
// gives the model information it will not have in deployment, and the resulting
// error estimate describes a situation that will never occur.
//
// Split assumes indices are already in chronological order; it does not sort
// them, because the caller's row order is the only available record of time.
type ForwardChaining struct {
	// Splits is the number of test blocks, and therefore the number of folds.
	Splits int

	// MinTrain is the smallest acceptable training prefix. Zero lets the first
	// training set be as small as the block size, which for a short series can
	// mean fitting on too little data to be meaningful.
	MinTrain int
}

// Name identifies the design.
func (f *ForwardChaining) Name() string {
	return fmt.Sprintf("forward-chaining (%d splits)", f.Splits)
}

// Split produces nested-prefix folds over indices, which are taken to be in
// chronological order.
//
// Algorithm complexity: O(n) in the number of indices.
func (f *ForwardChaining) Split(indices []int) ([]Fold, error) {
	if err := validateIndices(indices); err != nil {
		return nil, err
	}
	if f.Splits < 1 {
		return nil, fmt.Errorf("forward chaining needs at least 1 split, got %d", f.Splits)
	}
	if f.MinTrain < 0 {
		return nil, fmt.Errorf("minimum training size must not be negative, got %d", f.MinTrain)
	}

	n := len(indices)

	// The series is divided into Splits+1 blocks: the first is training-only
	// seed data, and each of the remaining Splits blocks is tested in turn.
	blocks := f.Splits + 1
	if n < blocks {
		return nil, fmt.Errorf("cannot make %d forward-chaining splits from %d observations: "+
			"each split needs at least one observation to test and one to train on",
			f.Splits, n)
	}

	minTrain := f.MinTrain
	if minTrain == 0 {
		minTrain = 1
	}
	if minTrain >= n {
		return nil, fmt.Errorf("minimum training size %d leaves nothing to test from %d observations",
			minTrain, n)
	}

	folds := make([]Fold, 0, f.Splits)
	for i := 1; i <= f.Splits; i++ {
		cut := i * n / blocks
		end := (i + 1) * n / blocks

		if cut < minTrain {
			// Skip folds whose training prefix is shorter than the caller is
			// willing to fit on, rather than emitting a fold they would have to
			// filter out themselves.
			continue
		}
		if end <= cut {
			continue
		}

		train := make([]int, cut)
		copy(train, indices[:cut])
		test := make([]int, end-cut)
		copy(test, indices[cut:end])

		folds = append(folds, Fold{Train: train, Test: test})
	}

	if len(folds) == 0 {
		return nil, fmt.Errorf("no usable forward-chaining folds: "+
			"a minimum training size of %d is too large for %d observations in %d splits",
			minTrain, n, f.Splits)
	}

	return folds, nil
}
